package main

import (
	"github.com/asim/go-micro/server"

	"./handler"

	log "github.com/golang/glog"
)

func main() {
	server.Name = "service.like"
	server.Init()

	server.Register(
		server.NewReceiver(
			new(handler.Example),
		),
	)

	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
