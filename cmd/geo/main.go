package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"net"
	"strings"
	"time"

	"github.com/harlow/go-micro-services/data"
	"github.com/harlow/go-micro-services/proto/geo"
	"github.com/harlow/go-micro-services/trace"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// newServer returns a server with initialization data loaded.
func newServer() *geoServer {
	s := &geoServer{serverName: "service.geo"}
	s.loadLocations(data.MustAsset("data/locations.json"))
	return s
}

type location struct {
	HotelID int32
	Point   *geo.Point
}

type geoServer struct {
	serverName string
	locations  []location
}

// BoundedBox returns all hotels contained within a given rectangle.
func (s *geoServer) BoundedBox(ctx context.Context, rect *geo.GeoRequest) (*geo.GeoReply, error) {
	md, _ := metadata.FromContext(ctx)
	traceID := strings.Join(md["traceID"], ",")
	fromName := strings.Join(md["fromName"], ",")

	t := trace.Tracer{TraceID: traceID}
	t.In(s.serverName, fromName)
	defer t.Out(fromName, s.serverName, time.Now())

	reply := new(geo.GeoReply)
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
func inRange(point *geo.Point, rect *geo.GeoRequest) bool {
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

func main() {
	var port = flag.Int("port", 8080, "The server port")
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	geo.RegisterGeoServer(grpcServer, newServer())
	grpcServer.Serve(lis)
}
