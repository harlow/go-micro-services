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
	"github.com/nu7hatch/gouuid"
)

var serviceID = "service1"

type Authenticator struct{}

func (a Authenticator) Valid(token string) bool {
	requestID, err := uuid.NewV4()
	if err != nil {
		log.Fatal(err)
	}

	req := user.AuthRequest{
		AuthToken: proto.String(token),
		CallerID:  proto.String(serviceID),
		RequestID: proto.String(requestID.String()),
	}
	resp := user.AuthResponse{}
	stub, client, err := user.DialUserService("tcp", ":"+os.Getenv("AUTH_SERVICE_PORT"))
	if err != nil {
		log.Fatalf("%s user.DialUserService error:", requestID.String(), err)
	}

	defer client.Close()
	log.Printf("%s rpc:auth_service status:begin\n", requestID.String())
	if err = stub.Auth(&req, &resp); err != nil {
		log.Printf("%s rpc:auth_service:error %v\n", requestID.String(), err)
	}
	log.Printf("%s rpc:auth_service status:complete success:%v\n", requestID.String(), resp.GetSuccess())

	return resp.GetSuccess()
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
	err := http.ListenAndServe(
		":"+os.Getenv("WEB_SERVICE_PORT"),
		alice.New(authHandler, timeoutHandler).Then(app),
	)
	if err != nil {
		fmt.Printf("http.ListenAndServe error: %v\n", err)
	}
}
