package handlers

import (
	"context"
	"net/http"
	"time"

	"blog-api/internal/database"
	"blog-api/internal/models"

	"github.com/rs/zerolog/log"
)

// PostHandler handles post-related HTTP requests
type PostHandler struct {
	db *database.DB
}

// NewPostHandler creates a new post handler
func NewPostHandler(db *database.DB) *PostHandler {
	return &PostHandler{db: db}
}

// CreatePost handles POST /posts
func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	var req models.PostRequest
	if err := parseJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	// Validate the request
	if err := ValidatePostRequest(&req); err != nil {
		writeValidationError(w, err)
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Verify that the user exists before creating the post
	_, err := h.db.GetUserByID(ctx, req.UserID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid user ID: user does not exist")
		return
	}

	// Create the post
	post, err := h.db.CreatePost(ctx, &req)
	if err != nil {
		handleDatabaseError(w, err, "create post")
		return
	}

	log.Info().Int("post_id", post.ID).Str("title", post.Title).Int("user_id", post.UserID).Msg("Post created successfully")
	writeJSON(w, http.StatusCreated, post)
}

// GetAllPosts handles GET /posts
func (h *PostHandler) GetAllPosts(w http.ResponseWriter, r *http.Request) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	posts, err := h.db.GetAllPosts(ctx)
	if err != nil {
		handleDatabaseError(w, err, "get all posts")
		return
	}

	writeJSON(w, http.StatusOK, posts)
}

// GetPost handles GET /posts/{id}
func (h *PostHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDFromURL(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	post, err := h.db.GetPostByID(ctx, id)
	if err != nil {
		handleDatabaseError(w, err, "get post")
		return
	}

	writeJSON(w, http.StatusOK, post)
}

// UpdatePost handles PUT /posts/{id}
func (h *PostHandler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDFromURL(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	var req models.PostRequest
	if err := parseJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	// Validate the request (for updates, fields are optional)
	if err := ValidatePostUpdateRequest(&req); err != nil {
		writeValidationError(w, err)
		return
	}

	// Check if at least one field is provided for update
	if req.Title == "" && req.Content == "" && req.UserID == 0 {
		writeError(w, http.StatusBadRequest, "At least one field must be provided for update")
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// If user_id is provided, verify that the user exists
	if req.UserID != 0 {
		_, err := h.db.GetUserByID(ctx, req.UserID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid user ID: user does not exist")
			return
		}
	}

	post, err := h.db.UpdatePost(ctx, id, &req)
	if err != nil {
		handleDatabaseError(w, err, "update post")
		return
	}

	log.Info().Int("post_id", post.ID).Str("title", post.Title).Msg("Post updated successfully")
	writeJSON(w, http.StatusOK, post)
}

// DeletePost handles DELETE /posts/{id}
func (h *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDFromURL(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	err = h.db.DeletePost(ctx, id)
	if err != nil {
		handleDatabaseError(w, err, "delete post")
		return
	}

	log.Info().Int("post_id", id).Msg("Post deleted successfully")
	writeSuccess(w, "Post deleted successfully", nil)
}
