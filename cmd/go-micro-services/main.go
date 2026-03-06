package main

import (
	"context"
	"flag"
	"log"
	"time"

	frontendsrv "github.com/harlow/go-micro-services/internal/services/frontend"
	geosrv "github.com/harlow/go-micro-services/internal/services/geo"
	profilesrv "github.com/harlow/go-micro-services/internal/services/profile"
	ratesrv "github.com/harlow/go-micro-services/internal/services/rate"
	searchsrv "github.com/harlow/go-micro-services/internal/services/search"
	"github.com/harlow/go-micro-services/internal/trace"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type server interface {
	Run(int) error
}

func main() {
	var (
		port        = flag.Int("port", 8080, "The service port")
		jaegeraddr  = flag.String("jaeger", "jaeger:4317", "OTLP endpoint")
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

	shutdownTrace, err := trace.New(cmd, *jaegeraddr)
	if err != nil {
		log.Fatalf("trace init error: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdownTrace(ctx); err != nil {
			log.Printf("trace shutdown error: %v", err)
		}
	}()

	var srv server

	switch cmd {
	case "geo":
		srv = geosrv.New()
	case "rate":
		srv = ratesrv.New()
	case "profile":
		srv = profilesrv.New()
	case "search":
		geoConn, err := dial(*geoaddr)
		if err != nil {
			log.Fatalf("dial geo error: %v", err)
		}
		rateConn, err := dial(*rateaddr)
		if err != nil {
			log.Fatalf("dial rate error: %v", err)
		}
		srv = searchsrv.New(geoConn, rateConn)
	case "frontend":
		searchConn, err := dial(*searchaddr)
		if err != nil {
			log.Fatalf("dial search error: %v", err)
		}
		profileConn, err := dial(*profileaddr)
		if err != nil {
			log.Fatalf("dial profile error: %v", err)
		}
		srv = frontendsrv.New(searchConn, profileConn)
	default:
		log.Fatalf("unknown cmd: %s", cmd)
	}

	if err := srv.Run(*port); err != nil {
		log.Fatalf("run %s error: %v", cmd, err)
	}
}

func dial(addr string) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	}

	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
