package api

import (
	"encoding/json"
	"net/http"

	"notion-notifier/internal/logging"
)

// respondJSON writes a JSON response with the given status code.
func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			logging.Error("HTTP", "json encode failed: %v", err)
		}
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

// requireMethod validates request method and writes 405 on mismatch.
func requireMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method == method {
		return true
	}
	respondError(w, http.StatusMethodNotAllowed, "method not allowed")
	return false
}
