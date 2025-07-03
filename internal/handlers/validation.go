package handlers

import (
	"fmt"
	"net/mail"
	"strings"

	"blog-api/internal/models"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors represents multiple validation errors
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

func (ve ValidationErrors) Error() string {
	var messages []string
	for _, err := range ve.Errors {
		messages = append(messages, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	return strings.Join(messages, ", ")
}

// ValidateUserRequest validates a user request
func ValidateUserRequest(req *models.UserRequest) error {
	var errors []ValidationError

	// Validate username
	if req.Username == "" {
		errors = append(errors, ValidationError{
			Field:   "username",
			Message: "username is required",
		})
	} else if len(req.Username) < 3 {
		errors = append(errors, ValidationError{
			Field:   "username",
			Message: "username must be at least 3 characters long",
		})
	} else if len(req.Username) > 50 {
		errors = append(errors, ValidationError{
			Field:   "username",
			Message: "username must be no more than 50 characters long",
		})
	}

	// Validate email
	if req.Email == "" {
		errors = append(errors, ValidationError{
			Field:   "email",
			Message: "email is required",
		})
	} else if !isValidEmail(req.Email) {
		errors = append(errors, ValidationError{
			Field:   "email",
			Message: "email format is invalid",
		})
	}

	// Validate password
	if req.Password == "" {
		errors = append(errors, ValidationError{
			Field:   "password",
			Message: "password is required",
		})
	} else if len(req.Password) < 6 {
		errors = append(errors, ValidationError{
			Field:   "password",
			Message: "password must be at least 6 characters long",
		})
	}

	if len(errors) > 0 {
		return ValidationErrors{Errors: errors}
	}

	return nil
}

// ValidateUserUpdateRequest validates a user update request
func ValidateUserUpdateRequest(req *models.UserRequest) error {
	var errors []ValidationError

	// For updates, fields are optional, but if provided, they must be valid
	if req.Username != "" {
		if len(req.Username) < 3 {
			errors = append(errors, ValidationError{
				Field:   "username",
				Message: "username must be at least 3 characters long",
			})
		} else if len(req.Username) > 50 {
			errors = append(errors, ValidationError{
				Field:   "username",
				Message: "username must be no more than 50 characters long",
			})
		}
	}

	if req.Email != "" && !isValidEmail(req.Email) {
		errors = append(errors, ValidationError{
			Field:   "email",
			Message: "email format is invalid",
		})
	}

	if req.Password != "" && len(req.Password) < 6 {
		errors = append(errors, ValidationError{
			Field:   "password",
			Message: "password must be at least 6 characters long",
		})
	}

	if len(errors) > 0 {
		return ValidationErrors{Errors: errors}
	}

	return nil
}

// ValidatePostRequest validates a post request
func ValidatePostRequest(req *models.PostRequest) error {
	var errors []ValidationError

	// Validate title
	if req.Title == "" {
		errors = append(errors, ValidationError{
			Field:   "title",
			Message: "title is required",
		})
	} else if len(req.Title) > 255 {
		errors = append(errors, ValidationError{
			Field:   "title",
			Message: "title must be no more than 255 characters long",
		})
	}

	// Validate content
	if req.Content == "" {
		errors = append(errors, ValidationError{
			Field:   "content",
			Message: "content is required",
		})
	}

	// Validate user_id
	if req.UserID <= 0 {
		errors = append(errors, ValidationError{
			Field:   "user_id",
			Message: "user_id must be a positive integer",
		})
	}

	if len(errors) > 0 {
		return ValidationErrors{Errors: errors}
	}

	return nil
}

// ValidatePostUpdateRequest validates a post update request
func ValidatePostUpdateRequest(req *models.PostRequest) error {
	var errors []ValidationError

	// For updates, fields are optional, but if provided, they must be valid
	if req.Title != "" && len(req.Title) > 255 {
		errors = append(errors, ValidationError{
			Field:   "title",
			Message: "title must be no more than 255 characters long",
		})
	}

	// Content can be empty in updates, so no validation needed

	if req.UserID < 0 {
		errors = append(errors, ValidationError{
			Field:   "user_id",
			Message: "user_id must be a non-negative integer",
		})
	}

	if len(errors) > 0 {
		return ValidationErrors{Errors: errors}
	}

	return nil
}

// isValidEmail checks if the email format is valid
func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
