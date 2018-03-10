package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/hailocab/go-geoindex"
	"github.com/harlow/go-micro-services/data"
	"github.com/harlow/go-micro-services/pb/geo"
	"github.com/harlow/go-micro-services/tracing"
	opentracing "github.com/opentracing/opentracing-go"
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

type server struct {
	index *geoindex.ClusteringIndex
}

// Nearby returns all hotels within a given distance.
func (s *server) Nearby(ctx context.Context, req *geo.Request) (*geo.Result, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Inner.Nearby")
	defer span.Finish()

	var (
		points = s.getNearbyPoints(ctx, float64(req.Lat), float64(req.Lon))
		res    = &geo.Result{}
	)
	for _, p := range points {
		res.HotelIds = append(res.HotelIds, p.Id())
	}
	return res, nil
}

func (s *server) getNearbyPoints(ctx context.Context, lat, lon float64) []geoindex.Point {
	center := &geoindex.GeoPoint{
		Pid:  "",
		Plat: lat,
		Plon: lon,
	}
	return s.index.KNearest(
		center,
		maxSearchResults,
		geoindex.Km(maxSearchRadius), func(p geoindex.Point) bool {
			return true
		},
	)
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
	var (
		port       = flag.String("port", "8080", "The server port")
		jaegerAddr = flag.String("jaegeraddr", "jaeger:6831", "Jaeger server addr")
	)
	flag.Parse()

	var tracer = tracing.Init("geo", *jaegerAddr)

	srv := grpc.NewServer(
		grpc.UnaryInterceptor(
			otgrpc.OpenTracingServerInterceptor(tracer),
		),
	)

	geo.RegisterGeoServer(srv, &server{
		index: newGeoIndex("data/geo.json"),
	})

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	srv.Serve(lis)
}
