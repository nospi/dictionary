package sse

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// Event represents an SSE event to be sent to clients
type Event struct {
	Name string // Event name (e.g., "game-state", "room-state")
	Data string // HTML data to send
}

// Client represents a single SSE connection
type Client struct {
	RoomCode string
	PlayerID string
	Events   chan Event
	mu       sync.Mutex
}

// BroadcastMessage contains an event to broadcast to a room
type BroadcastMessage struct {
	RoomCode string
	Event    Event
}

// Hub maintains the set of active SSE clients and broadcasts events to rooms
type Hub struct {
	// Clients grouped by room code
	rooms map[string]map[*Client]bool

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Broadcast messages to all clients in a room
	broadcast chan *BroadcastMessage

	// Mutex for thread-safe access to rooms map
	mu sync.RWMutex
}

// NewHub creates a new SSE Hub
func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *BroadcastMessage, 256),
	}
}

// Run starts the hub's main event loop
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("SSE Hub shutting down")
			return

		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastToRoom(message)
		}
	}
}

// registerClient adds a client to a room
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.rooms[client.RoomCode] == nil {
		h.rooms[client.RoomCode] = make(map[*Client]bool)
	}
	h.rooms[client.RoomCode][client] = true

	log.Printf("SSE client registered: player=%s room=%s (total in room: %d)",
		client.PlayerID, client.RoomCode, len(h.rooms[client.RoomCode]))
}

// unregisterClient removes a client from a room
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.rooms[client.RoomCode]; ok {
		if _, exists := clients[client]; exists {
			delete(clients, client)
			close(client.Events)

			log.Printf("SSE client unregistered: player=%s room=%s (remaining in room: %d)",
				client.PlayerID, client.RoomCode, len(clients))

			// Clean up empty rooms
			if len(clients) == 0 {
				delete(h.rooms, client.RoomCode)
				log.Printf("SSE room cleaned up: %s (no clients remaining)", client.RoomCode)
			}
		}
	}
}

// broadcastToRoom sends an event to all clients in a specific room
func (h *Hub) broadcastToRoom(message *BroadcastMessage) {
	h.mu.RLock()
	clients := h.rooms[message.RoomCode]
	h.mu.RUnlock()

	if clients == nil {
		log.Printf("Attempted to broadcast to non-existent room: %s", message.RoomCode)
		return
	}

	log.Printf("Broadcasting event '%s' to room %s (%d clients)",
		message.Event.Name, message.RoomCode, len(clients))

	for client := range clients {
		select {
		case client.Events <- message.Event:
			// Event sent successfully
		default:
			// Client's event buffer is full, unregister them
			log.Printf("Client buffer full, unregistering: player=%s room=%s",
				client.PlayerID, client.RoomCode)
			go func(c *Client) {
				h.unregister <- c
			}(client)
		}
	}
}

// BroadcastToRoom sends an event to all clients in a room (public method)
func (h *Hub) BroadcastToRoom(roomCode string, event Event) {
	h.broadcast <- &BroadcastMessage{
		RoomCode: roomCode,
		Event:    event,
	}
}

// BroadcastToPlayer sends an event to a specific player (useful for errors)
func (h *Hub) BroadcastToPlayer(roomCode, playerID string, event Event) {
	h.mu.RLock()
	clients := h.rooms[roomCode]
	h.mu.RUnlock()

	if clients == nil {
		return
	}

	for client := range clients {
		if client.PlayerID == playerID {
			select {
			case client.Events <- event:
				log.Printf("Event '%s' sent to player %s in room %s",
					event.Name, playerID, roomCode)
			default:
				log.Printf("Failed to send event to player %s (buffer full)", playerID)
			}
			return
		}
	}

	log.Printf("Player %s not found in room %s for event broadcast", playerID, roomCode)
}

// NewClient creates a new SSE client
func (h *Hub) NewClient(roomCode, playerID string) *Client {
	return &Client{
		RoomCode: roomCode,
		PlayerID: playerID,
		Events:   make(chan Event, 16), // Buffer 16 events
	}
}

// RegisterClient registers a client with the hub
func (h *Hub) RegisterClient(client *Client) {
	h.register <- client
}

// UnregisterClient unregisters a client from the hub
func (h *Hub) UnregisterClient(client *Client) {
	h.unregister <- client
}

// SendEvent sends an SSE event to the client's response writer
// This is called by the HTTP handler in a loop
func (client *Client) SendEvent(event Event, writer func(string) error) error {
	client.mu.Lock()
	defer client.mu.Unlock()

	// Format as SSE event
	var msg string
	if event.Name != "" {
		msg += fmt.Sprintf("event: %s\n", event.Name)
	}
	msg += fmt.Sprintf("data: %s\n\n", event.Data)

	return writer(msg)
}

// GetRoomClientCount returns the number of active clients in a room
func (h *Hub) GetRoomClientCount(roomCode string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.rooms[roomCode]; ok {
		return len(clients)
	}
	return 0
}

// Helper function to create common event types
func NewRoomStateEvent(html string) Event {
	return Event{Name: "room-state", Data: html}
}

func NewGameStateEvent(html string) Event {
	return Event{Name: "game-state", Data: html}
}

func NewVotingOptionsEvent(html string) Event {
	return Event{Name: "voting-options", Data: html}
}

func NewRoundResultsEvent(html string) Event {
	return Event{Name: "round-results", Data: html}
}

func NewErrorEvent(html string) Event {
	return Event{Name: "error", Data: html}
}

// KeepAliveEvent sends a comment to keep the connection alive
func NewKeepAliveEvent() Event {
	return Event{Name: "", Data: fmt.Sprintf(": keepalive %d", time.Now().Unix())}
}
