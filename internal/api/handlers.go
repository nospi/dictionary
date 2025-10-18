package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/nospi/dictionary/internal/dictionary"
	"github.com/nospi/dictionary/internal/game"
	"github.com/nospi/dictionary/internal/room"
	"github.com/nospi/dictionary/internal/sse"
)

// Server holds dependencies for API handlers
type Server struct {
	RoomManager *room.Manager
	SSEHub      *sse.Hub
	Dictionary  dictionary.Service
	Templates   *template.Template

	// Game engines (one per active game)
	engines map[string]*game.Engine
	enginesMu sync.RWMutex
}

// NewServer creates a new API server
func NewServer(roomMgr *room.Manager, sseHub *sse.Hub, dict dictionary.Service, tmpl *template.Template) *Server {
	return &Server{
		RoomManager: roomMgr,
		SSEHub:      sseHub,
		Dictionary:  dict,
		Templates:   tmpl,
		engines:     make(map[string]*game.Engine),
	}
}

// CreateRoomHandler handles room creation
func (s *Server) CreateRoomHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		s.renderError(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	nickname := strings.TrimSpace(r.FormValue("nickname"))
	if nickname == "" {
		s.renderError(w, "Nickname is required", http.StatusBadRequest)
		return
	}

	if len(nickname) < 2 || len(nickname) > 20 {
		s.renderError(w, "Nickname must be 2-20 characters", http.StatusBadRequest)
		return
	}

	// Create room (default 10 max players)
	room, err := s.RoomManager.CreateRoom(nickname, 10)
	if err != nil {
		log.Printf("Failed to create room: %v", err)
		s.renderError(w, "Failed to create room", http.StatusInternalServerError)
		return
	}

	log.Printf("Room created: %s by %s (player ID: %s)", room.Code, nickname, room.Players[0].ID)

	// Set player ID in cookie for SSE authentication
	http.SetCookie(w, &http.Cookie{
		Name:     "player_id",
		Value:    room.Players[0].ID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	// Redirect to room page
	w.Header().Set("HX-Redirect", fmt.Sprintf("/room/%s", room.Code))
	w.WriteHeader(http.StatusOK)
}

// JoinRoomHandler handles joining an existing room
func (s *Server) JoinRoomHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		s.renderError(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	roomCode := strings.ToUpper(strings.TrimSpace(r.FormValue("roomCode")))
	nickname := strings.TrimSpace(r.FormValue("nickname"))

	if roomCode == "" || nickname == "" {
		s.renderError(w, "Room code and nickname are required", http.StatusBadRequest)
		return
	}

	if len(nickname) < 2 || len(nickname) > 20 {
		s.renderError(w, "Nickname must be 2-20 characters", http.StatusBadRequest)
		return
	}

	// Join room
	_, player, err := s.RoomManager.JoinRoom(roomCode, nickname)
	if err != nil {
		if errors.Is(err, room.ErrRoomNotFound) {
			s.renderError(w, "Room not found", http.StatusNotFound)
		} else if errors.Is(err, room.ErrRoomFull) {
			s.renderError(w, "Room is full", http.StatusBadRequest)
		} else if errors.Is(err, room.ErrRoomLocked) {
			s.renderError(w, "Room is locked (game in progress)", http.StatusBadRequest)
		} else if errors.Is(err, room.ErrDuplicateNick) {
			s.renderError(w, "Nickname already taken in this room", http.StatusBadRequest)
		} else {
			log.Printf("Failed to join room: %v", err)
			s.renderError(w, "Failed to join room", http.StatusInternalServerError)
		}
		return
	}

	log.Printf("Player joined room: %s -> %s (player ID: %s)", nickname, roomCode, player.ID)

	// Set player ID in cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "player_id",
		Value:    player.ID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	// Broadcast room state update to all clients
	s.broadcastRoomState(roomCode)

	// Redirect to room page
	w.Header().Set("HX-Redirect", fmt.Sprintf("/room/%s", roomCode))
	w.WriteHeader(http.StatusOK)
}

// StartGameHandler handles starting a game (host only)
func (s *Server) StartGameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	roomCode, playerID, err := s.getRoomAndPlayer(r)
	if err != nil {
		s.renderError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Start game (room manager initializes game state)
	if err := s.RoomManager.StartGame(roomCode, playerID); err != nil {
		if errors.Is(err, room.ErrNotHost) {
			s.renderError(w, "Only the host can start the game", http.StatusForbidden)
		} else {
			log.Printf("Failed to start game: %v", err)
			s.renderError(w, "Failed to start game", http.StatusInternalServerError)
		}
		return
	}

	// Get room
	rm, err := s.RoomManager.GetRoom(roomCode)
	if err != nil {
		log.Printf("Failed to get room after starting game: %v", err)
		s.renderError(w, "Internal error", http.StatusInternalServerError)
		return
	}

	// Create game engine
	engine := game.NewEngine(rm, s.Dictionary)
	s.enginesMu.Lock()
	s.engines[roomCode] = engine
	s.enginesMu.Unlock()

	// Start first round
	if err := engine.StartGame(); err != nil {
		log.Printf("Failed to start game engine: %v", err)
		s.renderError(w, "Failed to start game", http.StatusInternalServerError)
		return
	}

	log.Printf("Game started in room %s by player %s", roomCode, playerID)

	// Broadcast game state to all clients
	s.broadcastGameState(roomCode)

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `<div>Game started!</div>`)
}

// SubmitDefinitionHandler handles definition submissions
func (s *Server) SubmitDefinitionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	roomCode, playerID, err := s.getRoomAndPlayer(r)
	if err != nil {
		s.renderError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Parse form
	if err := r.ParseForm(); err != nil {
		s.renderError(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	definition := strings.TrimSpace(r.FormValue("definition"))
	if definition == "" {
		s.renderError(w, "Definition is required", http.StatusBadRequest)
		return
	}

	if len(definition) > 200 {
		s.renderError(w, "Definition must be 200 characters or less", http.StatusBadRequest)
		return
	}

	// Get engine
	engine := s.getEngine(roomCode)
	if engine == nil {
		s.renderError(w, "Game not started", http.StatusBadRequest)
		return
	}

	// Submit definition
	if err := engine.SubmitDefinition(playerID, definition); err != nil {
		if errors.Is(err, game.ErrInvalidPhase) {
			s.renderError(w, "Not in writing phase", http.StatusBadRequest)
		} else if errors.Is(err, game.ErrAlreadySubmitted) {
			s.renderError(w, "You already submitted a definition", http.StatusBadRequest)
		} else {
			log.Printf("Failed to submit definition: %v", err)
			s.renderError(w, "Failed to submit definition", http.StatusInternalServerError)
		}
		return
	}

	log.Printf("Definition submitted by player %s in room %s", playerID, roomCode)

	// Broadcast game state update
	s.broadcastGameState(roomCode)

	// If phase changed to voting, broadcast voting options
	rm, _ := s.RoomManager.GetRoom(roomCode)
	if rm.GameState.Phase == game.PhaseVoting {
		s.broadcastVotingOptions(roomCode)
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `<div class="success">Definition submitted! Waiting for other players...</div>`)
}

// SubmitVoteHandler handles vote submissions
func (s *Server) SubmitVoteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	roomCode, playerID, err := s.getRoomAndPlayer(r)
	if err != nil {
		s.renderError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Parse form
	if err := r.ParseForm(); err != nil {
		s.renderError(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	definitionID := strings.TrimSpace(r.FormValue("definitionId"))
	if definitionID == "" {
		s.renderError(w, "Definition ID is required", http.StatusBadRequest)
		return
	}

	// Get engine
	engine := s.getEngine(roomCode)
	if engine == nil {
		s.renderError(w, "Game not started", http.StatusBadRequest)
		return
	}

	// Submit vote
	if err := engine.SubmitVote(playerID, definitionID); err != nil {
		if errors.Is(err, game.ErrInvalidPhase) {
			s.renderError(w, "Not in voting phase", http.StatusBadRequest)
		} else if errors.Is(err, game.ErrVoteForOwnDef) {
			s.renderError(w, "You cannot vote for your own definition", http.StatusBadRequest)
		} else {
			log.Printf("Failed to submit vote: %v", err)
			s.renderError(w, "Failed to submit vote", http.StatusInternalServerError)
		}
		return
	}

	log.Printf("Vote submitted by player %s in room %s", playerID, roomCode)

	// Broadcast game state update
	s.broadcastGameState(roomCode)

	// If phase changed to scoring, broadcast results
	rm, _ := s.RoomManager.GetRoom(roomCode)
	if rm.GameState.Phase == game.PhaseScoring {
		s.broadcastRoundResults(roomCode)
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `<div class="success">Vote submitted! Waiting for other players...</div>`)
}

// EndGameHandler handles ending a game (host only)
func (s *Server) EndGameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	roomCode, playerID, err := s.getRoomAndPlayer(r)
	if err != nil {
		s.renderError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Verify host
	rm, err := s.RoomManager.GetRoom(roomCode)
	if err != nil {
		s.renderError(w, "Room not found", http.StatusNotFound)
		return
	}

	if rm.HostID != playerID {
		s.renderError(w, "Only the host can end the game", http.StatusForbidden)
		return
	}

	// Get engine
	engine := s.getEngine(roomCode)
	if engine == nil {
		s.renderError(w, "Game not started", http.StatusBadRequest)
		return
	}

	// End game
	if err := engine.EndGame(); err != nil {
		log.Printf("Failed to end game: %v", err)
		s.renderError(w, "Failed to end game", http.StatusInternalServerError)
		return
	}

	log.Printf("Game ended in room %s by player %s", roomCode, playerID)

	// Broadcast final game state
	s.broadcastGameState(roomCode)

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `<div>Game ended!</div>`)
}

// SSE Broadcast helpers

func (s *Server) broadcastRoomState(roomCode string) {
	rm, err := s.RoomManager.GetRoom(roomCode)
	if err != nil {
		log.Printf("Failed to get room for broadcast: %v", err)
		return
	}

	// Render room state partial
	html, err := s.renderPartial("player-list", rm)
	if err != nil {
		log.Printf("Failed to render player list: %v", err)
		return
	}

	s.SSEHub.BroadcastToRoom(roomCode, sse.NewRoomStateEvent(html))
}

func (s *Server) broadcastGameState(roomCode string) {
	rm, err := s.RoomManager.GetRoom(roomCode)
	if err != nil {
		log.Printf("Failed to get room for broadcast: %v", err)
		return
	}

	// Determine which partial to render based on phase
	var partialName string
	switch rm.GameState.Phase {
	case game.PhaseLobby:
		partialName = "lobby"
	case game.PhaseWriting:
		partialName = "writing"
	case game.PhaseVoting:
		partialName = "voting"
	case game.PhaseScoring:
		partialName = "scoring"
	case game.PhaseEnded:
		partialName = "ended"
	default:
		log.Printf("Unknown game phase: %s", rm.GameState.Phase)
		return
	}

	html, err := s.renderPartial(partialName, rm)
	if err != nil {
		log.Printf("Failed to render game state: %v", err)
		return
	}

	s.SSEHub.BroadcastToRoom(roomCode, sse.NewGameStateEvent(html))
}

func (s *Server) broadcastVotingOptions(roomCode string) {
	engine := s.getEngine(roomCode)
	if engine == nil {
		return
	}

	rm, _ := s.RoomManager.GetRoom(roomCode)

	data := struct {
		Room        *game.Room
		Definitions []game.Definition
	}{
		Room:        rm,
		Definitions: engine.GetVotingOptions(),
	}

	html, err := s.renderPartial("voting", data)
	if err != nil {
		log.Printf("Failed to render voting options: %v", err)
		return
	}

	s.SSEHub.BroadcastToRoom(roomCode, sse.NewVotingOptionsEvent(html))
}

func (s *Server) broadcastRoundResults(roomCode string) {
	engine := s.getEngine(roomCode)
	if engine == nil {
		return
	}

	rm, _ := s.RoomManager.GetRoom(roomCode)

	data := struct {
		Room    *game.Room
		Results map[string]interface{}
	}{
		Room:    rm,
		Results: engine.GetRoundResults(),
	}

	html, err := s.renderPartial("scoring", data)
	if err != nil {
		log.Printf("Failed to render round results: %v", err)
		return
	}

	s.SSEHub.BroadcastToRoom(roomCode, sse.NewRoundResultsEvent(html))
}

// Helper methods

func (s *Server) getRoomAndPlayer(r *http.Request) (roomCode, playerID string, err error) {
	// Get room code from form or URL
	if err := r.ParseForm(); err != nil {
		return "", "", fmt.Errorf("invalid form data")
	}

	roomCode = strings.ToUpper(strings.TrimSpace(r.FormValue("roomCode")))
	if roomCode == "" {
		return "", "", fmt.Errorf("room code is required")
	}

	// Get player ID from cookie
	cookie, err := r.Cookie("player_id")
	if err != nil {
		return "", "", fmt.Errorf("player ID not found (cookie missing)")
	}
	playerID = cookie.Value

	return roomCode, playerID, nil
}

func (s *Server) getEngine(roomCode string) *game.Engine {
	s.enginesMu.RLock()
	defer s.enginesMu.RUnlock()
	return s.engines[roomCode]
}

func (s *Server) renderPartial(name string, data interface{}) (string, error) {
	var buf strings.Builder
	if err := s.Templates.ExecuteTemplate(&buf, name, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (s *Server) renderError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(statusCode)
	fmt.Fprintf(w, `<div class="error">%s</div>`, message)
}

// SSEHandler serves Server-Sent Events for real-time updates
func (s *Server) SSEHandler(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Get room code and player ID from query params
	roomCode := r.URL.Query().Get("roomCode")
	playerID := r.URL.Query().Get("playerId")

	if roomCode == "" || playerID == "" {
		http.Error(w, "roomCode and playerId required", http.StatusBadRequest)
		return
	}

	// Verify room exists
	_, err := s.RoomManager.GetRoom(roomCode)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Create SSE client
	client := s.SSEHub.NewClient(roomCode, playerID)
	s.SSEHub.RegisterClient(client)
	defer s.SSEHub.UnregisterClient(client)

	log.Printf("SSE connection established: player=%s room=%s", playerID, roomCode)

	// Flush interface for streaming
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Send initial connection confirmation
	fmt.Fprintf(w, ": connected\n\n")
	flusher.Flush()

	// Create ticker for keep-alive
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Stream events
	for {
		select {
		case <-r.Context().Done():
			// Client disconnected
			log.Printf("SSE connection closed: player=%s room=%s", playerID, roomCode)
			return

		case event, ok := <-client.Events:
			if !ok {
				// Channel closed
				return
			}

			// Write SSE event
			writer := func(msg string) error {
				_, err := fmt.Fprint(w, msg)
				if err != nil {
					return err
				}
				flusher.Flush()
				return nil
			}

			if err := client.SendEvent(event, writer); err != nil {
				log.Printf("Failed to send SSE event: %v", err)
				return
			}

		case <-ticker.C:
			// Send keep-alive comment
			fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()
		}
	}
}

// HealthHandler for health checks
func (s *Server) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "ok",
		"rooms":     s.RoomManager.GetRoomCount(),
		"timestamp": time.Now().Unix(),
	})
}
