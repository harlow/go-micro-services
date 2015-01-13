package shared

import (
	"log"
	"os"

	"./../protobufs/user"

	"code.google.com/p/goprotobuf/proto"
)

type TokenValidator struct {
	CallerID  string
	RequestID string
}

func (v TokenValidator) Valid(token string) bool {
	req := user.AuthRequest{
		AuthToken: proto.String(token),
		CallerID:  proto.String(v.CallerID),
		RequestID: proto.String(v.RequestID),
	}
	resp := user.AuthResponse{}
	stub, client, err := user.DialUserService("tcp", ":"+os.Getenv("AUTH_SERVICE_PORT"))

	if err != nil {
		log.Fatalf("%s user.DialUserService error:", v.RequestID, err)
	}

	defer client.Close()
	log.Printf("%s rpc:auth_service status:begin\n", v.RequestID)

	if err = stub.Auth(&req, &resp); err != nil {
		log.Printf("%s rpc:auth_service:error %v\n", v.RequestID, err)
	}

	log.Printf("%s rpc:auth_service status:complete success:%v\n", v.RequestID, resp.GetValid())

	// Return the response from auth service
	return resp.GetValid()
}
