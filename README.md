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

Config is stored at `%APPDATA%\chat\config.json` on Windows or `~/.config/chat/config.json` on Linux/macOS. Every terminal on the same machine shares that file, so they share the same default username.

### Commands

| Command | Description |
|---------|-------------|
| `signup [username]` | Save your username locally |
| `create-room [name]` | Create a room (you become host) |
| `view-rooms` | List all rooms |
| `join-room [name]` | Join a room and chat over WebSocket |
| `leave-room [name]` | Leave a room without an active session |
| `delete-room [name]` | Delete a room you host |

**Global flags** (work on any command):

| Flag | Description |
|------|-------------|
| `--username [name]` | Override the saved username for this command |
| `--server [url]` | Override the server URL (default `http://localhost:8080`) |

Use `--username` on `create-room` and `join-room` when you want a specific identity without changing your saved signup, or when testing with **two terminals on one machine** (each tab needs a different username or messages from the other tab will be hidden).

### Examples

```bash
# Create and list rooms
go run . create-room general --username alice
go run . view-rooms

# Chat in a room (type messages, /leave or Ctrl+C to exit)
go run . join-room general --username alice
```

Your own messages show as `> hello`. Other users show as `bob: hello`.

**Two terminals on one machine:**

```bash
# Terminal 1
go run . create-room general --username alice
go run . join-room general --username alice

# Terminal 2
go run . join-room general --username bob
```

Without `--username`, both tabs use the same saved name from `signup`, so each tab filters out all messages as its own.

```bash
# Leave or delete without an active chat session
go run . leave-room general --username bob
go run . delete-room general --username alice
```

Build a binary:

```bash
cd client
go build -o chat .
./chat signup bob
./chat join-room general --username bob
```

When you run `join-room`, the CLI joins via the API, opens a WebSocket, and prints messages from other users. Exiting with `/leave` or Ctrl+C disconnects the WebSocket and calls the leave API automatically.

## API overview

| Method | Path | Purpose |
|--------|------|---------|
| `POST` | `/rooms` | Create room |
| `GET` | `/rooms` | List rooms |
| `PUT` | `/rooms/join` | Join room |
| `PUT` | `/rooms/leave` | Leave room |
| `DELETE` | `/rooms` | Delete room (host only) |
| `GET` | `/ws?room_name=&username=` | WebSocket chat |
