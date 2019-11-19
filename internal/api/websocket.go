package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/schema"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	kafka "github.com/segmentio/kafka-go"
	"io"
	"net/http"
	"net/url"
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

type errRes struct {
	Error string `json:"error"`
}

var (
	decoder  = schema.NewDecoder()
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

func WebsocketHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	if r.Method != "GET" {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, `{"error": "GET-only endpoint"}`)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(errors.Wrap(err, "could not upgrade to websocket connection").Error())
		return
	}

	go serveWebsocket(r.URL.Query(), conn)
}

func serveWebsocket(q url.Values, conn *websocket.Conn) {
	ctx, cancel := context.WithCancel(context.Background())
	// Ticker will be used to send keepalive pings. 10 seconds seems like a reasonable interval?
	ticker := time.NewTicker(10 * time.Second)
	defer cancel()
	defer conn.Close()
	defer ticker.Stop()

	decoder := schema.NewDecoder()
	var params wsParams

	if err := decoder.Decode(&params, q); err != nil {
		errMsg, _ := json.Marshal(errRes{errors.Wrap(err, "malformed query").Error()})
		conn.WriteMessage(websocket.TextMessage, errMsg)
		return
	}

	kr := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   config.BrokerList,
		Topic:     params.Topic,
		Partition: 0,
		MinBytes:  10,
		MaxBytes:  10e6, // 10 MB
	})
	kr.SetOffset(params.Offset)

	mChan := make(chan kafka.Message)
	eChan := make(chan error)

	go readMessages(ctx, kr, mChan, eChan)
	go drainIncoming(ctx, conn)

	for {
		select {
		case m := <-mChan:
			msg, err := json.Marshal(formatMessage(m))
			if err != nil {
				msg = []byte(`{"error": "could not json-encode Kafka message"}`)
			}
			if err = conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				fmt.Println(errors.Wrap(err, "could not send websocket message").Error())
				return
			}
		case err := <-eChan:
			errMsg := errors.Wrap(err, "error reading from Kafka").Error()
			fmt.Println(errMsg)
			res, _ := json.Marshal(errRes{errMsg})
			conn.WriteMessage(websocket.TextMessage, res)
			return
		case <-ticker.C:
			if err := conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				fmt.Println(errors.Wrap(err, "couldn't send websocket message").Error())
				return
			}
		}
	}
}

/*
We don't care about incoming messages, but if they do arrive they can clog the incoming buffer. Just
drain it until there's an error or the context gets canceled.
*/

func drainIncoming(ctx context.Context, conn *websocket.Conn) {
	hadError := make(chan bool)
	readMessage := func() {
		_, _, err := conn.ReadMessage()
		hadError <- err != nil
	}

	go readMessage()

	for {
		select {
		case errored := <-hadError:
			if errored {
				return
			}
			go readMessage()
		case <-ctx.Done():
			return
		}
	}
}
