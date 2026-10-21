package supplier

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"agro-supply-erp/internal/response"
)

// Handler menangani request HTTP untuk modul supplier
type Handler struct {
	service Service
}

// NewHandler membuat instance Handler baru dengan menyuntikkan Service
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes mendaftarkan seluruh endpoint supplier ke ServeMux
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/suppliers", h.Create)
	mux.HandleFunc("GET /api/v1/suppliers", h.List)
	mux.HandleFunc("GET /api/v1/suppliers/{id}", h.GetByID)
	mux.HandleFunc("PUT /api/v1/suppliers/{id}", h.Update)
	mux.HandleFunc("DELETE /api/v1/suppliers/{id}", h.Delete)
}

// Create menangani POST /api/v1/suppliers
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateSupplierRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "Format JSON request tidak valid", nil)
		return
	}

	s, err := h.service.Create(r.Context(), req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	response.JSON(w, http.StatusCreated, s)
}

// List menangani GET /api/v1/suppliers (mendukung filter query ?active=true)
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	onlyActive := r.URL.Query().Get("active") == "true"

	suppliers, err := h.service.List(r.Context(), onlyActive)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memuat daftar supplier", nil)
		return
	}

	if suppliers == nil {
		suppliers = []*Supplier{} // kembalikan array kosong [] ketimbang null di JSON
	}

	response.JSON(w, http.StatusOK, suppliers)
}

// GetByID menangani GET /api/v1/suppliers/{id}
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "ID supplier harus berupa angka bulat positif", nil)
		return
	}

	s, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Supplier tidak ditemukan", nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal mengambil data supplier", nil)
		return
	}

	response.JSON(w, http.StatusOK, s)
}

// Update menangani PUT /api/v1/suppliers/{id}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "ID supplier harus berupa angka bulat positif", nil)
		return
	}

	var req UpdateSupplierRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "Format JSON request tidak valid", nil)
		return
	}

	s, err := h.service.Update(r.Context(), id, req)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Supplier tidak ditemukan", nil)
			return
		}
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	response.JSON(w, http.StatusOK, s)
}

// Delete menangani DELETE /api/v1/suppliers/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "ID supplier harus berupa angka bulat positif", nil)
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Supplier tidak ditemukan", nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal menonaktifkan supplier", nil)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "Supplier berhasil dinonaktifkan",
	})
}
