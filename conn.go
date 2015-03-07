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

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Respond to a websocket request the list of comments
func getComments(w http.ResponseWriter, r *http.Request) {
	log.Printf("WebSocket comments")
	// Try and upgrade to a websocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading to websocket connection: %s", err)
		return
	}
	go sendUpdates(conn)
	readComments(conn)
}

func sendUpdates(conn *websocket.Conn) {
	pingTicker := time.NewTicker(pingPeriod)
	sendUpdatesTicker := time.NewTicker(updateInterval)
	defer func() {
		pingTicker.Stop()
		sendUpdatesTicker.Stop()
		conn.Close()
	}()
	// Send update immediately with initial data
	if err := sendUpdate(conn); err != nil {
		return
	}
	for {
		select {
		case <-pingTicker.C:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		case <-sendUpdatesTicker.C:
			if err := sendUpdate(conn); err != nil {
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

func readComments(conn *websocket.Conn) {
	defer conn.Close()
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, r, err := conn.NextReader()
		if err != nil {
			log.Printf("Error reading socket: %v", err)
			return
		} else {
			newComment := decodeComment(r)
			log.Printf("Got new comment: %+v", newComment)
			addComment(newComment)
		}
	}
}
