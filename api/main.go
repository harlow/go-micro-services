package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"

	user "./../user_service/proto/user"

	"github.com/asim/go-micro/client"
	"github.com/golang/protobuf/proto"
)

func parseToken(auth string) (string, error) {
	const basicScheme string = "Basic "
	const bearerScheme string = "Bearer "
	var token string

	// Confirm the request is sending Basic Authentication credentials.
	if !strings.HasPrefix(auth, basicScheme) && !strings.HasPrefix(auth, bearerScheme) {
		return "", errors.New("auth: Type not supported")
	}

	// Get the token from the request header
	// The first six characters are skipped - e.g. "Basic ".
	if strings.HasPrefix(auth, basicScheme) {
		str, err := base64.StdEncoding.DecodeString(auth[len(basicScheme):])
		if err != nil {
			return "", errors.New("auth: Base64 encoding issue")
		}
		creds := strings.Split(string(str), ":")
		token = creds[0]
	} else {
		token = auth[len(bearerScheme):]
	}

	return token, nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	token, err := parseToken(r.Header.Get("Authorization"))

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
  http.ListenAndServe(":8080", nil)
}
