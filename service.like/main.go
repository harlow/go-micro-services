package main

import (
	"sync/atomic"
  "flag"
  "net"
  "fmt"
  "log"

	pb "../proto/like"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 10001, "The server port")
  name = "service.like"
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

// RecordLike incrments the like counter and returns total likes.
func (s *server) RecordLike(ctx context.Context, req *pb.LikeRequest) (*pb.LikeResponse, error) {
	counter.incr()
  like := &pb.Like{Count: counter.get(), PostID: req.PostID}
	return &pb.LikeResponse{Like: like, From: name}, nil
}

func main() {
  flag.Parse()
  lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
  if err != nil {
    log.Fatalf("failed to listen: %v", err)
  }
	s := new(server)
	grpcServer := grpc.NewServer()
	pb.RegisterLikeServiceServer(grpcServer, s)
	grpcServer.Serve(lis)
}
