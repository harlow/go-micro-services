package main

import (
	"errors"
	"log"
	"net/http"
	"net/rpc"
	"os"
	"time"

	"github.com/harlow/auth_token"
)

const (
	APIName         = "api.like"
	UserServiceName = "service.user"
	LikeServiceName = "service.like"
)

type AuthRequest struct {
	AuthToken string
	From      string
	RequestID string
}

type AuthResponse struct {
	From      string
	RequestID string
	User      User
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

type LikeRequest struct {
	From      string
	PostID    int
	RequestID string
	UserID    int
}

type LikeResponse struct {
	Count int
}

func getUser(token string) (User, error) {
	logRequest(UserServiceName)
	defer logResponse(UserServiceName, time.Now())

	args := AuthRequest{AuthToken: token, From: APIName, RequestID: "11111111"}
	reply := &AuthResponse{}
	client, err := rpc.DialHTTP("tcp", os.Getenv("USER_SERVICE_URL"))

	if err != nil {
		return reply.User, errors.New(err.Error())
	}

	err = client.Call("Service.Login", args, &reply)

	if err != nil {
		return reply.User, errors.New(err.Error())
	}

	return reply.User, nil
}

func likePost(user User, postID int) (LikeResponse, error) {
	logRequest(LikeServiceName)
	defer logResponse(LikeServiceName, time.Now())

	args := &LikeRequest{UserID: user.ID, PostID: postID}
	reply := &LikeResponse{}
	client, err := rpc.DialHTTP("tcp", os.Getenv("LIKE_SERVICE_URL"))

	if err != nil {
		return LikeResponse{}, errors.New(err.Error())
	}

	err = client.Call("Service.Like", args, &reply)

	if err != nil {
		return LikeResponse{}, errors.New(err.Error())
	}

	return *reply, nil
}

func logRequest(to string) {
	log.Printf("[REQ] %v → %v\n", APIName, to)
}

func logResponse(to string, start time.Time) {
	elapsed := time.Since(start)
	log.Printf("[REP] %v → %v - %v\n", APIName, to, elapsed)
}

func requestHandler(w http.ResponseWriter, r *http.Request) {
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

	like, err := likePost(user, 1234)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(user.FullName() + " w/ Likes: " + string(like.Count)))
}

func main() {
	http.HandleFunc("/", requestHandler)
	http.ListenAndServe(":"+os.Getenv("API_PORT"), nil)
}
