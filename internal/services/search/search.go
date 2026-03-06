package search

import (
	"fmt"
	"net"

	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	geo "github.com/harlow/go-micro-services/internal/services/geo/proto"
	rate "github.com/harlow/go-micro-services/internal/services/rate/proto"
	search "github.com/harlow/go-micro-services/internal/services/search/proto"
	opentracing "github.com/opentracing/opentracing-go"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

// New returns a new server
func New(t opentracing.Tracer, geoconn, rateconn *grpc.ClientConn) *Search {
	return &Search{
		geoClient:  geo.NewGeoClient(geoconn),
		rateClient: rate.NewRateClient(rateconn),
		tracer:     t,
	}
}

// Search implments the search service
type Search struct {
	geoClient  geo.GeoClient
	rateClient rate.RateClient
	tracer     opentracing.Tracer
}

// Run starts the server
func (s *Search) Run(port int) error {
	srv := grpc.NewServer(
		grpc.UnaryInterceptor(
			otgrpc.OpenTracingServerInterceptor(s.tracer),
		),
	)
	search.RegisterSearchServer(srv, s)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	return srv.Serve(lis)
}

// Nearby returns ids of nearby hotels ordered by ranking algo
func (s *Search) Nearby(ctx context.Context, req *search.NearbyRequest) (*search.SearchResult, error) {
	// find nearby hotels
	nearby, err := s.geoClient.Nearby(ctx, &geo.Request{
		Lat: req.Lat,
		Lon: req.Lon,
	})
	if err != nil {
		return nil, fmt.Errorf("nearby error: %w", err)
	}

	// find rates for hotels
	rates, err := s.rateClient.GetRates(ctx, &rate.Request{
		HotelIds: nearby.HotelIds,
		InDate:   req.InDate,
		OutDate:  req.OutDate,
	})
	if err != nil {
		return nil, fmt.Errorf("rates error: %w", err)
	}

	// build the response
	res := new(search.SearchResult)
	for _, ratePlan := range rates.RatePlans {
		res.HotelIds = append(res.HotelIds, ratePlan.HotelId)
	}

	return res, nil
}
