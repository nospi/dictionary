# Architecture Documentation

## Design Philosophy

This codebase prioritises clean separation of concerns, functional composition, and avoiding premature abstraction. Key principles:

1. **Separation of concerns**: Room lifecycle management is independent of game logic
2. **Composition over inheritance**: No wrapper classes, use interfaces sparingly
3. **Functional patterns**: Pure functions, immutable data where possible
4. **Pragmatic abstraction**: Only abstract when you have 2+ concrete implementations

## System Overview

```
┌─────────────────────────────────────────────────────────┐
│                      HTTP Server                        │
│  (cmd/server/main.go - net/http stdlib)               │
└───────────────────┬─────────────────────────────────────┘
                    │
        ┌───────────┴──────────┐
        │                      │
        ▼                      ▼
┌──────────────┐      ┌──────────────┐
│   Static     │      │  WebSocket   │
│   Files      │      │   Handler    │
└──────────────┘      └──────┬───────┘
                             │
                             ▼
                    ┌────────────────┐
                    │   Hub          │
                    │  (ws/hub.go)   │
                    └────┬───────────┘
                         │
         ┌───────────────┼───────────────┐
         │               │               │
         ▼               ▼               ▼
┌─────────────┐  ┌──────────────┐  ┌──────────────┐
│ Room        │  │ Game Engine  │  │ Dictionary   │
│ Manager     │  │ (game/       │  │ Service      │
│ (room/      │  │  engine.go)  │  │ (dictionary/)│
│  manager.go)│  │              │  │              │
└─────────────┘  └──────────────┘  └──────────────┘
```

## Core Components

### 1. Room Manager (`internal/room/manager.go`)

**Responsibility**: Room lifecycle only
- Create rooms with unique codes
- Player join/leave
- Host migration on disconnect
- Room cleanup when empty

**Key Methods**:
```go
CreateRoom(hostNickname string, maxPlayers int) (*Room, error)
JoinRoom(roomCode, nickname string) (*Room, *Player, error)
PlayerDisconnected(roomCode, playerID string) error
StartGame(roomCode, playerID string) error // Just initialises GameState
```

**Does NOT**:
- Handle game logic or scoring
- Know about definitions, votes, or phases
- Manage dictionary words

### 2. Game Engine (`internal/game/engine.go`)

**Responsibility**: Game logic only
- State machine: LOBBY → WRITING → VOTING → SCORING → ENDED
- Phase transitions (manual and automatic on timeout)
- Scoring calculation
- Validation rules

**Key Methods**:
```go
StartGame() error                                    // Begin first round
SubmitDefinition(playerID, text string) error      // Collect definitions
SubmitVote(playerID, definitionID string) error    // Collect votes
MoveToVoting() error                                // Transition phases
MoveToScoring() error                               // Calculate scores
```

**Does NOT**:
- Manage rooms or players
- Handle WebSocket communication
- Fetch words (uses injected DictionaryService)

### 3. WebSocket Hub (`internal/ws/hub.go`)

**Responsibility**: Connection management and message routing
- Maintain active client connections
- Register/unregister clients
- Route messages to appropriate handlers
- Broadcast to rooms

**Composition** (not inheritance):
```go
type Hub struct {
    clients     map[*Client]bool
    roomManager RoomManager         // Interface - room operations
    dictionary  DictionaryService    // Interface - word fetching
    engines     map[string]*game.Engine // One engine per active game
}
```

**Why this design?**
- Hub doesn't need to know room internals, just calls interface methods
- Game engines are created on-demand when games start
- Clean dependency injection - easy to test and swap implementations

### 4. Message Handlers (`internal/ws/handlers.go`)

**Responsibility**: WebSocket message processing
- Parse and validate incoming messages
- Coordinate between Room Manager and Game Engine
- Broadcast state changes to clients

