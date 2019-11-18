package api

import (
	"fmt"
	// "context"
	// "encoding/json"
	"github.com/gorilla/schema"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	kafka "github.com/segmentio/kafka-go"
	"io"
	"net/http"
	"time"

	"github.com/erincall/prafka/internal/config"
)

type wsParams struct {
	Topic  string `schema:"topic,required"`
	Offset int64  `schema:"offset"`
}

type message struct {
	Topic     string `json:"topic"`
	Partition int    `json:"partition"`
	Offset    int64  `json:"offset"`
	Value     string `json:"value"`
	Time      string `json:"time"`
}

var (
	decoder  = schema.NewDecoder()
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

func WebsocketHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Can I actually send http error responses like this with a ws:// request?
	var err error
	if r.Method != "GET" {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, `{"error": "GET-only endpoint"}`)
		return
	}

	decoder := schema.NewDecoder()
	var params wsParams
	err = decoder.Decode(&params, r.URL.Query())
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, errors.Wrap(err, "malformed request query").Error())
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(errors.Wrap(err, "could not upgrade to websocket connection").Error())
		return
	}

	go serveWebsocket(params, conn)
}

func serveWebsocket(params wsParams, conn *websocket.Conn) {
	_ = time.Second // TODO: remove me

	kr := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   config.BrokerList,
		Topic:     params.Topic,
		Partition: 0,
		MinBytes:  10,
		MaxBytes:  10e6, // 10 MB
	})
	kr.SetOffset(params.Offset)

	//
	// OK SO
	//
	// ReadMessage will block until it can find something. That means we won't be able to send pings,
	// so we probably need to set up a goroutine + channel just like in read.go
	//

	for {
		response := []byte("Behold, a manpage")
		if err := conn.WriteMessage(websocket.TextMessage, response); err != nil {
			fmt.Printf("Error from conn: '%s' (type %T)\n", err.Error(), err)
			return
		}
	}
}
