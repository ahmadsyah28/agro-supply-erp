package procurement

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"agro-supply-erp/internal/response"
)

// Handler menangani request HTTP untuk modul procurement
type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes mendaftarkan endpoint PR dan PO ke ServeMux
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// Purchase Requests
	mux.HandleFunc("POST /api/v1/purchase-requests", h.CreatePR)
	mux.HandleFunc("GET /api/v1/purchase-requests", h.ListPR)
	mux.HandleFunc("GET /api/v1/purchase-requests/{id}", h.GetPRByID)
	mux.HandleFunc("POST /api/v1/purchase-requests/{id}/submit", h.SubmitPR)
	mux.HandleFunc("POST /api/v1/purchase-requests/{id}/approve", h.ApprovePR)
	mux.HandleFunc("POST /api/v1/purchase-requests/{id}/reject", h.RejectPR)

	// Purchase Orders
	mux.HandleFunc("POST /api/v1/purchase-orders", h.CreatePO)
	mux.HandleFunc("GET /api/v1/purchase-orders", h.ListPO)
	mux.HandleFunc("GET /api/v1/purchase-orders/{id}", h.GetPOByID)
	mux.HandleFunc("POST /api/v1/purchase-orders/{id}/approve", h.ApprovePO)
	mux.HandleFunc("POST /api/v1/purchase-orders/{id}/cancel", h.CancelPO)
}

// --- Purchase Request Handlers ---

func (h *Handler) CreatePR(w http.ResponseWriter, r *http.Request) {
	var req CreatePRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "Format JSON request tidak valid", nil)
		return
	}

	pr, err := h.service.CreatePR(r.Context(), req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	response.JSON(w, http.StatusCreated, pr)
}

func (h *Handler) ListPR(w http.ResponseWriter, r *http.Request) {
	list, err := h.service.ListPR(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memuat daftar PR", nil)
		return
	}
	if list == nil {
		list = []*PurchaseRequest{}
	}
	response.JSON(w, http.StatusOK, list)
}

func (h *Handler) GetPRByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "ID dokumen tidak valid", nil)
		return
	}

	pr, err := h.service.GetPRByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Dokumen PR tidak ditemukan", nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}

	response.JSON(w, http.StatusOK, pr)
}

func (h *Handler) SubmitPR(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "ID dokumen tidak valid", nil)
		return
	}

	pr, err := h.service.SubmitPR(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Dokumen PR tidak ditemukan", nil)
			return
		}
		response.Error(w, http.StatusBadRequest, "INVALID_TRANSITION", err.Error(), nil)
		return
	}

	response.JSON(w, http.StatusOK, pr)
}

func (h *Handler) ApprovePR(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "ID dokumen tidak valid", nil)
		return
	}

	pr, err := h.service.ApprovePR(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Dokumen PR tidak ditemukan", nil)
			return
		}
		response.Error(w, http.StatusBadRequest, "INVALID_TRANSITION", err.Error(), nil)
		return
	}

	response.JSON(w, http.StatusOK, pr)
}

func (h *Handler) RejectPR(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "ID dokumen tidak valid", nil)
		return
	}

	pr, err := h.service.RejectPR(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Dokumen PR tidak ditemukan", nil)
			return
		}
		response.Error(w, http.StatusBadRequest, "INVALID_TRANSITION", err.Error(), nil)
		return
	}

	response.JSON(w, http.StatusOK, pr)
}

// --- Purchase Order Handlers ---

func (h *Handler) CreatePO(w http.ResponseWriter, r *http.Request) {
	var req CreatePORequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "Format JSON request tidak valid", nil)
		return
	}

	po, err := h.service.CreatePO(r.Context(), req)
	if err != nil {
		if errors.Is(err, ErrSupplierNotActive) || errors.Is(err, ErrPRNotApproved) || errors.Is(err, ErrProductNotActive) {
			response.Error(w, http.StatusUnprocessableEntity, "BUSINESS_RULE_VIOLATION", err.Error(), nil)
			return
		}
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	response.JSON(w, http.StatusCreated, po)
}

func (h *Handler) ListPO(w http.ResponseWriter, r *http.Request) {
	list, err := h.service.ListPO(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memuat daftar PO", nil)
		return
	}
	if list == nil {
		list = []*PurchaseOrder{}
	}
	response.JSON(w, http.StatusOK, list)
}

func (h *Handler) GetPOByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "ID dokumen tidak valid", nil)
		return
	}

	po, err := h.service.GetPOByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Dokumen PO tidak ditemukan", nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}

	response.JSON(w, http.StatusOK, po)
}

func (h *Handler) ApprovePO(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "ID dokumen tidak valid", nil)
		return
	}

	po, err := h.service.ApprovePO(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Dokumen PO tidak ditemukan", nil)
			return
		}
		response.Error(w, http.StatusBadRequest, "INVALID_TRANSITION", err.Error(), nil)
		return
	}

	response.JSON(w, http.StatusOK, po)
}

func (h *Handler) CancelPO(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "ID dokumen tidak valid", nil)
		return
	}

	po, err := h.service.CancelPO(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Dokumen PO tidak ditemukan", nil)
			return
		}
		response.Error(w, http.StatusBadRequest, "INVALID_TRANSITION", err.Error(), nil)
		return
	}

	response.JSON(w, http.StatusOK, po)
}
