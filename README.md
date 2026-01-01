# Collaborative Markdown Editor

A real-time collaborative markdown editor where multiple users can edit the same document simultaneously with instant synchronization.

## Features

- Real-time collaboration via WebSockets
- Live markdown preview
- Auto-save after 2 seconds of inactivity
- Document persistence in PostgreSQL
- Share documents via URL

## Quick Start

```bash
# Start PostgreSQL
docker-compose up -d

# Run the server
go run cmd/server/main.go

# Open browser
http://localhost:8080
```

## Architecture

```
md_editor/
├── cmd/server/          # Application entry point
├── internal/
│   ├── db/              # PostgreSQL integration
│   └── models/          # Data models
└── web/                 # HTTP/WebSocket handlers
    ├── hub.go           # Connection coordinator
    ├── client.go        # WebSocket client
    ├── handlers.go      # REST API
    └── static/          # Frontend (HTML/CSS/JS)
```

### How It Works

**Hub-Client Pattern:**
- Hub manages all WebSocket connections and document "rooms"
- Each client gets two goroutines: one for reading, one for writing
- Changes broadcast to all clients in the same document room
- Channels ensure thread-safe communication

**Data Flow:**
1. User types in editor
2. Change broadcasts via WebSocket to all connected clients
3. Auto-save triggers after 2 seconds of inactivity
4. Content persists to PostgreSQL

## Design Decisions

**Go + Goroutines:** Handles thousands of concurrent WebSocket connections efficiently with minimal overhead.

**WebSockets over HTTP polling:** Persistent bidirectional connection provides low-latency real-time updates without constant polling overhead.

**PostgreSQL:** ACID compliance ensures data integrity. Simple schema stores documents by UUID.

**Hub-Client Pattern:** Centralized hub coordinates all connections, broadcasts messages to document rooms, and cleanly separates connection management from business logic.

**Debounced Auto-save:** 2-second delay prevents excessive database writes while ensuring changes aren't lost.

## API

**REST Endpoints:**
- `POST /documents` - Create new document
- `GET /documents/{id}` - Get document by ID
- `PUT /documents/{id}` - Update document content

**WebSocket:**
- `WS /ws?documentId={id}&userId={userId}&userName={name}`
- Message format: `{"type": "content", "content": "...", "userId": "...", "userName": "..."}`
