package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
)

type Comment struct {
	Author string `json:"author"`
	Text   string `json:"text"`
}

var commentsFile = "comments.json"

// Respond to an HTTP request the list of comments
func getComments(w http.ResponseWriter, r *http.Request) {
	log.Printf("GET comments")
	http.ServeFile(w, r, commentsFile)
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

func decodeComment(r io.ReadCloser) Comment {
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
