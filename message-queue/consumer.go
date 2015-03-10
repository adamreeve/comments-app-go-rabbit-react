package main

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	tag     string
	done    chan error
}

func newConsumer(amqpUri, exchange, exchangeType, queueName, key, ctag string) (*Consumer, error) {
	var err error
	c := &Consumer{
		conn:    nil,
		channel: nil,
		tag:     ctag,
		done:    make(chan error),
	}

	log.Printf("Consumer dialing %s", amqpUri)
	c.conn, err = amqp.Dial(amqpUri)
	if err != nil {
		return nil, fmt.Errorf("Dial: %s", err)
	}

	go func() {
		fmt.Printf("Consumer closing: %s", <-c.conn.NotifyClose(make(chan *amqp.Error)))
	}()

	log.Printf("Consumer getting channel")
	c.channel, err = c.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("Channel: %s", err)
	}

	log.Printf("Consumer declaring exchange %s", exchange)
	if err = c.channel.ExchangeDeclare(
		exchange,
		exchangeType,
		true,  // durable
		false, // delete when complete
		false, // internal
		false, // no wait
		nil,   // arguments
	); err != nil {
		return nil, fmt.Errorf("Exchange declare: %s", err)
	}

	log.Printf("Consumer declaring queue %s", queueName)
	queue, err := c.channel.QueueDeclare(
		queueName,
		true,  // durable
		false, // auto delete
		false, // exclusive
		false, // no wait
		nil,   // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("Queue declare: %s", err)
	}

	log.Printf("Declared Queue (%q %d messages, %d consumers), binding to Exchange (key %q)",
		queue.Name, queue.Messages, queue.Consumers, key)
	if err = c.channel.QueueBind(
		queue.Name,
		key,      // binding key
		exchange, // source exchange
		false,    // no wait
		nil,      // arguments
	); err != nil {
		return nil, fmt.Errorf("Queue bind: %s", err)
	}

	log.Printf("Queue bound to exchange, beginning consume (consumer tag %q)",
		c.tag)
	deliveries, err := c.channel.Consume(
		queue.Name,
		c.tag,
		false, // no ack
		false, // exclusive
		false, // no local
		false, // no wait
		nil,   // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("Consume: %s", err)
	}

	go handle(deliveries, c.done, ctag)

	return c, err
}

func handle(deliveries <-chan amqp.Delivery, done chan error, tag string) {
	for d := range deliveries {
		log.Printf("Got delivery in consumer %s (len %d): %q", tag, len(d.Body), d.Body)
		d.Ack(false)
	}
	log.Printf("handle: deliveries channel closed")
	done <- nil
}

func (c *Consumer) Shutdown() error {
	// Close the deliveries channel
	if err := c.channel.Cancel(c.tag, true); err != nil {
		return fmt.Errorf("Consumer cancel failed: %s", err)
	}

	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("Channel close failed: %s", err)
	}

	defer log.Printf("AMQP Consumer shutdown OK")

	// wait for handle to exit
	return <-c.done
}
