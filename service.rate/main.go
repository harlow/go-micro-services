package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"

	pb "github.com/harlow/go-micro-services/service.rate/proto"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	port       = flag.Int("port", 10004, "The server port")
	jsonDBFile = flag.String("json_db_file", "data/rates.json", "A json file containing a list of rate plans")
	serverName = "service.rate"
)

type stay struct {
	HotelId int32
	InDate  string
	OutDate string
}

type rateServer struct {
	rates map[stay]*pb.RatePlan
}

// GetRates gets rates for hotels for specific date range.
func (s *rateServer) GetRates(ctx context.Context, args *pb.Args) (*pb.Reply, error) {
	reply := new(pb.Reply)
	for _, hotelId := range args.HotelIds {
		k := stay{hotelId, args.InDate, args.OutDate}
		if s.rates[k] == nil {
			continue
		}
		reply.Rates = append(reply.Rates, s.rates[k])
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

// newServer returns a server with initialization data loaded.
func newServer() *rateServer {
	s := new(rateServer)
	s.loadRates(*jsonDBFile)
	return s
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterRateServer(grpcServer, newServer())
	grpcServer.Serve(lis)
}
