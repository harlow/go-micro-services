package search

import (
	"fmt"
	"log"
	"net"

	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	geo "github.com/harlow/go-micro-services/services/geo/proto"
	rate "github.com/harlow/go-micro-services/services/rate/proto"
	pb "github.com/harlow/go-micro-services/services/search/proto"
	opentracing "github.com/opentracing/opentracing-go"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

// Server implments the search service
type Server struct {
	GeoClient  geo.GeoClient
	RateClient rate.RateClient
	Tracer     opentracing.Tracer
	Port       string
}

// Run starts the server
func (s *Server) Run() error {
	if s.Port == "" {
		return fmt.Errorf("server port must be set")
	}

	srv := grpc.NewServer(
		grpc.UnaryInterceptor(
			otgrpc.OpenTracingServerInterceptor(s.Tracer),
		),
	)

	pb.RegisterSearchServer(srv, s)

	lis, err := net.Listen("tcp", ":"+s.Port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	return srv.Serve(lis)
}

// Nearby returns ids of nearby hotels ordered by ranking algo
func (s *Server) Nearby(ctx context.Context, req *pb.NearbyRequest) (*pb.SearchResult, error) {
	// find nearby hotels
	nearby, err := s.GeoClient.Nearby(ctx, &geo.Request{
		Lat: req.Lat,
		Lon: req.Lon,
	})
	if err != nil {
		log.Fatalf("nearby error: %v", err)
	}

	// find rates for hotels
	rates, err := s.RateClient.GetRates(ctx, &rate.Request{
		HotelIds: nearby.HotelIds,
		InDate:   req.InDate,
		OutDate:  req.OutDate,
	})
	if err != nil {
		log.Fatalf("rates error: %v", err)
	}

	// TODO(hw): add simple ranking algo to order hotel ids:
	// * geo distance
	// * price (best discount?)
	// * reviews

	// build the response
	res := new(pb.SearchResult)
	for _, ratePlan := range rates.RatePlans {
		res.HotelIds = append(res.HotelIds, ratePlan.HotelId)
	}
	return res, nil
}
