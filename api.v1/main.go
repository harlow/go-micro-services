package main

import (
	"encoding/json"
	_ "expvar"
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"

	auth "github.com/harlow/go-micro-services/service.auth/lib"
	geo "github.com/harlow/go-micro-services/service.geo/lib"
	profile "github.com/harlow/go-micro-services/service.profile/lib"

	profile_pb "github.com/harlow/go-micro-services/service.profile/proto"
	rate "github.com/harlow/go-micro-services/service.rate/proto"

	"github.com/harlow/auth_token"
	"github.com/harlow/go-micro-services/trace"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	serverName     = "api.v1"
	port           = flag.String("port", "5000", "The server port")
	rateServerAddr = flag.String("rate_server_addr", "127.0.0.1:10004", "The Rate Code server address in the format of host:port")
)

type inventory struct {
	Hotels []*profile_pb.Hotel `json:"hotels"`
	Rates  []*rate.RatePlan    `json:"rates"`
}

func getRates(traceID string, serverName string, hotelIDs []int32, inDate string, outDate string) ([]*rate.RatePlan, error) {
	// dial server connection
	conn, err := grpc.Dial(*rateServerAddr)
	if err != nil {
		return []*rate.RatePlan{}, err
	}
	defer conn.Close()

	// set up args
	args := &rate.Args{
		TraceId:  traceID,
		From:     serverName,
		HotelIds: hotelIDs,
		InDate:   inDate,
		OutDate:  outDate,
	}
	client := rate.NewRateClient(conn)

	// get rates
	reply, err := client.GetRates(context.Background(), args)
	if err != nil {
		return []*rate.RatePlan{}, err
	}

	return reply.Rates, nil
}

type api struct {
	authClient    *auth.Client
	geoClient     *geo.Client
	profileClient *profile.Client
}

func (api api) requestHandler(w http.ResponseWriter, r *http.Request) {
	t := trace.NewTracer()
	t.In("www", "api.v1")
	defer t.Out("api.v1", "www", time.Now())

	md := metadata.Pairs("traceID", t.TraceID, "from", serverName)
	ctx := context.Background()
	ctx = metadata.NewContext(ctx, md)

	// parse token from Authorization header
	authToken, err := auth_token.Parse(r.Header.Get("Authorization"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	// read and validate in/out arguments
	inDate := r.URL.Query().Get("inDate")
	outDate := r.URL.Query().Get("outDate")
	if inDate == "" || outDate == "" {
		http.Error(w, "Please specify inDate / outDate", http.StatusBadRequest)
		return
	}

	// t.Req(serverName, "service.auth", "VerifyToken")
	// t.Rep("service.auth", serverName, time.Now())
	// verify auth token
	err = api.authClient.VerifyToken(ctx, serverName, authToken)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// get hotels within geo rectangle
	hotelIDs, err := api.geoClient.HotelsWithinBoundedBox(ctx, 100, 100)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	hotelsCh := api.getHotels(ctx, hotelIDs)
	ratePlansReady := api.getRates(t.TraceID, serverName, hotelIDs, inDate, outDate)

	hotelsReply := <-hotelsCh
	if err := hotelsReply.err; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ratePlanResp := <-ratePlansReady
	if err := ratePlanResp.err; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	inventory := inventory{
		Hotels: hotelsReply.hotels,
		Rates:  ratePlanResp.ratePlans,
	}

	body, err := json.Marshal(inventory)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(body)
}

type ratePlanResults struct {
	ratePlans []*rate.RatePlan
	err       error
}

func (api api) getRates(traceID string, serverName string, hotelIDs []int32, inDate string, outDate string) chan ratePlanResults {
	ch := make(chan ratePlanResults, 1)

	go func() {
		ratePlans, err := getRates(traceID, serverName, hotelIDs, inDate, outDate)

		ch <- ratePlanResults{
			ratePlans: ratePlans,
			err:       err,
		}
	}()

	return ch
}

type profileResults struct {
	hotels []*profile_pb.Hotel
	err    error
}

func (api api) getHotels(ctx context.Context, hotelIDs []int32) chan profileResults {
	ch := make(chan profileResults, 1)

	go func() {
		hotels, err := api.profileClient.GetHotels(ctx, hotelIDs)

		ch <- profileResults{
			hotels: hotels,
			err:    err,
		}
	}()

	return ch
}

func main() {
	authServerAddr := flag.String("auth_server_addr", "127.0.0.1:10001", "The Auth server address in the format of host:port")
	geoServerAddr := flag.String("geo_server_addr", "127.0.0.1:10002", "The Geo server address in the format of host:port")
	profileServerAddr := flag.String("profile_server_addr", "127.0.0.1:10003", "The Pofile server address in the format of host:port")
	flag.Parse()

	authClient, err := auth.NewClient(*authServerAddr)
	if err != nil {
		log.Fatal("AuthClient error:", err)
	}
	defer authClient.Close()

	geoClient, err := geo.NewClient(*geoServerAddr)
	if err != nil {
		log.Fatal("GeoClient error:", err)
	}
	defer geoClient.Close()

	profileClient, err := profile.NewClient(*profileServerAddr)
	if err != nil {
		log.Fatal("ProfileClient error:", err)
	}
	defer profileClient.Close()

	api := api{
		authClient:    authClient,
		geoClient:     geoClient,
		profileClient: profileClient,
	}

	http.HandleFunc("/", api.requestHandler)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
