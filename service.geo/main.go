package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net"

	pb "github.com/harlow/go-micro-services/service.geo/proto"

	"google.golang.org/grpc"
)

var (
	port       = flag.Int("port", 10002, "The server port")
	jsonDBFile = flag.String("json_db_file", "data/locations.json", "A json file containing hotel locations")
	serverName = "service.geo"
)

type geoServer struct {
	locations []*pb.Location
}

// NearbyLocations returns all hotels contained within bounding BoundingBox.
func (s *geoServer) NearbyLocations(args *pb.Args, stream pb.Geo_NearbyLocationsServer) error {
	for _, loc := range s.locations {
		if inRange(loc.Point, args) {
			if err := stream.Send(loc); err != nil {
				return err
			}
		}
	}
	return nil
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
func inRange(point *pb.Point, args *pb.Args) bool {
	left := math.Min(float64(args.Lo.Longitude), float64(args.Hi.Longitude))
	right := math.Max(float64(args.Lo.Longitude), float64(args.Hi.Longitude))
	top := math.Max(float64(args.Lo.Latitude), float64(args.Hi.Latitude))
	bottom := math.Min(float64(args.Lo.Latitude), float64(args.Hi.Latitude))

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
	s.loadLocations(*jsonDBFile)
	return s
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterGeoServer(grpcServer, newServer())
	grpcServer.Serve(lis)
}
