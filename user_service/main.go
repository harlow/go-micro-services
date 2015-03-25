package main

import (
	"database/sql"
	"errors"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"time"

	_ "github.com/lib/pq"
)

const ServiceName = "com.go-micro.services.user"

type AuthRequest struct {
	AuthToken string
	From      string
	RequestID string
}

type User struct {
	Email     string
	FirstName string
	ID        int
	LastName  string
}

type AuthResponse struct {
	From      string
	RequestID string
	User      User
}

type UserService int

func logRequest(from string) {
	log.Printf("[IN] %v → %v\n", ServiceName, from)
}

func logResponse(from string, start time.Time) {
	elapsed := time.Since(start)
	log.Printf("[OUT] %v → %v - %v\n", ServiceName, from, elapsed)
}

func (u *UserService) Login(args *AuthRequest, reply *AuthResponse) error {
	logRequest(args.From)
	defer logResponse(args.From, time.Now())
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))

	if err != nil {
		return errors.New(err.Error())
	}

	defer db.Close()
	stmt := "SELECT id, first_name, last_name, email FROM users WHERE auth_token=$1"
	err = db.QueryRow(stmt, args.AuthToken).Scan(
		&reply.User.ID,
		&reply.User.FirstName,
		&reply.User.LastName,
		&reply.User.Email,
	)

	switch {
	case err == sql.ErrNoRows:
		return errors.New("Unknown User")
	case err != nil:
		return errors.New(err.Error())
	}

	return nil
}

func main() {
	srv := new(UserService)
	rpc.Register(srv)
	rpc.HandleHTTP()
	ln, err := net.Listen("tcp", ":"+os.Getenv("PORT"))

	if err != nil {
		log.Fatalf("net.Listen tcp :%v: %v", os.Getenv("PORT"), err)
	}

	http.Serve(ln, nil)
}
