package main

import (
  "log"
  "net"
  "net/http"
  "net/rpc"
  "os"
  "sync/atomic"
  "time"

  "../shared/req"
  "../shared/like"
)

type count32 int32

func (c *count32) incr() int32 {
    return atomic.AddInt32((*int32)(c), 1)
}

func (c *count32) get() int32 {
    return atomic.LoadInt32((*int32)(c))
}

var counter count32

type Service int

func (s Service) Like(args *like.Args, reply *like.Reply) error {
  req.LogIn(like.ServiceID, args.ServiceID)
  defer req.LogOut(like.ServiceID, args.ServiceID, time.Now())

  counter.incr()
  reply.Like.Count = counter.get()
  reply.Like.PostID = args.PostID
  return nil
}

func main() {
	s := new(Service)
	rpc.Register(s)
	rpc.HandleHTTP()
	ln, err := net.Listen("tcp", ":"+os.Getenv("LIKE_SERVICE_PORT"))

	if err != nil {
		log.Fatalf("net.Listen tcp :%v: %v", os.Getenv("LIKE_SERVICE_PORT"), err)
	}

	http.Serve(ln, nil)
}
