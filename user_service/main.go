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

const ServiceName = "service.user"

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

type Service int

func (u *Service) Login(args *AuthRequest, reply *AuthResponse) error {
	logRequest(args.From)
	defer logResponse(args.From, time.Now())
	db, err := sql.Open("postgres", os.Getenv("USER_SERVICE_DATABASE_URL"))

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
		return errors.New("Token not found")
	case err != nil:
		return errors.New(err.Error())
	}

	return nil
}

func logRequest(from string) {
	log.Printf("[IN]  %v → %v\n", ServiceName, from)
}

func logResponse(from string, start time.Time) {
	elapsed := time.Since(start)
	log.Printf("[OUT] %v → %v - %v\n", ServiceName, from, elapsed)
}

func main() {
	srv := new(Service)
	rpc.Register(srv)
	rpc.HandleHTTP()
	ln, err := net.Listen("tcp", ":"+os.Getenv("USER_SERVICE_PORT"))

	if err != nil {
		log.Fatalf("net.Listen tcp :%v: %v", os.Getenv("USER_SERVICE_PORT"), err)
	}

	http.Serve(ln, nil)
}
