package game

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidPhase      = errors.New("invalid phase for this action")
	ErrPlayerNotInGame   = errors.New("player not in game")
	ErrAlreadySubmitted  = errors.New("already submitted for this round")
	ErrVoteForOwnDef     = errors.New("cannot vote for your own definition")
	ErrNotEnoughPlayers  = errors.New("not enough players to start game")
)

const (
	MinPlayers = 3

	// Phase timeouts
	WritingPhaseDuration = 90 * time.Second
	VotingPhaseDuration  = 60 * time.Second
	ScoringPhaseDuration = 15 * time.Second
)

// DictionaryService defines the interface for getting words
type DictionaryService interface {
	GetRandomWord() (Word, error)
}

// Engine manages the game logic and state transitions
type Engine struct {
	room       *Room
	dictionary DictionaryService
}

// NewEngine creates a new game engine for a room
func NewEngine(room *Room, dictionary DictionaryService) *Engine {
	return &Engine{
		room:       room,
		dictionary: dictionary,
	}
}

// StartGame initializes the game and moves to first round
func (e *Engine) StartGame() error {
	if len(e.room.Players) < MinPlayers {
		return ErrNotEnoughPlayers
	}

	// Already initialized by room manager, now start first round
	return e.StartNextRound()
}

// StartNextRound begins a new round with a new word
func (e *Engine) StartNextRound() error {
	gs := e.room.GameState

	// Get random word from dictionary
	word, err := e.dictionary.GetRandomWord()
	if err != nil {
		return fmt.Errorf("failed to get word: %w", err)
	}

	// Increment round
	gs.CurrentRound++

	// Set word picker (rotate through turn order)
	pickerIndex := (gs.CurrentRound - 1) % len(gs.TurnOrder)
	gs.WordPickerID = gs.TurnOrder[pickerIndex]

	// Set current word
	gs.CurrentWord = word

	// Clear round state
	gs.Definitions = []Definition{}
	gs.Votes = make(map[string]string)

	// Move to writing phase
	gs.Phase = PhaseWriting
	gs.PhaseStartedAt = time.Now()
	gs.PhaseTimeout = WritingPhaseDuration

	return nil
}

// SubmitDefinition handles a player submitting their fake definition
func (e *Engine) SubmitDefinition(playerID, text string) error {
	gs := e.room.GameState

	if gs.Phase != PhaseWriting {
		return ErrInvalidPhase
	}

	// Verify player is in game
	if !e.isPlayerInGame(playerID) {
		return ErrPlayerNotInGame
	}

	// Check if already submitted
	for _, def := range gs.Definitions {
		if def.AuthorID == playerID {
			return ErrAlreadySubmitted
		}
	}

	// Create definition
	def := Definition{
		ID:       uuid.New().String(),
		Text:     text,
		AuthorID: playerID,
		IsReal:   false,
	}

	gs.Definitions = append(gs.Definitions, def)

	// Check if all players have submitted
	if e.allPlayersSubmitted() {
		e.MoveToVoting()
	}

	return nil
}

// MoveToVoting transitions from writing to voting phase
func (e *Engine) MoveToVoting() error {
	gs := e.room.GameState

	if gs.Phase != PhaseWriting {
		return ErrInvalidPhase
	}

	// Add real definition to the pool
	realDef := Definition{
		ID:       "REAL",
		Text:     gs.CurrentWord.Definition,
		AuthorID: "REAL",
		IsReal:   true,
	}
	gs.Definitions = append(gs.Definitions, realDef)

	// Shuffle definitions
	e.shuffleDefinitions()

	// Move to voting phase
	gs.Phase = PhaseVoting
	gs.PhaseStartedAt = time.Now()
	gs.PhaseTimeout = VotingPhaseDuration

	return nil
}

// SubmitVote handles a player voting for a definition
func (e *Engine) SubmitVote(playerID, definitionID string) error {
	gs := e.room.GameState

	if gs.Phase != PhaseVoting {
		return ErrInvalidPhase
	}

	// Verify player is in game
	if !e.isPlayerInGame(playerID) {
		return ErrPlayerNotInGame
	}

	// Check if voting for own definition
	for _, def := range gs.Definitions {
		if def.ID == definitionID && def.AuthorID == playerID {
			return ErrVoteForOwnDef
		}
	}

	// Record vote
	gs.Votes[playerID] = definitionID

	// Check if all players have voted
	if e.allPlayersVoted() {
		return e.MoveToScoring()
	}

	return nil
}

// MoveToScoring calculates scores and transitions to scoring phase
func (e *Engine) MoveToScoring() error {
	gs := e.room.GameState

	if gs.Phase != PhaseVoting {
		return ErrInvalidPhase
	}

	// Calculate scores
	e.calculateScores()

	// Move to scoring phase
	gs.Phase = PhaseScoring
	gs.PhaseStartedAt = time.Now()
	gs.PhaseTimeout = ScoringPhaseDuration

	return nil
}

