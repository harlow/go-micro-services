package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/harlow/auth_token"
	"github.com/harlow/go-micro-services/proto/like"
	"github.com/harlow/go-micro-services/proto/user"
	"github.com/harlow/go-micro-services/trace"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	serverName     = "api.v1"
	likeServerAddr = flag.String("like_server_addr", "127.0.0.1:10001", "The Like server address in the format of host:port")
	userServerAddr = flag.String("user_server_addr", "127.0.0.1:10002", "The User server address in the format of host:port")
)

type Response struct {
	Status string `json:"status"`
	Count  int32  `json:"count"`
}

func getUser(t trace.Tracer, token string) (user.User, error) {
	conn, err := grpc.Dial(*userServerAddr)
	if err != nil {
		return user.User{}, err
	}

	t.Request("service.user")
	defer t.Reply("service.user", time.Now())
	defer conn.Close()

	ctx := context.Background()
	context.WithValue(ctx, "traceID", "SOME STRING")
	client := user.NewUserLookupClient(conn)
	u, err := client.GetUser(ctx, &user.Args{token})

	if err != nil {
		return user.User{}, err
	}

	return *u, nil
}

func likePost(t trace.Tracer, userID int32, postID int32) (like.Like, error) {
	conn, err := grpc.Dial(*likeServerAddr)
	if err != nil {
		return like.Like{}, err
	}

	t.Request("service.like")
	defer t.Reply("service.like", time.Now())
	defer conn.Close()

	ctx := context.Background()
	context.WithValue(ctx, "traceID", "SOME STRING")

	client := like.NewLikeServiceClient(conn)
	l, err := client.RecordLike(ctx, &like.Args{userID, postID})

	if err != nil {
		return like.Like{}, err
	}

	return *l, nil
}

func requestHandler(w http.ResponseWriter, r *http.Request) {
	token, err := auth_token.Parse(r.Header.Get("Authorization"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	t := trace.NewTracer(serverName)
	u, err := getUser(t, token)

	if err != nil {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	like, err := likePost(t, u.ID, 1234)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s := &Response{Status: "success", Count: like.Count}
	b, err := json.Marshal(s)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(b)
}

func main() {
	http.HandleFunc("/", requestHandler)
	log.Fatal(http.ListenAndServe(":"+os.Getenv("API_V1_PORT"), nil))
}
