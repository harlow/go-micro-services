package main

import (
  "net/http"

  "github.com/harlow/go-micro-services/proto/auth"

  "github.com/harlow/authtoken"
  "golang.org/x/net/context"
  "google.golang.org/grpc/metadata"
)

func NewAuthHandler(authAddr *string) func(http.Handler) http.Handler {
  authClient := auth.NewAuthClient(mustDial(authAddr))

  return func(handler http.Handler) http.Handler {
    return authMiddleware{handler, authClient}
  }
}

type authMiddleware struct {
  next http.Handler
  auth.AuthClient
}

func (m authMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  // get token from Authorization header
  authToken, err := authtoken.FromRequest(r)
  if err != nil {
    http.Error(w, err.Error(), http.StatusForbidden)
    return
  }

  // context and metadata
  md := metadata.Pairs("traceID", "ABC", "fromName", "api")
  ctx := context.Background()
  ctx = metadata.NewContext(ctx, md)

  // verify token w/ auth service
  _, err = m.VerifyToken(ctx, &auth.AuthRequest{
    AuthToken: authToken,
  })

  if err != nil {
    http.Error(w, "Unauthorized", http.StatusForbidden)
    return
  }

  // Call the next handler on success.
  m.next.ServeHTTP(w, r)
}
