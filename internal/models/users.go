package models

import (
	"log"

	"github.com/jmoiron/sqlx"
)

type Users struct {
	ID        int    `db:"id"`
	FirstName string `db:"first_name"`
	LastName  string `db:"last_name"`
	Email     string `db:"email"`
	Password  string `db:"password"`
	Avatar    string `db:"avatar"`
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

func CreateUsersTable(db *sqlx.DB) error {
	schema := `
		CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		first_name TEXT,
		last_name TEXT,
		email TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		avatar TEXT,
		created_at TEXT DEFAULT CURRENT_TIMESTAMP,
		updated_at TEXT DEFAULT CURRENT_TIMESTAMP
    ) strict;`

	_, err := db.Exec(schema)
	if err != nil {
		log.Printf("Error creating users table: %v", err)
		return err
	}

	return nil
}
