package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// Serve a static file out of the web directory
func serveFile(w http.ResponseWriter, r *http.Request) {
	log.Printf("GET file: %s", r.RequestURI)
	http.ServeFile(w, r, strings.Join([]string{"web", r.RequestURI}, "/"))
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

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/comments", getComments).Methods("GET")
	router.HandleFunc("/comments", postComments).Methods("POST")

	// Default to the file handler for anything that doesn't match above
	router.PathPrefix("/").HandlerFunc(serveFile)

	log.Printf("Listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
