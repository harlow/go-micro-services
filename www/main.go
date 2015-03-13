package main

import (
	"fmt"
	"net/http"
	"os"

	user "./../user_service/proto/user"

	"github.com/asim/go-micro/client"
	"github.com/golang/protobuf/proto"
	"github.com/harlow/auth_token"
)

func handler(w http.ResponseWriter, r *http.Request) {
	token, err := auth_token.Parse(r.Header.Get("Authorization"))

	if err != nil {
		http.Error(w, "Auth Token Required", http.StatusForbidden)
		return
	}

	req := client.NewRequest("service.user", "Authentication.Call", &user.AuthRequest{
		AuthToken: proto.String(token),
	})

	rsp := &user.AuthResponse{}

	if err := client.Call(req, rsp); err != nil {
		fmt.Println(err)
		return
	}

	if rsp.GetValid() == true {
		w.Write([]byte("Hello world!"))
	} else {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
}

func main() {
	http.HandleFunc("/", handler)
  http.ListenAndServe(":"+os.Getenv("PORT"), nil)
}
