package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write updates to the client.
	writeWait = 10 * time.Second
	// Time allowed to read the next pong message from the client.
	pongWait = 60 * time.Second
	// Send pings to client with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
	// Send updates every:
	updateInterval = 5 * time.Second
)

type connection struct {
	// Websocket connection
	ws *websocket.Conn
	// Message queue consumer
	consumer *Consumer
	// Message queue publisher
	publisher *Publisher
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Respond to a websocket request the list of comments
func getComments(w http.ResponseWriter, r *http.Request) {
	log.Printf("WebSocket comments")
	// Try and upgrade to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading to websocket connection: %s", err)
		return
	}

	messagesToSend := make(chan []byte)

	consumer, err := newConsumer(
		mqUri,
		mqExchangeName,
		mqExchangeType,
		"",
		mqBindingKey,
		"",
		messagesToSend,
	)
	if err != nil {
		panic(err)
	}

	publisher, err := newPublisher(
		mqUri,
		mqExchangeName,
		mqExchangeType,
		mqSendReliable,
	)
	if err != nil {
		panic(err)
	}

	c := &connection{
		ws:        ws,
		consumer:  consumer,
		publisher: publisher,
	}
	go sendMessages(c, messagesToSend)
	readMessages(c)
	consumer.Shutdown()
}

// Reads incoming messages from Rabbit through the messages
// channel and then sends them over the websocket to the client.
func sendMessages(c *connection, messages chan []byte) {
	pingTicker := time.NewTicker(pingPeriod)
	defer func() {
		pingTicker.Stop()
		c.ws.Close()
	}()
	// Send update immediately with initial data
	if err := sendInitialData(c.ws); err != nil {
		return
	}
	for {
		select {
		case <-pingTicker.C:
			c.ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		case message, ok := <-messages:
			if !ok {
				c.ws.WriteMessage(websocket.CloseMessage, []byte{})
			}
			if err := sendMessage(c.ws, message); err != nil {
				return
			}
		}
	}
}

func sendInitialData(conn *websocket.Conn) error {
	conn.SetWriteDeadline(time.Now().Add(writeWait))
	messageType := websocket.TextMessage

	comments := loadComments()

	for _, comment := range comments {
		if err := conn.WriteMessage(messageType, encodeComment(comment)); err != nil {
			return err
		}
	}

	return nil
}

func sendMessage(conn *websocket.Conn, message []byte) error {
	conn.SetWriteDeadline(time.Now().Add(writeWait))
	messageType := websocket.TextMessage

	if err := conn.WriteMessage(messageType, message); err != nil {
		return err
	}
	return nil
}

// Reads messages from the client and then publishes them
// to all other clients via Rabbit MQ.
func readMessages(c *connection) {
	commentChan := make(chan Comment)
	go readWebsocket(c, commentChan)

	for {
		newComment, ok := <-commentChan
		if !ok {
			// Channel closed
			return
		}
		log.Printf("Got new comment: %+v", newComment)
		c.publisher.Publish(mqBindingKey, encodeComment(newComment))
	}
}

func readWebsocket(c *connection, commentChan chan Comment) {
	defer func() {
		c.ws.Close()
		close(commentChan)
	}()
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error {
		c.ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, r, err := c.ws.NextReader()
		if err != nil {
			log.Printf("Error reading socket: %v", err)
			return
		} else {
			newComment := decodeComment(r)
			commentChan <- newComment
		}
	}
}
