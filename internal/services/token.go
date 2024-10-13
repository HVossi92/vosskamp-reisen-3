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

func (s *TokenService) ValidateToken(token string) (int, error) {
	// Query to check if the token exists and is still valid (not expired)
	query := `SELECT token, expires, user_id FROM tokens WHERE token = ?`

	// Store the result in a struct
	var tokenData struct {
		Token   string `db:"token"`
		Expires int64  `db:"expires"`
		UserId  int    `db:"user_id"`
	}

	// Execute the query with token as a regular parameter
	err := s.db.Db().Get(&tokenData, query, token)
	if err != nil {
		if err == sql.ErrNoRows {
			return -1, errors.New("invalid token") // Token not found
		}
		return -1, err // Other error during query
	}

	// Check if the token has expired
	if time.Now().Unix() > tokenData.Expires {
		deleteQuery := `DELETE FROM tokens WHERE token = ?`
		_, deleteErr := s.db.Db().Exec(deleteQuery, token)
		if deleteErr != nil {
			return -1, errors.New("token expired, but failed to delete from database") // Handle deletion error
		}
		return -1, errors.New("token has expired")
	}

	// If the token is valid and not expired, return nil (no error)
	return tokenData.UserId, nil
}

func (s *TokenService) InsertToken(token string, expires time.Time, userId int) error {
	deleteQuery := `DELETE FROM tokens WHERE user_id = :user_id`
	deleteParams := map[string]interface{}{
		"user_id": userId,
	}

	_, err := s.db.Db().NamedExec(deleteQuery, deleteParams)
	if err != nil {
		return err // Return error if deletion fails
	}

	query := `INSERT INTO tokens (token, expires, user_id) 
              VALUES (:token, :expires, :user_id)`

	params := map[string]interface{}{
		"token":   token,
		"expires": expires.Unix(), // Convert time.Time to Unix timestamp
		"user_id": userId,
	}

	_, err = s.db.Db().NamedExec(query, params)
	if err != nil {
		return err
	}

	return nil
}

func (s *TokenService) RemoveToken(userId int) error {
	deleteQuery := `DELETE FROM tokens WHERE user_id = :user_id`
	deleteParams := map[string]interface{}{
		"user_id": userId,
	}

	_, err := s.db.Db().NamedExec(deleteQuery, deleteParams)
	if err != nil {
		return err // Return error if deletion fails
	}

	return nil
}
