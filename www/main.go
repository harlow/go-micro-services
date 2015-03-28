package main

import (
	"errors"
	"net/http"
	"net/rpc"
	"os"
	"time"
	"encoding/json"

	"../shared/like"
	"../shared/req"
	"../shared/user"

	"github.com/harlow/auth_token"
)

const APIName = "api.like"

func main() {
	http.HandleFunc("/", requestHandler)
	http.ListenAndServe(":"+os.Getenv("API_PORT"), nil)
}

type Response struct {
    Status string `json:"status"`
    Count int32 `json:"count"`
}

func requestHandler(w http.ResponseWriter, r *http.Request) {
	req.LogReq("www", APIName)
	defer req.LogRep(APIName, "www", time.Now())
	token, err := auth_token.Parse(r.Header.Get("Authorization"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	user, err := getUser(token)

	if err != nil {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	like, err := likePost(user.ID, 1234)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s := &Response{
	  Status: "success",
	  Count: like.Count,
	}

  b, err := json.Marshal(s)

  if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

  w.Write(b)
}

func getUser(token string) (user.User, error) {
	req.LogReq(APIName, user.ServiceID)
	defer req.LogRep(user.ServiceID, APIName, time.Now())

	client, err := rpc.DialHTTP("tcp", os.Getenv("USER_SERVICE_URL"))

	if err != nil {
		return user.User{}, errors.New(err.Error())
	}

	args := user.Args{AuthToken: token, ServiceID: APIName}
	reply := &user.Reply{}
	err = client.Call("Service.Login", args, &reply)

	if err != nil {
		return user.User{}, errors.New(err.Error())
	}

	return reply.User, nil
}

func likePost(userID int, postID int) (like.Like, error) {
	req.LogReq(APIName, like.ServiceID)
	defer req.LogRep(like.ServiceID, APIName, time.Now())

	client, err := rpc.DialHTTP("tcp", os.Getenv("LIKE_SERVICE_URL"))

	if err != nil {
		return like.Like{}, errors.New(err.Error())
	}

	args := &like.Args{UserID: userID, PostID: postID, ServiceID: APIName}
	reply := &like.Reply{}
	err = client.Call("Service.Like", args, &reply)

	if err != nil {
		return like.Like{}, errors.New(err.Error())
	}

	return reply.Like, nil
}
