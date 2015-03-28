package main

import (
	"database/sql"
	"errors"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"

	"../shared/user"

	_ "github.com/lib/pq"
)

type Service int

func (u *Service) Login(args *user.Args, reply *user.Reply) error {
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
		return errors.New("Invalid Token")
	case err != nil:
		return errors.New(err.Error())
	}

	return nil
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
