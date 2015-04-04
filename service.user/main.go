package main

import (
	"database/sql"
	"errors"
	"log"
	"net"
	"os"
	"flag"
	"fmt"

	pb "../proto/user"

	_ "github.com/lib/pq"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 10002, "The server port")
	name = "service.user"
)

type server int

// GetUser lookups up a User by token and returns if found.
func (s *server) GetUser(ctx context.Context, req *pb.UserRequest) (*pb.UserResponse, error) {
	db, err := sql.Open("postgres", os.Getenv("USER_SERVICE_DATABASE_URL"))
	if err != nil {
		return &pb.UserResponse{}, errors.New(err.Error())
	}

	defer db.Close()
	u := &pb.User{}

	stmt := "SELECT id, first_name, last_name, email FROM users WHERE auth_token=$1"
	err = db.QueryRow(stmt, req.Token).Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email)

	switch {
	case err == sql.ErrNoRows:
		log.Println("sql.ErrNoRows")
		return &pb.UserResponse{}, errors.New("Invalid Token")
	case err != nil:
		log.Println(err.Error())
		return &pb.UserResponse{}, errors.New(err.Error())
	}

	return &pb.UserResponse{User: u, From: name}, nil
}

func main() {
  flag.Parse()
  lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
  if err != nil {
    log.Fatalf("failed to listen: %v", err)
  }
	s := new(server)
	grpcServer := grpc.NewServer()
	pb.RegisterUserServiceServer(grpcServer, s)
	grpcServer.Serve(lis)
}
