package main

import (
	"context"
	"encoding/json"
	_ "expvar"
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/harlow/go-micro-services/pb/geo"
	"github.com/harlow/go-micro-services/pb/profile"
	"github.com/harlow/go-micro-services/pb/rate"

	uuid "github.com/nu7hatch/gouuid"

	"golang.org/x/net/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type response struct {
	Type     string    `json:"type"`
	Features []feature `json:"features"`
}

type feature struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	Properties struct {
		Name        string `json:"name"`
		PhoneNumber string `json:"phone_number"`
	} `json:"properties"`
	Geometry struct {
		Type        string    `json:"type"`
		Coordinates []float32 `json:"coordinates"`
	} `json:"geometry"`
}

type client struct {
	geo.GeoClient
	profile.ProfileClient
	rate.RateClient
}

func geoJSONResponse(hotels []*profile.Hotel) response {
	r := response{
		Type: "FeatureCollection",
	}

	for _, hotel := range hotels {
		f := feature{
			Type: "Feature",
			ID:   hotel.Id,
		}
		f.Properties.Name = hotel.Name
		f.Properties.PhoneNumber = hotel.PhoneNumber
		f.Geometry.Type = "Point"
		f.Geometry.Coordinates = []float32{
			hotel.Address.Lon,
			hotel.Address.Lat,
		}
		r.Features = append(r.Features, f)
	}

	return r
}

func requestHandler(c client, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// tracing
	tr := trace.New("api.v1", r.URL.Path)
	defer tr.Finish()

	// context
	ctx := context.Background()
	ctx = trace.NewContext(ctx, tr)

	// add a unique request id to context
	if traceID, err := uuid.NewV4(); err == nil {
		ctx = metadata.NewContext(ctx, metadata.Pairs(
			"traceID", traceID.String(),
			"fromName", "api.v1",
		))
	}

	// checkin and checkout date query params
	inDate, outDate := r.URL.Query().Get("inDate"), r.URL.Query().Get("outDate")
	if inDate == "" || outDate == "" {
		http.Error(w, "Please specify inDate/outDate params", http.StatusBadRequest)
		return
	}

	// finds nearby hotels
	// TODO(hw): use lat/lon from request params
	nearby, err := c.Nearby(ctx, &geo.Request{
		Lat: 37.7749,
		Lon: -122.4194,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// make reqeusts for profiles and rates
	profileCh := getHotelProfiles(c, ctx, nearby.HotelIds)
	rateCh := getRatePlans(c, ctx, nearby.HotelIds, inDate, outDate)

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

	// GeoJSON response
	json.NewEncoder(w).Encode(
		geoJSONResponse(profileReply.hotels),
	)
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
		geoAddr     = flag.String("geo", "geo:8080", "The Geo server address in the format of host:port")
		profileAddr = flag.String("profile", "profile:8080", "The Pofile server address in the format of host:port")
		rateAddr    = flag.String("rate", "rate:8080", "The Rate Code server address in the format of host:port")
	)
	flag.Parse()

	// client with all grpc connections
	c := client{
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
