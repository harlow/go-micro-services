package rate

import (
	"encoding/json"
	"fmt"
	"log"
	"net"

	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/harlow/go-micro-services/data"
	pb "github.com/harlow/go-micro-services/services/rate/proto"
	opentracing "github.com/opentracing/opentracing-go"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// Server implements the rate service
type Server struct {
	rateTable map[stay]*pb.RatePlan

	Tracer opentracing.Tracer
	Port   string
}

// Run starts the server
func (s *Server) Run() error {
	if s.Port == "" {
		return fmt.Errorf("server port must be set")
	}

	if s.rateTable == nil {
		s.rateTable = loadRateTable("data/inventory.json")
	}

	srv := grpc.NewServer(
		grpc.UnaryInterceptor(
			otgrpc.OpenTracingServerInterceptor(s.Tracer),
		),
	)

	pb.RegisterRateServer(srv, s)

	lis, err := net.Listen("tcp", ":"+s.Port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	return srv.Serve(lis)
}

// GetRates gets rates for hotels for specific date range.
func (s *Server) GetRates(ctx context.Context, req *pb.Request) (*pb.Result, error) {
	res := new(pb.Result)

	for _, hotelID := range req.HotelIds {
		stay := stay{
			HotelID: hotelID,
			InDate:  req.InDate,
			OutDate: req.OutDate,
		}
		if s.rateTable[stay] != nil {
			res.RatePlans = append(res.RatePlans, s.rateTable[stay])
		}
	}

	return res, nil
}

// loadRates loads rate codes from JSON file.
func loadRateTable(path string) map[stay]*pb.RatePlan {
	file := data.MustAsset(path)

	rates := []*pb.RatePlan{}
	if err := json.Unmarshal(file, &rates); err != nil {
		log.Fatalf("Failed to load json: %v", err)
	}

	rateTable := make(map[stay]*pb.RatePlan)
	for _, ratePlan := range rates {
		stay := stay{
			HotelID: ratePlan.HotelId,
			InDate:  ratePlan.InDate,
			OutDate: ratePlan.OutDate,
		}
		rateTable[stay] = ratePlan
	}

	return rateTable
}

type stay struct {
	HotelID string
	InDate  string
	OutDate string
}
