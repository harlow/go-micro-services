package handler

import (
	"database/sql"
	"os"

	"code.google.com/p/goprotobuf/proto"
	"github.com/asim/go-micro/server"
	"golang.org/x/net/context"
	_ "github.com/lib/pq"
	log "github.com/golang/glog"
	user "./../proto/user"
)

type Authentication struct{}

func (e *Authentication) Call(ctx context.Context, req *user.AuthRequest, rsp *user.AuthResponse) error {
	log.Infof("Received Authentication.Call request")

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
		log.Errorf("%s status=error request_id=% caller_id=%s %v\n", server.Id, requestID, callerID, err)
	}

	defer db.Close()
	err = db.QueryRow(selectStmt, *req.AuthToken).Scan(&id, &firstName, &lastName, &email)

	switch {
	case err == sql.ErrNoRows:
		log.Infof("%s status=failed request_id=%s caller_id=%s\n", server.Id, requestID, callerID)
		rsp.Valid = proto.Bool(false)
	case err != nil:
		log.Infof("%s status=error request_id=%s caller_id=%s error=%v\n", server.Id, requestID, callerID, err)
		rsp.Valid = proto.Bool(false)
	default:
		log.Infof("%s status=success request_id=%s caller_id=%s user_id=%d\n", server.Id, requestID, callerID, id)
		rsp.Valid = proto.Bool(true)
		rsp.User = &user.User{
			AuthToken: proto.String(authToken),
			Email:     proto.String(email),
			FirstName: proto.String(firstName),
			Id:        proto.Int32(id),
			LastName:  proto.String(lastName),
		}
	}

	return nil
}
