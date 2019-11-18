package main

import (
	"context"
	"fmt"
	kafka "github.com/segmentio/kafka-go"
	"os"
)

func main() {
	topic := "prafka"
	partition := 0
	brokerList := []string{"localhost:32769"}
	ctx := context.Background()

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   brokerList,
		Topic:     topic,
		Partition: partition,
		MinBytes:  10,
		MaxBytes:  10e6,
	})
	r.SetOffset(0)

	for {
		m, err := r.ReadMessage(ctx)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Printf("message at offset %d: %s\n", m.Offset, m.Value)
	}

	fmt.Println("Good so far...")
}
