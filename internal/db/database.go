package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// Database wraps a PostgreSQL database connection
type Database struct {
	conn *sql.DB
}

// New creates a new database connection with the provided credentials
func New(host, port, user, password, dbname string) (*Database, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Verify connection is established
	if err := conn.Ping(); err != nil {
		return nil, err
	}

	return &Database{conn: conn}, nil
}

// Close closes the database connection
func (db *Database) Close() error {
	return db.conn.Close()
}
