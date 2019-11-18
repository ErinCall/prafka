package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	kafka "github.com/segmentio/kafka-go"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/erincall/prafka/internal/config"
)

type sendBody struct {
	Topic   string `json:"topic"`
	Message string `json:"message"`
}

func SendHandler(w http.ResponseWriter, r *http.Request) {
	_ = fmt.Println // May want fmt around for stdout debugging, so let it stay in the import list

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, "POST-only endpoint")
		return
	}

	rawBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, err.Error())
		return
	}

	var body sendBody
	err = json.Unmarshal(rawBody, &body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, errors.Wrap(err, "malformed JSON in body").Error())
		return
	}

	if body.Topic == "" || body.Message == "" {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "topic and message are mandatory")
		return
	}

	ctx := context.Background()
	// TODO: ctx = context.WithTimeout(ctx, ...)

	kw := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  config.BrokerList,
		Topic:    body.Topic,
		Balancer: &kafka.LeastBytes{},
	})

	kw.WriteMessages(ctx, kafka.Message{
		Key:   []byte{},
		Value: []byte(body.Message),
	})

	io.WriteString(w, `{"status": "success"}`)
}
