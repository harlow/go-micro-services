package main

import (
	"errors"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"fmt"
	"log"

	"github.com/garyburd/redigo/redis"
)

type Args struct {
	UserID int
	PostID int
}

type Reply struct {
	Count int
}

type LikeService int

func (s LikeService) Like(args *Args, reply *Reply) error {
	conn, err := redis.Dial("tcp", ":6379")
  defer conn.Close()

  if err != nil {
    fmt.Println(err)
    return errors.New(err.Error())
  }

  reply.Count = 4
  return nil
}

func main() {
	srv := new(LikeService)
	rpc.Register(srv)
	rpc.HandleHTTP()
	ln, err := net.Listen("tcp", ":"+os.Getenv("PORT"))

	if err != nil {
		log.Fatalf("net.Listen tcp :%v: %v", os.Getenv("PORT"), err)
	}

	http.Serve(ln, nil)
}
