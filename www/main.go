package main

import (
	"errors"
	"net/http"
	"os"

	user "./../user_service/proto/user"

	"github.com/asim/go-micro/client"
	"github.com/golang/protobuf/proto"
	"github.com/harlow/auth_token"
)

func lookupUserByToken(u *user.User, authHeader string) (error) {
	token, err := auth_token.Parse(authHeader)

	if err != nil {
		return err
	}

	req := client.NewRequest("service.user", "Authentication.Call", &user.AuthRequest{
		AuthToken: proto.String(token),
		CallerID: proto.String("www"),
	})

	rsp := &user.AuthResponse{}

	if err := client.Call(req, rsp); err != nil {
		return err
	}

	if rsp.GetValid() == false {
		return errors.New("Unauthorized")
	}

	u = rsp.User
	return nil
}

func requestHandler(w http.ResponseWriter, r *http.Request) {
	u := &user.User{}
	err := lookupUserByToken(u, r.Header.Get("Authorization"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	w.Write([]byte("Hello world!"))
}

func main() {
	http.HandleFunc("/", requestHandler)
  http.ListenAndServe(":"+os.Getenv("PORT"), nil)
}
