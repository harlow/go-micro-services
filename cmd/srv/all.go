package main

import (
	"github.com/harlow/go-micro-services/registry"
)

func runAll(port int, registry *registry.Client, jaegeraddr string) error {
	go runGeo(8001, registry, jaegeraddr)
	go runProfile(8002, registry, jaegeraddr)
	go runRate(8003, registry, jaegeraddr)
	go runSearch(8004, registry, jaegeraddr)

	return runFrontend(5000, registry, jaegeraddr)
}
