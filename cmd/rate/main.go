package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"cloud.google.com/go/trace"
	"github.com/harlow/go-micro-services/data"
	"github.com/harlow/go-micro-services/lib"
	"github.com/harlow/go-micro-services/pb/rate"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type stay struct {
	HotelID string
	InDate  string
	OutDate string
}

type rateServer struct {
	traceClient *trace.Client
	rateTable   map[stay]*rate.RatePlan
}

// GetRates gets rates for hotels for specific date range.
func (s *rateServer) GetRates(ctx context.Context, req *rate.Request) (*rate.Result, error) {
	res := new(rate.Result)
	for _, hotelID := range req.HotelIds {
		stay := stay{
			HotelID: hotelID,
			InDate:  req.InDate,
			OutDate: req.OutDate,
		}
		if s.rateTable[stay] != nil {
			res.RatePlans = append(res.RatePlans, s.rateTable[stay])
		}
	}

	// add some artifical time so traces display nicely
	time.Sleep(1 * time.Millisecond)

	return res, nil
}

// loadRates loads rate codes from JSON file.
func loadRateTable(path string) map[stay]*rate.RatePlan {
	file := data.MustAsset("data/rates.json")

	rates := []*rate.RatePlan{}
	if err := json.Unmarshal(file, &rates); err != nil {
		log.Fatalf("Failed to load json: %v", err)
	}

	rateTable := make(map[stay]*rate.RatePlan)
	for _, ratePlan := range rates {
		stay := stay{
			HotelID: ratePlan.HotelId,
			InDate:  ratePlan.InDate,
			OutDate: ratePlan.OutDate,
		}
		rateTable[stay] = ratePlan
	}
	return rateTable
}

func main() {
	// server port
	var port = flag.Int("port", 8080, "The server port")
	flag.Parse()

	// tcp listener
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	tc := lib.NewTraceClient(
		os.Getenv("TRACE_PROJECT_ID"),
		os.Getenv("TRACE_JSON_CONFIG"),
	)

	// grpc server with rate endpoint
	srv := grpc.NewServer(
		grpc.UnaryInterceptor(trace.GRPCServerInterceptor(tc)),
	)
	rate.RegisterRateServer(srv, &rateServer{
		rateTable:   loadRateTable("data/rates.json"),
		traceClient: tc,
	})
	srv.Serve(lis)
}
