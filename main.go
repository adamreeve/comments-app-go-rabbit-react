package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"code.google.com/p/go-uuid/uuid"
	"github.com/gorilla/mux"
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

type Comment struct {
	Author string    `json:"author"`
	Text   string    `json:"text"`
	ID     uuid.UUID `json:"id"`
}

func (c Comment) String() string {
	return fmt.Sprintf("{Author:\"%s\", Text:\"%s\", ID:\"%s\"}", c.Author, c.Text, c.ID)
}

var commentsFile = "comments.json"

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

// Respond to an HTTP POST request for uploading a comment
func postComments(w http.ResponseWriter, r *http.Request) {
	log.Printf("POST comments")

	comment := decodeComment(r.Body)

	addComment(comment)

	comments := loadComments()

	encoder := json.NewEncoder(w)
	encoder.Encode(comments)
}

func loadJsonComments() []byte {
	contents, err := ioutil.ReadFile(commentsFile)
	if err != nil {
		panic(err)
	}
	return contents
}

func loadComments() []Comment {
	var comments []Comment
	file, err := os.Open(commentsFile)
	if err != nil {
		panic(err)
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&comments)
	if err != nil {
		panic(err)
	}
	file.Close()
	return comments
}

func saveComments(comments []Comment) {
	file, err := os.Create(commentsFile)
	if err != nil {
		panic(err)
	}
	encoder := json.NewEncoder(file)
	encoder.Encode(comments)
	file.Close()
}

func decodeComment(r io.Reader) Comment {
	var comment Comment
	decoder := json.NewDecoder(r)
	err := decoder.Decode(&comment)
	if err != nil {
		panic(err)
	}
	return comment
}

func addComment(comment Comment) {
	allComments := loadComments()
	allComments = append(allComments, comment)
	saveComments(allComments)
}

// Serve a static file out of the web directory
func serveFile(w http.ResponseWriter, r *http.Request) {
	log.Printf("GET file: %s", r.RequestURI)
	http.ServeFile(w, r, strings.Join([]string{"web", r.RequestURI}, "/"))
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/comments", getComments).Methods("GET")
	router.HandleFunc("/comments", postComments).Methods("POST")

	// Default to the file handler for anything that doesn't match above
	router.PathPrefix("/").HandlerFunc(serveFile)

	log.Printf("Listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
