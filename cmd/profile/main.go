package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/harlow/go-micro-services/data"
	"github.com/harlow/go-micro-services/proto/profile"

	"golang.org/x/net/context"
	"golang.org/x/net/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// newServer returns a server with initialization data loaded.
func newServer() *profileServer {
	s := new(profileServer)
	s.loadProfiles(data.MustAsset("data/profiles.json"))
	return s
}

type profileServer struct {
	hotels map[int32]*profile.Hotel
}

// VerifyToken finds a customer by authentication token.
func (s *profileServer) GetProfiles(ctx context.Context, req *profile.Request) (*profile.Result, error) {
	md, _ := metadata.FromContext(ctx)
	traceID := strings.Join(md["traceID"], ",")

	if tr, ok := trace.FromContext(ctx); ok {
  	tr.LazyPrintf("traceID %s", traceID)
  }

	res := new(profile.Result)
	for _, i := range req.HotelIds {
		res.Hotels = append(res.Hotels, s.hotels[i])
	}

	return res, nil
}

// loadProfiles loads hotel profiles from a JSON file.
func (s *profileServer) loadProfiles(file []byte) {
	hotels := []*profile.Hotel{}
	if err := json.Unmarshal(file, &hotels); err != nil {
		log.Fatalf("Failed to load json: %v", err)
	}
	s.hotels = make(map[int32]*profile.Hotel)
	for _, hotel := range hotels {
		s.hotels[hotel.Id] = hotel
	}
}

func main() {
	var port = flag.Int("port", 8080, "The server port")
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	g := grpc.NewServer()
	profile.RegisterProfileServer(g, newServer())
	g.Serve(lis)
}
