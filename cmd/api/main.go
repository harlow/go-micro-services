package main

import (
	"encoding/json"
	_ "expvar"
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/harlow/go-micro-services/proto/auth"
	"github.com/harlow/go-micro-services/proto/geo"
	"github.com/harlow/go-micro-services/proto/profile"
	"github.com/harlow/go-micro-services/proto/rate"
	"github.com/harlow/go-micro-services/trace"

	"github.com/harlow/authtoken"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type inventory struct {
	Hotels    []*profile.Hotel `json:"hotels"`
	RatePlans []*rate.RatePlan `json:"ratePlans"`
}

// newServer returns a server with initialization data loaded.
func newServer(authAddr, geoAddr, profileAddr, rateAddr *string) apiServer {
	return apiServer{
		serverName:    "api.v1",
		AuthClient:    auth.NewAuthClient(mustDial(authAddr)),
		GeoClient:     geo.NewGeoClient(mustDial(geoAddr)),
		ProfileClient: profile.NewProfileClient(mustDial(profileAddr)),
		RateClient:    rate.NewRateClient(mustDial(rateAddr)),
	}
}

type apiServer struct {
	serverName string
	auth.AuthClient
	geo.GeoClient
	profile.ProfileClient
	rate.RateClient
}

// mustDial ensures the tcp connection to specified address.
func mustDial(addr *string) *grpc.ClientConn {
	conn, err := grpc.Dial(*addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to dial: %v", err)
		panic(err)
	}
	return conn
}

func (s apiServer) requestHandler(w http.ResponseWriter, r *http.Request) {
	t := trace.NewTracer()
	t.In("www", s.serverName)
	defer t.Out(s.serverName, "www", time.Now())

	// context and metadata
	md := metadata.Pairs("traceID", t.TraceID, "fromName", s.serverName)
	ctx := context.Background()
	ctx = metadata.NewContext(ctx, md)

	// grab auth token from request
	token, err := authtoken.FromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	// verify token w/ auth service
	_, err = s.VerifyToken(ctx, &auth.Request{token})
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
	geoRes, err := s.BoundedBox(ctx, &geo.Request{
		Lo: &geo.Point{Latitude: 400000000, Longitude: -750000000},
		Hi: &geo.Point{Latitude: 420000000, Longitude: -730000000},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// make reqeusts for profiles and rates
	profileCh := s.getProfiles(ctx, geoRes.HotelIds)
	rateCh := s.getRatePlans(ctx, geoRes.HotelIds, inDate, outDate)

	// wait on profiles reply
	profileReply := <-profileCh
	if err := profileReply.err; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// wait on rates reply
	rateReply := <-rateCh
	if err := rateReply.err; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// build the final inventory response
	inventory := inventory{
		Hotels:    profileReply.hotels,
		RatePlans: rateReply.ratePlans,
	}

	// encode JSON for rendering
	encoder := json.NewEncoder(w)
	if err = encoder.Encode(inventory); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s apiServer) getRatePlans(ctx context.Context, hotelIDs []int32, inDate string, outDate string) chan rateResults {
	ch := make(chan rateResults, 1)

	go func() {
		req := &rate.Request{hotelIDs, inDate, outDate}
		res, err := s.GetRates(ctx, req)
		ch <- rateResults{res.RatePlans, err}
	}()

	return ch
}

type rateResults struct {
	ratePlans []*rate.RatePlan
	err       error
}

func (s apiServer) getProfiles(ctx context.Context, hotelIDs []int32) chan profileResults {
	ch := make(chan profileResults, 1)

	go func() {
		req := &profile.Request{hotelIDs, "en"}
		res, err := s.GetProfiles(ctx, req)
		ch <- profileResults{res.Hotels, err}
	}()

	return ch
}

type profileResults struct {
	hotels []*profile.Hotel
	err    error
}

func main() {
	var (
		port        = flag.String("port", "8080", "The server port")
		authAddr    = flag.String("auth", "auth:8080", "The Auth server address in the format of host:port")
		geoAddr     = flag.String("geo", "geo:8080", "The Geo server address in the format of host:port")
		profileAddr = flag.String("profile", "profile:8080", "The Pofile server address in the format of host:port")
		rateAddr    = flag.String("rate", "rate:8080", "The Rate Code server address in the format of host:port")
	)
	flag.Parse()

	s := newServer(authAddr, geoAddr, profileAddr, rateAddr)
	http.HandleFunc("/", s.requestHandler)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
