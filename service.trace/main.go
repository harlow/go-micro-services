package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"time"

	pb "github.com/harlow/go-micro-services/service.trace/proto"
	trace "github.com/harlow/go-micro-services/trace"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	port       = flag.Int("port", 10005, "The server port")
	serverName = "service.trace"
)

type traceServer struct {
	events map[string]*pb.Trace
}

// GetRates gets rates for hotels for specific date range.
func (s *traceServer) Track(ctx context.Context, trace *pb.Trace) (*pb.Reply, error) {
	trace.events[trace.TraceID] = trace
	return *pb.Reply{}, nil
}

// newServer returns a server.
func newServer() *rateServer {
	s := new(rateServer)
	s.events = make(map[string]*pb.Event)
	return s
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterTrackServer(grpcServer, newServer())
	grpcServer.Serve(lis)
}
