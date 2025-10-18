package ws

import "github.com/nospi/dictionary/internal/game"

// Type aliases to avoid import cycles and keep ws package clean
type (
	Room       = game.Room
	Player     = game.Player
	GameState  = game.GameState
	Definition = game.Definition
	Word       = game.Word
	Phase      = game.Phase
)
