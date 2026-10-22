package inventory

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"agro-supply-erp/internal/response"
)

// Handler menangani request HTTP untuk modul inventory
type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes mendaftarkan endpoint goods receipts dan inventory ledger ke ServeMux
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/goods-receipts", h.CreateGoodsReceipt)
	mux.HandleFunc("GET /api/v1/goods-receipts", h.ListGoodsReceipts)
	mux.HandleFunc("GET /api/v1/goods-receipts/{id}", h.GetGoodsReceiptByID)
	mux.HandleFunc("GET /api/v1/inventory/stock", h.GetStockSummary)
	mux.HandleFunc("GET /api/v1/inventory/ledger", h.ListTransactions)
}

func (h *Handler) CreateGoodsReceipt(w http.ResponseWriter, r *http.Request) {
	var req CreateGRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "Format JSON request tidak valid", nil)
		return
	}

	gr, err := h.service.CreateGoodsReceipt(r.Context(), req)
	if err != nil {
		if errors.Is(err, ErrPOInvalidState) || errors.Is(err, ErrOverReceipt) || errors.Is(err, ErrPOItemNotFound) {
			response.Error(w, http.StatusUnprocessableEntity, "BUSINESS_RULE_VIOLATION", err.Error(), nil)
			return
		}
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	response.JSON(w, http.StatusCreated, gr)
}

func (h *Handler) ListGoodsReceipts(w http.ResponseWriter, r *http.Request) {
	list, err := h.service.ListGoodsReceipts(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memuat daftar Goods Receipt", nil)
		return
	}
	if list == nil {
		list = []*GoodsReceipt{}
	}
	response.JSON(w, http.StatusOK, list)
}

func (h *Handler) GetGoodsReceiptByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "ID dokumen tidak valid", nil)
		return
	}

	gr, err := h.service.GetGoodsReceiptByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Dokumen Goods Receipt tidak ditemukan", nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}

	response.JSON(w, http.StatusOK, gr)
}

func (h *Handler) GetStockSummary(w http.ResponseWriter, r *http.Request) {
	summaries, err := h.service.GetStockSummary(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memuat saldo stok", nil)
		return
	}
	if summaries == nil {
		summaries = []*StockSummary{}
	}
	response.JSON(w, http.StatusOK, summaries)
}

func (h *Handler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	var productIDPtr *int64
	productIDStr := r.URL.Query().Get("product_id")
	if productIDStr != "" {
		if id, err := strconv.ParseInt(productIDStr, 10, 64); err == nil && id > 0 {
			productIDPtr = &id
		}
	}

	txs, err := h.service.ListTransactions(r.Context(), productIDPtr)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memuat riwayat transaksi stok", nil)
		return
	}
	if txs == nil {
		txs = []*InventoryTransaction{}
	}
	response.JSON(w, http.StatusOK, txs)
}
