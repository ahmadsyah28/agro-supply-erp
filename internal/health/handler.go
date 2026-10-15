package health

import (
	"context"
	"net/http"
	"time"

	"agro-supply-erp/internal/database"
	"agro-supply-erp/internal/response"
)

type Handler struct {
	db *database.DB
}

func NewHandler(db *database.DB) *Handler {
	return &Handler{db: db}
}

// Live handles GET /health/live - checks if the application server is alive
func (h *Handler) Live(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", nil)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"status":    "alive",
		"timestamp": time.Now().UTC(),
	})
}

// Ready handles GET /health/ready - checks if the application dependencies (e.g. database) are ready
func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", nil)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	if err := h.db.Ping(ctx); err != nil {
		response.Error(w, http.StatusServiceUnavailable, "DATABASE_UNAVAILABLE", "Database connection is unhealthy", nil)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"status":    "ready",
		"database":  "connected",
		"timestamp": time.Now().UTC(),
	})
}
