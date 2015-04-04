package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"../proto/like"
	"../proto/user"
	"../trace"

	"github.com/harlow/auth_token"
	"github.com/nu7hatch/gouuid"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	name = "api.v1"
	likeServerAddr = flag.String("like_server_addr", "127.0.0.1:10001", "The Like server address in the format of host:port")
	userServerAddr = flag.String("user_server_addr", "127.0.0.1:10002", "The Like server address in the format of host:port")
)

type Response struct {
    Status string `json:"status"`
    Count int32 `json:"count"`
}

func getUser(traceID string, token string) (user.User, error) {
	trace.Request(traceID, name, "service.user")
	defer trace.Reply(traceID, "service.user", name, time.Now())

	conn, err := grpc.Dial(*userServerAddr)
	if err != nil {
		return user.User{}, err
	}

	defer conn.Close()
	client := user.NewUserServiceClient(conn)
	reply, err := client.GetUser(
		context.Background(),
		&user.UserRequest{Token: token, From: name},
	)

	if err != nil {
		return user.User{}, err
	}

	return *reply.User, nil
}

func likePost(traceID string, userID int32, postID int32) (like.Like, error) {
	trace.Request(traceID, name, "service.like")
	defer trace.Reply(traceID, "service.like", name, time.Now())

	conn, err := grpc.Dial(*likeServerAddr)
	if err != nil {
		return like.Like{}, err
	}

	defer conn.Close()
	client := like.NewLikeServiceClient(conn)
	reply, err := client.RecordLike(
		context.Background(),
		&like.LikeRequest{UserID: userID, PostID: postID, From: name},
	)

	if err != nil {
		return like.Like{}, err
	}

	return *reply.Like, nil
}

func requestHandler(w http.ResponseWriter, r *http.Request) {
	traceID, err := uuid.NewV4()

  if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
			return
  }

	trace.Request(traceID.String(), "www", name)
	defer trace.Reply(traceID.String(), name, "www", time.Now())
	token, err := auth_token.Parse(r.Header.Get("Authorization"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	user, err := getUser(traceID.String(), token)

	if err != nil {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	like, err := likePost(traceID.String(), user.ID, 1234)

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

func main() {
	http.HandleFunc("/", requestHandler)
	log.Fatal(http.ListenAndServe(":"+os.Getenv("API_PORT"), nil))
}
