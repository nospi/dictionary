package ws

import (
	"encoding/json"
	"log"

	"github.com/nospi/dictionary/internal/game"
)

// handleJoinRoom processes JOIN_ROOM messages
func (h *Hub) handleJoinRoom(cm *ClientMessage) {
	var payload JoinRoomPayload
	if err := unmarshalPayload(cm.Message.Payload, &payload); err != nil {
		cm.Client.SendError("Invalid payload")
		return
	}

	log.Printf("Client %s attempting to join room %s as %s", cm.Client.ID, payload.RoomCode, payload.Nickname)

	// If room code is empty, create a new room
	var room *game.Room
	var player *game.Player
	var err error

	if payload.RoomCode == "" {
		// Create new room
		room, err = h.roomManager.CreateRoom(payload.Nickname, 10) // Default max 10 players
		if err != nil {
			log.Printf("Failed to create room: %v", err)
			cm.Client.SendError("Failed to create room")
			return
		}
		player = room.Players[0] // Host is first player
		log.Printf("Created new room %s for client %s", room.Code, cm.Client.ID)
	} else {
		// Join existing room
		room, player, err = h.roomManager.JoinRoom(payload.RoomCode, payload.Nickname)
		if err != nil {
			log.Printf("Failed to join room: %v", err)
			cm.Client.SendError(err.Error())
			return
		}
		log.Printf("Client %s joined room %s as player %s", cm.Client.ID, room.Code, player.ID)
	}

	// Associate client with player and room
	cm.Client.PlayerID = player.ID
	cm.Client.RoomCode = room.Code

	// Send room state to all clients in room
	h.broadcastRoomState(room)
}

// handleStartGame processes START_GAME messages
func (h *Hub) handleStartGame(cm *ClientMessage) {
	if cm.Client.RoomCode == "" || cm.Client.PlayerID == "" {
		cm.Client.SendError("Not in a room")
		return
	}

	log.Printf("Client %s (%s) attempting to start game in room %s",
		cm.Client.ID, cm.Client.PlayerID, cm.Client.RoomCode)

	// Start game state in room manager
	if err := h.roomManager.StartGame(cm.Client.RoomCode, cm.Client.PlayerID); err != nil {
		log.Printf("Failed to start game: %v", err)
		cm.Client.SendError(err.Error())
		return
	}

	// Get room
	room, err := h.roomManager.GetRoom(cm.Client.RoomCode)
	if err != nil {
		cm.Client.SendError("Room not found")
		return
	}

	// Create game engine
	engine := game.NewEngine(room, h.dictionary)
	h.engines[room.Code] = engine

	// Start the game
	if err := engine.StartGame(); err != nil {
		log.Printf("Failed to start game engine: %v", err)
		cm.Client.SendError(err.Error())
		return
	}

	log.Printf("Game started in room %s, round %d", room.Code, room.GameState.CurrentRound)

	// Broadcast game state to all clients
	h.broadcastGameState(room)
	h.broadcastRoomState(room)
}

// handleSubmitDefinition processes SUBMIT_DEFINITION messages
func (h *Hub) handleSubmitDefinition(cm *ClientMessage) {
	if cm.Client.RoomCode == "" || cm.Client.PlayerID == "" {
		cm.Client.SendError("Not in a room")
		return
	}

	var payload SubmitDefinitionPayload
	if err := unmarshalPayload(cm.Message.Payload, &payload); err != nil {
		cm.Client.SendError("Invalid payload")
		return
	}

	log.Printf("Player %s submitted definition in room %s", cm.Client.PlayerID, cm.Client.RoomCode)

	// Get game engine
	engine, ok := h.engines[cm.Client.RoomCode]
	if !ok || engine == nil {
		cm.Client.SendError("Game not started")
		return
	}

	// Submit definition
	if err := engine.SubmitDefinition(cm.Client.PlayerID, payload.Text); err != nil {
		log.Printf("Failed to submit definition: %v", err)
		cm.Client.SendError(err.Error())
		return
	}

	// Get room
	room, _ := h.roomManager.GetRoom(cm.Client.RoomCode)

	// Broadcast updated game state
	h.broadcastGameState(room)

	// If moved to voting phase, send voting options
	if room.GameState.Phase == game.PhaseVoting {
		h.broadcastVotingOptions(room, engine)
	}
}

