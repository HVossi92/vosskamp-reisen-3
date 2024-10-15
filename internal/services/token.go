package services

import (
	"database/sql"
	"errors"
	"time"
	"vosskamp-reisen-3/internal/database"
)

type TokenService struct {
	db database.Service
}

func NewTokenService(db database.Service) *TokenService {
	return &TokenService{db: db}
}

func (s *TokenService) ValidateToken(token string) error {
	// Query to check if the token exists and is still valid (not expired)
	query := `SELECT token, expires FROM tokens WHERE token = ?`

	// Store the result in a struct
	var tokenData struct {
		Token   string `db:"token"`
		Expires int64  `db:"expires"`
	}

	// Execute the query with token as a regular parameter
	err := s.db.Db().Get(&tokenData, query, token)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("invalid token") // Token not found
		}
		return err // Other error during query
	}

	// Check if the token has expired
	if time.Now().Unix() > tokenData.Expires {
		return errors.New("token has expired")
	}

	// If the token is valid and not expired, return nil (no error)
	return nil
}

func (s *TokenService) InsertToken(token string, expires time.Time, userId int) error {
	query := `INSERT INTO tokens (token, expires, user_id) 
              VALUES (:token, :expires, :user_id)`

	params := map[string]interface{}{
		"token":   token,
		"expires": expires.Unix(), // Convert time.Time to Unix timestamp
		"user_id": userId,
	}

	_, err := s.db.Db().NamedExec(query, params)
	if err != nil {
		return err
	}

	return nil
}
