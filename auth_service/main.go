package main

import (
	"database/sql"
	"log"
	"os"

	"./../protobufs/user"

	"code.google.com/p/goprotobuf/proto"
	_ "github.com/lib/pq"
)

const serviceID = "auth_service"

type AuthHandler int

func (t *AuthHandler) Auth(req *user.AuthRequest, resp *user.AuthResponse) error {
	requestID := req.GetRequestID()
	callerID := req.GetCallerID()
	authToken := req.GetAuthToken()

	var email string
	var firstName string
	var id int32
	var lastName string

	selectStmt := "SELECT id, first_name, last_name, email FROM users WHERE auth_token=$1"
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Printf("%s %s.status.error caller_id=%s %v\n", requestID, serviceID, callerID, err)
	}

	defer db.Close()
	err = db.QueryRow(selectStmt, *req.AuthToken).Scan(&id, &firstName, &lastName, &email)

	switch {
	case err == sql.ErrNoRows:
		log.Printf("%s %s.status.failed caller_id=%s\n", requestID, serviceID, callerID)
		resp.Valid = proto.Bool(false)
	case err != nil:
		log.Printf("%s %s.status.error  %v\n", requestID, serviceID, callerID, err)
		resp.Valid = proto.Bool(false)
	default:
		log.Printf("%s %s.status.success caller_id=%s user_id=%d\n", requestID, serviceID, callerID, id)
		resp.User = &user.User{
			AuthToken: proto.String(authToken),
			Email:     proto.String(email),
			FirstName: proto.String(firstName),
			Id:        proto.Int32(id),
			LastName:  proto.String(lastName),
		}
		resp.Valid = proto.Bool(true)
	}
	return nil
}

func main() {
	user.ListenAndServeUserService(
		"tcp",
		":"+os.Getenv("AUTH_SERVICE_PORT"),
		new(AuthHandler),
	)
}
