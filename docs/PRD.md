# Dictionary Game - Product Requirements Document

**Version:** 1.0 (MVP)
**Last Updated:** 2025-10-18
**Status:** Initial Release

---

## Executive Summary

A multiplayer word game where players submit fake definitions for obscure words and vote to identify the real definition. Designed for both remote and couch play sessions. The MVP focuses on private room-based gameplay with a clean architecture that supports future expansion to multiple game types, public matchmaking, and persistent user accounts.

---

## Product Vision

### Long-term Goals
- **Multi-game platform**: Support for various party games (Dictionary, Codenames, Spyfall, etc.)
- **Flexible play modes**: Remote, couch/table play, or hybrid sessions
- **Social features**: Persistent leaderboards, friend groups, achievement tracking
- **Rich communication**: Integrated voice/video comms for remote play
- **Cross-platform**: Works seamlessly on mobile, desktop, web, and native apps

### MVP Scope (v1.0)
Focus on delivering a polished, playable Dictionary game with private rooms. Build clean architectural boundaries that will support future expansion without requiring rewrites.

---

## Target Users

### Primary Personas
1. **Remote Friend Groups**: 3-8 friends who want to play party games together online
2. **Family Game Night**: Households looking for screen-based party games
3. **Party Hosts**: People who organize regular game sessions with mixed remote/local players

### User Requirements
- No account creation required (localStorage persistence)
- Works on phones, tablets, and computers
- Simple room joining via codes (no complex matchmaking)
- Can use external voice chat (Discord, Zoom) for MVP

---

## MVP Features (v1.0)

### ✅ In Scope

#### Core Gameplay
- **Dictionary game mechanics**
  - One player's turn per round (word picker rotates)
  - All players (including word picker) submit fake definitions
  - Vote on definitions (real definition always included)
  - Scoring: 1 point for correct vote, 1 point per person fooled
  - Players cannot vote for their own definition
- **Round-robin turn order** with clear indication of current word picker
- **Host can manually end game** at any time
- **Final scores displayed** when game ends

#### Room Management
- **Private rooms only** (room code join)
- **3-10 players per room** (configurable by host)
- **Rooms locked once game starts** (no mid-game joins)
- **Host controls**: Start game, end game
- **Room persists** until manually closed or all players disconnect

