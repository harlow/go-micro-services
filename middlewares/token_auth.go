package shared

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
)

type Validator interface {
	Valid(token string) bool
}

// TokenAuth provides HTTP middleware for protecting URIs with HTTP Token Authentication
// the server authenticates token provided in the "Authorization" HTTP header.
func TokenAuth(v Validator) func(http.Handler) http.Handler {
	fn := func(h http.Handler) http.Handler {
		return tokenAuth{h, v}
	}
	return fn
}

type tokenAuth struct {
	http     http.Handler
	valdator Validator
}

// Satisfies the http.Handler interface for tokenAuth.
// Write and unauthorized message and returns if user has not successfully authenticated.
func (a tokenAuth) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	token, err := a.getToken(r.Header.Get("Authorization"))

	// Check that the token is valid
	if err != nil || a.valdator.Valid(token) == false {
		w.Header().Set("WWW-Authenticate", `Basic realm="Application"`)
		http.Error(w, http.StatusText(401), 401)
		return
	}

	// Call the next handler on success.
	a.http.ServeHTTP(w, r)
}

// getToken retrieves the Authorization token.
func (a tokenAuth) getToken(auth string) (string, error) {
	const basicScheme string = "Basic "
	const bearerScheme string = "Bearer "
	var token string

	// Confirm the request is sending Basic Authentication credentials.
	if !strings.HasPrefix(auth, basicScheme) && !strings.HasPrefix(auth, bearerScheme) {
		return "", errors.New("auth: Type not supported")
	}

	// Get the token from the request header
	// The first six characters are skipped - e.g. "Basic ".
	if strings.HasPrefix(auth, basicScheme) {
		str, err := base64.StdEncoding.DecodeString(auth[len(basicScheme):])
		if err != nil {
			return "", errors.New("auth: Base64 encoding issue")
		}
		creds := strings.Split(string(str), ":")
		token = creds[0]
	} else {
		token = auth[len(bearerScheme):]
	}

	// Return token
	return token, nil
}
