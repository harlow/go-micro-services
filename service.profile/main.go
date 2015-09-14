package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	// "time"

	pb "github.com/harlow/go-micro-services/service.profile/proto"
	// trace "github.com/harlow/go-micro-services/api.trace/client"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// newServer returns a server with initialization data loaded.
func newServer(dataPath string) *profileServer {
	s := &profileServer{}
	s.loadProfiles(dataPath)
	return s
}

type profileServer struct {
	hotels map[int32]*pb.Hotel
}

// VerifyToken finds a customer by authentication token.
func (s *profileServer) GetHotels(ctx context.Context, args *pb.Args) (*pb.Reply, error) {
	md, _ := metadata.FromContext(ctx)
	log.Printf("traceID=%s", md["traceID"])
	// trace.Req(md["traceID"], "service.profile", md["from"], "")
	// defer trace.Rep(md["traceID"], md["from"], "service.profile", time.Now())

	reply := new(pb.Reply)
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

	hotels := []*pb.Hotel{}
	if err := json.Unmarshal(file, &hotels); err != nil {
		log.Fatalf("Failed to load json: %v", err)
	}

	s.hotels = make(map[int32]*pb.Hotel)
	for _, hotel := range hotels {
		s.hotels[hotel.Id] = hotel
	}
}

func main() {
	var (
		port     = flag.Int("port", 10003, "The server port")
		dataPath = flag.String("data_path", "data/profiles.json", "A json file containing a list of customers")
	)
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterProfileServer(grpcServer, newServer(*dataPath))
	grpcServer.Serve(lis)
}
