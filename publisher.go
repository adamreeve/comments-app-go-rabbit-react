package main

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

type Publisher struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	reliable bool
	ack      chan uint64
	nack     chan uint64
	exchange string
}

func newPublisher(amqpUri, exchange, exchangeType string, reliable bool) (*Publisher, error) {
	var err error
	p := &Publisher{
		conn:     nil,
		channel:  nil,
		reliable: mqSendReliable,
		exchange: exchange,
	}

	log.Printf("Publisher dialing %s", amqpUri)
	p.conn, err = amqp.Dial(amqpUri)
	if err != nil {
		return nil, fmt.Errorf("Dial: %s", err)
	}

	log.Printf("Publisher got connection, getting channel")
	p.channel, err = p.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("Channel: %s", err)
	}

	log.Printf("Publisher declaring exchange %s", p.exchange)
	if err = p.channel.ExchangeDeclare(
		p.exchange,
		exchangeType,
		true,  // durable
		false, // delete when complete
		false, // internal
		false, // no wait
		nil,   // arguments
	); err != nil {
		return nil, fmt.Errorf("Exchange declare: %s", err)
	}

	if reliable {
		log.Printf("Enabling reliable publishing")
		if err := p.channel.Confirm(false); err != nil {
			return nil, fmt.Errorf("Channel could not be put in confirm mode: %s", err)
		}

		p.ack, p.nack = p.channel.NotifyConfirm(make(chan uint64, 1), make(chan uint64, 1))
	}

	return p, err
}

func (p *Publisher) Publish(routingKey string, body []byte) error {
	log.Printf("Publishing %dB body (%q)", len(body), body)

	publishing := amqp.Publishing{
		Headers:         amqp.Table{},
		ContentType:     "text/plain",
		ContentEncoding: "",
		Body:            []byte(body),
		DeliveryMode:    amqp.Transient,
		Priority:        0, // 0 - 9
	}

	var confirmed = false
	for !confirmed {

		if err := p.channel.Publish(
			p.exchange,
			routingKey,
			false, // mandatory
			false, // immediate,
			publishing,
		); err != nil {
			return fmt.Errorf("Publishing: %s", err)
		}

		if p.reliable {
			select {
			case tag := <-p.ack:
				confirmed = true
				log.Printf("Confirmed delivery with tag: %d", tag)
			case tag := <-p.nack:
				log.Printf("Failed delivery of tag: %d", tag)
			}
		} else {
			confirmed = true
		}

	}

	return nil
}
