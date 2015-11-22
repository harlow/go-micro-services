package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"net"
	"strings"

	"github.com/harlow/go-micro-services/data"
	"github.com/harlow/go-micro-services/proto/geo"

	"golang.org/x/net/context"
	"golang.org/x/net/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// newServer returns a server with initialization data loaded.
func newServer() *geoServer {
	s := new(geoServer)
	s.loadLocations(data.MustAsset("data/locations.json"))
	return s
}

type location struct {
	HotelID int32
	Point   *geo.Point
}

type geoServer struct {
	locations []location
}

// BoundedBox returns all hotels contained within a given rectangle.
func (s *geoServer) BoundedBox(ctx context.Context, req *geo.Request) (*geo.Result, error) {
	md, _ := metadata.FromContext(ctx)
	traceID := strings.Join(md["traceID"], ",")

	if tr, ok := trace.FromContext(ctx); ok {
  	tr.LazyPrintf("traceID %s", traceID)
  }

  res := new(geo.Result)
	for _, loc := range s.locations {
		if inRange(loc.Point, req) {
			res.HotelIds = append(res.HotelIds, loc.HotelID)
		}
	}

	return res, nil
}

// loadLocations loads hotel locations from a JSON file.
func (s *geoServer) loadLocations(file []byte) {
	if err := json.Unmarshal(file, &s.locations); err != nil {
		log.Fatalf("Failed to load hotels: %v", err)
	}
}

// inRange calculates if a point appears within a BoundingBox.
func inRange(point *geo.Point, req *geo.Request) bool {
	left := math.Min(float64(req.Lo.Longitude), float64(req.Hi.Longitude))
	right := math.Max(float64(req.Lo.Longitude), float64(req.Hi.Longitude))
	top := math.Max(float64(req.Lo.Latitude), float64(req.Hi.Latitude))
	bottom := math.Min(float64(req.Lo.Latitude), float64(req.Hi.Latitude))

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

	g := grpc.NewServer()
	geo.RegisterGeoServer(g, newServer())
	g.Serve(lis)
}
