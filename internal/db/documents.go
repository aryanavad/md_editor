package db

import (
	"database/sql"

	"github.com/aryanavad/md_editor/internal/models"
)

// CreateDocument inserts a new document into the database
func (db *Database) CreateDocument(doc *models.Document) error {
	query := `
		INSERT INTO documents (id, title, content, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
	`
	_, err := db.conn.Exec(query, doc.ID, doc.Title, doc.Content)
	return err
}

// GetDocument retrieves a document by ID from the database
// Returns nil if the document is not found
func (db *Database) GetDocument(id string) (*models.Document, error) {
	query := `
		SELECT id, title, content, created_at, updated_at
		FROM documents
		WHERE id = $1
	`
	doc := &models.Document{}
	err := db.conn.QueryRow(query, id).Scan(
		&doc.ID,
		&doc.Title,
		&doc.Content,
		&doc.CreatedAt,
		&doc.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return doc, nil
}

// UpdateDocument updates a document's content and timestamp
func (db *Database) UpdateDocument(id, content string) error {
	query := `
		UPDATE documents
		SET content = $1, updated_at = NOW()
		WHERE id = $2
	`
	_, err := db.conn.Exec(query, content, id)
	return err
}
