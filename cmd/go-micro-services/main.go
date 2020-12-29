package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	services "github.com/harlow/go-micro-services"
	"github.com/harlow/go-micro-services/internal/trace"
	opentracing "github.com/opentracing/opentracing-go"
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

	t, err := trace.New("search", *jaegeraddr)
	if err != nil {
		log.Fatalf("trace new error: %v", err)
	}

	var srv server
	var cmd = os.Args[1]

	switch cmd {
	case "geo":
		srv = services.NewGeo(t)
	case "rate":
		srv = services.NewRate(t)
	case "profile":
		srv = services.NewProfile(t)
	case "search":
		srv = services.NewSearch(
			t,
			dial(*geoaddr, t),
			dial(*rateaddr, t),
		)
	case "frontend":
		srv = services.NewFrontend(
			t,
			dial(*searchaddr, t),
			dial(*profileaddr, t),
		)
	default:
		log.Fatalf("unknown cmd: %s", cmd)
	}

	if err := srv.Run(*port); err != nil {
		log.Fatalf("run %s error: %v", cmd, err)
	}
}

func dial(addr string, t opentracing.Tracer) *grpc.ClientConn {
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(otgrpc.OpenTracingClientInterceptor(t)),
	}

	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		panic(fmt.Sprintf("ERROR: dial error: %v", err))
	}

	return conn
}
