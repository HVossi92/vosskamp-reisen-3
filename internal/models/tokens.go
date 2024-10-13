package models

import (
	"log"
	"time"

	"github.com/jmoiron/sqlx"
)

type Tokens struct {
	ID      int       `db:"id"`
	Expires time.Time `db:"expires"`
	UserId  int       `db:"user_id"`
	Token   string    `db:"token"`
}

func CreateTokenTable(db *sqlx.DB) error {
	schema := `
    CREATE TABLE IF NOT EXISTS tokens (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        expires INTEGER NOT NULL,
        user_id INTEGER NOT NULL,
        token TEXT NOT NULL UNIQUE,
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
    ) STRICT;`

	_, err := db.Exec(schema)
	if err != nil {
		log.Printf("Error creating tokens table: %v", err)
		return err
	}

	return nil
}
