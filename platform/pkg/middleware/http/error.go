package http

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse is a structure for returning errors in JSON format
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail is the error detail
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// writeErrorResponse writes a structured JSON error response
func writeErrorResponse(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	}

	// the status is already sent, so a failed write can only be dropped
	_ = json.NewEncoder(w).Encode(response)
}
