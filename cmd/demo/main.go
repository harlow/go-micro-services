package main

import (
	"flag"
	"log"

	"github.com/harlow/go-micro-services/frontend"
	"github.com/harlow/go-micro-services/geo"
	"github.com/harlow/go-micro-services/profile"
	"github.com/harlow/go-micro-services/rate"
	"github.com/harlow/go-micro-services/search"
	"github.com/harlow/go-micro-services/trace"
)

type server interface {
	Run(int) error
}

func main() {
	var (
		port        = flag.Int("port", 8080, "The service port")
		service     = flag.String("service", "", "The service to run")
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

	switch *service {
	case "geo":
		srv = geo.NewServer(tracer)
	case "search":
		srv = search.NewServer(tracer, *geoaddr, *rateaddr)
	case "rate":
		srv = rate.NewServer(tracer)
	case "profile":
		srv = profile.NewServer(tracer)
	case "frontend":
		srv = frontend.NewServer(tracer, *searchaddr, *profileaddr)
	default:
		log.Fatalf("unknown command %s", *service)
	}

	srv.Run(*port)
}
