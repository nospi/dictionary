package ws

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/coder/websocket"
)

// Client represents a WebSocket client connection
type Client struct {
	ID       string
	PlayerID string
	RoomCode string
	conn     *websocket.Conn
	hub      *Hub
	send     chan []byte
	mu       sync.Mutex
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewClient creates a new WebSocket client
func NewClient(conn *websocket.Conn, hub *Hub) *Client {
	ctx, cancel := context.WithCancel(context.Background())

	return &Client{
		conn:   conn,
		hub:    hub,
		send:   make(chan []byte, 256),
		ctx:    ctx,
		cancel: cancel,
	}
}

// ReadPump reads messages from the WebSocket connection
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close(websocket.StatusNormalClosure, "")
	}()

	for {
		_, data, err := c.conn.Read(c.ctx)
		if err != nil {
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
				log.Printf("Client %s closed connection normally", c.ID)
			} else {
				log.Printf("Read error for client %s: %v", c.ID, err)
			}
			break
		}

		// Parse message
		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Printf("JSON unmarshal error: %v", err)
			c.SendError("Invalid message format")
			continue
		}

		// Handle message
		c.hub.handleMessage <- &ClientMessage{
			Client:  c,
			Message: msg,
		}
	}
}

// WritePump writes messages to the WebSocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(54 * time.Second) // Ping every 54 seconds
	defer func() {
		ticker.Stop()
		c.cancel()
	}()

	for {
		select {
		case <-c.ctx.Done():
			return

		case message, ok := <-c.send:
			if !ok {
				// Channel closed
				c.conn.Close(websocket.StatusNormalClosure, "")
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			err := c.conn.Write(ctx, websocket.MessageText, message)
			cancel()

			if err != nil {
				log.Printf("Write error for client %s: %v", c.ID, err)
				return
			}

		case <-ticker.C:
			// Send ping
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			err := c.conn.Ping(ctx)
			cancel()

			if err != nil {
				log.Printf("Ping error for client %s: %v", c.ID, err)
				return
			}
		}
	}
}

// SendMessage sends a message to the client
func (c *Client) SendMessage(msgType MessageType, payload interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	msg := Message{
		Type:      msgType,
		Payload:   payload,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	select {
	case c.send <- data:
		return nil
	case <-c.ctx.Done():
		return c.ctx.Err()
	default:
		// Channel full, client too slow
		log.Printf("Client %s send buffer full, closing connection", c.ID)
		close(c.send)
		return nil
	}
}

// SendError sends an error message to the client
func (c *Client) SendError(message string) {
	c.SendMessage(MsgError, ErrorPayload{Message: message})
}

// Close closes the client connection
func (c *Client) Close() {
	c.cancel()
	close(c.send)
}
