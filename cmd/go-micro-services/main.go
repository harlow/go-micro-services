package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	services "github.com/harlow/go-micro-services"
	"github.com/harlow/go-micro-services/internal/dialer"
	"github.com/harlow/go-micro-services/internal/trace"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
)

type server interface {
	Run(int) error
}

func main() {
	var (
		port        = flag.Int("port", 8080, "The service port")
		jaegeraddr  = flag.String("jaeger", "jaeger:6831", "Jaeger address")
		profileaddr = flag.String("profileaddr", "profile:8080", "Profile service addr")
		geoaddr     = flag.String("geoaddr", "geo:8080", "Geo server addr")
		rateaddr    = flag.String("rateaddr", "rate:8080", "Rate server addr")
		searchaddr  = flag.String("searchaddr", "search:8080", "Search service addr")
	)
	flag.Parse()

	tracer, err := trace.New("search", *jaegeraddr)
	if err != nil {
		log.Fatalf("trace new error: %v", err)
	}

	var srv server

	switch os.Args[1] {
	case "geo":
		srv = services.NewGeo(tracer)
	case "rate":
		srv = services.NewRate(tracer)
	case "profile":
		srv = services.NewProfile(tracer)
	case "search":
		srv = services.NewSearch(
			tracer,
			initGRPCConn(*geoaddr, tracer),
			initGRPCConn(*rateaddr, tracer),
		)
	case "frontend":
		srv = services.NewFrontend(
			tracer,
			initGRPCConn(*searchaddr, tracer),
			initGRPCConn(*profileaddr, tracer),
		)
	default:
		log.Fatalf("unknown command %s", os.Args[1])
	}

	srv.Run(*port)
}

func initGRPCConn(addr string, tracer opentracing.Tracer) *grpc.ClientConn {
	conn, err := dialer.Dial(addr, dialer.WithTracer(tracer))
	if err != nil {
		panic(fmt.Sprintf("ERROR: dial error: %v", err))
	}
	return conn
}
