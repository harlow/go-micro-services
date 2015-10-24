package main

import (
	"encoding/base64"
	"errors"
	"log"
	"net/http"
	"strings"

	auth "github.com/harlow/go-micro-services/service.auth/lib"

	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

const (
	BASIC_SCHEMA  string = "Basic "
	BEARER_SCHEMA string = "Bearer "
)

func NewAuthMiddleware(addr string) func(http.Handler) http.Handler {
	authClient, err := auth.NewClient(addr)
	if err != nil {
		log.Fatal("AuthClient error:", err)
	}
	defer authClient.Close()

	fn := func(h http.Handler) http.Handler {
		return authMiddleware{h, authClient}
	}

	return fn
}

type authMiddleware struct {
	next       http.Handler
	authClient *auth.Client
}

func (b authMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	authToken, err := b.parseToken(r.Header.Get("Authorization"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	// context and metadata
	md := metadata.Pairs("traceID", "ABC", "from", "api.v1")
	ctx := context.Background()
	ctx = metadata.NewContext(ctx, md)

	// verify token w/ auth service
	err = b.authClient.VerifyToken(ctx, authToken)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Call the next handler on success.
	b.next.ServeHTTP(w, r)
}

// Parse takes a HTTP Authoirzation header and parses out
// a Basic or Bearer auth token
func (b authMiddleware) parseToken(auth_header string) (string, error) {
	var token string

	// Confirm the request is sending Basic Authentication credentials.
	if !strings.HasPrefix(auth_header, BASIC_SCHEMA) && !strings.HasPrefix(auth_header, BEARER_SCHEMA) {
		return "", errors.New("Auth type not supported")
	}

	// Get the token from the request header
	// The first six characters are skipped - e.g. "Basic ".
	if strings.HasPrefix(auth_header, BASIC_SCHEMA) {
		str, err := base64.StdEncoding.DecodeString(auth_header[len(BASIC_SCHEMA):])
		if err != nil {
			return "", errors.New("Base64 encoding issue")
		}
		creds := strings.Split(string(str), ":")
		token = creds[0]
	} else {
		token = auth_header[len(BEARER_SCHEMA):]
	}

	return token, nil
}
