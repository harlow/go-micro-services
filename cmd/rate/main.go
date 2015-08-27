package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/harlow/go-micro-services/rate"
	"github.com/harlow/go-micro-services/trace"

	"github.com/harlow/go-micro-services/data"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"strings"
)

var (
	port       = flag.Int("port", 8080, "The server port")
	serverName = "service.rate"
)

type stay struct {
	HotelID int32
	InDate  string
	OutDate string
}

type rateServer struct {
	rates map[stay]*rate.RatePlan
}

// GetRates gets rates for hotels for specific date range.
func (s *rateServer) GetRates(ctx context.Context, args *rate.Args) (*rate.Reply, error) {
	md, _ := metadata.FromContext(ctx)
	t := trace.Tracer{TraceID: strings.Join(md["traceID"], ",")}
	t.In(serverName, strings.Join(md["from"], ","))
	defer t.Out(strings.Join(md["from"], ","), serverName, time.Now())

	reply := new(rate.Reply)
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
func (s *rateServer) loadRates(file []byte) {
	rates := []*rate.RatePlan{}
	if err := json.Unmarshal(file, &rates); err != nil {
		log.Fatalf("Failed to load json: %v", err)
	}
	s.rates = make(map[stay]*rate.RatePlan)
	for _, ratePlan := range rates {
		k := stay{ratePlan.HotelId, ratePlan.InDate, ratePlan.OutDate}
		s.rates[k] = ratePlan
	}
}

// newServer returns a server with initialization data loaded.
func newServer() *rateServer {
	s := new(rateServer)
	s.loadRates(data.MustAsset("data/rates.json"))
	return s
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	rate.RegisterRateServer(grpcServer, newServer())
	grpcServer.Serve(lis)
}
