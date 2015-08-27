package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"time"

	"github.com/harlow/go-micro-services/profile"
	"github.com/harlow/go-micro-services/trace"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"strings"
)

var (
	port       = flag.Int("port", 10003, "The server port")
	jsonDBFile = flag.String("json_db_file", "data/profiles.json", "A json file containing a list of customers")
	serverName = "service.profile"
)

type profileServer struct {
	hotels map[int32]*profile.Hotel
}

// VerifyToken finds a customer by authentication token.
func (s *profileServer) GetHotels(ctx context.Context, args *profile.Args) (*profile.Reply, error) {
	md, _ := metadata.FromContext(ctx)
	t := trace.Tracer{TraceID: strings.Join(md["traceID"], ",")}
	t.In(serverName, strings.Join(md["from"], ","))
	defer t.Out(strings.Join(md["from"], ","), serverName, time.Now())

	reply := new(profile.Reply)
	for _, i := range args.HotelIds {
		reply.Hotels = append(reply.Hotels, s.hotels[i])
	}

	return reply, nil
}

// loadProfiles loads hotel profiles from a JSON file.
func (s *profileServer) loadProfiles(filePath string) {
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Failed to load file: %v", err)
	}
	hotels := []*profile.Hotel{}
	if err := json.Unmarshal(file, &hotels); err != nil {
		log.Fatalf("Failed to load json: %v", err)
	}
	s.hotels = make(map[int32]*profile.Hotel)
	for _, hotel := range hotels {
		s.hotels[hotel.Id] = hotel
	}
}

// newServer returns a server with initialization data loaded.
func newServer() *profileServer {
	s := new(profileServer)
	s.loadProfiles(*jsonDBFile)
	return s
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	profile.RegisterProfileServer(grpcServer, newServer())
	grpcServer.Serve(lis)
}
