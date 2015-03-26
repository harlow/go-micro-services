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

const ServiceName = "service.like"

type Args struct {
	UserID int
	PostID int
}

type Reply struct {
	Count int
}

type Service int

func (s Service) Like(args *Args, reply *Reply) error {
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
	s := new(Service)
	rpc.Register(s)
	rpc.HandleHTTP()
	ln, err := net.Listen("tcp", ":"+os.Getenv("LIKE_SERVICE_PORT"))

	if err != nil {
		log.Fatalf("net.Listen tcp :%v: %v", os.Getenv("LIKE_SERVICE_PORT"), err)
	}

	http.Serve(ln, nil)
}
