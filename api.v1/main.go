package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"time"

	auth "github.com/harlow/go-micro-services/service.auth/proto"
	geo "github.com/harlow/go-micro-services/service.geo/proto"
	profile "github.com/harlow/go-micro-services/service.profile/proto"
	rate "github.com/harlow/go-micro-services/service.rate/proto"

	"github.com/harlow/auth_token"
	"github.com/harlow/go-micro-services/trace"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	serverName        = "api.v1"
	port              = flag.String("port", "5000", "The server port")
	authServerAddr    = flag.String("auth_server_addr", "127.0.0.1:10001", "The Auth server address in the format of host:port")
	geoServerAddr     = flag.String("geo_server_addr", "127.0.0.1:10002", "The Geo server address in the format of host:port")
	profileServerAddr = flag.String("profile_server_addr", "127.0.0.1:10003", "The Pofile server address in the format of host:port")
	rateServerAddr    = flag.String("rate_server_addr", "127.0.0.1:10004", "The Rate Code server address in the format of host:port")
)

type inventory struct {
	Hotels []*profile.Hotel `json:"hotels"`
	Rates  []*rate.RatePlan `json:"rates"`
}

func verifyToken(traceID string, serverName string, authToken string) error {
	// dial server connection
	conn, err := grpc.Dial(*authServerAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	// set up args and client
	args := &auth.Args{TraceID: traceID, From: serverName, AuthToken: authToken}
	client := auth.NewAuthClient(conn)

	// verify auth token
	_, err = client.VerifyToken(context.Background(), args)
	if err != nil {
		return err
	}

	return nil
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
		&geo.Point{400000000, -750000000},
		&geo.Point{420000000, -730000000},
	}
	args := &geo.Args{TraceID: traceID, From: serverName, Rect: rect}
	client := geo.NewGeoClient(conn)

	// get hotels within bounded bob
	reply, err := client.BoundedBox(context.Background(), args)
	if err != nil {
		return []int32{}, err
	}

	return reply.HotelIds, nil
}

func hotelProfiles(traceID string, serverName string, hotelIds []int32) ([]*profile.Hotel, error) {
	// dial server connection
	conn, err := grpc.Dial(*profileServerAddr)
	if err != nil {
		return []*profile.Hotel{}, err
	}
	defer conn.Close()

	// set up args
	args := &profile.Args{TraceID: traceID, From: serverName, HotelIds: hotelIds}
	client := profile.NewProfileClient(conn)

	// get profile data
	reply, err := client.GetProfiles(context.Background(), args)
	if err != nil {
		return []*profile.Hotel{}, err
	}

	return reply.Hotels, nil
}

func getRates(traceID string, serverName string, hotelIds []int32, inDate string, outDate string) ([]*rate.RatePlan, error) {
	// dial server connection
	conn, err := grpc.Dial(*rateServerAddr)
	if err != nil {
		return []*rate.RatePlan{}, err
	}
	defer conn.Close()

	// set up args
	args := &rate.Args{
		TraceID:  traceID,
		From:     serverName,
		HotelIds: hotelIds,
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

func requestHandler(w http.ResponseWriter, r *http.Request) {
	t := trace.NewTracer()
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
	err = verifyToken(t.TraceID, serverName, authToken)
	t.Rep("service.auth", serverName, time.Now())

	if err != nil {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// search for hotels within geo rectangle
	t.Req(serverName, "service.geo", "BoundedBox")
	hotelIds, err := hotelsWithinBoundedBox(t.TraceID, serverName, 100, 100)
	t.Rep("service.geo", serverName, time.Now())

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// fetch hotel profiles
	t.Req(serverName, "service.profile", "GetProfiles")
	profiles, err := hotelProfiles(t.TraceID, serverName, hotelIds)
	t.Rep("service.profile", serverName, time.Now())

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// fetch hotel rate plans
	t.Req(serverName, "service.rate", "GetRates")
	ratePlans, err := getRates(t.TraceID, serverName, hotelIds, inDate, outDate)
	t.Rep("service.rate", serverName, time.Now())

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// marshal json body
	inventory := inventory{Hotels: profiles, Rates: ratePlans}
	body, err := json.Marshal(inventory)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(body)
}

func main() {
	flag.Parse()
	http.HandleFunc("/", requestHandler)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
