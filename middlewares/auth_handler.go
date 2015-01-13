package middlewares

import (
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"strings"

	"./../protobufs/user"

	"code.google.com/p/goprotobuf/proto"
)

// TokenAuth provides HTTP middleware for protecting URIs with HTTP Token Authentication
// the server authenticates token provided in the "Authorization" HTTP header.
func TokenAuth(requestID string, callerID string) func(http.Handler) http.Handler {
	fn := func(h http.Handler) http.Handler {
		return tokenAuth{h, callerID, requestID}
	}
	return fn
}

type tokenAuth struct {
	h         http.Handler
	callerID  string
	requestID string
}

// Satisfies the http.Handler interface for tokenAuth.
func (b tokenAuth) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Check that the token is valid
	if b.authenticate(r) == false {
		w.Header().Set("WWW-Authenticate", `Basic realm="Application"`)
		http.Error(w, http.StatusText(401), 401)
		return
	}

	// Call the next handler on success.
	b.h.ServeHTTP(w, r)
}

// authenticate retrieves and then validates the token provided in the request header.
// Returns 'false' if the user has not successfully authenticated.
func (b *tokenAuth) authenticate(r *http.Request) bool {
	const basicScheme string = "Basic "
	const bearerScheme string = "Bearer "
	var token string

	// Confirm the request is sending Basic Authentication credentials.
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, basicScheme) && !strings.HasPrefix(auth, bearerScheme) {
		return false
	}

	// Get the token from the request header
	// The first six characters are skipped - e.g. "Basic ".
	if strings.HasPrefix(auth, basicScheme) {
		str, err := base64.StdEncoding.DecodeString(auth[len(basicScheme):])
		if err != nil {
			return false
		}
		creds := strings.Split(string(str), ":")
		token = creds[0]
	} else {
		token = auth[len(bearerScheme):]
	}

	// Set up protobuf for the auth request
	req := user.AuthRequest{
		AuthToken: proto.String(token),
		CallerID:  proto.String(b.callerID),
		RequestID: proto.String(b.requestID),
	}
	resp := user.AuthResponse{}
	stub, client, err := user.DialUserService("tcp", ":"+os.Getenv("AUTH_SERVICE_PORT"))

	if err != nil {
		log.Fatalf("%s user.DialUserService error:", b.requestID, err)
	}

	defer client.Close()
	log.Printf("%s rpc:auth_service status:begin\n", b.requestID)

	if err = stub.Auth(&req, &resp); err != nil {
		log.Printf("%s rpc:auth_service:error %v\n", b.requestID, err)
	}

	log.Printf("%s rpc:auth_service status:complete success:%v\n", b.requestID, resp.GetSuccess())

	// Return the response from auth service
	return resp.GetSuccess()
}
