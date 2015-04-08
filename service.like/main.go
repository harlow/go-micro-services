package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"sync/atomic"

  "github.com/harlow/go-micro-services/proto/like"
  "golang.org/x/net/context"
  "google.golang.org/grpc"
)

var (
	port       = flag.Int("port", 10001, "The server port")
	serverName = "service.like"
)

type count32 int32

func (c *count32) incr() int32 {
	return atomic.AddInt32((*int32)(c), 1)
}

func (c *count32) get() int32 {
	return atomic.LoadInt32((*int32)(c))
}

var counter count32

type server int

// RecordLike records a like for a post.
func (s *server) RecordLike(ctx context.Context, req *like.Req) (*like.Like, error) {
	counter.incr()
	l := &like.Like{Count: counter.get(), PostID: req.PostID}
	return l, nil
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := new(server)
	grpcServer := grpc.NewServer()
	like.RegisterLikeServiceServer(grpcServer, s)
	grpcServer.Serve(lis)
}
