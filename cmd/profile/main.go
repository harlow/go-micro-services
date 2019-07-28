package main

import (
	"flag"
	"log"

	"github.com/harlow/go-micro-services/services/profile"
	"github.com/harlow/go-micro-services/trace"
)

func main() {
	var (
		port       = flag.Int("port", 8080, "The server port")
		jaegeraddr = flag.String("jaeger_addr", "jaeger:6831", "Jaeger address")
	)
	flag.Parse()

	tracer, err := trace.New("profile", *jaegeraddr)
	if err != nil {
		log.Fatalf("trace new error: %v", err)
	}

	srv := profile.NewServer(tracer)
	srv.Run(*port)
}
