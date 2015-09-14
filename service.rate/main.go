package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	// "time"

	pb "github.com/harlow/go-micro-services/service.rate/proto"
	// trace "github.com/harlow/go-micro-services/api.trace/client"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// newServer returns a server with initialization data loaded.
func newServer(dataPath string) *rateServer {
	s := &rateServer{}
	s.loadRates(dataPath)
	return s
}

type stay struct {
	HotelID int32
	InDate  string
	OutDate string
}

type rateServer struct {
	serverName string
	rates      map[stay]*pb.RatePlan
}

// GetRates gets rates for hotels for specific date range.
func (s *rateServer) GetRates(ctx context.Context, args *pb.Args) (*pb.Reply, error) {
	md, _ := metadata.FromContext(ctx)
	log.Printf("traceID=%s", md["traceID"])

	reply := new(pb.Reply)
	for _, hotelID := range args.HotelIds {
		k := stay{hotelID, args.InDate, args.OutDate}
		if s.rates[k] == nil {
			continue
		}

		reply.RatePlans = append(reply.RatePlans, s.rates[k])
	}

	return reply, nil
}

// loadRates loads rate codes from JSON file.
func (s *rateServer) loadRates(filePath string) {
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Failed to load file: %v", err)
	}

	rates := []*pb.RatePlan{}
	if err := json.Unmarshal(file, &rates); err != nil {
		log.Fatalf("Failed to load json: %v", err)
	}

	s.rates = make(map[stay]*pb.RatePlan)
	for _, ratePlan := range rates {
		k := stay{ratePlan.HotelId, ratePlan.InDate, ratePlan.OutDate}
		s.rates[k] = ratePlan
	}
}

func main() {
	var (
		port     = flag.Int("port", 10004, "The server port")
		dataPath = flag.String("data_path", "data/rates.json", "A json file containing a list of rate plans")
	)
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterRateServer(grpcServer, newServer(*dataPath))
	grpcServer.Serve(lis)
}
