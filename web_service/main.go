package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"./../middlewares"

	"github.com/justinas/alice"
	"github.com/nu7hatch/gouuid"
)

const serviceID = "service1"

func timeoutHandler(h http.Handler) http.Handler {
	return http.TimeoutHandler(h, 1*time.Second, "timed out")
}

func appHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello world!"))
}

func main() {
	rID, err := uuid.NewV4()

	if err != nil {
		log.Fatal(err)
	}

	authHandler := middlewares.TokenAuth(rID.String(), serviceID)
	app := http.HandlerFunc(appHandler)
	chain := alice.New(authHandler, timeoutHandler).Then(app)
	err = http.ListenAndServe(":"+os.Getenv("WEB_SERVICE_PORT"), chain)

	if err != nil {
		fmt.Printf("http.ListenAndServe error: %v\n", err)
	}
}
