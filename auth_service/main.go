package main

import (
	"database/sql"
	"log"
	"os"

	"./../protobufs/user"
	"code.google.com/p/goprotobuf/proto"
	_ "github.com/lib/pq"
)

type UserService int

func (t *UserService) Auth(req *user.AuthRequest, resp *user.AuthResponse) error {
	authToken := *req.Token
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()
	var id int32
	var firstName string
	var lastName string
	var email string
	err = db.
		QueryRow("SELECT id, first_name, last_name, email FROM users WHERE auth_token=$1", authToken).
		Scan(&id, &firstName, &lastName, &email)

	switch {
	case err == sql.ErrNoRows:
		log.Println("No user with that token.")
		resp.Valid = proto.Bool(false)
	default:
		resp.Valid = proto.Bool(true)
		resp.User = &user.User{
			Id:        proto.Int32(id),
			FirstName: proto.String(firstName),
			LastName:  proto.String(lastName),
			Email:     proto.String(email),
			AuthToken: proto.String(authToken),
		}
	}
	return nil
}

func main() {
	log.Println("running")
	user.ListenAndServeUserService("tcp", ":1984", new(UserService))
}
