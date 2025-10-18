package ws

import (
	"log"

	"github.com/google/uuid"
	"github.com/nospi/dictionary/internal/game"
)

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
	dictionary    DictionaryService
	engines       map[string]*game.Engine // roomCode -> engine
}

// RoomManager defines room operations
type RoomManager interface {
	CreateRoom(hostNickname string, maxPlayers int) (*Room, error)
	JoinRoom(roomCode, nickname string) (*Room, *Player, error)
	GetRoom(roomCode string) (*Room, error)
	PlayerDisconnected(roomCode, playerID string) error
	StartGame(roomCode, playerID string) error
}

// DictionaryService defines dictionary operations
type DictionaryService interface {
	GetRandomWord() (Word, error)
}

// NewHub creates a new WebSocket hub
func NewHub(roomManager RoomManager, dictionary DictionaryService) *Hub {
	return &Hub{
		clients:       make(map[*Client]bool),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		handleMessage: make(chan *ClientMessage),
		roomManager:   roomManager,
		dictionary:    dictionary,
		engines:       make(map[string]*game.Engine),
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

				// Notify room manager about disconnection
				if client.PlayerID != "" && client.RoomCode != "" {
					if err := h.roomManager.PlayerDisconnected(client.RoomCode, client.PlayerID); err != nil {
						log.Printf("Error handling player disconnect: %v", err)
					}
					// Clean up engine if needed
					delete(h.engines, client.RoomCode)
				}
			}

		case clientMsg := <-h.handleMessage:
			h.processMessage(clientMsg)
		}
	}
}

// processMessage handles incoming messages from clients
func (h *Hub) processMessage(cm *ClientMessage) {
	log.Printf("Processing message type %s from client %s", cm.Message.Type, cm.Client.ID)

	switch cm.Message.Type {
	case MsgJoinRoom:
		h.handleJoinRoom(cm)

	case MsgStartGame:
		h.handleStartGame(cm)

	case MsgSubmitDefinition:
		h.handleSubmitDefinition(cm)

	case MsgSubmitVote:
		h.handleSubmitVote(cm)

	case MsgEndGame:
		h.handleEndGame(cm)

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
