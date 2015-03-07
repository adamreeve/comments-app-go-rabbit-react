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
	// Channel telling the connection to send new data
	send chan bool
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
	c := &connection{
		// Does this channel need to be buffered?
		send: make(chan bool, 256),
		ws:   ws,
	}
	h.register <- c
	go sendUpdates(c)
	readComments(c)
}

func sendUpdates(c *connection) {
	pingTicker := time.NewTicker(pingPeriod)
	defer func() {
		pingTicker.Stop()
		c.ws.Close()
	}()
	// Send update immediately with initial data
	if err := sendUpdate(c.ws); err != nil {
		return
	}
	for {
		select {
		case <-pingTicker.C:
			c.ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		case _, ok := <-c.send:
			if !ok {
				// channel has been closed by the hub
				c.ws.WriteMessage(websocket.CloseMessage, []byte{})
			}
			if err := sendUpdate(c.ws); err != nil {
				return
			}
		}
	}
}

func sendUpdate(conn *websocket.Conn) error {
	conn.SetWriteDeadline(time.Now().Add(writeWait))
	messageType := websocket.TextMessage
	p := loadJsonComments()

	if err := conn.WriteMessage(messageType, p); err != nil {
		return err
	}
	return nil
}

func readComments(c *connection) {
	defer func() {
		h.unregister <- c
		c.ws.Close()
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
			log.Printf("Got new comment: %+v", newComment)
			addComment(newComment)
			h.broadcast <- true
		}
	}
}
