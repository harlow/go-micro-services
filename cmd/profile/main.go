package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"cloud.google.com/go/trace"
	"github.com/harlow/go-micro-services/data"
	"github.com/harlow/go-micro-services/lib"
	"github.com/harlow/go-micro-services/pb/profile"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type server struct {
	traceClient *trace.Client
	hotels      map[string]*profile.Hotel
}

// GetProfiles returns hotel profiles for requested IDs
func (s *server) GetProfiles(ctx context.Context, req *profile.Request) (*profile.Result, error) {
	md, _ := metadata.FromContext(ctx)
	span := s.traceClient.SpanFromHeader("/svc.Profile/GetProfiles", strings.Join(md["trace"], ""))
	defer span.Finish()

	res := new(profile.Result)
	for _, i := range req.HotelIds {
		res.Hotels = append(res.Hotels, s.hotels[i])
	}
	return res, nil
}

// loadProfiles loads hotel profiles from a JSON file.
func loadProfiles(path string) map[string]*profile.Hotel {
	file := data.MustAsset(path)

	// unmarshal json profiles
	hotels := []*profile.Hotel{}
	if err := json.Unmarshal(file, &hotels); err != nil {
		log.Fatalf("Failed to load json: %v", err)
	}

	profiles := make(map[string]*profile.Hotel)
	for _, hotel := range hotels {
		profiles[hotel.Id] = hotel
	}
	return profiles
}

func main() {
	// service port
	var port = flag.Int("port", 8080, "The server port")
	flag.Parse()

	// tcp listener
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	traceClient := lib.NewTraceClient(
		os.Getenv("TRACE_PROJECT_ID"),
		os.Getenv("TRACE_JSON_CONFIG"),
	)

	// grpc server with profiles endpoint
	srv := grpc.NewServer()
	profile.RegisterProfileServer(srv, &server{
		hotels:      loadProfiles("data/profiles.json"),
		traceClient: traceClient,
	})
	srv.Serve(lis)
}
