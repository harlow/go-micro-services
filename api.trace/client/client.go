package client

import (
	"fmt"
	"time"
	"log"
	"net/http"
	"net/url"

	"github.com/nu7hatch/gouuid"
)

func NewTraceID() string {
	t, _ := uuid.NewV4()
	return t.String()
}

func post(traceID string, msg string) {
	_, err := http.PostForm("http://localhost:5001", url.Values{
		"traceId": {traceID},
		"msg": {msg},
	})
	if err != nil {
      log.Panic("Could not connect")
  }
}

func Req(traceID string, from string, to string, action string) {
	msg := fmt.Sprintf("%v->%v: %v\n", from, to, action)
	post(traceID, msg)
}

func Rep(traceID string, from string, to string, startTime time.Time) {
	elapsedTime := time.Since(startTime)
	msg := fmt.Sprintf("%v-->%v: %v\n", from, to, elapsedTime)
	post(traceID, msg)
}
