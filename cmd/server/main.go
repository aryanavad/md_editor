package main

import (
	"log"
	"net/http"

	"github.com/aryanavad/md_editor/internal/db"
	"github.com/aryanavad/md_editor/web"
)

func main() {
	// Connect to PostgreSQL database
	database, err := db.New("localhost", "5432", "markdown", "markdown", "markdown")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	// Initialize and start the websocket hub for real-time collaboration
	hub := web.NewHub()
	go hub.Run()

	// Serve static files (HTML, CSS, JS)
	http.Handle("/", http.FileServer(http.Dir("./web/static")))

	// WebSocket endpoint for real-time document editing
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		web.ServeWs(hub, w, r)
	})

	// API endpoint for creating new documents
	http.HandleFunc("/documents", func(w http.ResponseWriter, r *http.Request) {
		web.HandleDocuments(database, w, r)
	})

	// API endpoint for getting/updating specific documents
	http.HandleFunc("/documents/", func(w http.ResponseWriter, r *http.Request) {
		web.HandleDocument(database, w, r)
	})

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
