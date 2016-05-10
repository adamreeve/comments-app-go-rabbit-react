package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/pborman/uuid"
)

type Comment struct {
	Author string    `json:"author"`
	Text   string    `json:"text"`
	ID     uuid.UUID `json:"id"`
}

const (
	commentsFile = "comments.json"
)

func (c Comment) String() string {
	return fmt.Sprintf("{Author:\"%s\", Text:\"%s\", ID:\"%s\"}", c.Author, c.Text, c.ID)
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

func encodeComment(comment Comment) []byte {
	enc, err := json.Marshal(comment)
	if err != nil {
		panic(err)
	}
	return enc
}

func addComment(comment Comment) {
	allComments := loadComments()
	allComments = append(allComments, comment)
	saveComments(allComments)
}
