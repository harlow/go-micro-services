package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/harlow/go-micro-services/proto/user"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	_ "github.com/lib/pq"
)

var (
	port       = flag.Int("port", 10002, "The server port")
	serverName = "service.user"
)

type server int

// GetUser finds a User by authentication token.
func (s *server) GetUser(ctx context.Context, args *user.Args) (*user.User, error) {
	db, err := sql.Open("postgres", os.Getenv("USER_SERVICE_DATABASE_URL"))

	if err != nil {
		return &user.User{}, errors.New(err.Error())
	}

	defer db.Close()
	u := &user.User{}
	stmt := "SELECT id, first_name, last_name, email FROM users WHERE auth_token=$1"
	err = db.QueryRow(stmt, args.Token).Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email)

	switch {
	case err == sql.ErrNoRows:
		return &user.User{}, errors.New("Invalid Token")
	case err != nil:
		return &user.User{}, errors.New(err.Error())
	}

	return u, nil
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := new(server)
	grpcServer := grpc.NewServer()
	user.RegisterUserLookupServer(grpcServer, s)
	grpcServer.Serve(lis)
}
