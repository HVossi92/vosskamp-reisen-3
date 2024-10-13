package services

import (
	"fmt"
	"time"
	"vosskamp-reisen-3/internal/database"
	"vosskamp-reisen-3/internal/models"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	db database.Service
}

func NewUserService(db database.Service) *UserService {
	return &UserService{db: db}
}

func (s *UserService) FetchAllUsers() (*[]models.Users, error) {
	var users []models.Users
	err := s.db.Db().Select(&users, "SELECT * FROM users")
	if err != nil {
		return nil, err
	}
	return &users, nil
}

func (s *UserService) FetchUserById(id int) (*models.Users, error) {
	var user models.Users
	err := s.db.Db().Get(&user, "SELECT * FROM users WHERE id = $1 LIMIT 1", id)
	if err != nil {
		return nil, err
	}
	user = *s.convertUserCreatedDate(&user)
	return &user, nil
}

func (s *UserService) FetchUserByEmail(email string) (*models.Users, error) {
	var user models.Users
	err := s.db.Db().Get(&user, "SELECT * FROM users WHERE email = $1 LIMIT 1", email)
	if err != nil {
		return nil, err
	}
	user = *s.convertUserCreatedDate(&user)
	return &user, nil
}

func (s *UserService) CreateUser(user models.Users) (*models.Users, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user.Password = string(passwordHash)

	query := `INSERT INTO users (first_name, last_name, email, password, avatar) 
              VALUES (:first_name, :last_name, :email, :password, "")`
	result, err := s.db.Db().NamedExec(query, user)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Scan the result into the user struct
	dbo, err := s.FetchUserById(int(id))
	if err != nil {
		return nil, err
	}

	return dbo, nil
}

func (s *UserService) UpdateUser(user models.Users) (*models.Users, error) {
	query := `UPDATE users 
	          SET first_name = :first_name, 
	              last_name = :last_name, 
	              email = :email 
	          WHERE id = :id 
	          RETURNING *`

	// Use NamedQuery to update the user and return the updated row(s)
	rows, err := s.db.Db().NamedQuery(query, user)
	if err != nil {
		return nil, err
	}
	defer rows.Close() // Ensure the rows are properly closed

	// Check if there's a result and scan it
	if rows.Next() {
		var updatedUser models.Users
		err := rows.StructScan(&updatedUser)
		if err != nil {
			return nil, fmt.Errorf("failed to scan updated user: %w", err)
		}
		return &updatedUser, nil
	}

	// If no rows were returned, this means the update likely failed
	return nil, fmt.Errorf("no user found with the provided ID")
}

func (s *UserService) DeleteUser(id int) error {
	params := map[string]interface{}{
		"id": id,
	}
	result, err := s.db.Db().NamedExec("DELETE FROM users WHERE id = :id", params)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user with ID %d not found", id)
	}
	return nil
}

func (s *UserService) UpdateUserAvatar(id int, filePath string) error {
	params := struct {
		FilePath string `db:"avatar"`
		ID       int    `db:"id"`
	}{
		FilePath: filePath,
		ID:       id,
	}

	// Use NamedExec instead of NamedQuery for update operations
	_, err := s.db.Db().NamedExec(`UPDATE users SET avatar = :avatar WHERE id = :id`, params)
	if err != nil {
		return err
	}
	return nil
}

func (s *UserService) convertUserCreatedDate(user *models.Users) *models.Users {
	created, err := time.Parse("2006-01-02 15:04", user.CreatedAt)
	if err != nil {
		created = time.Now()
	}
	user.CreatedAt = created.Format("02.01.2006")
	return user
}
