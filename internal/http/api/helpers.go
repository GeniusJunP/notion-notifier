package api

import (
	"encoding/json"
	"net/http"
)

// respondJSON writes a JSON response with the given status code.
func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// respondError writes a JSON error response.
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

// respondValidationError writes a 422 error with structured details.
func respondValidationError(w http.ResponseWriter, message string, details map[string]string) {
	respondJSON(w, http.StatusUnprocessableEntity, map[string]any{
		"error":   message,
		"details": details,
	})
}
