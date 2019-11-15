package main

import (
	"io"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/erincall/prafka/internal/kafka"
	"github.com/erincall/prafka/internal/web/api"
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
	err := kafka.Initialize()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	http.HandleFunc("/status", statusHandler)
	http.HandleFunc("/api", api.APIHandler)
	log.Fatal(http.ListenAndServe("localhost:8000", nil))
}