**Flow Example** (START_GAME):
```go
1. Validate client is in a room and is host
2. Call roomManager.StartGame() → initialises GameState
3. Create new game.Engine(room, dictionary)
4. Store engine in hub.engines[roomCode]
5. Call engine.StartGame() → begins first round
6. Broadcast updated game state to all clients in room
```

**Key Pattern**: Handlers orchestrate, they don't contain logic
- Room Manager handles room state
- Game Engine handles game state
- Handlers just connect them and broadcast results

## Design Decisions

### Why No GameManager Wrapper?

**Initial attempt** (rejected):
```go
type GameManager struct {
    manager    *Manager
    dictionary DictionaryService
    engines    map[string]*Engine
}

// Wrapper methods
func (gm *GameManager) CreateRoom(...) { return gm.manager.CreateRoom(...) }
func (gm *GameManager) JoinRoom(...) { return gm.manager.JoinRoom(...) }
```

**Problems**:
- Duplicates all room methods (violates DRY)
- Unclear responsibility - is it for rooms or games?
- Leaky abstraction - exposes room details
- Makes testing harder (more mocking needed)

**Better approach** (current):
- Hub composes RoomManager + Dictionary directly
- Hub creates Engine instances when games start
- Clear boundaries: rooms vs games vs connections

### Why Interfaces Are Minimal?

We only define interfaces where:
1. **Multiple implementations exist or planned**: `DictionaryService` (mock vs real API)
2. **Dependency inversion needed**: `RoomManager` interface lets Hub not depend on concrete room package

We DON'T create interfaces for:
- Game Engine (only one implementation)
- Client (concrete type, no need to swap)
- Models (structs, not behaviour)

### Functional Patterns Used

**Pure helper functions**:
```go
func (e *Engine) allPlayersSubmitted() bool {
    // Read-only check, no side effects
}

func (e *Engine) shuffleDefinitions() {
    // Mutates in place, but localised
}
```

**Composition**:
```go
// Hub composes behaviour, doesn't inherit
hub := NewHub(roomManager, dictionary)
```

**Explicit error handling**:
```go
// No exceptions, explicit error returns
if err := engine.SubmitVote(playerID, defID); err != nil {
    client.SendError(err.Error())
    return
}
```

## WebSocket Protocol

### Message Format
All messages use JSON over WebSocket:
```json
{
  "type": "MESSAGE_TYPE",
  "payload": { ... },
  "timestamp": "2025-10-18T12:00:00Z"
}
```

### Client → Server Messages
- `JOIN_ROOM`: Create or join a room
- `START_GAME`: Begin the game (host only)
- `SUBMIT_DEFINITION`: Submit fake definition
- `SUBMIT_VOTE`: Vote for a definition
- `END_GAME`: End the game (host only)

### Server → Client Messages
- `ROOM_STATE`: Room info (players, locked status)
- `GAME_STATE`: Current phase, round, word, scores
- `VOTING_OPTIONS`: Shuffled definitions for voting
- `ROUND_RESULTS`: Scores and correct answer reveal
- `ERROR`: Error message

### State Synchronisation
- Server is source of truth
- Broadcast after every state change
- Clients can reconnect and receive latest state

## Phase Transitions

### Automatic Transitions
When all players complete an action:
```
WRITING phase:
  All submit definition → auto-advance to VOTING

VOTING phase:
  All submit vote → auto-advance to SCORING
```

### Timeout Transitions
Each phase has a timeout:
```go
WritingPhaseDuration = 90 seconds
VotingPhaseDuration  = 60 seconds
ScoringPhaseDuration = 15 seconds
```

If timeout expires:
- Missing submissions/votes are ignored
- Game advances anyway
- Players who didn't participate get 0 points

### Manual Transitions
Host can end game at any time:
```
Any phase → END_GAME → ENDED phase
```

## Scoring Algorithm

```go
func calculateScores() {
    for each definition:
        if definition.IsReal:
            // Award 1 point to each player who voted for it
            for each voter:
                scores[voterID]++
        else:
            // Award definition author 1 point per person fooled
            voteCount = number of votes for this definition
            scores[definition.AuthorID] += voteCount
}
```

