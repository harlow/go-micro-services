package main

import (
	"database/sql"
	"errors"
	"net"
	"net/http"
	"net/rpc"
	"os"

	log "github.com/golang/glog"
	_ "github.com/lib/pq"
)

type AuthRequest struct {
	Token     string
	RequestID string
}

type User struct {
	Email     string
	FirstName string
	ID        int32
	LastName  string
}

type UserService int

func (u *UserService) Login(args *AuthRequest, reply *User) error {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))

	if err != nil {
		log.Infof("Error: %v", err)
		return errors.New(err.Error())
	}

	defer db.Close()
	stmt := "SELECT id, first_name, last_name, email FROM users WHERE auth_token=$1"
	err = db.QueryRow(stmt, args.Token).Scan(
		&reply.Email,
		&reply.FirstName,
		&reply.ID,
		&reply.LastName,
	)

	switch {
	case err == sql.ErrNoRows:
		log.Infof("User not found")
		return errors.New("Unknown User")
	case err != nil:
		log.Infof("Error: %v", err)
		return errors.New(err.Error())
	default:
		log.Infof("Success!")
	}

	return nil
}

func main() {
	userService := new(UserService)
	rpc.Register(userService)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":"+os.Getenv("PORT"))

	if e != nil {
		log.Fatalf("net.Listen tcp :%v: %v", os.Getenv("PORT"), e)
	}

	http.Serve(l, nil)
}
