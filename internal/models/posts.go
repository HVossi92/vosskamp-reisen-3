package models

import (
	"log"

	"github.com/jmoiron/sqlx"
)

type Posts struct {
	ID        int    `db:"id"`
	Title     string `db:"title"`
	Body      string `db:"body"`
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

func CreatePostTable(db *sqlx.DB) error {
	schema := `
    CREATE TABLE IF NOT EXISTS posts (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        title TEXT NOT NULL,
		body TEXT,
        created_at TEXT DEFAULT CURRENT_TIMESTAMP,
        updated_at TEXT DEFAULT CURRENT_TIMESTAMP
    ) strict;`

	_, err := db.Exec(schema)
	if err != nil {
		log.Printf("Error creating posts table: %v", err)
		return err
	}

	return nil
}
