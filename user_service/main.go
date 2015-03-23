package main

import (
	"database/sql"
	"errors"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"

	_ "github.com/lib/pq"
)

type AuthRequest struct {
	Token     string
	RequestID string
}

type User struct {
	Email     string
	FirstName string
	ID        int
	LastName  string
}

type UserService int

func (u *UserService) Login(args *AuthRequest, reply *User) error {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))

	if err != nil {
		log.Printf("Error: %v\n", err)
		return errors.New(err.Error())
	}

	defer db.Close()
	stmt := "SELECT id, first_name, last_name, email FROM users WHERE auth_token=$1"
	err = db.QueryRow(stmt, args.Token).Scan(&reply.ID, &reply.FirstName, &reply.LastName, &reply.Email)

	switch {
	case err == sql.ErrNoRows:
		log.Println("User not found")
		return errors.New("Unknown User")
	case err != nil:
		log.Printf("Error: %v\n", err)
		return errors.New(err.Error())
	default:
		log.Printf("Successful login for %v\n", reply.ID)
	}

	return nil
}

func main() {
	srvc := new(UserService)
	rpc.Register(srvc)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":"+os.Getenv("PORT"))

	if e != nil {
		log.Fatalf("net.Listen tcp :%v: %v", os.Getenv("PORT"), e)
	}

	http.Serve(l, nil)
}
