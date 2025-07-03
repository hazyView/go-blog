package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"blog-api/internal/models"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Error().Err(err).Msg("Failed to encode JSON response")
	}
}

// writeError writes an error response
func writeError(w http.ResponseWriter, status int, message string) {
	response := models.ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
		Code:    status,
	}
	writeJSON(w, status, response)
}

// writeValidationError writes a validation error response
func writeValidationError(w http.ResponseWriter, err error) {
	if validationErr, ok := err.(ValidationErrors); ok {
		response := map[string]interface{}{
			"error":   "Validation failed",
			"code":    http.StatusBadRequest,
			"details": validationErr.Errors,
		}
		writeJSON(w, http.StatusBadRequest, response)
	} else {
		writeError(w, http.StatusBadRequest, err.Error())
	}
}

// writeSuccess writes a success response
func writeSuccess(w http.ResponseWriter, message string, data interface{}) {
	response := models.SuccessResponse{
		Message: message,
		Data:    data,
	}
	writeJSON(w, http.StatusOK, response)
}

// parseIDFromURL extracts and validates an ID from the URL path
func parseIDFromURL(r *http.Request, paramName string) (int, error) {
	vars := mux.Vars(r)
	idStr, exists := vars[paramName]
	if !exists {
		return 0, http.ErrMissingFile
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, err
	}

	if id <= 0 {
		return 0, http.ErrMissingFile
	}

	return id, nil
}

// parseJSON parses JSON from request body
func parseJSON(r *http.Request, dst interface{}) error {
	if r.Body == nil {
		return http.ErrMissingFile
	}
	
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(dst)
}

// handleDatabaseError converts database errors to appropriate HTTP responses
func handleDatabaseError(w http.ResponseWriter, err error, operation string) {
	log.Error().Err(err).Str("operation", operation).Msg("Database operation failed")
	
	errMsg := err.Error()
	
	// Check for common error patterns
	switch {
	case contains(errMsg, "not found"):
		writeError(w, http.StatusNotFound, "Resource not found")
	case contains(errMsg, "duplicate") || contains(errMsg, "unique"):
		writeError(w, http.StatusConflict, "Resource already exists")
	case contains(errMsg, "foreign key"):
		writeError(w, http.StatusBadRequest, "Invalid reference to related resource")
	case contains(errMsg, "invalid") || contains(errMsg, "check constraint"):
		writeError(w, http.StatusBadRequest, "Invalid data provided")
	default:
		writeError(w, http.StatusInternalServerError, "Internal server error")
	}
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] && s[i+j] != substr[j]-32 && s[i+j] != substr[j]+32 {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
