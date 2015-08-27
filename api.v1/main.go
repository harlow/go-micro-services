package main

import (
	"encoding/json"
	_ "expvar"
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"

	auth "github.com/harlow/go-micro-services/service.auth/proto"
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
	serverName = "api.v1"
	port       = flag.String("port", "5000", "The server port")
)

type inventory struct {
	Hotels    []*profile.Hotel `json:"hotels"`
	RatePlans []*rate.RatePlan `json:"ratePlans"`
}

type api struct {
	auth.AuthClient
	geo.GeoClient
	profile.ProfileClient
	rate.RateClient
}

func (api api) requestHandler(w http.ResponseWriter, r *http.Request) {
	t := trace.NewTracer()
	t.In("www", "api.v1")
	defer t.Out("api.v1", "www", time.Now())

	// context and metadata
	md := metadata.Pairs("traceID", t.TraceID, "from", serverName)
	ctx := context.Background()
	ctx = metadata.NewContext(ctx, md)

	// parse token from Authorization header
	authToken, err := auth_token.Parse(r.Header.Get("Authorization"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	// verify auth token
	_, err = api.VerifyToken(ctx, &auth.Args{
		From:      serverName,
		AuthToken: authToken,
	})
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
	reply, err := api.BoundedBox(ctx, &geo.Rectangle{
		Lo: &geo.Point{Latitude: 400000000, Longitude: -750000000},
		Hi: &geo.Point{Latitude: 420000000, Longitude: -730000000},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	inventory := &inventory{}
	for i := 0; i < 2; i++ {
		select {
		case profileReply := <-api.getHotels(ctx, reply.HotelIds):
			if err := profileReply.err; err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			inventory.Hotels = profileReply.hotels
		case rateReply := <-api.getRatePlans(ctx, reply.HotelIds, inDate, outDate):
			if err := rateReply.err; err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			inventory.RatePlans = rateReply.ratePlans
		}
	}
	encoder := json.NewEncoder(w)
	if err = encoder.Encode(inventory); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type rateResults struct {
	ratePlans []*rate.RatePlan
	err       error
}

func (api api) getRatePlans(ctx context.Context, hotelIDs []int32, inDate string, outDate string) chan rateResults {
	ch := make(chan rateResults, 1)

	go func() {
		reply, err := api.GetRates(ctx,
			&rate.Args{
				HotelIds: hotelIDs,
				InDate:   inDate,
				OutDate:  outDate,
			})

		ch <- rateResults{
			ratePlans: reply.RatePlans,
			err:       err,
		}
	}()

	return ch
}

type profileResults struct {
	hotels []*profile.Hotel
	err    error
}

func (api api) getHotels(ctx context.Context, hotelIDs []int32) chan profileResults {
	ch := make(chan profileResults, 1)

	go func() {
		reply, err := api.GetHotels(ctx, &profile.Args{HotelIds: hotelIDs})

		ch <- profileResults{
			hotels: reply.Hotels,
			err:    err,
		}
	}()

	return ch
}

func main() {
	var (
		authServerAddr    = flag.String("auth_server_addr", "127.0.0.1:10001", "The Auth server address in the format of host:port")
		geoServerAddr     = flag.String("geo_server_addr", "127.0.0.1:10002", "The Geo server address in the format of host:port")
		profileServerAddr = flag.String("profile_server_addr", "127.0.0.1:10003", "The Pofile server address in the format of host:port")
		rateServerAddr    = flag.String("rate_server_addr", "127.0.0.1:10004", "The Rate Code server address in the format of host:port")
	)

	flag.Parse()

	api := newAPI(authServerAddr, geoServerAddr, profileServerAddr, rateServerAddr)
	http.HandleFunc("/", api.requestHandler)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}

func mustDial(addr *string) *grpc.ClientConn {
	conn, err := grpc.Dial(*addr)
	if err != nil {
		log.Fatalf("dial failed: %s", err)
		panic(err)
	}
	return conn
}

func newAPI(authAddr, geoAddr, profileAddr, rateAddr *string) api {
	return api{
		AuthClient:    auth.NewAuthClient(mustDial(authAddr)),
		GeoClient:     geo.NewGeoClient(mustDial(geoAddr)),
		ProfileClient: profile.NewProfileClient(mustDial(profileAddr)),
		RateClient:    rate.NewRateClient(mustDial(rateAddr)),
	}
}
