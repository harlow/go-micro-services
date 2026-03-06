package main

import (
	"flag"
	"log"

	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	frontendsrv "github.com/harlow/go-micro-services/internal/services/frontend"
	geosrv "github.com/harlow/go-micro-services/internal/services/geo"
	profilesrv "github.com/harlow/go-micro-services/internal/services/profile"
	ratesrv "github.com/harlow/go-micro-services/internal/services/rate"
	searchsrv "github.com/harlow/go-micro-services/internal/services/search"
	"github.com/harlow/go-micro-services/internal/trace"
	opentracing "github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	if flag.NArg() < 1 {
		log.Fatalf("usage: go-micro-services <frontend|search|profile|geo|rate> [flags]")
	}
	cmd := flag.Arg(0)

	t, err := trace.New(cmd, *jaegeraddr)
	if err != nil {
		log.Fatalf("trace new error: %v", err)
	}

	var srv server

	switch cmd {
	case "geo":
		srv = geosrv.New(t)
	case "rate":
		srv = ratesrv.New(t)
	case "profile":
		srv = profilesrv.New(t)
	case "search":
		geoConn, err := dial(*geoaddr, t)
		if err != nil {
			log.Fatalf("dial geo error: %v", err)
		}
		rateConn, err := dial(*rateaddr, t)
		if err != nil {
			log.Fatalf("dial rate error: %v", err)
		}
		srv = searchsrv.New(
			t,
			geoConn,
			rateConn,
		)
	case "frontend":
		searchConn, err := dial(*searchaddr, t)
		if err != nil {
			log.Fatalf("dial search error: %v", err)
		}
		profileConn, err := dial(*profileaddr, t)
		if err != nil {
			log.Fatalf("dial profile error: %v", err)
		}
		srv = frontendsrv.New(
			t,
			searchConn,
			profileConn,
		)
	default:
		log.Fatalf("unknown cmd: %s", cmd)
	}

	if err := srv.Run(*port); err != nil {
		log.Fatalf("run %s error: %v", cmd, err)
	}
}

func dial(addr string, t opentracing.Tracer) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(otgrpc.OpenTracingClientInterceptor(t)),
	}

	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
