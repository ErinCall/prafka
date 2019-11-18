package api

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	kafka "github.com/segmentio/kafka-go"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/erincall/prafka/internal/config"
)

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
		io.WriteString(w, "GET-only endpoint")
		return
	}

	topics, ok := r.URL.Query()["topic"]
	if !ok || len(topics) != 1 {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "Must provide exactly one topic")
		return
	}
	topic := topics[0]

	offset := int64(0)
	offsets, ok := r.URL.Query()["offset"]
	if ok && len(offsets) > 0 {
		// Could throw an error on offsets > 1, but we have enough to return something, so...eh
		offset, err = strconv.ParseInt(offsets[0], 0, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, errors.Wrap(err, "offset parameter must be an integer").Error())
			return
		}
	}

	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, 5*time.Second)

	kr := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   config.BrokerList,
		Topic:     topic,
		Partition: 0,
		MinBytes:  10,
		MaxBytes:  10e6, // 10 MB
	})
	kr.SetOffset(offset)

	mChan := make(chan kafka.Message)
	eChan := make(chan error)

	go readMessages(ctx, kr, mChan, eChan)

	messages := make([]kafka.Message, 0)
ReadLoop:
	for {
		select {
		case m := <-mChan:
			messages = append(messages, m)
			/*
				TODO: if len(messages) > req.Query()["maxMessages"] {break}
			*/
		case err = <-eChan:
			if errors.Cause(err) != context.DeadlineExceeded {
				w.WriteHeader(http.StatusInternalServerError)
				io.WriteString(w, errors.Wrap(err, "error reading from kafka").Error())
				return
			}
			break ReadLoop
		case <-ctx.Done():
			break ReadLoop
		}
	}

	responses := make([]response, 0)
	for _, msg := range messages {
		responses = append(responses, formatMessage(msg))
	}

	resBody, err := json.Marshal(responses)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, errors.Wrap(err, "could not json-serialize response").Error())
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
		} else {
			mChan <- m
		}
	}
}
