package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

// newServer returns a trace server with an
// initialzed slice of events
func newServer() *traceServer {
	s := &traceServer{}
	s.events = make(map[string][]string)
	return s
}

type traceServer struct {
	events map[string][]string
}

func (s traceServer) requestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()

		traceID := r.PostFormValue("traceId")
		if traceID == "" {
			http.Error(w, "Param traceId required", http.StatusBadRequest)
			return
		}

		msg := r.PostFormValue("msg")
		if msg == "" {
			http.Error(w, "Param msg required", http.StatusBadRequest)
			return
		}

		s.events[traceID] = append(s.events[traceID], msg)
	} else {
		traceID := r.URL.Query().Get("traceId")
		if traceID == "" {
			http.Error(w, "Param traceID required", http.StatusBadRequest)
			return
		}

		fmt.Fprintln(w, "<div class=wsd wsd_style=\"default\" ><pre>")
		for _, msg := range s.events[traceID] {
     	fmt.Fprintln(w, msg)
  	}
  	fmt.Fprintln(w, "</pre></div>")
  	fmt.Fprintln(w, "<script type=\"text/javascript\" src=\"http://www.websequencediagrams.com/service.js\"></script>")
  }
}

func main() {
	var port = flag.String("port", "5001", "The server port")
	flag.Parse()

	s := newServer()
	http.HandleFunc("/", s.requestHandler)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
