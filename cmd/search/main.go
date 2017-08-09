package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/harlow/go-micro-services/pb/geo"
	"github.com/harlow/go-micro-services/pb/rate"
	"github.com/harlow/go-micro-services/pb/search"
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
		port     = flag.Int("port", 8080, "The server port")
		geoAddr  = flag.String("geoaddr", "geo:8080", "Geo server addr")
		rateAddr = flag.String("rateaddr", "rate:8080", "Rate server addr")
	)
	flag.Parse()

	// tcp listener
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// grpc server
	srv := grpc.NewServer()
	search.RegisterSearchServer(srv, &server{
		geoClient:  geo.NewGeoClient(mustDial(*geoAddr)),
		rateClient: rate.NewRateClient(mustDial(*rateAddr)),
	})
	srv.Serve(lis)
}

// mustDial ensures a tcp connection to specified address.
func mustDial(addr string) *grpc.ClientConn {
	conn, err := grpc.Dial(
		addr,
		grpc.WithInsecure(),
	)
	if err != nil {
		log.Fatalf("failed to dial: %v", err)
		panic(err)
	}
	return conn
}
