package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"blog-api/internal/models"
)

// CreatePost creates a new post in the database
func (db *DB) CreatePost(ctx context.Context, req *models.PostRequest) (*models.Post, error) {
	query := `
		INSERT INTO posts (title, content, user_id, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, title, content, user_id, created_at`

	var post models.Post
	err := db.QueryRowContext(ctx, query, req.Title, req.Content, req.UserID, time.Now()).Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.UserID,
		&post.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	return &post, nil
}

// GetAllPosts retrieves all posts from the database with user information
func (db *DB) GetAllPosts(ctx context.Context) ([]models.Post, error) {
	query := `
		SELECT p.id, p.title, p.content, p.user_id, p.created_at, u.username
		FROM posts p
		JOIN users u ON p.user_id = u.id
		ORDER BY p.created_at DESC`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query posts: %w", err)
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.UserID,
			&post.CreatedAt,
			&post.Username,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}
		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return posts, nil
}

// GetPostByID retrieves a post by its ID with user information
func (db *DB) GetPostByID(ctx context.Context, id int) (*models.Post, error) {
	query := `
		SELECT p.id, p.title, p.content, p.user_id, p.created_at, u.username
		FROM posts p
		JOIN users u ON p.user_id = u.id
		WHERE p.id = $1`

	var post models.Post
	err := db.QueryRowContext(ctx, query, id).Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.UserID,
		&post.CreatedAt,
		&post.Username,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("post not found")
		}
		return nil, fmt.Errorf("failed to get post: %w", err)
	}

	return &post, nil
}

// UpdatePost updates an existing post
func (db *DB) UpdatePost(ctx context.Context, id int, req *models.PostRequest) (*models.Post, error) {
	// Start building the query dynamically based on what fields are provided
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.Title != "" {
		setParts = append(setParts, fmt.Sprintf("title = $%d", argIndex))
		args = append(args, req.Title)
		argIndex++
	}

	if req.Content != "" {
		setParts = append(setParts, fmt.Sprintf("content = $%d", argIndex))
		args = append(args, req.Content)
		argIndex++
	}

	if req.UserID != 0 {
		setParts = append(setParts, fmt.Sprintf("user_id = $%d", argIndex))
		args = append(args, req.UserID)
		argIndex++
	}

	if len(setParts) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}

	// Add the post ID as the last argument
	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE posts 
		SET %s 
		WHERE id = $%d
		RETURNING id, title, content, user_id, created_at`,
		joinStrings(setParts, ", "),
		argIndex,
	)

	var post models.Post
	err := db.QueryRowContext(ctx, query, args...).Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.UserID,
		&post.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("post not found")
		}
		return nil, fmt.Errorf("failed to update post: %w", err)
	}

	return &post, nil
}

// DeletePost deletes a post by its ID
func (db *DB) DeletePost(ctx context.Context, id int) error {
	query := `DELETE FROM posts WHERE id = $1`

	result, err := db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("post not found")
	}

	return nil
}

// GetPostsByUserID retrieves all posts by a specific user
func (db *DB) GetPostsByUserID(ctx context.Context, userID int) ([]models.Post, error) {
	query := `
		SELECT p.id, p.title, p.content, p.user_id, p.created_at, u.username
		FROM posts p
		JOIN users u ON p.user_id = u.id
		WHERE p.user_id = $1
		ORDER BY p.created_at DESC`

	rows, err := db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query posts by user: %w", err)
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.UserID,
			&post.CreatedAt,
			&post.Username,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}
		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return posts, nil
}
