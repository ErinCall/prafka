package main

import (
	"fmt"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	// "io"
	"math/rand"
	"os"
)

// This should probably be...something. It's unclear what I should do with the information.
// Currently I'm just using a single broker and consumer, so rebalancing shouldn't be an issue
// ...right?
func assignedPartition(c *kafka.Consumer, e kafka.Event) error {
	switch event := e.(type) {
	case kafka.AssignedPartitions:
		fmt.Printf("received AssignedPartitions event. New partitions: %v\n", event.Partitions)
		current, err := c.Assignment()
		if err != nil {
			return err
		}
		fmt.Printf("current assigned partitions (type %T): %v\n", current)
		c.Assign(event.Partitions)
	}
	return nil
}

func main() {
	confMap := kafka.ConfigMap{
		"bootstrap.servers": "localhost:32768",
		"group.id": fmt.Sprintf("%f", rand.Float64),
		// "go.events.channel.enable": true,
		"go.application.rebalance.enable": false,
	}
	consumer, err := kafka.NewConsumer(&confMap)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	consumer.Subscribe("prafka", assignedPartition)

	for {
		e := consumer.Poll(10)
		if e != nil {
			fmt.Printf("received an event from poll: %v\n", e)
		}
	}

	// events := consumer.Events()

	// for {
	// 	select {
	// 	case e := <- events:
	// 		fmt.Printf("received an event on the events channel: %s", e)
	// 	}
	// }

	fmt.Println("yeah ok")
}
