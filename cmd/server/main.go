package main

import (
	"io"
	"log"
	"net/http"
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
	log.Fatal(http.ListenAndServe("localhost:8000", nil))
}
