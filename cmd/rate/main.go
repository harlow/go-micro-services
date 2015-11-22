package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/harlow/go-micro-services/data"
	"github.com/harlow/go-micro-services/proto/rate"

	"golang.org/x/net/context"
	"golang.org/x/net/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// newServer returns a server with initialization data loaded.
func newServer() *rateServer {
	s := new(rateServer)
	s.loadRates(data.MustAsset("data/rates.json"))
	return s
}

type stay struct {
	HotelID int32
	InDate  string
	OutDate string
}

type rateServer struct {
	rates map[stay]*rate.RatePlan
}

// GetRates gets rates for hotels for specific date range.
func (s *rateServer) GetRates(ctx context.Context, req *rate.Request) (*rate.Result, error) {
	md, _ := metadata.FromContext(ctx)
	traceID := strings.Join(md["traceID"], ",")

	if tr, ok := trace.FromContext(ctx); ok {
  	tr.LazyPrintf("traceID %s", traceID)
  }

	res := new(rate.Result)
	for _, hotelID := range req.HotelIds {
		k := stay{hotelID, req.InDate, req.OutDate}
		if s.rates[k] == nil {
			continue
		}
		res.RatePlans = append(res.RatePlans, s.rates[k])
	}

	return res, nil
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

func main() {
	var port = flag.Int("port", 8080, "The server port")
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	g := grpc.NewServer()
	rate.RegisterRateServer(g, newServer())
	g.Serve(lis)
}
