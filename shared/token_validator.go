package shared

import (
	"log"
	"os"

	"./../protobufs/user"

	"code.google.com/p/goprotobuf/proto"
	"github.com/nu7hatch/gouuid"
)

type TokenValidator struct {
	CallerID string
}

func (v TokenValidator) Valid(token string) bool {
	// Create Request ID
	requestID, err := uuid.NewV4()
	if err != nil {
		log.Fatal(err)
	}

	req := user.AuthRequest{
		AuthToken: proto.String(token),
		CallerID:  proto.String(v.CallerID),
		RequestID: proto.String(requestID.String()),
	}
	resp := user.AuthResponse{}
	stub, client, err := user.DialUserService("tcp", ":"+os.Getenv("AUTH_SERVICE_PORT"))

	if err != nil {
		log.Fatalf("%s user.DialUserService error:", requestID.String(), err)
	}

	defer client.Close()
	log.Printf("%s proto.user.auth.begin\n", requestID.String())

	if err = stub.Auth(&req, &resp); err != nil {
		log.Printf("%s proto.user.auth.error %v\n", requestID.String(), err)
	}

	log.Printf("%s proto.user.auth.complete valid=%v\n", requestID.String(), resp.GetValid())

	// Return the response from auth service
	return resp.GetValid()
}
