package ws

import (
	"log"

	"github.com/google/uuid"
)

// RoomManager defines the interface for room operations
type RoomManager interface {
	// Define room manager interface methods here
	// Will be implemented when we create the room package
}

// ClientMessage combines a client with their message
type ClientMessage struct {
	Client  *Client
	Message Message
}

// Hub maintains active clients and broadcasts messages
type Hub struct {
	clients       map[*Client]bool
	register      chan *Client
	unregister    chan *Client
	handleMessage chan *ClientMessage
	roomManager   RoomManager
}

// NewHub creates a new WebSocket hub
func NewHub(roomManager RoomManager) *Hub {
	return &Hub{
		clients:       make(map[*Client]bool),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		handleMessage: make(chan *ClientMessage),
		roomManager:   roomManager,
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			client.ID = uuid.New().String()
			log.Printf("Client registered: %s (total: %d)", client.ID, len(h.clients))

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("Client unregistered: %s (total: %d)", client.ID, len(h.clients))

				// TODO: Notify room manager about disconnection
				// if client.PlayerID != "" && client.RoomCode != "" {
				//     h.roomManager.PlayerDisconnected(client.RoomCode, client.PlayerID)
				// }
			}

		case clientMsg := <-h.handleMessage:
			h.processMessage(clientMsg)
		}
	}
}

// processMessage handles incoming messages from clients
func (h *Hub) processMessage(cm *ClientMessage) {
	log.Printf("Processing message type %s from client %s", cm.Message.Type, cm.Client.ID)

	// TODO: Implement message handlers
	// For now, just log the message type
	switch cm.Message.Type {
	case MsgJoinRoom:
		log.Printf("JOIN_ROOM request from client %s", cm.Client.ID)
		// TODO: Handle room join
		cm.Client.SendError("Not yet implemented")

	case MsgStartGame:
		log.Printf("START_GAME request from client %s", cm.Client.ID)
		// TODO: Handle game start
		cm.Client.SendError("Not yet implemented")

	case MsgSubmitDefinition:
		log.Printf("SUBMIT_DEFINITION from client %s", cm.Client.ID)
		// TODO: Handle definition submission
		cm.Client.SendError("Not yet implemented")

	case MsgSubmitVote:
		log.Printf("SUBMIT_VOTE from client %s", cm.Client.ID)
		// TODO: Handle vote submission
		cm.Client.SendError("Not yet implemented")

	case MsgEndGame:
		log.Printf("END_GAME request from client %s", cm.Client.ID)
		// TODO: Handle game end
		cm.Client.SendError("Not yet implemented")

	default:
		log.Printf("Unknown message type: %s", cm.Message.Type)
		cm.Client.SendError("Unknown message type")
	}
}

// BroadcastToRoom sends a message to all clients in a specific room
func (h *Hub) BroadcastToRoom(roomCode string, msgType MessageType, payload interface{}) {
	for client := range h.clients {
		if client.RoomCode == roomCode {
			client.SendMessage(msgType, payload)
		}
	}
}

// SendToClient sends a message to a specific client by player ID
func (h *Hub) SendToClient(playerID string, msgType MessageType, payload interface{}) bool {
	for client := range h.clients {
		if client.PlayerID == playerID {
			return client.SendMessage(msgType, payload) == nil
		}
	}
	return false
}
