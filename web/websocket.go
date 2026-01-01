package web

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// upgrader configures the websocket connection upgrade parameters
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// CheckOrigin allows all connections (configure appropriately for production)
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// ServeWs handles websocket connection requests from clients
// It upgrades the HTTP connection to a websocket and registers the client with the hub
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to websocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// Extract connection parameters from query string
	documentId := r.URL.Query().Get("documentId")
	userId := r.URL.Query().Get("userId")
	userName := r.URL.Query().Get("userName")

	// Require documentId to establish connection
	if documentId == "" {
		conn.Close()
		return
	}

	// Create new client instance
	client := &Client{
		hub:        hub,
		conn:       conn,
		send:       make(chan []byte, 256),
		documentId: documentId,
		userId:     userId,
		userName:   userName,
	}

	// Register client with the hub
	client.hub.register <- client

	// Start client goroutines for reading and writing
	go client.writePump()
	go client.readPump()
}
