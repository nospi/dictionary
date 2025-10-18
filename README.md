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

- **Backend**: Go (net/http + gorilla/websocket)
- **Frontend**: Go templates + Vanilla JS + HTMX
- **Deployment**: Docker Compose
- **Dictionary**: Free Dictionary API (with mock fallback)

## Quick Start

*Coming soon - project under active development*

```bash
# Clone the repo
git clone https://github.com/nospi/dictionary.git
cd dictionary

# Run with Docker Compose
docker compose up

# Visit http://localhost:8080
```

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

**Current Phase**: Initial planning & documentation

See [PRD](docs/PRD.md) for:
- Complete game rules and mechanics
- Technical architecture details
- WebSocket message formats
- Future features roadmap

## Roadmap

### v1.0 (MVP) - In Progress
- ✅ PRD completed
- ⬜ Backend implementation
- ⬜ Frontend implementation
- ⬜ Docker deployment setup
- ⬜ Testing & polish

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
