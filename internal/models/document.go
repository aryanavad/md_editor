package models

import "time"

// Document represents a markdown document in the system
type Document struct {
	ID        string    `json:"id"`        // Unique identifier
	Title     string    `json:"title"`     // Document title
	Content   string    `json:"content"`   // Markdown content
	CreatedAt time.Time `json:"createdAt"` // Creation timestamp
	UpdatedAt time.Time `json:"updatedAt"` // Last update timestamp
}
