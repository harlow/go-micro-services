package main

import (
	"log"
	"net/http"
	"net/rpc"
	"os"

	"github.com/harlow/auth_token"
)

type AuthRequest struct {
	Token string
}

type User struct {
	Email     string
	FirstName string
	ID        int32
	LastName  string
}

func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}

func handler(w http.ResponseWriter, r *http.Request) {
	token, err := auth_token.Parse(r.Header.Get("Authorization"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	url := os.Getenv("USER_SERVICE_URL")
	client, err := rpc.DialHTTP("tcp", url)

	if err != nil {
		log.Fatal("dialing:", err)
	}

	args := &AuthRequest{Token: token}
	user := &User{}

	err = client.Call("UserService.Login", args, &user)

	if err != nil {
		log.Print("service error: ", err)
		return
	}

	w.Write([]byte(user.FullName()))
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":"+os.Getenv("PORT"), nil)
}
