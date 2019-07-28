package main

import (
	"flag"
	"log"

	"github.com/harlow/go-micro-services/dialer"
	"github.com/harlow/go-micro-services/services/frontend"
	profile "github.com/harlow/go-micro-services/services/profile/proto"
	search "github.com/harlow/go-micro-services/services/search/proto"
	"github.com/harlow/go-micro-services/trace"
)

func main() {
	var (
		port        = flag.Int("port", 5000, "The server port")
		searchAddr  = flag.String("searchaddr", "search:8080", "Search service addr")
		profileAddr = flag.String("profileaddr", "profile:8080", "Profile service addr")
		jaegeraddr  = flag.String("jaeger_addr", "jaeger:6831", "Jaeger address")
	)
	flag.Parse()

	tracer, err := trace.New("frontend", *jaegeraddr)
	if err != nil {
		log.Fatalf("trace new error: %v", err)
	}

	// dial search service
	sc, err := dialer.Dial(*searchAddr, dialer.WithTracer(tracer))
	if err != nil {
		log.Fatalf("dialer error: %v", err)
	}

	// dial profile service
	pc, err := dialer.Dial(*profileAddr, dialer.WithTracer(tracer))
	if err != nil {
		log.Fatalf("dialer error: %v", err)
	}

	srv := frontend.NewServer(
		search.NewSearchClient(sc),
		profile.NewProfileClient(pc),
		tracer,
	)
	srv.Run(*port)
}