// handleSubmitVote processes SUBMIT_VOTE messages
func (h *Hub) handleSubmitVote(cm *ClientMessage) {
	if cm.Client.RoomCode == "" || cm.Client.PlayerID == "" {
		cm.Client.SendError("Not in a room")
		return
	}

	var payload SubmitVotePayload
	if err := unmarshalPayload(cm.Message.Payload, &payload); err != nil {
		cm.Client.SendError("Invalid payload")
		return
	}

	log.Printf("Player %s voted for definition %s in room %s",
		cm.Client.PlayerID, payload.DefinitionID, cm.Client.RoomCode)

	// Get game engine
	engine, ok := h.engines[cm.Client.RoomCode]
	if !ok || engine == nil {
		cm.Client.SendError("Game not started")
		return
	}

	// Submit vote
	if err := engine.SubmitVote(cm.Client.PlayerID, payload.DefinitionID); err != nil{
		log.Printf("Failed to submit vote: %v", err)
		cm.Client.SendError(err.Error())
		return
	}

	// Get room
	room, _ := h.roomManager.GetRoom(cm.Client.RoomCode)

	// Broadcast updated game state
	h.broadcastGameState(room)

	// If moved to scoring phase, send round results
	if room.GameState.Phase == game.PhaseScoring {
		h.broadcastRoundResults(room, engine)
	}
}

// handleEndGame processes END_GAME messages
func (h *Hub) handleEndGame(cm *ClientMessage) {
	if cm.Client.RoomCode == "" || cm.Client.PlayerID == "" {
		cm.Client.SendError("Not in a room")
		return
	}

	log.Printf("Client %s requesting to end game in room %s", cm.Client.ID, cm.Client.RoomCode)

	// Get game engine
	engine, ok := h.engines[cm.Client.RoomCode]
	if !ok || engine == nil {
		cm.Client.SendError("Game not started")
		return
	}

	// End the game
	if err := engine.EndGame(); err != nil {
		cm.Client.SendError(err.Error())
		return
	}

	// Get room
	room, _ := h.roomManager.GetRoom(cm.Client.RoomCode)

	// Broadcast final state
	h.broadcastGameState(room)
	h.broadcastRoundResults(room, engine)
}

// Broadcast helper methods

func (h *Hub) broadcastRoomState(room *game.Room) {
	payload := map[string]interface{}{
		"roomCode":   room.Code,
		"players":    room.Players,
		"isLocked":   room.IsLocked,
		"maxPlayers": room.MaxPlayers,
	}

	h.BroadcastToRoom(room.Code, MsgRoomState, payload)
}

func (h *Hub) broadcastGameState(room *game.Room) {
	if room.GameState == nil {
		return
	}

	gs := room.GameState
	payload := map[string]interface{}{
		"phase":        gs.Phase,
		"currentRound": gs.CurrentRound,
		"currentWord":  gs.CurrentWord.Text, // Don't send definition yet
		"wordPickerId": gs.WordPickerID,
		"scores":       gs.Scores,
		"phaseEndsAt":  gs.PhaseStartedAt.Add(gs.PhaseTimeout),
	}

	h.BroadcastToRoom(room.Code, MsgGameState, payload)
}

func (h *Hub) broadcastVotingOptions(room *game.Room, engine *game.Engine) {
	options := engine.GetVotingOptions()

	payload := map[string]interface{}{
		"definitions": options,
	}

	h.BroadcastToRoom(room.Code, MsgVotingOptions, payload)
}

func (h *Hub) broadcastRoundResults(room *game.Room, engine *game.Engine) {
	results := engine.GetRoundResults()
	h.BroadcastToRoom(room.Code, MsgRoundResults, results)
}

// unmarshalPayload is a helper to unmarshal message payloads
func unmarshalPayload(payload interface{}, target interface{}) error {
	// payload comes in as map[string]interface{} from JSON
	// Re-marshal and unmarshal to convert to target type
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}
