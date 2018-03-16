package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/harlow/go-micro-services/registry"
)

func usage() {
	fmt.Fprintf(os.Stderr, "USAGE\n")
	fmt.Fprintf(os.Stderr, "  serve <mode> [flags]\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Services\n")
	fmt.Fprintf(os.Stderr, "  all          Boots all services\n")
	fmt.Fprintf(os.Stderr, "  frontend     Web UI and JSON API\n")
	fmt.Fprintf(os.Stderr, "  geo          Geo service\n")
	fmt.Fprintf(os.Stderr, "  profile      Profile service\n")
	fmt.Fprintf(os.Stderr, "  rate         Rate service\n")
	fmt.Fprintf(os.Stderr, "  search       Search service\n")
	fmt.Fprintf(os.Stderr, "\n")
}

func main() {
	var (
		port       = flag.Int("port", 5000, "The server port")
		jaegeraddr = flag.String("jaegeraddr", "jaeger:6831", "Jaeger address")
		consuladdr = flag.String("consuladdr", "consul:8500", "Consul address")
	)
	flag.Parse()

	var run func(int, *registry.Client, string) error

	switch strings.ToLower(os.Args[1]) {
	case "all":
		run = runAll
	case "frontend":
		run = runFrontend
	case "geo":
		run = runGeo
	case "profile":
		run = runProfile
	case "rate":
		run = runRate
	case "search":
		run = runSearch
	default:
		usage()
		os.Exit(1)
	}

	registry, err := registry.NewClient(*consuladdr)
	if err != nil {
		panic(err)
	}

	if err := run(*port, registry, *jaegeraddr); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
