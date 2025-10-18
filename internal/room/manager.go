package room

import (
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nospi/dictionary/internal/game"
)

var (
	ErrRoomNotFound    = errors.New("room not found")
	ErrRoomFull        = errors.New("room is full")
	ErrRoomLocked      = errors.New("room is locked")
	ErrNotHost         = errors.New("only host can perform this action")
	ErrPlayerNotFound  = errors.New("player not found in room")
	ErrDuplicateNick   = errors.New("nickname already taken in room")
)

// Manager handles room lifecycle and player management
type Manager struct {
	rooms map[string]*game.Room
	mu    sync.RWMutex
}

// NewManager creates a new room manager
func NewManager() *Manager {
	return &Manager{
		rooms: make(map[string]*game.Room),
	}
}

// CreateRoom creates a new room and returns the room code
func (m *Manager) CreateRoom(hostNickname string, maxPlayers int) (*game.Room, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Generate unique room code
	code := m.generateRoomCode()
	for m.rooms[code] != nil {
		code = m.generateRoomCode()
	}

	// Create host player
	host := &game.Player{
		ID:          uuid.New().String(),
		Nickname:    hostNickname,
		RoomCode:    code,
		ConnectedAt: time.Now(),
		IsHost:      true,
		IsActive:    true,
	}

	// Create room
	room := &game.Room{
		Code:       code,
		HostID:     host.ID,
		Players:    []*game.Player{host},
		MaxPlayers: maxPlayers,
		CreatedAt:  time.Now(),
		IsLocked:   false,
		GameState:  nil, // No game state until started
	}

	m.rooms[code] = room

	return room, nil
}

// JoinRoom adds a player to an existing room
func (m *Manager) JoinRoom(roomCode, nickname string) (*game.Room, *game.Player, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	room, exists := m.rooms[roomCode]
	if !exists {
		return nil, nil, ErrRoomNotFound
	}

	if room.IsLocked {
		return nil, nil, ErrRoomLocked
	}

	if len(room.Players) >= room.MaxPlayers {
		return nil, nil, ErrRoomFull
	}

	// Check for duplicate nickname
	for _, p := range room.Players {
		if strings.EqualFold(p.Nickname, nickname) {
			return nil, nil, ErrDuplicateNick
		}
	}

	// Create new player
	player := &game.Player{
		ID:          uuid.New().String(),
		Nickname:    nickname,
		RoomCode:    roomCode,
		ConnectedAt: time.Now(),
		IsHost:      false,
		IsActive:    true,
	}

	room.Players = append(room.Players, player)

	return room, player, nil
}

// GetRoom retrieves a room by code
func (m *Manager) GetRoom(roomCode string) (*game.Room, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	room, exists := m.rooms[roomCode]
	if !exists {
		return nil, ErrRoomNotFound
	}

	return room, nil
}

// PlayerDisconnected handles a player disconnecting
func (m *Manager) PlayerDisconnected(roomCode, playerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	room, exists := m.rooms[roomCode]
	if !exists {
		return ErrRoomNotFound
	}

	// Find and mark player as inactive
	var player *game.Player
	for _, p := range room.Players {
		if p.ID == playerID {
			p.IsActive = false
			player = p
			break
		}
	}

	if player == nil {
		return ErrPlayerNotFound
	}

	// If host disconnected, migrate host to oldest remaining active player
	if player.IsHost {
		for _, p := range room.Players {
			if p.IsActive && p.ID != playerID {
				p.IsHost = true
				room.HostID = p.ID
				break
			}
		}
	}

	// If no active players remain, schedule room cleanup
	hasActivePlayers := false
	for _, p := range room.Players {
		if p.IsActive {
			hasActivePlayers = true
			break
		}
	}

	if !hasActivePlayers {
		// TODO: Schedule room deletion after grace period
		// For now, just delete immediately
		delete(m.rooms, roomCode)
	}

	return nil
}

// StartGame initializes the game state for a room
func (m *Manager) StartGame(roomCode, playerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	room, exists := m.rooms[roomCode]
	if !exists {
		return ErrRoomNotFound
	}

	// Verify player is host
	if room.HostID != playerID {
		return ErrNotHost
	}

	// Lock room
	room.IsLocked = true

	// Initialize game state
	room.GameState = &game.GameState{
		Phase:          game.PhaseLobby,
		CurrentRound:   0,
		TurnOrder:      m.buildTurnOrder(room),
		Scores:         make(map[string]int),
		Votes:          make(map[string]string),
		PhaseStartedAt: time.Now(),
	}

	// Initialize scores for all players
	for _, player := range room.Players {
		if player.IsActive {
			room.GameState.Scores[player.ID] = 0
		}
	}

	return nil
}

// buildTurnOrder creates a turn order from active players
func (m *Manager) buildTurnOrder(room *game.Room) []string {
	var turnOrder []string
	for _, player := range room.Players {
		if player.IsActive {
			turnOrder = append(turnOrder, player.ID)
		}
	}
	return turnOrder
}

// generateRoomCode generates a 6-character alphanumeric room code
func (m *Manager) generateRoomCode() string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // Exclude ambiguous chars
	const codeLength = 6

	b := make([]byte, codeLength)
	rand.Read(b)

	code := make([]byte, codeLength)
	for i := 0; i < codeLength; i++ {
		code[i] = chars[int(b[i])%len(chars)]
	}

	return string(code)
}

// GetRoomCount returns the total number of active rooms
func (m *Manager) GetRoomCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.rooms)
}

// ListRooms returns all room codes (for debugging)
func (m *Manager) ListRooms() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	codes := make([]string, 0, len(m.rooms))
	for code := range m.rooms {
		codes = append(codes, code)
	}
	return codes
}

// String returns a string representation of the manager state
func (m *Manager) String() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return fmt.Sprintf("RoomManager{rooms: %d}", len(m.rooms))
}
