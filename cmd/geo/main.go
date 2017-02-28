package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"

	"time"

	"cloud.google.com/go/trace"
	"github.com/hailocab/go-geoindex"
	"github.com/harlow/go-micro-services/data"
	"github.com/harlow/go-micro-services/lib"
	"github.com/harlow/go-micro-services/pb/geo"
	"github.com/harlow/grpc-google-cloud-trace/interceptor"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	maxSearchRadius  = 10
	maxSearchResults = 5
)

type point struct {
	Pid  string  `json:"hotelId"`
	Plat float64 `json:"lat"`
	Plon float64 `json:"lon"`
}

// Implement Point interface
func (p *point) Lat() float64 { return p.Plat }
func (p *point) Lon() float64 { return p.Plon }
func (p *point) Id() string   { return p.Pid }

type geoServer struct {
	traceClient *trace.Client
	index       *geoindex.ClusteringIndex
}

// Nearby returns all hotels within a given distance.
func (s *geoServer) Nearby(ctx context.Context, req *geo.Request) (*geo.Result, error) {
	points := s.getNearbyPoints(ctx, float64(req.Lat), float64(req.Lon))

	// add some artifical time so traces display nicely
	time.Sleep(time.Duration(rand.Int31n(5)) * time.Millisecond)

	res := &geo.Result{}
	for _, p := range points {
		res.HotelIds = append(res.HotelIds, p.Id())
	}

	return res, nil
}

func (s *geoServer) getNearbyPoints(ctx context.Context, lat, lon float64) []geoindex.Point {
	span := trace.FromContext(ctx).NewChild("getNearbyPoints")
	defer span.Finish()

	// add some artifical time so traces display nicely
	time.Sleep(1 * time.Millisecond)

	center := &geoindex.GeoPoint{Pid: "", Plat: lat, Plon: lon}
	points := s.index.KNearest(center, maxSearchResults, geoindex.Km(maxSearchRadius), func(p geoindex.Point) bool {
		return true
	})
	return points
}

// newGeoIndex returns a geo index with points loaded
func newGeoIndex(path string) *geoindex.ClusteringIndex {
	file := data.MustAsset(path)

	// unmarshal json points
	var points []*point
	if err := json.Unmarshal(file, &points); err != nil {
		log.Fatalf("Failed to load hotels: %v", err)
	}

	// add points to index
	index := geoindex.NewClusteringIndex()
	for _, point := range points {
		index.Add(point)
	}
	return index
}

func main() {
	// port number
	var port = flag.Int("port", 8080, "The server port")
	flag.Parse()

	// tcp listener
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	traceClient := lib.NewTraceClient(
		os.Getenv("TRACE_PROJECT_ID"),
		os.Getenv("TRACE_JSON_CONFIG"),
	)

	// grpc server
	srv := grpc.NewServer(
		grpc.UnaryInterceptor(interceptor.Server(traceClient)),
	)
	geo.RegisterGeoServer(srv, &geoServer{
		index:       newGeoIndex("data/locations.json"),
		traceClient: traceClient,
	})
	srv.Serve(lis)
}
