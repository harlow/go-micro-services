package profile

import (
	"encoding/json"
	"fmt"
	"log"
	"net"

	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/harlow/go-micro-services/data"
	pb "github.com/harlow/go-micro-services/services/profile/proto"
	opentracing "github.com/opentracing/opentracing-go"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// Server implements the profile service
type Server struct {
	hotels map[string]*pb.Hotel

	Tracer opentracing.Tracer
	Port   string
}

// Run starts the server
func (s *Server) Run() error {
	if s.Port == "" {
		return fmt.Errorf("server port must be set")
	}

	if s.hotels == nil {
		s.hotels = loadProfiles("data/hotels.json")
	}

	srv := grpc.NewServer(
		grpc.UnaryInterceptor(
			otgrpc.OpenTracingServerInterceptor(s.Tracer),
		),
	)

	pb.RegisterProfileServer(srv, s)

	lis, err := net.Listen("tcp", ":"+s.Port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	return srv.Serve(lis)
}

// GetProfiles returns hotel profiles for requested IDs
func (s *Server) GetProfiles(ctx context.Context, req *pb.Request) (*pb.Result, error) {
	res := new(pb.Result)
	for _, i := range req.HotelIds {
		res.Hotels = append(res.Hotels, s.hotels[i])
	}
	return res, nil
}

// loadProfiles loads hotel profiles from a JSON file.
func loadProfiles(path string) map[string]*pb.Hotel {
	file := data.MustAsset(path)

	// unmarshal json profiles
	hotels := []*pb.Hotel{}
	if err := json.Unmarshal(file, &hotels); err != nil {
		log.Fatalf("Failed to load json: %v", err)
	}

	profiles := make(map[string]*pb.Hotel)
	for _, hotel := range hotels {
		profiles[hotel.Id] = hotel
	}

	return profiles
}
