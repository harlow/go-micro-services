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

	geo "github.com/harlow/go-micro-services/service.geo/proto"
	profile "github.com/harlow/go-micro-services/service.profile/proto"
	rate "github.com/harlow/go-micro-services/service.rate/proto"

	"github.com/harlow/auth_token"
	"github.com/harlow/go-micro-services/trace"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	serverName        = "api.v1"
	port              = flag.String("port", "5000", "The server port")
	geoServerAddr     = flag.String("geo_server_addr", "127.0.0.1:10002", "The Geo server address in the format of host:port")
	profileServerAddr = flag.String("profile_server_addr", "127.0.0.1:10003", "The Pofile server address in the format of host:port")
	rateServerAddr    = flag.String("rate_server_addr", "127.0.0.1:10004", "The Rate Code server address in the format of host:port")
)

type inventory struct {
	Hotels []*profile.Hotel `json:"hotels"`
	Rates  []*rate.RatePlan `json:"rates"`
}

func hotelsWithinBoundedBox(traceID string, serverName string, latitude int32, longitude int32) ([]int32, error) {
	// dial server connection
	conn, err := grpc.Dial(*geoServerAddr)
	if err != nil {
		return []int32{}, err
	}
	defer conn.Close()

	// set up args and client
	rect := &geo.Rectangle{
		Lo: &geo.Point{Latitude: 400000000, Longitude: -750000000},
		Hi: &geo.Point{Latitude: 420000000, Longitude: -730000000},
	}
	args := &geo.Args{TraceId: traceID, From: serverName, Rect: rect}
	client := geo.NewGeoClient(conn)

	// get hotels within bounded bob
	reply, err := client.BoundedBox(context.Background(), args)
	if err != nil {
		return []int32{}, err
	}

	return reply.HotelIds, nil
}

func hotelProfiles(traceID string, serverName string, hotelIDs []int32) ([]*profile.Hotel, error) {
	// dial server connection
	conn, err := grpc.Dial(*profileServerAddr)
	if err != nil {
		return []*profile.Hotel{}, err
	}
	defer conn.Close()

	// set up args
	args := &profile.Args{TraceId: traceID, From: serverName, HotelIds: hotelIDs}
	client := profile.NewProfileClient(conn)

	// get profile data
	reply, err := client.GetProfiles(context.Background(), args)
	if err != nil {
		return []*profile.Hotel{}, err
	}

	return reply.Hotels, nil
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
	authClient *auth.Client
}

func (api api) requestHandler(w http.ResponseWriter, r *http.Request) {
	t := trace.NewTracer()

	md := metadata.Pairs("traceID", t.TraceID)
	ctx := context.Background()
	ctx = metadata.NewContext(ctx, md)

	t.In("www", "api.v1")
	defer t.Out("api.v1", "www", time.Now())

	// extract authentication token from Authorization header
	authToken, err := auth_token.Parse(r.Header.Get("Authorization"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	// grab and verify in and out date
	inDate := r.URL.Query().Get("inDate")
	outDate := r.URL.Query().Get("outDate")

	if inDate == "" || outDate == "" {
		http.Error(w, "Please specify inDate / outDate", http.StatusBadRequest)
		return
	}

	// verify auth token
	t.Req(serverName, "service.auth", "VerifyToken")
	err = api.authClient.VerifyToken(ctx, serverName, authToken)
	t.Rep("service.auth", serverName, time.Now())

	if err != nil {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// search for hotels within geo rectangle
	t.Req(serverName, "service.geo", "BoundedBox")
	hotelIDs, err := hotelsWithinBoundedBox(t.TraceID, serverName, 100, 100)
	t.Rep("service.geo", serverName, time.Now())

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	hotelProfilesReady := api.getHotelProfiles(t.TraceID, serverName, hotelIDs)
	ratePlansReady := api.getRates(t.TraceID, serverName, hotelIDs, inDate, outDate)

	hotelProfileResp := <-hotelProfilesReady
	if err := hotelProfileResp.err; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ratePlanResp := <-ratePlansReady
	if err := ratePlanResp.err; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	inventory := inventory{
		Hotels: hotelProfileResp.hotelProfiles,
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

type hotelProfileResults struct {
	hotelProfiles []*profile.Hotel
	err           error
}

func (api api) getHotelProfiles(traceID string, serverName string, hotelIDs []int32) chan hotelProfileResults {
	ch := make(chan hotelProfileResults, 1)

	go func() {
		hotelProfiles, err := hotelProfiles(traceID, serverName, hotelIDs)

		ch <- hotelProfileResults{
			hotelProfiles: hotelProfiles,
			err:           err,
		}
	}()

	return ch
}

func main() {
	authServerAddr := flag.String("auth_server_addr", "127.0.0.1:10001", "The Auth server address in the format of host:port")

	flag.Parse()

	authClient, err := auth.NewClient(*authServerAddr)

	if err != nil {
		log.Fatal("AuthClient error:", err)
	}

	defer authClient.Close()

	api := api{
		authClient: authClient,
	}

	http.HandleFunc("/", api.requestHandler)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
