package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"./../protobufs/user"

	"code.google.com/p/goprotobuf/proto"
	"github.com/harlow/go-auth-middleware"
	"github.com/justinas/alice"
)

func authServiceUrl() string {
	return os.Getenv("AUTH_SERVICE_ADDRESS") + ":" + os.Getenv("AUTH_SERVICE_PORT")
}

type Authenticator struct{}

func (a Authenticator) Valid(token string) bool {
	stub, client, err := user.DialUserService("tcp", authServiceUrl())
	if err != nil {
		log.Fatal(`user.DialUserService error:`, err)
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
	log.Println("service1: running")
	app := http.HandlerFunc(appHandler)
	err := http.ListenAndServe(":"+os.Getenv("SERVICE1_PORT"), alice.New(authHandler, timeoutHandler).Then(app))
	if err != nil {
		fmt.Printf("http.ListenAndServe error: %v\n", err)
	}
}
