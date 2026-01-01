package web

import (
	"encoding/json"
	"log"
)

// Hub manages active document rooms and client connections
// It handles registration, unregistration, and message broadcasting
type Hub struct {
	// documents maps document IDs to their active document rooms
	documents map[string]*Document
	// register channel for new client connections
	register chan *Client
	// unregister channel for disconnecting clients
	unregister chan *Client
}

// Document represents a collaborative document session
// It tracks all clients currently editing the same document
type Document struct {
	// id is the unique document identifier
	id string
	// clients maps client connections to their active status
	clients map[*Client]bool
}

// NewHub creates and initializes a new Hub instance
func NewHub() *Hub {
	return &Hub{
		documents:  make(map[string]*Document),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the hub's main event loop
// It processes client registration, unregistration, and manages document rooms
func (h *Hub) Run() {
	for {
		select {
		// Handle new client registration
		case client := <-h.register:
			doc, exists := h.documents[client.documentId]
			if !exists {
				// Create new document room if it doesn't exist
				doc = &Document{
					id:      client.documentId,
					clients: make(map[*Client]bool),
				}
				h.documents[client.documentId] = doc
			}
			// Add client to the document room
			doc.clients[client] = true

			// Broadcast updated user list to all clients in the document
			h.broadcastUserList(client.documentId)

		// Handle client disconnection
		case client := <-h.unregister:
			if doc, exists := h.documents[client.documentId]; exists {
				if _, ok := doc.clients[client]; ok {
					// Remove client from document room
					delete(doc.clients, client)
					close(client.send)

					// Broadcast updated user list to remaining clients
					h.broadcastUserList(client.documentId)

					// Clean up empty document rooms
					if len(doc.clients) == 0 {
						delete(h.documents, client.documentId)
					}
				}
			}
		}
	}
}

// Broadcast sends a message to all clients in a document room except the sender
// This enables real-time collaboration by distributing changes to all participants
func (h *Hub) Broadcast(documentId string, message []byte, sender *Client) {
	if doc, exists := h.documents[documentId]; exists {
		for client := range doc.clients {
			// Skip sending the message back to the sender
			if client != sender {
				select {
				case client.send <- message:
				default:
					// Client's send channel is full or closed, clean up
					close(client.send)
					delete(doc.clients, client)
				}
			}
		}
	}
}

// User represents a connected user's information
type User struct {
	UserId   string `json:"userId"`
	UserName string `json:"userName"`
}

// UserListMessage represents the message sent to clients with active users
type UserListMessage struct {
	Type  string `json:"type"`
	Users []User `json:"users"`
}

// broadcastUserList sends the current list of active users to all clients in a document
func (h *Hub) broadcastUserList(documentId string) {
	doc, exists := h.documents[documentId]
	if !exists {
		return
	}

	// Build list of active users
	users := make([]User, 0, len(doc.clients))
	for client := range doc.clients {
		users = append(users, User{
			UserId:   client.userId,
			UserName: client.userName,
		})
	}

	// Create user list message
	message := UserListMessage{
		Type:  "userList",
		Users: users,
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling user list: %v", err)
		return
	}

	// Send to all clients in the document
	for client := range doc.clients {
		select {
		case client.send <- messageBytes:
		default:
			// Client's send channel is full or closed, clean up
			close(client.send)
			delete(doc.clients, client)
		}
	}
}
