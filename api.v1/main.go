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
	rate "github.com/harlow/go-micro-services/service.rate/lib"
	rate_pb "github.com/harlow/go-micro-services/service.rate/proto"

	"github.com/harlow/auth_token"
	"github.com/harlow/go-micro-services/trace"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

type inventory struct {
	Hotels    []*profile_pb.Hotel `json:"hotels"`
	RatePlans []*rate_pb.RatePlan `json:"ratePlans"`
}

type api struct {
	authClient    *auth.Client
	geoClient     *geo.Client
	profileClient *profile.Client
	rateClient    *rate.Client
}

func (api api) requestHandler(w http.ResponseWriter, r *http.Request) {
	t := trace.NewTracer()
	t.In("www", "api.v1")
	defer t.Out("api.v1", "www", time.Now())

	// context and metadata
	md := metadata.Pairs("traceID", t.TraceID, "from", "api.v1")
	ctx := context.Background()
	ctx = metadata.NewContext(ctx, md)

	// parse token from Authorization header
	authToken, err := auth_token.Parse(r.Header.Get("Authorization"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	// verify auth token
	err = api.authClient.VerifyToken(ctx, authToken)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// read and validate in/out arguments
	inDate := r.URL.Query().Get("inDate")
	outDate := r.URL.Query().Get("outDate")
	if inDate == "" || outDate == "" {
		http.Error(w, "Please specify inDate / outDate", http.StatusBadRequest)
		return
	}

	// get hotels within geo box
	hotelIDs, err := api.geoClient.HotelsWithinBoundedBox(ctx, 100, 100)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	profileCh := api.getHotels(ctx, hotelIDs)
	rateCh := api.getRatePlans(ctx, hotelIDs, inDate, outDate)

	profileReply := <-profileCh
	if err := profileReply.err; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rateReply := <-rateCh
	if err := rateReply.err; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	inventory := inventory{
		Hotels:    profileReply.hotels,
		RatePlans: rateReply.ratePlans,
	}

	body, err := json.Marshal(inventory)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(body)
}

type rateResults struct {
	ratePlans []*rate_pb.RatePlan
	err       error
}

func (api api) getRatePlans(ctx context.Context, hotelIDs []int32, inDate string, outDate string) chan rateResults {
	ch := make(chan rateResults, 1)

	go func() {
		ratePlans, err := api.rateClient.GetRatePlans(ctx, hotelIDs, inDate, outDate)

		ch <- rateResults{
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
	var (
		port              = flag.String("port", "5000", "The server port")
		authServerAddr    = flag.String("auth_server_addr", "127.0.0.1:10001", "The Auth server address in the format of host:port")
		geoServerAddr     = flag.String("geo_server_addr", "127.0.0.1:10002", "The Geo server address in the format of host:port")
		profileServerAddr = flag.String("profile_server_addr", "127.0.0.1:10003", "The Pofile server address in the format of host:port")
		rateServerAddr    = flag.String("rate_server_addr", "127.0.0.1:10004", "The Rate Code server address in the format of host:port")
	)
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

	rateClient, err := rate.NewClient(*rateServerAddr)
	if err != nil {
		log.Fatal("RateClient error:", err)
	}
	defer rateClient.Close()

	api := api{
		authClient:    authClient,
		geoClient:     geoClient,
		profileClient: profileClient,
		rateClient:    rateClient,
	}

	http.HandleFunc("/", api.requestHandler)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
