-- Database initialization script for the collaborative markdown editor
-- This script creates the necessary tables and indexes

-- Create documents table
CREATE TABLE IF NOT EXISTS documents (
    id VARCHAR(36) PRIMARY KEY,
    title VARCHAR(255) NOT NULL DEFAULT 'Untitled',
    content TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create index on created_at for efficient sorting
CREATE INDEX IF NOT EXISTS idx_documents_created_at ON documents(created_at);

-- Create index on updated_at for efficient queries
CREATE INDEX IF NOT EXISTS idx_documents_updated_at ON documents(updated_at);