// calculateScores awards points based on votes
func (e *Engine) calculateScores() {
	gs := e.room.GameState

	// Count votes for each definition
	voteCounts := make(map[string]int)
	for _, defID := range gs.Votes {
		voteCounts[defID]++
	}

	// Award points
	for _, def := range gs.Definitions {
		voteCount := voteCounts[def.ID]

		if def.IsReal {
			// Players who voted for real definition get 1 point each
			for voterID, votedFor := range gs.Votes {
				if votedFor == def.ID {
					gs.Scores[voterID]++
				}
			}
		} else {
			// Author gets 1 point per person who voted for their fake definition
			gs.Scores[def.AuthorID] += voteCount
		}
	}
}

// EndGame ends the current game
func (e *Engine) EndGame() error {
	gs := e.room.GameState
	gs.Phase = PhaseEnded
	gs.PhaseStartedAt = time.Now()
	return nil
}

// CheckPhaseTimeout checks if the current phase has timed out
func (e *Engine) CheckPhaseTimeout() bool {
	gs := e.room.GameState

	if gs.Phase == PhaseLobby || gs.Phase == PhaseEnded {
		return false
	}

	elapsed := time.Since(gs.PhaseStartedAt)
	return elapsed >= gs.PhaseTimeout
}

// HandlePhaseTimeout handles automatic phase transitions on timeout
func (e *Engine) HandlePhaseTimeout() error {
	gs := e.room.GameState

	switch gs.Phase {
	case PhaseWriting:
		// Move to voting even if not everyone submitted
		return e.MoveToVoting()

	case PhaseVoting:
		// Move to scoring even if not everyone voted
		return e.MoveToScoring()

	case PhaseScoring:
		// Automatically start next round
		return e.StartNextRound()

	default:
		return nil
	}
}

// Helper methods

func (e *Engine) isPlayerInGame(playerID string) bool {
	for _, p := range e.room.Players {
		if p.ID == playerID && p.IsActive {
			return true
		}
	}
	return false
}

func (e *Engine) allPlayersSubmitted() bool {
	gs := e.room.GameState
	activePlayerCount := 0

	for _, p := range e.room.Players {
		if p.IsActive {
			activePlayerCount++
		}
	}

	return len(gs.Definitions) == activePlayerCount
}

func (e *Engine) allPlayersVoted() bool {
	gs := e.room.GameState
	activePlayerCount := 0

	for _, p := range e.room.Players {
		if p.IsActive {
			activePlayerCount++
		}
	}

	return len(gs.Votes) == activePlayerCount
}

func (e *Engine) shuffleDefinitions() {
	gs := e.room.GameState
	rand.Shuffle(len(gs.Definitions), func(i, j int) {
		gs.Definitions[i], gs.Definitions[j] = gs.Definitions[j], gs.Definitions[i]
	})
}

// GetVotingOptions returns definitions for voting (without author info)
func (e *Engine) GetVotingOptions() []Definition {
	gs := e.room.GameState

	// Return definitions without revealing authors
	options := make([]Definition, len(gs.Definitions))
	for i, def := range gs.Definitions {
		options[i] = Definition{
			ID:   def.ID,
			Text: def.Text,
			// Don't include AuthorID or IsReal
		}
	}

	return options
}

// GetRoundResults returns detailed results for the scoring phase
func (e *Engine) GetRoundResults() map[string]interface{} {
	gs := e.room.GameState

	// Count votes for each definition
	voteCounts := make(map[string][]string) // defID -> []voterIDs
	for voterID, defID := range gs.Votes {
		voteCounts[defID] = append(voteCounts[defID], voterID)
	}

	// Build results
	defResults := make([]map[string]interface{}, 0, len(gs.Definitions))
	for _, def := range gs.Definitions {
		voters := voteCounts[def.ID]
		pointsEarned := 0

		if def.IsReal {
			pointsEarned = len(voters) // Each correct voter got a point
		} else {
			pointsEarned = len(voters) // Author got points
		}

		defResults = append(defResults, map[string]interface{}{
			"id":           def.ID,
			"text":         def.Text,
			"authorId":     def.AuthorID,
			"isReal":       def.IsReal,
			"votes":        voters,
			"pointsEarned": pointsEarned,
		})
	}

	// Find real definition ID
	realDefID := ""
	for _, def := range gs.Definitions {
		if def.IsReal {
			realDefID = def.ID
			break
		}
	}

	return map[string]interface{}{
		"realDefinitionId": realDefID,
		"definitions":      defResults,
		"scores":           gs.Scores,
	}
}
