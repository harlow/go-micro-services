package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/hailocab/go-geoindex"
	"github.com/harlow/go-micro-services/data"
	"github.com/harlow/go-micro-services/pb/geo"

	"golang.org/x/net/context"
	"golang.org/x/net/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
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
	index *geoindex.ClusteringIndex
}

// Nearby returns all hotels within a given distance.
func (s *geoServer) Nearby(ctx context.Context, req *geo.Request) (*geo.Result, error) {
	md, _ := metadata.FromContext(ctx)
	traceID := strings.Join(md["traceID"], ",")

	if tr, ok := trace.FromContext(ctx); ok {
		tr.LazyPrintf("traceID %s", traceID)
	}

	// create center point for query
	center := &geoindex.GeoPoint{
		Pid:  "",
		Plat: float64(req.Lat),
		Plon: float64(req.Lon),
	}

	// find points around center point
	points := s.index.KNearest(center, maxSearchResults, geoindex.Km(maxSearchRadius), func(p geoindex.Point) bool {
		return true
	})

	res := &geo.Result{}
	for _, p := range points {
		res.HotelIds = append(res.HotelIds, p.Id())
	}
	return res, nil
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

	// grpc server
	srv := grpc.NewServer()
	geo.RegisterGeoServer(srv, &geoServer{
		index: newGeoIndex("data/locations.json"),
	})
	srv.Serve(lis)
}
