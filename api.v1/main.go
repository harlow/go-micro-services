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
	rateServerAddr = flag.String("rate_server_addr", "127.0.0.1:10004", "The Rate Code server address in the format of host:port")
)

type Inventory struct {
	Hotels []*profile.Hotel `json:"hotels"`
	Rates []*rate.RatePlan `json:"rates"`
}

func authenticateCustomer(t trace.Tracer, authToken string) error {
	t.Req(serverName, "service.auth", "AuthenticateCustomer")
	defer t.Rep("service.auth", serverName, time.Now())

	conn, err := grpc.Dial(*authServerAddr)
	if err != nil {
		return err
	}

	defer conn.Close()
	client := auth.NewAuthClient(conn)
	args := &auth.Args{TraceId: t.TraceID, From: "api.v1", AuthToken: authToken}
	_, err = client.VerifyToken(context.Background(), args)

	if err != nil {
		return err
	}
	return nil
}

func nearbyHotels(t trace.Tracer) ([]int32, error) {
	t.Req(serverName, "service.geo", "BoundingBox")
	defer t.Rep("service.geo", serverName, time.Now())

	conn, err := grpc.Dial(*geoServerAddr)
	if err != nil {
		return []int32{}, err
	}

	rect := &geo.Rectangle{
		&geo.Point{400000000, -750000000},
		&geo.Point{420000000, -730000000},
	}
	args := &geo.Args{t.TraceID, rect}
	client := geo.NewGeoClient(conn)
	reply, err := client.BoundedBox(context.Background(), args)
	if err != nil {
		log.Fatalf("%v.BoundedBox(_) = _, %v", conn, err)
	}

	return reply.HotelIds, nil
}

func hotelProfiles(t trace.Tracer, hotelIds []int32) ([]*profile.Hotel, error) {
	t.Req(serverName, "service.profile", "GetProfiles")
	defer t.Rep("service.profile", serverName, time.Now())

	conn, err := grpc.Dial(*profileServerAddr)
	if err != nil {
		return []*profile.Hotel{}, err
	}

	defer conn.Close()
	args := &profile.Args{
		TraceId: t.TraceID,
		From: "api.v1",
		HotelIds: hotelIds,
	}
	client := profile.NewProfileClient(conn)
	reply, err := client.GetProfiles(context.Background(), args)
	if err != nil {
		return []*profile.Hotel{}, err
	}

	return reply.Hotels, nil
}

func ratePlans(t trace.Tracer, args *rate.Args) ([]*rate.RatePlan, error) {
	t.Req(serverName, "service.rate", "GetRates")
	defer t.Rep("service.rate", serverName, time.Now())

	conn, err := grpc.Dial(*rateServerAddr)
	if err != nil {
		return []*rate.RatePlan{}, err
	}

	defer conn.Close()
	client := rate.NewRateClient(conn)
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

	authToken, err := auth_token.Parse(r.Header.Get("Authorization"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	err = authenticateCustomer(t, authToken)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	hotelIds, err := nearbyHotels(t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	hotels, err := hotelProfiles(t, hotelIds)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rateArgs := &rate.Args{
		TraceId: t.TraceID,
		HotelIds: hotelIds,
		InDate: "2015-04-09",
		OutDate: "2015-04-10",
	}
	rates, err := ratePlans(t, rateArgs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	inventory := Inventory{Hotels: hotels, Rates: rates}
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
