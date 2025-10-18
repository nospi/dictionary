package ws

import (
	"log"
	"net/http"

	"github.com/coder/websocket"
)

// ServeWS handles WebSocket upgrade requests
func ServeWS(hub *Hub, w http.ResponseWriter, r *http.Request) {
	// Accept WebSocket connection with permissive origin (TODO: tighten for production)
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // Allow all origins for development
	})
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		http.Error(w, "Could not upgrade connection", http.StatusBadRequest)
		return
	}

	// Create new client
	client := NewClient(conn, hub)

	// Register client with hub
	hub.register <- client

	// Start client goroutines
	go client.WritePump()
	go client.ReadPump()

	log.Printf("WebSocket connection established from %s", r.RemoteAddr)
}
