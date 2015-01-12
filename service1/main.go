package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"./../protobufs/user"

	"code.google.com/p/goprotobuf/proto"
	"github.com/harlow/go-auth-middleware"
	"github.com/justinas/alice"
)

type Authenticator struct{}

func (a Authenticator) Valid(token string) bool {
	stub, client, err := user.DialUserService("tcp", "127.0.0.1:1984")
	if err != nil {
		log.Fatal(`user.DialUserService("tcp", "127.0.0.1:1984"):`, err)
	}
	defer client.Close()
	var req user.AuthRequest
	var resp user.AuthResponse
	req.Token = proto.String(token)
	if err = stub.Auth(&req, &resp); err != nil {
		log.Fatal("stub.User error:", err)
	}
	return resp.GetValid()
}

func authHandler(h http.Handler) http.Handler {
	a := &Authenticator{}
	return auth.AuthHandler(h, a)
}

func timeoutHandler(h http.Handler) http.Handler {
	return http.TimeoutHandler(h, 1*time.Second, "timed out")
}

func appHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello world!"))
}

func main() {
	app := http.HandlerFunc(appHandler)
	err := http.ListenAndServe(":8000", alice.New(authHandler, timeoutHandler).Then(app))
	if err != nil {
		fmt.Printf("http.ListenAndServe error: %v\n", err)
	}
}
