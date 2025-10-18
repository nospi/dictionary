package game

import "time"

// Phase represents the current phase of the game
type Phase string

const (
	PhaseLobby   Phase = "LOBBY"   // Waiting for players/host to start
	PhaseWriting Phase = "WRITING" // Players writing definitions
	PhaseVoting  Phase = "VOTING"  // Players voting
	PhaseScoring Phase = "SCORING" // Results displayed
	PhaseEnded   Phase = "ENDED"   // Game finished
)

// Player represents a player in the game
type Player struct {
	ID          string    `json:"id"`
	Nickname    string    `json:"nickname"`
	RoomCode    string    `json:"roomCode"`
	ConnectedAt time.Time `json:"connectedAt"`
	IsHost      bool      `json:"isHost"`
	IsActive    bool      `json:"isActive"` // false if disconnected
}

// Room represents a game room
type Room struct {
	Code       string     `json:"code"`
	HostID     string     `json:"hostId"`
	Players    []*Player  `json:"players"`
	GameState  *GameState `json:"gameState,omitempty"`
	MaxPlayers int        `json:"maxPlayers"`
	CreatedAt  time.Time  `json:"createdAt"`
	IsLocked   bool       `json:"isLocked"` // true when game starts
}

// GameState represents the current state of a game
type GameState struct {
	Phase          Phase                `json:"phase"`
	CurrentRound   int                  `json:"currentRound"`
	CurrentWord    Word                 `json:"currentWord"`
	WordPickerID   string               `json:"wordPickerId"`
	TurnOrder      []string             `json:"turnOrder"` // Player IDs in rotation
	Definitions    []Definition         `json:"definitions,omitempty"`
	Votes          map[string]string    `json:"votes,omitempty"`          // PlayerID → DefinitionID
	Scores         map[string]int       `json:"scores"`                   // PlayerID → total score
	PhaseStartedAt time.Time            `json:"phaseStartedAt"`
	PhaseTimeout   time.Duration        `json:"phaseTimeout"`
}

// Definition represents a player-submitted or real definition
type Definition struct {
	ID       string `json:"id"`
	Text     string `json:"text"`
	AuthorID string `json:"authorId"` // PlayerID (or "REAL" for dictionary definition)
	IsReal   bool   `json:"isReal"`
}

// Word represents a word from the dictionary
type Word struct {
	Text       string `json:"text"`
	Definition string `json:"definition"` // Real definition from dictionary
	Source     string `json:"source"`     // Dictionary API source
}
