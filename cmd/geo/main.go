package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"net"
	"time"

	"github.com/harlow/go-micro-services/data"
	"github.com/harlow/go-micro-services/protos/geo"
	"github.com/harlow/go-micro-services/trace"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	port       = flag.Int("port", 8080, "The server port")
	serverName = "service.geo"
)

type location struct {
	HotelID int32
	Point   *geo.Point
}

type geoServer struct {
	locations []location
}

// BoundedBox returns all hotels contained within a given rectangle.
func (s *geoServer) BoundedBox(ctx context.Context, rect *geo.Rectangle) (*geo.Reply, error) {
	md, _ := metadata.FromContext(ctx)
	t := trace.Tracer{TraceID: md["traceID"]}
	t.In(serverName, md["from"])
	defer t.Out(md["from"], serverName, time.Now())

	reply := new(geo.Reply)
	for _, loc := range s.locations {
		if inRange(loc.Point, rect) {
			reply.HotelIds = append(reply.HotelIds, loc.HotelID)
		}
	}

	return reply, nil
}

// loadLocations loads hotel locations from a JSON file.
func (s *geoServer) loadLocations(file []byte) {
	if err := json.Unmarshal(file, &s.locations); err != nil {
		log.Fatalf("Failed to load hotels: %v", err)
	}
}

// inRange calculates if a point appears within a BoundingBox.
func inRange(point *geo.Point, rect *geo.Rectangle) bool {
	left := math.Min(float64(rect.Lo.Longitude), float64(rect.Hi.Longitude))
	right := math.Max(float64(rect.Lo.Longitude), float64(rect.Hi.Longitude))
	top := math.Max(float64(rect.Lo.Latitude), float64(rect.Hi.Latitude))
	bottom := math.Min(float64(rect.Lo.Latitude), float64(rect.Hi.Latitude))

	if float64(point.Longitude) >= left &&
		float64(point.Longitude) <= right &&
		float64(point.Latitude) >= bottom &&
		float64(point.Latitude) <= top {
		return true
	}
	return false
}

func newServer() *geoServer {
	s := new(geoServer)
	s.loadLocations(data.MustAsset("data/locations.json"))
	return s
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	geo.RegisterGeoServer(grpcServer, newServer())
	grpcServer.Serve(lis)
}
