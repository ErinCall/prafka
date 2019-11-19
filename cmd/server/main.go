package main

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/erincall/prafka/internal/api"
)

func statusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	io.WriteString(w, `
		{
			"status": "Everything is awesome! Everything is cool over HTTP!"
		}
	`)
}

func main() {
	http.HandleFunc("/status", statusHandler)
	http.HandleFunc("/api/send", api.SendHandler)
	http.HandleFunc("/api/read", api.ReadHandler)
	http.HandleFunc("/api/websocket", api.WebsocketHandler)
	host := "localhost:8000"
	fmt.Printf("Now listening on %s\n", host)
	log.Fatal(http.ListenAndServe(host, nil))
}
