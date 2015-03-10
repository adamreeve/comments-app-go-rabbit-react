package main

import (
	"fmt"
	"log"
	"time"
)

const (
	uri          = "amqp://guest:guest@localhost:5672/"
	exchangeName = "test-exchange-fanout"
	exchangeType = "fanout"
	bindingKey   = "test-key"

	reliable = true
)

func main() {
	go receive("consumer1", "consumer1")

	go receive("consumer2", "consumer2")

	go publish()

	log.Printf("Running forever")
	select {}
}

func receive(queueName, consumerTag string) {
	_, err := newConsumer(
		uri,
		exchangeName,
		exchangeType,
		queueName,
		bindingKey,
		consumerTag,
	)
	if err != nil {
		panic(err)
	}
}

func publish() {
	publisher, err := newPublisher(
		uri,
		exchangeName,
		exchangeType,
		reliable,
	)
	if err != nil {
		panic(err)
	}

	for i := 0; i < 10; i++ {
		message := fmt.Sprintf("Hello with message %d", i)
		publisher.Publish(bindingKey, []byte(message))
		time.Sleep(time.Second)
	}
}
