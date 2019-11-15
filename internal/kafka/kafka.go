package kafka

import (
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

var (
	Consumer *kafka.Consumer
)

// OK SO

// hnnnngh it's unclear how to interact with `kafka.Consumer`s
// make a cmd/futzing/ or something; see how that treats you

func Initialize() error {
	var err error
	confMap := kafka.ConfigMap{
		"bootstrap.servers": "localhost:32768",
		"group.id": "aoeuidhtns",
	}
	Consumer, err = kafka.NewConsumer(&confMap)
	return err
}
