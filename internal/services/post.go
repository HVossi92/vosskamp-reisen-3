package services

import (
	"fmt"
	"vosskamp-reisen-3/internal/database"
	"vosskamp-reisen-3/internal/helpers"
	"vosskamp-reisen-3/internal/models"
)

type PostService struct {
	db database.Service
}

func NewPostService(db database.Service) *PostService {
	return &PostService{db: db}
}

func (s *PostService) FetchPaginatedPosts(page int, limit int) (*[]models.Posts, int, error) {
	var posts []models.Posts
	var totalPosts int

	// Get total count for pagination metadata
	err := s.db.Db().Get(&totalPosts, "SELECT COUNT(*) FROM posts")
	if err != nil {
		return nil, 0, err
	}

	// Calculate offset
	offset := (page - 1) * limit

	// Fetch paginated posts
	err = s.db.Db().Select(&posts, `
        SELECT * FROM posts 
        ORDER BY created_at DESC 
        LIMIT $1 OFFSET $2
    `, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	for i := range posts {
		post := &posts[i]
		post.CreatedAt = helpers.GetDayMonthYearFrom(post.CreatedAt)
		post.UpdatedAt = helpers.GetDayMonthYearFrom(post.UpdatedAt)
		if len(post.Body) > 32 {
			// post.Body = post.Body[:256] + "..."
		}
	}

	return &posts, totalPosts, nil
}

func (s *PostService) FetchPostById(id int) (*models.Posts, error) {
	var post models.Posts
	err := s.db.Db().Get(&post, "SELECT * FROM posts WHERE id = $1 LIMIT 1", id)
	if err != nil {
		return nil, err
	}
	post.CreatedAt = helpers.GetDayMonthYearFrom(post.CreatedAt)
	post.UpdatedAt = helpers.GetDayMonthYearFrom(post.UpdatedAt)
	return &post, nil
}

func (s *PostService) CreatePost(post models.Posts) (*models.Posts, error) {
	query := `INSERT INTO posts (title, body, picture) 
              VALUES (:title, :body, :picture)`
	result, err := s.db.Db().NamedExec(query, post)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Scan the result into the post struct
	dbo, err := s.FetchPostById(int(id))
	if err != nil {
		return nil, err
	}

	return dbo, nil
}

func (s *PostService) UpdatePost(post models.Posts) (*models.Posts, error) {
	query := `UPDATE posts 
	          SET title = :title, 
	              body = :body, 
	              picture = :picture 
	          WHERE id = :id 
	          RETURNING *`

	// Use NamedQuery to update the post and return the updated row(s)
	rows, err := s.db.Db().NamedQuery(query, post)
	if err != nil {
		return nil, err
	}
	defer rows.Close() // Ensure the rows are properly closed

	// Check if there's a result and scan it
	if rows.Next() {
		var updatedPost models.Posts
		err := rows.StructScan(&updatedPost)
		if err != nil {
			return nil, fmt.Errorf("failed to scan updated post: %w", err)
		}
		return &updatedPost, nil
	}

	// If no rows were returned, this means the update likely failed
	return nil, fmt.Errorf("no post found with the provided ID")
}

func (s *PostService) DeletePost(id int) error {
	params := map[string]interface{}{
		"id": id,
	}
	result, err := s.db.Db().NamedExec("DELETE FROM posts WHERE id = :id", params)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("post with ID %d not found", id)
	}
	return nil
}

func (s *PostService) UpdatePostAvatar(id int, filePath string) error {
	params := struct {
		FilePath string `db:"avatar"`
		ID       int    `db:"id"`
	}{
		FilePath: filePath,
		ID:       id,
	}

	// Use NamedExec instead of NamedQuery for update operations
	_, err := s.db.Db().NamedExec(`UPDATE posts SET avatar = :avatar WHERE id = :id`, params)
	if err != nil {
		return err
	}
	return nil
}
