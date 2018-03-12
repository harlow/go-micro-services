package main

import "github.com/harlow/go-micro-services/services/frontend"

func main() {
	srv := frontend.Server{}
	srv.Run()
}
