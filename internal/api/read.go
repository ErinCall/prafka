package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/schema"
	"github.com/pkg/errors"
	kafka "github.com/segmentio/kafka-go"
	"io"
	"net/http"
	"time"

	"github.com/erincall/prafka/internal/config"
)

type readParams struct {
	Topic       string `schema:"topic,required"`
	Offset      int64  `schema:"offset"`
	MaxMessages int    `schema:"max"`
}

type response struct {
	Topic     string `json:"topic"`
	Partition int    `json:"partition"`
	Offset    int64  `json:"offset"`
	Value     string `json:"value"`
	Time      string `json:"time"`
}

func ReadHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, `{"error": "GET-only endpoint"}`)
		return
	}

	decoder := schema.NewDecoder()
	var params readParams
	err = decoder.Decode(&params, r.URL.Query())
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, fmt.Sprintf(`{"error": "%s"}`,
			errors.Wrap(err, "malformed request query")))
		return
	}

	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, 5*time.Second)

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

	responses := make([]response, 0)
ReadLoop:
	for {
		select {
		case m := <-mChan:
			responses = append(responses, formatMessage(m))
			if len(responses) >= params.MaxMessages {
				break ReadLoop
			}
		case err = <-eChan:
			if !isContextErr(err) {
				w.WriteHeader(http.StatusInternalServerError)
				io.WriteString(w, fmt.Sprintf(`{"error": "%s"}`,
					errors.Wrap(err, "error reading from kafka")))
				return
			}
			break ReadLoop
		case <-ctx.Done():
			break ReadLoop
		}
	}

	resBody, err := json.Marshal(responses)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, fmt.Sprintf(`{"error": "%s"}`,
			errors.Wrap(err, "could not json-serialize response")))
		return
	}

	w.Write(resBody)
}

func formatMessage(msg kafka.Message) response { // ,error?
	return response{
		Topic:     msg.Topic,
		Partition: msg.Partition,
		Offset:    msg.Offset,
		Value:     string(msg.Value),
		Time:      msg.Time.Format(time.RFC3339),
	}
}

func readMessages(ctx context.Context, kr *kafka.Reader, mChan chan kafka.Message, eChan chan error) {
	for {
		m, err := kr.ReadMessage(ctx)
		if err != nil {
			eChan <- err
			if isContextErr(err) {
				return
			}
		} else {
			mChan <- m
		}
	}
}

func isContextErr(err error) bool {
	return errors.Cause(err) == context.Canceled || errors.Cause(err) == context.DeadlineExceeded
}
