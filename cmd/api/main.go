package main

import (
	"context"
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

	uuid "github.com/nu7hatch/gouuid"

	"github.com/harlow/authtoken"
	"golang.org/x/net/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type inventory struct {
	Hotels    []*profile.Hotel `json:"hotels"`
	RatePlans []*rate.RatePlan `json:"ratePlans"`
}

type client struct {
	auth.AuthClient
	geo.GeoClient
	profile.ProfileClient
	rate.RateClient
}

func requestHandler(c client, w http.ResponseWriter, r *http.Request) {
	// tracing
	tr := trace.New("api.v1", r.URL.Path)
	defer tr.Finish()

	// context
	ctx := context.Background()
	ctx = trace.NewContext(ctx, tr)

	// add metadata to context for grpc calls
	traceID, _ := uuid.NewV4()
	ctx = metadata.NewContext(ctx, metadata.Pairs(
		"traceID", traceID.String(),
		"fromName", "api.v1",
	))

	// token from request headers
	token, err := authtoken.FromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	// verify token w/ auth service
	_, err = c.VerifyToken(ctx, &auth.Request{
		AuthToken: token,
	})
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// query params
	inDate := r.URL.Query().Get("inDate")
	outDate := r.URL.Query().Get("outDate")
	if inDate == "" || outDate == "" {
		http.Error(w, "Please specify inDate / outDate", http.StatusBadRequest)
		return
	}

	// finds nearby hotels
	// TODO(hw): use lat/lon from request params
	geoRes, err := c.Nearby(ctx, &geo.Request{
		Lat: 51.502973,
		Lon: -0.114723,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// make reqeusts for profiles and rates
	profileCh := getHotelProfiles(c, ctx, geoRes.HotelIds)
	rateCh := getRatePlans(c, ctx, geoRes.HotelIds, inDate, outDate)

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

	// render inventory response
	p := inventory{
		Hotels:    profileReply.hotels,
		RatePlans: rateReply.ratePlans,
	}
	json.NewEncoder(w).Encode(p)
}

type rateResults struct {
	ratePlans []*rate.RatePlan
	err       error
}

func getRatePlans(c client, ctx context.Context, hotelIDs []string, inDate string, outDate string) chan rateResults {
	ch := make(chan rateResults, 1)

	go func() {
		res, err := c.GetRates(ctx, &rate.Request{
			HotelIds: hotelIDs,
			InDate:   inDate,
			OutDate:  outDate,
		})
		ch <- rateResults{res.RatePlans, err}
	}()

	return ch
}

type profileResults struct {
	hotels []*profile.Hotel
	err    error
}

func getHotelProfiles(c client, ctx context.Context, hotelIDs []string) chan profileResults {
	ch := make(chan profileResults, 1)

	go func() {
		res, err := c.GetProfiles(ctx, &profile.Request{
			HotelIds: hotelIDs,
			Locale:   "en",
		})
		ch <- profileResults{res.Hotels, err}
	}()

	return ch
}

// mustDial ensures a tcp connection to specified address.
func mustDial(addr *string) *grpc.ClientConn {
	conn, err := grpc.Dial(*addr, grpc.WithInsecure())

	if err != nil {
		log.Fatalf("failed to dial: %v", err)
		panic(err)
	}

	return conn
}

func main() {
	// trace library patched for demo purposes.
	// https://github.com/golang/net/blob/master/trace/trace.go#L94
	trace.AuthRequest = func(req *http.Request) (any, sensitive bool) {
		return true, true
	}

	// ports for grpc connections (default uses docker-compose links)
	var (
		port        = flag.String("port", "8080", "The server port")
		authAddr    = flag.String("auth", "auth:8080", "The Auth server address in the format of host:port")
		geoAddr     = flag.String("geo", "geo:8080", "The Geo server address in the format of host:port")
		profileAddr = flag.String("profile", "profile:8080", "The Pofile server address in the format of host:port")
		rateAddr    = flag.String("rate", "rate:8080", "The Rate Code server address in the format of host:port")
	)
	flag.Parse()

	// client with all grpc connections
	c := client{
		AuthClient:    auth.NewAuthClient(mustDial(authAddr)),
		GeoClient:     geo.NewGeoClient(mustDial(geoAddr)),
		ProfileClient: profile.NewProfileClient(mustDial(profileAddr)),
		RateClient:    rate.NewRateClient(mustDial(rateAddr)),
	}

	// handle http requests
	http.HandleFunc("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestHandler(c, w, r)
	}))
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
