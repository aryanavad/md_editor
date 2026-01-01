package web

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// writeWait is the time allowed to write a message to the peer
	writeWait = 10 * time.Second
	// pongWait is the time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second
	// pingPeriod is the interval for sending ping messages (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10
	// maxMessageSize is the maximum message size allowed from peer (512KB)
	maxMessageSize = 512 * 1024
)

// Client represents a websocket connection for a single user editing a document
type Client struct {
	// hub is the central message hub for broadcasting
	hub *Hub
	// conn is the websocket connection
	conn *websocket.Conn
	// send is a buffered channel for outbound messages
	send chan []byte
	// documentId identifies which document this client is editing
	documentId string
	// userId is the unique identifier for the user
	userId string
	// userName is the display name of the user
	userName string
}

// readPump reads messages from the websocket connection and broadcasts them to other clients
// It runs in its own goroutine and ensures only one reader per connection
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		// Broadcast the received message to all other clients in the document
		c.hub.Broadcast(c.documentId, message, c)
	}
}

// writePump writes messages from the send channel to the websocket connection
// It runs in its own goroutine and ensures only one writer per connection
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			// Send periodic ping to keep connection alive
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
