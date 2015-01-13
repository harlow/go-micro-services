package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"./../shared"

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
	requestID, err := uuid.NewV4()
	if err != nil {
		log.Fatal(err)
	}

	validator := shared.TokenValidator{serviceID, requestID.String()}
	authHandler := shared.TokenAuth(validator)
	app := http.HandlerFunc(appHandler)
	chain := alice.New(authHandler, timeoutHandler).Then(app)

	err = http.ListenAndServe(":"+os.Getenv("WEB_SERVICE_PORT"), chain)
	if err != nil {
		fmt.Printf("http.ListenAndServe error: %v\n", err)
	}
}
