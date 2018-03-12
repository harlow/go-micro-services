package main

import (
	"flag"

	"github.com/harlow/go-micro-services/services/api"
	profile "github.com/harlow/go-micro-services/services/profile/proto"
	search "github.com/harlow/go-micro-services/services/search/proto"
	"github.com/harlow/go-micro-services/tracing"
)

func main() {
	var (
		port        = flag.String("port", "8080", "The server port")
		searchAddr  = flag.String("searchaddr", "search:8080", "Search service addr")
		profileAddr = flag.String("profileaddr", "profile:8080", "Profile service addr")
		jaegerAddr  = flag.String("jaegeraddr", "jaeger:6831", "Jaeger server addr")
	)
	flag.Parse()

	var (
		tracer        = tracing.Init("api", *jaegerAddr)
		searchClient  = search.NewSearchClient(tracing.MustDial(*searchAddr, tracer))
		profileClient = profile.NewProfileClient(tracing.MustDial(*profileAddr, tracer))
	)

	srv := &api.Server{
		SearchClient:  searchClient,
		ProfileClient: profileClient,
		Tracer:        tracer,
		Port:          *port,
	}
	srv.Run()
}
