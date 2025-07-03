package handlers

import (
	"context"
	"net/http"
	"time"

	"blog-api/internal/database"
	"blog-api/internal/models"

	"github.com/rs/zerolog/log"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	db *database.DB
}

// NewUserHandler creates a new user handler
func NewUserHandler(db *database.DB) *UserHandler {
	return &UserHandler{db: db}
}

// CreateUser handles POST /users
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req models.UserRequest
	if err := parseJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	// Validate the request
	if err := ValidateUserRequest(&req); err != nil {
		writeValidationError(w, err)
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Create the user
	user, err := h.db.CreateUser(ctx, &req)
	if err != nil {
		handleDatabaseError(w, err, "create user")
		return
	}

	log.Info().Int("user_id", user.ID).Str("username", user.Username).Msg("User created successfully")
	writeJSON(w, http.StatusCreated, user)
}

// GetAllUsers handles GET /users
func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	users, err := h.db.GetAllUsers(ctx)
	if err != nil {
		handleDatabaseError(w, err, "get all users")
		return
	}

	writeJSON(w, http.StatusOK, users)
}

// GetUser handles GET /users/{id}
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDFromURL(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	user, err := h.db.GetUserByID(ctx, id)
	if err != nil {
		handleDatabaseError(w, err, "get user")
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// UpdateUser handles PUT /users/{id}
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDFromURL(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req models.UserRequest
	if err := parseJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	// Validate the request (for updates, fields are optional)
	if err := ValidateUserUpdateRequest(&req); err != nil {
		writeValidationError(w, err)
		return
	}

	// Check if at least one field is provided for update
	if req.Username == "" && req.Email == "" && req.Password == "" {
		writeError(w, http.StatusBadRequest, "At least one field must be provided for update")
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	user, err := h.db.UpdateUser(ctx, id, &req)
	if err != nil {
		handleDatabaseError(w, err, "update user")
		return
	}

	log.Info().Int("user_id", user.ID).Str("username", user.Username).Msg("User updated successfully")
	writeJSON(w, http.StatusOK, user)
}

// DeleteUser handles DELETE /users/{id}
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDFromURL(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	err = h.db.DeleteUser(ctx, id)
	if err != nil {
		handleDatabaseError(w, err, "delete user")
		return
	}

	log.Info().Int("user_id", id).Msg("User deleted successfully")
	writeSuccess(w, "User deleted successfully", nil)
}
