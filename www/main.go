package main

import (
	"errors"
	"net/http"
	"net/rpc"
	"os"

	"github.com/harlow/auth_token"
)

type AuthRequest struct {
	Token string
}

type LikeRequest struct {
	PostID int
	UserID int
}

type LikeReply struct {
	Count int
}

type User struct {
	Email     string
	FirstName string
	ID        int
	LastName  string
}

func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}

func getUser(token string) (User, error) {
	args := AuthRequest{Token: token}
	user := &User{}
	url := os.Getenv("USER_SERVICE_URL")
	client, err := rpc.DialHTTP("tcp", url)

	if err != nil {
		return *user, errors.New(err.Error())
	}

	err = client.Call("UserService.Login", args, &user)

	if err != nil {
		return *user, errors.New(err.Error())
	}

	return *user, nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	token, err := auth_token.Parse(r.Header.Get("Authorization"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	user, err := getUser(token)

	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	url := os.Getenv("LIKE_SERVICE_URL")
	client, err := rpc.DialHTTP("tcp", url)
	args := &LikeRequest{UserID: user.ID, PostID: 1234}
	reply := &LikeReply{}

	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	err = client.Call("LikeService.Like", args, &reply)

	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	w.Write([]byte(user.FullName() + " w/ Likes: " + string(reply.Count)))
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":"+os.Getenv("PORT"), nil)
}
