package profile

import (
	"encoding/json"
	"fmt"
	"log"
	"net"

	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/harlow/go-micro-services/data"
	pb "github.com/harlow/go-micro-services/profile/proto"
	opentracing "github.com/opentracing/opentracing-go"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// NewServer returns a new server
func NewServer(tr opentracing.Tracer) *Server {
	return &Server{
		tracer:   tr,
		profiles: loadProfiles("data/hotels.json"),
	}
}

// Server implements the profile service
type Server struct {
	profiles map[string]*pb.Hotel
	tracer   opentracing.Tracer
}

// Run starts the server
func (s *Server) Run(port int) error {
	srv := grpc.NewServer(
		grpc.UnaryInterceptor(
			otgrpc.OpenTracingServerInterceptor(s.tracer),
		),
	)
	pb.RegisterProfileServer(srv, s)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	return srv.Serve(lis)
}

// GetProfiles returns hotel profiles for requested IDs
func (s *Server) GetProfiles(ctx context.Context, req *pb.Request) (*pb.Result, error) {
	res := new(pb.Result)
	for _, id := range req.HotelIds {
		res.Hotels = append(res.Hotels, s.getProfile(id))
	}
	return res, nil
}

func (s *Server) getProfile(id string) *pb.Hotel {
	return s.profiles[id]
}

// loadProfiles loads hotel profiles from a JSON file.
func loadProfiles(path string) map[string]*pb.Hotel {
	var (
		file   = data.MustAsset(path)
		hotels []*pb.Hotel
	)

	if err := json.Unmarshal(file, &hotels); err != nil {
		log.Fatalf("Failed to load json: %v", err)
	}

	profiles := make(map[string]*pb.Hotel)
	for _, hotel := range hotels {
		profiles[hotel.Id] = hotel
	}
	return profiles
}
