package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/harlow/go-micro-services/pb/geo"
	"github.com/harlow/go-micro-services/pb/rate"
	"github.com/harlow/go-micro-services/pb/search"
	"github.com/harlow/go-micro-services/tracing"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

type server struct {
	geoClient  geo.GeoClient
	rateClient rate.RateClient
}

// Nearby returns ids of nearby hotels ordered by ranking algo
func (s *server) Nearby(ctx context.Context, req *search.NearbyRequest) (*search.SearchResult, error) {
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

	// TODO(hw): add simple ranking algo to order hotel ids:
	// * geo distance
	// * price (best discount?)
	// * reviews

	// build the response
	res := new(search.SearchResult)
	for _, ratePlan := range rates.RatePlans {
		res.HotelIds = append(res.HotelIds, ratePlan.HotelId)
	}
	return res, nil
}

func main() {
	var (
		port       = flag.String("port", "8080", "The server port")
		geoAddr    = flag.String("geoaddr", "geo:8080", "Geo server addr")
		rateAddr   = flag.String("rateaddr", "rate:8080", "Rate server addr")
		jaegerAddr = flag.String("jaegeraddr", "jaeger:6831", "Jaeger server addr")
	)
	flag.Parse()

	// grpc server w/ tracing iterceptor
	var tracer = tracing.Init("search", *jaegerAddr)
	srv := grpc.NewServer(
		grpc.UnaryInterceptor(
			otgrpc.OpenTracingServerInterceptor(tracer),
		),
	)

	// register impl for pb definition
	var (
		geoClient  = geo.NewGeoClient(tracing.MustDial(*geoAddr, tracer))
		rateClient = rate.NewRateClient(tracing.MustDial(*rateAddr, tracer))
	)
	search.RegisterSearchServer(srv, &server{
		geoClient:  geoClient,
		rateClient: rateClient,
	})

	// listener
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	srv.Serve(lis)
}
