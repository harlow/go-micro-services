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

// newServer returns a server with initialization data loaded.
func newServer() *geoServer {
	s := new(geoServer)
	s.index = geoindex.NewClusteringIndex()
	s.loadLocations(data.MustAsset("data/locations.json"))
	return s
}

type geoServer struct {
	index *geoindex.ClusteringIndex
}

// BoundedBox returns all hotels contained within a given rectangle.
func (s *geoServer) BoundedBox(ctx context.Context, req *geo.Request) (*geo.Result, error) {
	md, _ := metadata.FromContext(ctx)
	traceID := strings.Join(md["traceID"], ",")

	if tr, ok := trace.FromContext(ctx); ok {
		tr.LazyPrintf("traceID %s", traceID)
	}

	point := &geoindex.GeoPoint{"", float64(req.Lat), float64(req.Lon)}
	points := s.index.KNearest(point, maxSearchResults, geoindex.Km(maxSearchRadius), func(p geoindex.Point) bool {
		return true
	})

	res := new(geo.Result)
	for _, p := range points {
		res.HotelIds = append(res.HotelIds, p.Id())
	}
	return res, nil
}

// loadLocations loads hotel locations from a JSON file.
func (s *geoServer) loadLocations(file []byte) {
	var points []*point
	if err := json.Unmarshal(file, &points); err != nil {
		log.Fatalf("Failed to load hotels: %v", err)
	}

	for _, point := range points {
		s.index.Add(point)
	}
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
