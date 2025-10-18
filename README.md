# Dictionary Game

A multiplayer word game where players submit fake definitions for obscure words and vote to identify the real definition. Designed for both remote and couch play.

## Overview

Dictionary Game is a web-based party game that works on mobile and desktop browsers. Players join private rooms, take turns picking words, write believable fake definitions, and score points by fooling others or guessing correctly.

**Status**: MVP in development - see [PRD](docs/PRD.md) for full specification.

## Features (MVP v1.0)

- **Private room-based multiplayer** (3-10 players)
- **No accounts required** - just enter a nickname
- **Real-time gameplay** via WebSockets
- **Progressive Web App** - install on mobile or desktop
- **Simple scoring**: 1 point for correct guess, 1 point per person fooled
- **Clean architecture** - ready for future game types and features

## Tech Stack

- **Backend**: Go 1.23 (stdlib net/http + coder/websocket)
- **Frontend**: Go templates + Vanilla JS + HTMX
- **Deployment**: Docker Compose
- **Dictionary**: Free Dictionary API (with mock fallback for testing)

## Quick Start

### Local Development (Recommended)

```bash
# Clone the repo
git clone https://github.com/nospi/dictionary.git
cd dictionary

# Install dependencies
go mod download

# Run the server
make run

# Visit http://127.0.0.1:8080
```

### Docker Compose

```bash
# Build and run
docker compose up --build

# Visit http://localhost:8080
```

**Note**: If using a reverse proxy like Traefik, you may need to clear browser cache or use `http://127.0.0.1:8080` instead of `localhost`.

## Project Structure

```
dictionary/
├── docs/
│   └── PRD.md              # Complete product requirements
├── cmd/
│   └── server/             # Main application entry point
├── internal/
│   ├── game/               # Dictionary game logic
│   ├── room/               # Room management
│   ├── ws/                 # WebSocket handling
│   └── dictionary/         # Dictionary service abstraction
├── web/
│   ├── templates/          # Go HTML templates
│   └── static/             # JS, CSS, assets
├── docker-compose.yml      # Deployment configuration
└── README.md
```

## How to Play

1. **Create a room** - Host generates a 6-character room code
2. **Invite friends** - Share the code (3-10 players needed)
3. **Start game** - Host begins when ready

**Each round**:
- One player's turn (word picker rotates)
- Everyone sees an obscure word
- All players write fake definitions (90 seconds)
- Vote for the definition you think is real (60 seconds)
- Earn points for guessing right or fooling others
- Repeat until host ends game

## Development Status

**Current Phase**: Backend complete (Phases 1 & 2), frontend next

**Completed**:
- ✅ Project structure and build system
- ✅ Core data models (Player, Room, GameState, Definition, Word)
- ✅ Room manager with join/leave, host migration
- ✅ WebSocket infrastructure (coder/websocket with context support)
- ✅ Game engine with full state machine (5 phases)
- ✅ Phase transitions and scoring logic
- ✅ Mock dictionary service (30 obscure words)
- ✅ WebSocket message handlers for all game actions

**Architecture Highlights**:
- Clean separation: Room lifecycle vs Game logic
- Functional composition over inheritance
- Interfaces only where needed (no over-abstraction)
- Ready for multiple game types without refactoring

See [PRD](docs/PRD.md) for complete specification and [ARCHITECTURE.md](docs/ARCHITECTURE.md) for design decisions.

## Roadmap

### v1.0 (MVP) - In Progress
- ✅ PRD completed
- ✅ Backend implementation (Phases 1 & 2)
- ⬜ Frontend implementation (Phase 3)
- ⬜ End-to-end testing
- ⬜ Production deployment

### v2.0 (Future)
- Public matchmaking
- User accounts (OAuth)
- Voice/video comms
- Persistent leaderboards

### v3.0+ (Future)
- Multiple game types (Codenames, Spyfall, etc.)
- Configurable rulesets
- Native mobile apps

## Contributing

*Guidelines coming soon*

## License

MIT License - see [LICENSE.md](LICENSE.md)

## Acknowledgements

- Dictionary data from [Free Dictionary API](https://dictionaryapi.dev/)