**Example**:
- Real definition: "A large knife" (3 players voted) → those 3 players +1 point each
- Alice's fake: "A type of bird" (2 players fooled) → Alice +2 points
- Bob's fake: "A dance move" (0 players fooled) → Bob +0 points

## Adding New Game Types (Future)

### Current Approach (Pragmatic)
- Room and Game are separate
- But no generic "Game" interface yet
- Why? We only have 1 game type, abstraction would be premature

### When Adding Game #2
1. **Extract common interface**:
```go
type GameEngine interface {
    Start() error
    HandleAction(playerID, actionType string, data interface{}) error
    GetState() interface{}
    IsComplete() bool
}
```

2. **Room Manager doesn't change** - it just manages rooms

3. **Hub creates appropriate engine**:
```go
func (h *Hub) startGame(roomCode, gameType string) {
    switch gameType {
    case "dictionary":
        engine = game.NewDictionaryEngine(room, h.dictionary)
    case "codenames":
        engine = game.NewCodenamesEngine(room)
    }
    h.engines[roomCode] = engine
}
```

4. **Keep game-specific logic in packages**:
```
internal/
├── game/
│   ├── dictionary/    # Dictionary-specific
│   ├── codenames/     # Codenames-specific
│   └── common.go      # Shared interfaces
```

## Testing Strategy

### Unit Tests
- Game Engine: Phase transitions, scoring, validation
- Room Manager: Join/leave, host migration
- Helper functions: Pure functions easy to test

### Integration Tests
- WebSocket message flow
- Room lifecycle with multiple clients
- Game flow from start to end

### Manual Testing
- Browser-based WebSocket client
- Multiple tabs simulating players
- Edge cases: disconnects, timeouts

## Performance Considerations

### Concurrency
- WebSocket Hub runs in single goroutine (channel-based)
- Each client has 2 goroutines: ReadPump, WritePump
- Room Manager uses mutex for thread-safe access

### Memory
- Rooms cleaned up when all players leave
- Game engines deleted when rooms close
- Client send buffers: 256 messages (prevents slow clients blocking)

### Scalability (Future)
Current design supports:
- ~100 concurrent rooms on single VPS (estimate)
- Limited by WebSocket connections, not CPU

For larger scale:
- Add Redis for room state (horizontal scaling)
- Separate WebSocket servers from game logic
- Load balance with sticky sessions

## Technology Choices

### Why coder/websocket (nhooyr)?
- **Recommended by Go authors**
- Context-aware (proper cancellation)
- Minimal, idiomatic API
- Battle-tested (used by Traefik, Vault, Cloudflare)
- Better than gorilla/websocket (no longer actively maintained)

### Why stdlib over frameworks?
- `net/http`: Simple, performant, well-documented
- `html/template`: Secure by default (XSS protection)
- `encoding/json`: Fast enough for our use case
- Fewer dependencies = easier maintenance

### Why Go 1.23?
- Latest stable release
- Improved iterator support
- Better performance
- Modern tooling (go.mod, go.sum)

## Common Patterns

### Error Handling
```go
// Always explicit
if err != nil {
    log.Printf("Context about error: %v", err)
    return err // Or handle and continue
}
```

### Logging
```go
// Structured, actionable
log.Printf("Client %s joined room %s as player %s", clientID, roomCode, playerID)
```

### Validation
```go
// Early returns, clear error messages
if gs.Phase != PhaseVoting {
    return ErrInvalidPhase
}
```

## Future Improvements

### Short Term
- Add phase timeout goroutines
- Better error types (custom errors with codes)
- Reconnection handling (preserve player state)

### Medium Term
- Frontend implementation
- Automated tests
- Performance profiling

### Long Term
- Redis for state persistence
- Multiple game types
- Voice comms integration

---

**Last Updated**: 2025-10-18
**Phase**: Backend complete (Phases 1 & 2)
