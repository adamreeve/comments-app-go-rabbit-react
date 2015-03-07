package main

import (
	"log"
)

type hub struct {
	// Registered connections
	connections map[*connection]bool

	// Send updated data to clients
	broadcast chan bool

	// Registration requests from connections
	register chan *connection

	// Unregister requests from connections
	unregister chan *connection
}

var h = hub{
	register:    make(chan *connection),
	unregister:  make(chan *connection),
	broadcast:   make(chan bool),
	connections: make(map[*connection]bool),
}

func (h *hub) run() {
	for {
		select {
		case c := <-h.register:
			h.connections[c] = true
			log.Printf("Registering new channel")
		case c := <-h.unregister:
			if _, ok := h.connections[c]; ok {
				log.Printf("Unregistering channel")
				delete(h.connections, c)
				close(c.send)
			}
		case b := <-h.broadcast:
			for c := range h.connections {
				select {
				case c.send <- b:
				default:
					// Closed or no room in the buffer
					close(c.send)
					delete(h.connections, c)
				}
			}
		}
	}
}
