package web

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/aryanavad/md_editor/internal/db"
	"github.com/aryanavad/md_editor/internal/models"
	"github.com/google/uuid"
)

// HandleDocuments manages document creation operations
// POST: Creates a new document and returns its ID
func HandleDocuments(database *db.Database, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	doc := &models.Document{
		ID:      uuid.New().String(),
		Title:   "Untitled",
		Content: "",
	}

	if err := database.CreateDocument(doc); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(doc)
}

// HandleDocument manages individual document operations
// GET: Retrieves a document by ID
// PUT: Updates a document's content
func HandleDocument(database *db.Database, w http.ResponseWriter, r *http.Request) {
	// Extract document ID from URL path
	id := strings.TrimPrefix(r.URL.Path, "/documents/")

	switch r.Method {
	case http.MethodGet:
		doc, err := database.GetDocument(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if doc == nil {
			http.Error(w, "Document not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(doc)

	case http.MethodPut:
		var req struct {
			Content string `json:"content"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := database.UpdateDocument(id, req.Content); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
