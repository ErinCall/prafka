package api

import (
	"net/http"
	"fmt"
	"io"
	"github.com/erincall/prafka/internal/kafka"
)

func APIHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	io.WriteString(w, fmt.Sprintf(`
		{
			"consumer initialized": %t
		}
	`, kafka.Consumer != nil))
}
