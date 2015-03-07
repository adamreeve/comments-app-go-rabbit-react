package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"strings"
)

type Comment struct {
	Author string `json:"author"`
	Text   string `json:"text"`
}

var commentsFile = "comments.json"

// Respond to an HTTP request for a page
func getComments(w http.ResponseWriter, r *http.Request) {
	log.Printf("GET comments")
	http.ServeFile(w, r, commentsFile)
}

// Respond to an HTTP request for a page
func postComments(w http.ResponseWriter, r *http.Request) {
	log.Printf("POST comments")
	var comment Comment
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&comment)
	if err != nil {
		panic(err)
	}

	var allComments []Comment
	file, err := os.Open(commentsFile)
	if err != nil {
		panic(err)
	}
	decoder = json.NewDecoder(file)
	err = decoder.Decode(&allComments)
	if err != nil {
		panic(err)
	}
	file.Close()
	allComments = append(allComments, comment)

	file, err = os.Create(commentsFile)
	if err != nil {
		panic(err)
	}
	encoder := json.NewEncoder(file)
	encoder.Encode(allComments)
	file.Close()

	encoder = json.NewEncoder(w)
	encoder.Encode(allComments)
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
