package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/harlow/go-micro-services/data"
	"github.com/harlow/go-micro-services/pb/rate"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type stay struct {
	HotelID string
	InDate  string
	OutDate string
}

type server struct {
	rateTable map[stay]*rate.RatePlan
}

// GetRates gets rates for hotels for specific date range.
func (s *server) GetRates(ctx context.Context, req *rate.Request) (*rate.Result, error) {
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
	return res, nil
}

// loadRates loads rate codes from JSON file.
func loadRateTable(path string) map[stay]*rate.RatePlan {
	file := data.MustAsset(path)

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

	// grpc server
	srv := grpc.NewServer()
	rate.RegisterRateServer(srv, &server{
		rateTable: loadRateTable("data/inventory.json"),
	})
	srv.Serve(lis)
}
