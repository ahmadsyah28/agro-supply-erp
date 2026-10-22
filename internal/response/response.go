package response

import (
	"encoding/json"
	"log"
	"math"
	"net/http"
	"strconv"
)

type APIResponse struct {
	Status string          `json:"status"`
	Data   interface{}     `json:"data,omitempty"`
	Meta   *PaginationMeta `json:"meta,omitempty"`
	Error  *APIError       `json:"error,omitempty"`
}

type APIError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

type PaginationMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
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

func JSONWithMeta(w http.ResponseWriter, statusCode int, data interface{}, meta *PaginationMeta) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	resp := APIResponse{
		Status: "success",
		Data:   data,
		Meta:   meta,
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("[ERROR] Failed to encode response with meta: %v", err)
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

// ParsePagination membaca parameter query ?page=1&limit=20 dari URL
func ParsePagination(r *http.Request) (page int, limit int, offset int) {
	page = 1
	limit = 20

	if pStr := r.URL.Query().Get("page"); pStr != "" {
		if p, err := strconv.Atoi(pStr); err == nil && p > 0 {
			page = p
		}
	}

	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			if l > 100 {
				limit = 100 // Batas atas maksimal 100 per halaman
			} else {
				limit = l
			}
		}
	}

	offset = (page - 1) * limit
	return page, limit, offset
}

// NewPaginationMeta membuat metadata pagination yang rapi
func NewPaginationMeta(page, limit int, totalItems int64) *PaginationMeta {
	totalPages := int(math.Ceil(float64(totalItems) / float64(limit)))
	if totalPages == 0 && totalItems == 0 {
		totalPages = 1
	}

	return &PaginationMeta{
		Page:       page,
		Limit:      limit,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}
}