#### Dictionary Integration
- **Static word list** for testing (20-30 obscure words hardcoded)
- **Abstracted dictionary service** that can be swapped for real API
- **Free Dictionary API** (https://dictionaryapi.dev/) as target integration
- Words should be obscure enough that most players won't know them

#### Technical Implementation
- **Go backend** (net/http + gorilla/websocket)
- **PWA frontend** (installable on mobile + desktop)
- **Go html/template** for server-side rendering
- **Vanilla JavaScript** for client interactivity
- **HTMX** for dynamic HTML updates where appropriate
- **WebSocket** for real-time game state synchronisation
- **localStorage** for client-side persistence (nickname, room history)
- **Docker Compose** deployment (single VPS setup)

#### User Experience
- **No accounts required** (just enter nickname to play)
- **Room code sharing** (simple 6-character codes)
- **Responsive design** (mobile-first, works on all screen sizes)
- **Clear game state indication** (waiting, writing, voting, scoring phases)
- **Simple, clean UI** (focus on readability and quick actions)

### ❌ Out of Scope (Future Versions)

#### v2.0+ Features
- **Public matchmaking** (join random games with strangers)
- **User accounts + OAuth** (Google, Discord, email/password)
- **Persistent leaderboards** (global or friend-group based)
- **Voice/video comms** (WebRTC P2P or server-mediated)
- **Multiple game types** (Codenames, Spyfall, etc.)
- **Configurable rulesets** (custom scoring, time limits, etc.)
- **Native mobile apps** (Flutter or React Native)
- **Skill-based matching** for public games
- **Achievements/badges** system
- **Replay/game history** storage
- **Spectator mode** for ended games
- **Advanced host powers** (kick players, pause/resume, change settings mid-game)

---

## Game Rules (Dictionary - v1.0)

### Setup
1. Host creates private room, receives 6-character room code
2. Players join via room code, enter nickname
3. Host starts game when 3-10 players ready (minimum 3)

### Round Flow

#### Phase 1: Word Selection (Automatic)
- System picks random word from dictionary
- Current turn player is designated as "word picker"
- Word + real definition displayed only to backend (for voting later)
- Word (no definition) shown to all players

#### Phase 2: Definition Writing (Timed - 90 seconds suggested)
- **All players** (including word picker) write fake definitions
- Timer counts down for all players
- Players can submit early (marked as "ready")
- Phase ends when: all submitted OR timer expires
- Players who don't submit get no definition (cannot score that round)

#### Phase 3: Voting (Timed - 60 seconds suggested)
- All definitions shuffled randomly:
  - Real definition from dictionary
  - All submitted fake definitions
- **All players** (including word picker) vote for one definition
- Players cannot vote for their own definition (option disabled/hidden)
- Phase ends when: all voted OR timer expires
- Players who don't vote get no points

#### Phase 4: Scoring & Results (15 seconds display)
- Reveal which definition was real
- Highlight who wrote each fake definition
- Show who voted for what
- Award points:
  - +1 point for each player who voted for real definition
  - +1 point for each player who voted for your fake definition
  - Word picker scores normally (can earn points for fooling others)
- Display updated leaderboard

#### Phase 5: Next Round
- Turn passes to next player (round-robin)
- Return to Phase 1 with new word
- Repeat until host ends game

### Game End
- Host clicks "End Game" button
- Final scores displayed
- Option to "Play Again" (new game, same players)
- Option to "Leave Room"

### Edge Cases
- **Player disconnects during round**: Definition/vote doesn't count, round continues
- **Player disconnects between rounds**: Removed from game, turn order adjusts
- **Host disconnects**: Oldest remaining player becomes new host
- **All players disconnect**: Room closes after 5 minutes
- **Tie in final scores**: Both players shown as winners (no tiebreaker)

---

## Technical Architecture

### System Overview

```
┌─────────────────┐
│   PWA Frontend  │
│  (Go Templates  │
│   + Vanilla JS  │
│   + HTMX)       │
└────────┬────────┘
         │ HTTP + WebSocket
         ↓
┌─────────────────┐
│   Go Backend    │
│  ┌───────────┐  │
│  │ HTTP      │  │ (Static files, templates, API)
│  │ Server    │  │
│  └───────────┘  │
│  ┌───────────┐  │
│  │ WebSocket │  │ (Real-time game state)
│  │ Manager   │  │
│  └───────────┘  │
│  ┌───────────┐  │
│  │   Room    │  │ (Game state, player management)
│  │  Manager  │  │
│  └───────────┘  │
│  ┌───────────┐  │
│  │   Game    │  │ (Dictionary game logic)
│  │  Engine   │  │
│  └───────────┘  │
│  ┌───────────┐  │
│  │Dictionary │  │ (Word fetching - abstracted)
│  │  Service  │  │
│  └───────────┘  │
└─────────────────┘
         │
         ↓
┌─────────────────┐
│ Free Dictionary │
│      API        │
│ (or mock data)  │
└─────────────────┘
```

### Backend Architecture Principles

**Clear Separation of Concerns**:
1. **HTTP Server**: Routes, templates, static files
2. **WebSocket Manager**: Connection handling, message routing
3. **Room Manager**: Room lifecycle, player joins/leaves, state persistence
4. **Game Engine**: Dictionary game rules, scoring, turn management
5. **Dictionary Service**: Word fetching (swappable implementation)

**Why This Separation Matters**:
- Room Manager is game-agnostic (can support future games)
- Game Engine is isolated (easy to test, swap, or run multiple game types)
- Dictionary Service abstraction allows mock → real API with zero changes to game logic
- WebSocket Manager doesn't know about game rules (just passes messages)

### Data Models

#### Player
```go
type Player struct {
    ID          string    // UUID
    Nickname    string    // User-chosen display name
    RoomCode    string    // Current room
    ConnectedAt time.Time
    IsHost      bool
    IsActive    bool      // false if disconnected
}
```

#### Room
```go
type Room struct {
    Code        string    // 6-character code
    HostID      string    // Current host player ID
    Players     []Player
    GameState   GameState // Current game state
    MaxPlayers  int       // 3-10
    CreatedAt   time.Time
    IsLocked    bool      // true when game starts
}
```

#### GameState
```go
type GameState struct {
    Phase          Phase          // LOBBY, WRITING, VOTING, SCORING, ENDED
    CurrentRound   int
    CurrentWord    Word
    WordPickerID   string         // Current turn player
    TurnOrder      []string       // Player IDs in rotation
    Definitions    []Definition   // Submitted this round
    Votes          map[string]string // PlayerID → DefinitionID
    Scores         map[string]int    // PlayerID → total score
    PhaseStartedAt time.Time
    PhaseTimeout   time.Duration
}

type Phase string
const (
    LOBBY   Phase = "LOBBY"   // Waiting for players/host to start
    WRITING Phase = "WRITING" // Players writing definitions
    VOTING  Phase = "VOTING"  // Players voting
    SCORING Phase = "SCORING" // Results displayed
    ENDED   Phase = "ENDED"   // Game finished
)
```

#### Definition
```go
type Definition struct {
    ID       string // UUID
    Text     string
    AuthorID string // PlayerID (or "REAL" for dictionary definition)
    IsReal   bool
}
```

#### Word
```go
type Word struct {
    Text       string
    Definition string // Real definition from dictionary
    Source     string // Dictionary API source
}
```

### WebSocket Message Format

All messages use JSON over WebSocket:

```json
{
  "type": "MESSAGE_TYPE",
  "payload": { ... },
  "timestamp": "2025-10-18T12:00:00Z"
}
```

#### Client → Server Messages

**JOIN_ROOM**
```json
{
  "type": "JOIN_ROOM",
  "payload": {
    "roomCode": "ABC123",
    "nickname": "PlayerName"
  }
}
```

**START_GAME** (host only)
```json
{
  "type": "START_GAME",
  "payload": {}
}
```

**SUBMIT_DEFINITION**
```json
{
  "type": "SUBMIT_DEFINITION",
  "payload": {
    "text": "A small purple fruit native to Antarctica"
  }
}
```

**SUBMIT_VOTE**
```json
{
  "type": "SUBMIT_VOTE",
  "payload": {
    "definitionId": "def-uuid-here"
  }
}
```

**END_GAME** (host only)
```json
{
  "type": "END_GAME",
  "payload": {}
}
```

#### Server → Client Messages

**ROOM_STATE** (sent on join, and after any state change)
```json
{
  "type": "ROOM_STATE",
  "payload": {
    "roomCode": "ABC123",
    "players": [
      { "id": "player-uuid", "nickname": "Alice", "isHost": true, "isActive": true },
      { "id": "player-uuid-2", "nickname": "Bob", "isHost": false, "isActive": true }
    ],
    "isLocked": false,
    "maxPlayers": 10
  }
}
```

**GAME_STATE** (sent when game starts, phase changes, or state updates)
```json
{
  "type": "GAME_STATE",
  "payload": {
    "phase": "WRITING",
    "currentRound": 1,
    "currentWord": "scurryfunge",
    "wordPickerId": "player-uuid",
    "scores": {
      "player-uuid": 5,
      "player-uuid-2": 3
    },
    "phaseEndsAt": "2025-10-18T12:01:30Z"
  }
}
```

**VOTING_OPTIONS** (sent when voting phase starts)
```json
{
  "type": "VOTING_OPTIONS",
  "payload": {
    "definitions": [
      { "id": "def-uuid-1", "text": "A small purple fruit..." },
      { "id": "def-uuid-2", "text": "The act of tidying..." },
      { "id": "REAL", "text": "A hasty tidying of the house..." }
    ]
  }
}
```

**ROUND_RESULTS** (sent in scoring phase)
```json
{
  "type": "ROUND_RESULTS",
  "payload": {
    "realDefinitionId": "REAL",
    "definitions": [
      {
        "id": "def-uuid-1",
        "text": "A small purple fruit...",
        "authorId": "player-uuid",
        "authorNickname": "Alice",
        "votes": ["player-uuid-3", "player-uuid-4"],
        "pointsEarned": 2
      }
    ],
    "scores": {
      "player-uuid": 7,
      "player-uuid-2": 5
    }
  }
}
```

**ERROR**
```json
{
  "type": "ERROR",
  "payload": {
    "message": "Cannot vote for your own definition"
  }
}
```

### Frontend Architecture

**PWA Structure**:
- `manifest.json`: App metadata for installation
- `service-worker.js`: Offline caching (cache static assets only for MVP)
- `index.html`: Main entry point (Go template)
- `/static/js/game.js`: WebSocket handling, UI updates
- `/static/css/styles.css`: Responsive styling

**Go Templates**:
- Server renders initial HTML with room/game state
- HTMX used for partial updates (player list, scores)
- WebSocket messages trigger JS to update DOM directly for real-time changes

**State Management** (Client-side):
- JavaScript maintains minimal state (connected, currentPhase, playerId)
- Server is source of truth (all game state in backend)
- localStorage stores: nickname preference, recent room codes

### Deployment Architecture

**Docker Compose Setup**:
```yaml
services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DICT_API_URL=https://api.dictionaryapi.dev/api/v2/entries/en/
      - USE_MOCK_DICTIONARY=true  # Toggle for testing
    volumes:
      - ./data:/data  # Room persistence (optional)
    restart: unless-stopped

  # Future: Add Nginx reverse proxy, Let's Encrypt, etc.
```

**Single VPS Deployment**:
- Modern Docker Compose (not docker-compose hyphenated)
- All config in repo (reproducible deployments)
- Env vars for API keys, feature flags

---

## Dictionary Service Abstraction

### Interface Design

```go
type DictionaryService interface {
    GetRandomWord() (Word, error)
    GetWordDefinition(word string) (Word, error)
}

// Mock implementation for testing
type MockDictionary struct {
    words []Word
}

// Real API implementation
type FreeDictionaryAPI struct {
    baseURL string
    client  *http.Client
}
```

### Mock Word List (20-30 words)

Initial hardcoded words for testing:
- scurryfunge, borborygmus, nudiustertian, kakorrhaphiophobia
- floccinaucinihilipilification, hippopotomonstrosesquippedaliophobia
- pneumonoultramicroscopicsilicovolcanoconiosis, etc.

(Full list to be determined during implementation)

### API Integration Notes

**Free Dictionary API**:
- Endpoint: `https://api.dictionaryapi.dev/api/v2/entries/en/{word}`
- No API key required
- Rate limits: Unknown (implement request throttling)
- Fallback: If word not found, pick different word

---

## Future Expansion Considerations

### Voice Comms (v2.0+)

**Option A: WebRTC Peer-to-Peer**
- Pros: Low latency, no server bandwidth cost
- Cons: NAT traversal issues, needs STUN/TURN servers
- Docker setup: Add `coturn` container for TURN server
- Good for: Small groups (3-5 players)

**Option B: Server-Mediated (Janus/mediasoup)**
- Pros: More reliable, works behind restrictive NATs
- Cons: Higher server bandwidth, more complex setup
- Docker setup: Add media server container
- Good for: Larger groups (6-10 players)

**Recommendation**: Start with "use Discord/Zoom" guidance, add WebRTC P2P in v2.0 if user feedback demands it.

### Multiple Game Types (v3.0+)

**Abstraction Strategy** (don't implement yet, but keep in mind):

```go
type Game interface {
    Start(players []Player) error
    HandleAction(playerID string, action Action) error
    GetState() GameState
    IsPhaseComplete() bool
    AdvancePhase() error
}

// Dictionary implements Game
type DictionaryGame struct { ... }

// Future: Codenames implements Game
type CodenamesGame struct { ... }
```

Room Manager would be game-agnostic:
```go
type Room struct {
    // ...
    Game Game  // Interface, not concrete type
}
```

**Don't build this abstraction yet** - wait until we have a second game to inform the design. For now, just ensure Room Manager and Game Logic are cleanly separated.

### User Accounts (v2.0+)

**Auth Layer** (future):
- OAuth providers: Google, Discord, GitHub
- Link localStorage progress to account on first login
- Optional - players can still play anonymously

**Database Schema** (future):
```sql
-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE,
    display_name VARCHAR(100),
    created_at TIMESTAMP
);

-- OAuth connections
CREATE TABLE oauth_accounts (
    user_id UUID REFERENCES users(id),
    provider VARCHAR(50),
    provider_user_id VARCHAR(255),
    PRIMARY KEY (user_id, provider)
);

-- Game history
CREATE TABLE games (
    id UUID PRIMARY KEY,
    game_type VARCHAR(50),
    created_at TIMESTAMP,
    ended_at TIMESTAMP
);

-- Player game participation
CREATE TABLE game_players (
    game_id UUID REFERENCES games(id),
    user_id UUID REFERENCES users(id),
    final_score INT,
    PRIMARY KEY (game_id, user_id)
);
```

### Leaderboards (v2.0+)

**Private Leaderboards** (friend groups):
- User creates "league" and invites friends
- Track scores across multiple game sessions
- Weekly/monthly resets optional
- Requires accounts to persist

**Global Leaderboards**:
- Anonymous players can't participate (need account)
- Per game type (Dictionary, Codenames, etc.)
- Per ruleset variant
- Anti-cheat considerations (rate limiting, detection)

---

## Success Metrics (Post-MVP)

### Key Performance Indicators
- **Player Retention**: % of players who join 2nd+ game
- **Session Length**: Average time from join to leave
- **Room Fill Rate**: % of rooms that reach 4+ players
- **Completion Rate**: % of started games that finish (not abandoned)

### Technical Metrics
- **WebSocket Latency**: p95 < 200ms
- **Room Creation Time**: < 500ms
- **Message Delivery**: 99.9% success rate
- **Uptime**: 99% (single VPS acceptable for MVP)

---

## Open Questions / Decisions Needed

### Before Implementation Starts
- [x] Word picker scoring rules - **DECIDED: Word picker can score**
- [x] Real definition in voting pool - **DECIDED: Always included**
- [x] Point values - **DECIDED: 1 point each (simple)**
- [x] Can vote for own definition - **DECIDED: No**
- [x] Dictionary API choice - **DECIDED: Free Dictionary API**

### During Implementation
- [ ] Exact timer durations (writing: 90s, voting: 60s, scoring: 15s?)
- [ ] Room code format (6 chars alphanumeric? Exclude ambiguous letters?)
- [ ] Minimum word obscurity threshold (how to filter dictionary API results?)
- [ ] Disconnection grace period (how long to wait before removing player?)
- [ ] Host migration strategy (oldest player? Let remaining players vote?)

### Nice-to-Have Explorations
- [ ] Mobile keyboard optimisations (definition input UX)
- [ ] Accessibility features (screen reader support, high contrast mode)
- [ ] Internationalisation (i18n) architecture (even if English-only for MVP)

---

## Risks & Mitigations

### Technical Risks

**Risk: WebSocket connection drops**
- Mitigation: Auto-reconnect with exponential backoff
- Mitigation: Server buffers recent messages for reconnecting clients

**Risk: Dictionary API rate limits or downtime**
- Mitigation: Use mock dictionary by default for MVP
- Mitigation: Cache fetched words locally (future enhancement)

**Risk: Room state desync between clients**
- Mitigation: Server is source of truth, periodic full state broadcasts
- Mitigation: Client validates actions against local state before sending

**Risk: Single VPS becomes bottleneck**
- Mitigation: Profile and optimise before scaling
- Mitigation: Architecture supports horizontal scaling (future: Redis for state)

### Product Risks

**Risk: Players cheat by looking up words during game**
- Mitigation: Not solvable - trust-based game
- Mitigation: Shorter timers reduce cheating window (future enhancement)

**Risk: Players prefer existing party game apps (Jackbox, etc.)**
- Mitigation: Free and open-source (no paywall)
- Mitigation: Focus on customisation/flexibility (future rulesets)

**Risk: Low adoption without marketing**
- Mitigation: MVP is for friend groups who will invite others
- Mitigation: Simple room codes make sharing easy

---

## Timeline Estimate (MVP)

**Week 1-2: Backend Foundation**
- Go project setup, Docker Compose config
- WebSocket server + connection handling
- Room manager + player join/leave logic
- Mock dictionary service

**Week 3-4: Game Engine**
- Dictionary game state machine
- Phase transitions (lobby → writing → voting → scoring)
- Scoring logic + turn rotation
- WebSocket message handlers

**Week 5-6: Frontend**
- Go templates + static file serving
- PWA setup (manifest, service worker)
- WebSocket client + game UI
- Responsive design (mobile + desktop)

**Week 7-8: Polish & Testing**
- End-to-end testing (manual + automated)
- Bug fixes, edge case handling
- Deployment to VPS
- Documentation (setup guide, API docs)

**Total: ~8 weeks for 1-2 developers**

---

## Appendix: Technical Stack Details

### Backend Dependencies
- `net/http`: Standard library HTTP server
- `gorilla/websocket`: WebSocket implementation
- `google/uuid`: UUID generation
- Standard library: `html/template`, `encoding/json`, `time`

### Frontend Dependencies
- **HTMX** (v1.9+): Dynamic HTML updates
- **Vanilla JS** (ES6+): WebSocket client, UI logic
- No build step required (serve raw JS/CSS)

### Development Tools
- **Air**: Live reload for Go development
- **Docker Compose**: Local dev environment
- **Make**: Build automation (optional)

### Testing Strategy
- **Unit tests**: Game logic, scoring, state transitions
- **Integration tests**: WebSocket message handling, room lifecycle
- **Manual testing**: UI/UX, cross-browser, mobile devices
- **Load testing**: Future - see how many rooms one VPS can handle

---

## Changelog

### v1.0 (2025-10-18)
- Initial PRD created
- MVP scope defined
- Architecture documented
- Future features catalogued
