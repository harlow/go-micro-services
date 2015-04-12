package main

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	auth "github.com/harlow/go-micro-services/service.auth/proto"
	geo "github.com/harlow/go-micro-services/service.geo/proto"
	profile "github.com/harlow/go-micro-services/service.profile/proto"

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
)

type Inventory struct {
	Points []*geo.Point `json:"point"`
	Profiles []*geo.Profile `json:"profiles"`
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
	_, err = client.VerifyToken(
		context.Background(),
		&auth.Args{
			TraceId:   t.TraceID,
			From:      "api.v1",
			AuthToken: authToken,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func nearbyHotels(t trace.Tracer) ([]int32, error) {
	t.Req(serverName, "service.geo", "NearbyLocations")
	defer t.Rep("service.geo", serverName, time.Now())

	conn, err := grpc.Dial(*geoServerAddr)
	if err != nil {
		return []*Hotel{}, err
	}

	rect := &geo.Rectangle{
		&geo.Point{400000000, -750000000},
		&geo.Point{420000000, -730000000},
	}
	geoArgs := &geo.Args{t.TraceID, rect}

	client := geo.NewGeoClient(conn)
	geoReply, err := client.BoundedBox(context.Background(), geoArgs)
	if err != nil {
		log.Fatalf("%v.BoundedBox(_) = _, %v", conn, err)
	}

	var wg sync.WaitGroup
	var hotels []Hotel

	for {
		loc, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("%v.NearbyLocations(_) = _, %v", conn, err)
		}
		h := &Hotel{ID: loc.HotelId, Address: loc.Address, Point: loc.Point}
		hotels = append(hotels, h)
		// wg.Add(1)
		// go setHotelDetails(t, h, &wg)
	}

	// wg.Wait()
	hotels = getHotelProfiles(t, hotels)
	return reply.HotelIds, nil
}

func getHotelProfiles(t trace.Tracer, h *Hotel) []Hotel {
	t.Req(serverName, "service.profile", "MultiProfile")
	defer t.Rep("service.profile", serverName, time.Now())
	defer wg.Done()

	conn, err := grpc.Dial(*profileServerAddr)
	if err != nil {
		return err
	}

	defer conn.Close()
	args := &profile.Args{TraceId: t.TraceID, From: "api.v1", HotelIds: h.ID}
	client := profile.NewHotelProfileClient(conn)
	reply, err := client.GetProfile(context.Background(), args)
	if err != nil {
		return err
	}

	h.Name = reply.Name
	h.PhoneNumber = reply.PhoneNumber
	h.Description = reply.Description
	return nil
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

	hotels, err := nearbyHotels(t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	body, err := json.Marshal(hotels)
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
