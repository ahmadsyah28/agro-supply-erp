package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"agro-supply-erp/internal/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// Login menangani POST /api/v1/auth/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "Format JSON request tidak valid", nil)
		return
	}

	authResp, err := h.service.Login(r.Context(), req)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			response.Error(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", err.Error(), nil)
			return
		}
		if errors.Is(err, ErrUserInactive) {
			response.Error(w, http.StatusForbidden, "USER_INACTIVE", err.Error(), nil)
			return
		}
		response.Error(w, http.StatusBadRequest, "LOGIN_FAILED", err.Error(), nil)
		return
	}

	response.JSON(w, http.StatusOK, authResp)
}

// Register menangani POST /api/v1/auth/register (Admin or authorized)
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "Format JSON request tidak valid", nil)
		return
	}

	user, err := h.service.Register(r.Context(), req)
	if err != nil {
		if errors.Is(err, ErrUsernameExists) || errors.Is(err, ErrEmailExists) {
			response.Error(w, http.StatusConflict, "USER_EXISTS", err.Error(), nil)
			return
		}
		if errors.Is(err, ErrInvalidRole) {
			response.Error(w, http.StatusBadRequest, "INVALID_ROLE", err.Error(), nil)
			return
		}
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	response.JSON(w, http.StatusCreated, user)
}

// GetMe menangani GET /api/v1/auth/me
func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	// Diambil dari context yang telah divalidasi oleh Auth middleware
	claims, ok := GetClaims(r.Context())
	if !ok || claims == nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Sesi tidak ditemukan", nil)
		return
	}

	profile, err := h.service.GetProfile(r.Context(), claims.UserID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "Profil pengguna tidak ditemukan", nil)
		return
	}

	response.JSON(w, http.StatusOK, profile)
}

// ListUsers menangani GET /api/v1/users (Admin only)
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.service.ListUsers(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memuat daftar pengguna", nil)
		return
	}

	if users == nil {
		users = []UserDTO{}
	}

	response.JSON(w, http.StatusOK, users)
}

// UpdateUserStatus menangani PUT /api/v1/users/{id}/status (Admin only)
func (h *Handler) UpdateUserStatus(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "ID user tidak valid", nil)
		return
	}

	var req struct {
		IsActive bool `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "Format JSON request tidak valid", nil)
		return
	}

	if err := h.service.UpdateUserStatus(r.Context(), id, req.IsActive); err != nil {
		if errors.Is(err, ErrUserNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Pengguna tidak ditemukan", nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memperbarui status pengguna", nil)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"id":        id,
		"is_active": req.IsActive,
		"message":   "Status pengguna berhasil diperbarui",
	})
}
