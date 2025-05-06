package database

import (
	"context"
	"fmt"
	"log"
	"time"
)

// This file provides usage examples for the database package

// Example shows how to use the database package
func Example() {
	// Create a database with default configuration but with a file path
	cfg := DefaultConfig()
	cfg.Path = "example.db"

	// Open connection to the database
	db, err := Open(cfg)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create a context with timeout
	ctx, cancel := WithContext(context.Background(), 5*time.Second)
	defer cancel()

	// Create a table with JSON and FTS5 support
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS documents (
		id INTEGER PRIMARY KEY,
		title TEXT NOT NULL,
		content TEXT,
		metadata JSON
	);

	-- Create FTS5 virtual table for full-text search
	CREATE VIRTUAL TABLE IF NOT EXISTS documents_fts USING fts5(
		title, content, content='documents', content_rowid='id'
	);

	-- Create trigger to update FTS when documents table is updated
	CREATE TRIGGER IF NOT EXISTS documents_ai AFTER INSERT ON documents BEGIN
		INSERT INTO documents_fts(rowid, title, content) VALUES (new.id, new.title, new.content);
	END;
	`

	// Execute the SQL using a transaction
	tx, err := db.BeginTx(ctx)
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
	}

	if _, err := tx.Exec(createTableSQL); err != nil {
		tx.Rollback()
		log.Fatalf("Failed to create tables: %v", err)
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("Failed to commit transaction: %v", err)
	}

	// Insert a document with JSON metadata
	insertSQL := `
	INSERT INTO documents (title, content, metadata) 
	VALUES (?, ?, json(?))
	`
	
	metadata := `{"author": "John Doe", "tags": ["example", "sqlite", "golang"]}`
	
	result, err := db.ExecContext(ctx, insertSQL, 
		"Example Document", 
		"This is an example document for full-text search and JSON capabilities.", 
		metadata)
	if err != nil {
		log.Fatalf("Failed to insert document: %v", err)
	}
	
	id, _ := result.LastInsertId()
	fmt.Printf("Inserted document with ID: %d\n", id)
	
	// Query using FTS
	rows, err := db.QueryContext(ctx, 
		"SELECT id, title FROM documents_fts WHERE documents_fts MATCH ?", 
		"example")
	if err != nil {
		log.Fatalf("Failed to query: %v", err)
	}
	defer rows.Close()
	
	fmt.Println("Search results:")
	for rows.Next() {
		var id int64
		var title string
		if err := rows.Scan(&id, &title); err != nil {
			log.Fatalf("Failed to scan row: %v", err)
		}
		fmt.Printf("- %d: %s\n", id, title)
	}
	
	// Query using JSON
	rows, err = db.QueryContext(ctx, 
		"SELECT id, title FROM documents WHERE json_extract(metadata, '$.author') = ?",
		"John Doe")
	if err != nil {
		log.Fatalf("Failed to query with JSON: %v", err)
	}
	defer rows.Close()
	
	fmt.Println("JSON query results:")
	for rows.Next() {
		var id int64
		var title string
		if err := rows.Scan(&id, &title); err != nil {
			log.Fatalf("Failed to scan row: %v", err)
		}
		fmt.Printf("- %d: %s\n", id, title)
	}
}
