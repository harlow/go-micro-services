package shared

import (
	"encoding/base64"
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
func (a tokenAuth) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Check that the token is valid
	if a.authenticate(r) == false {
		w.Header().Set("WWW-Authenticate", `Basic realm="Application"`)
		http.Error(w, http.StatusText(401), 401)
		return
	}

	// Call the next handler on success.
	a.http.ServeHTTP(w, r)
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

	// Return result from validator
	return b.valdator.Valid(token)
}
