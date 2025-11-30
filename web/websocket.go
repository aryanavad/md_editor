package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	documentId := r.URL.Query().Get("documentId")
	userId := r.URL.Query().Get("userId")
	userName := r.URL.Query().Get("userName")

	if documentId == "" {
		conn.Close()
		return
	}

	client := &Client{
		hub:        hub,
		conn:       conn,
		send:       make(chan []byte, 256),
		documentId: documentId,
		userId:     userId,
		userName:   userName,
	}

	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}
