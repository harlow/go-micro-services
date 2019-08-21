package search

import (
	"fmt"
	"log"
	"net"

	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/harlow/go-micro-services/dialer"
	geo "github.com/harlow/go-micro-services/geo/proto"
	rate "github.com/harlow/go-micro-services/rate/proto"
	pb "github.com/harlow/go-micro-services/search/proto"
	opentracing "github.com/opentracing/opentracing-go"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

// NewServer returns a new server
func NewServer(t opentracing.Tracer, geoaddr, rateaddr string) *Server {
	// dial geo srv
	gc, err := dialer.Dial(geoaddr, dialer.WithTracer(t))
	if err != nil {
		log.Fatalf("dialer error: %v", err)
	}

	// dial rate srv
	rc, err := dialer.Dial(rateaddr, dialer.WithTracer(t))
	if err != nil {
		log.Fatalf("dialer error: %v", err)
	}

	return &Server{
		geoClient:  geo.NewGeoClient(gc),
		rateClient: rate.NewRateClient(rc),
		tracer:     t,
	}
}

// Server implments the search service
type Server struct {
	geoClient  geo.GeoClient
	rateClient rate.RateClient
	tracer     opentracing.Tracer
}

// Run starts the server
func (s *Server) Run(port int) error {
	srv := grpc.NewServer(
		grpc.UnaryInterceptor(
			otgrpc.OpenTracingServerInterceptor(s.tracer),
		),
	)
	pb.RegisterSearchServer(srv, s)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	return srv.Serve(lis)
}

// Nearby returns ids of nearby hotels ordered by ranking algo
func (s *Server) Nearby(ctx context.Context, req *pb.NearbyRequest) (*pb.SearchResult, error) {
	// find nearby hotels
	nearby, err := s.geoClient.Nearby(ctx, &geo.Request{
		Lat: req.Lat,
		Lon: req.Lon,
	})
	if err != nil {
		log.Fatalf("nearby error: %v", err)
	}

	// find rates for hotels
	rates, err := s.rateClient.GetRates(ctx, &rate.Request{
		HotelIds: nearby.HotelIds,
		InDate:   req.InDate,
		OutDate:  req.OutDate,
	})
	if err != nil {
		log.Fatalf("rates error: %v", err)
	}

	// build the response
	res := new(pb.SearchResult)
	for _, ratePlan := range rates.RatePlans {
		res.HotelIds = append(res.HotelIds, ratePlan.HotelId)
	}

	return res, nil
}
