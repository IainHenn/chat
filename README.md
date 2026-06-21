# chat

A simple chat app with a Go backend and CLI client.

## Prerequisites

- Go 1.25+
- Docker (for Redis)

## Backend

Start Redis:

```bash
docker compose up -d
```

Run the API server from `backend/`:

```bash
cd backend
go run .
```

The server listens on `http://localhost:8080`.

## CLI

Install or run the client from `client/`:

```bash
cd client
go run . --help
```

### Setup

Pick a username (saved locally on your machine):

```bash
go run . signup alice
```

Optional: point at a different server:

```bash
go run . signup alice --server http://localhost:8080
```

Config is stored at `%APPDATA%\chat\config.json` on Windows or `~/.config/chat/config.json` on Linux/macOS.

### Commands

| Command | Description |
|---------|-------------|
| `signup [username]` | Save your username locally |
| `create-room [name]` | Create a room (you become host) |
| `view-rooms` | List all rooms |
| `join-room [name]` | Join a room and chat over WebSocket |
| `leave-room [name]` | Leave a room without an active session |
| `delete-room [name]` | Delete a room you host |

### Examples

```bash
# Create and list rooms
go run . create-room general
go run . view-rooms

# Chat in a room (type messages, /leave or Ctrl+C to exit)
go run . join-room general

# Leave or delete without an active chat session
go run . leave-room general
go run . delete-room general
```

When you run `join-room`, the CLI joins via the API, opens a WebSocket, and prints messages from other users. Exiting with `/leave` or Ctrl+C disconnects the WebSocket and calls the leave API automatically.

Build a binary:

```bash
cd client
go build -o chat .
./chat signup bob
./chat join-room general
```

## API overview

| Method | Path | Purpose |
|--------|------|---------|
| `POST` | `/rooms` | Create room |
| `GET` | `/rooms` | List rooms |
| `PUT` | `/rooms/join` | Join room |
| `PUT` | `/rooms/leave` | Leave room |
| `DELETE` | `/rooms` | Delete room (host only) |
| `GET` | `/ws?room_name=&username=` | WebSocket chat |
