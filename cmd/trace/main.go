package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/harlow/go-micro-services/trace"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	port       = flag.Int("port", 8080, "The server port")
	serverName = "service.trace"
)

type traceServer struct {
	events map[string]*trace.Trace
}

func (s *traceServer) Track(ctx context.Context, tracer *trace.Trace) (*trace.Reply, error) {
	s.events[tracer.TraceId] = tracer
	return &trace.Reply{}, nil
}

// newServer returns a server.
func newServer() *traceServer {
	return &traceServer{
		events: make(map[string]*trace.Trace),
	}
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	trace.RegisterTracerServer(grpcServer, newServer())
	grpcServer.Serve(lis)
}
