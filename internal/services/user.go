package services

import (
	"fmt"
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
	fmt.Println(err)
	if err != nil {
		return nil, err
	}
	return &users, nil
}

func (s *UserService) FetchUserById(id int) (*models.Users, error) {
	var user models.Users
	err := s.db.Db().Get(&user, "SELECT * FROM users WHERE id = $1 LIMIT 1", id)
	fmt.Println(err)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserService) FetchUserByEmail(email string) (*models.Users, error) {
	var user models.Users
	err := s.db.Db().Get(&user, "SELECT * FROM users WHERE email = $1 LIMIT 1", email)
	fmt.Println(err)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserService) CreateUser(user models.Users) (*models.Users, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user.Password = string(passwordHash)

	fmt.Print(user)

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
	          SET username = :username, 
	              first_name = :first_name, 
	              last_name = :last_name, 
	              email = :email 
	          WHERE id = :id 
	          RETURNING *`

	// Use NamedQueryRow to update the user and return the updated row
	row, err := s.db.Db().NamedQuery(query, user)
	if err != nil {
		return nil, err
	}

	var dbo models.Users
	err = row.StructScan(&dbo)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return &user, nil
}

func (s *UserService) DeleteUser(id int) (*models.Users, error) {
	var user models.Users
	rows, err := s.db.Db().NamedQuery("DELETE FROM users WHERE id = :id RETURNING *", id)
	if err != nil {
		return nil, err
	}
	var dbo models.Users
	err = rows.StructScan(&dbo)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	return &user, nil
}

func (s *UserService) UpdateUserAvatar(id int, filePath string) error {
	params := struct {
		FilePath string `db:"filePath"`
		ID       int    `db:"id"`
	}{
		FilePath: filePath,
		ID:       id,
	}
	_, err := s.db.Db().NamedQuery(`UPDATE users SET avatar = :filePath WHERE id = :id`, params)
	if err != nil {
		return err
	}
	return nil
}
