package main

import (
	"encoding/json"
	_ "expvar"
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/harlow/go-micro-services/pb/auth"
	"github.com/harlow/go-micro-services/pb/geo"
	"github.com/harlow/go-micro-services/pb/profile"
	"github.com/harlow/go-micro-services/pb/rate"

	"github.com/harlow/authtoken"
	"golang.org/x/net/context"
	"golang.org/x/net/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// "github.com/nu7hatch/gouuid"
// 	traceID, _ := uuid.NewV4()
// 	traceID.String()

type inventory struct {
	Hotels    []*profile.Hotel `json:"hotels"`
	RatePlans []*rate.RatePlan `json:"ratePlans"`
}

type service struct {
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

func (s service) requestHandler(w http.ResponseWriter, r *http.Request) {
	// tracing
	tr := trace.New(s.serverName, "URL PATH!")
	defer tr.Finish()

	// metadata
	md := metadata.Pairs("traceID", "TRACEID", "fromName", s.serverName)

	// context
	ctx := context.Background()
	ctx = trace.NewContext(ctx, tr)
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
		Lat: 51.502973,
		Lon: -0.114723,
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

func (s service) getRatePlans(ctx context.Context, hotelIDs []string, inDate string, outDate string) chan rateResults {
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

func (s service) getProfiles(ctx context.Context, hotelIDs []string) chan profileResults {
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
	// trace library patched for demo purposes.
	// https://github.com/golang/net/blob/master/trace/trace.go#L94
	trace.AuthRequest = func(req *http.Request) (any, sensitive bool) {
		return true, true
	}

	var (
		port        = flag.String("port", "8080", "The server port")
		authAddr    = flag.String("auth", "auth:8080", "The Auth server address in the format of host:port")
		geoAddr     = flag.String("geo", "geo:8080", "The Geo server address in the format of host:port")
		profileAddr = flag.String("profile", "profile:8080", "The Pofile server address in the format of host:port")
		rateAddr    = flag.String("rate", "rate:8080", "The Rate Code server address in the format of host:port")
	)
	flag.Parse()

	svc := service{
		serverName:    "api.v1",
		AuthClient:    auth.NewAuthClient(mustDial(authAddr)),
		GeoClient:     geo.NewGeoClient(mustDial(geoAddr)),
		ProfileClient: profile.NewProfileClient(mustDial(profileAddr)),
		RateClient:    rate.NewRateClient(mustDial(rateAddr)),
	}
	http.HandleFunc("/", svc.requestHandler)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
