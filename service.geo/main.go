package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net"
	"time"

	pb "github.com/harlow/go-micro-services/service.geo/proto"
	trace "github.com/harlow/go-micro-services/trace"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// newServer creates a new geoServer
// loads the locations from JSON data file
func newServer(dataPath string) *geoServer {
	s := &geoServer{serverName: "service.geo"}
	s.loadLocations(dataPath)
	return s
}

type location struct {
	HotelID int32
	Point   *pb.Point
}

type geoServer struct {
	serverName string
	locations  []location
}

// BoundedBox returns all hotels contained within a given rectangle.
func (s *geoServer) BoundedBox(ctx context.Context, rect *pb.Rectangle) (*pb.Reply, error) {
	md, _ := metadata.FromContext(ctx)

	t := trace.Tracer{TraceID: md["traceID"]}
	t.In(s.serverName, md["from"])
	defer t.Out(md["from"], s.serverName, time.Now())

	reply := new(pb.Reply)
	for _, loc := range s.locations {
		if inRange(loc.Point, rect) {
			reply.HotelIds = append(reply.HotelIds, loc.HotelID)
		}
	}

	return reply, nil
}

// loadLocations loads hotel locations from a JSON file.
func (s *geoServer) loadLocations(filePath string) {
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Failed to load file: %v", err)
	}
	if err := json.Unmarshal(file, &s.locations); err != nil {
		log.Fatalf("Failed to load hotels: %v", err)
	}
}

// inRange calculates if a point appears within a BoundingBox.
func inRange(point *pb.Point, rect *pb.Rectangle) bool {
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
	var (
		port     = flag.Int("port", 10002, "The server port")
		dataPath = flag.String("data_path", "data/locations.json", "A json file containing hotel locations")
	)
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterGeoServer(grpcServer, newServer(*dataPath))
	grpcServer.Serve(lis)
}
