package auth

import (
	"encoding/json"
	"log"
	"net/http"
)

type AuthError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeJSONError(w http.ResponseWriter, statusCode int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(AuthError{
		Code:    code,
		Message: message,
	}); err != nil {
		log.Printf("Failed to write JSON error: %v", err)
	}
}
