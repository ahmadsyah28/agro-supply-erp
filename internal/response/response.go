package response

import (
	"encoding/json"
	"log"
	"net/http"
)

type APIResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data,omitempty"`
	Error  *APIError   `json:"error,omitempty"`
}

type APIError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func JSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	resp := APIResponse{
		Status: "success",
		Data:   data,
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("[ERROR] Failed to encode response: %v", err)
	}
}

func Error(w http.ResponseWriter, statusCode int, code string, message string, details interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	resp := APIResponse{
		Status: "error",
		Error: &APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("[ERROR] Failed to encode error response: %v", err)
	}
}

func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
