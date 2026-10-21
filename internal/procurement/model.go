package procurement

import (
	"errors"
	"strings"
	"time"
)

// Status untuk Purchase Request (PR)
const (
	PRStatusDraft     = "DRAFT"
	PRStatusSubmitted = "SUBMITTED"
	PRStatusApproved  = "APPROVED"
	PRStatusRejected  = "REJECTED"
)

// Status untuk Purchase Order (PO)
const (
	POStatusDraft             = "DRAFT"
	POStatusApproved          = "APPROVED"
	POStatusPartiallyReceived = "PARTIALLY_RECEIVED"
	POStatusCompleted         = "COMPLETED"
	POStatusCancelled         = "CANCELLED"
)

// PurchaseRequest merepresentasikan dokumen pengajuan pembelian
type PurchaseRequest struct {
	ID        int64                 `json:"id"`
	PRNumber  string                `json:"pr_number"`
	Status    string                `json:"status"`
	Notes     string                `json:"notes"`
	Items     []PurchaseRequestItem `json:"items,omitempty"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
}

// PurchaseRequestItem merepresentasikan rincian produk dalam PR
type PurchaseRequestItem struct {
	ID                int64     `json:"id"`
	PurchaseRequestID int64     `json:"purchase_request_id"`
	ProductID         int64     `json:"product_id"`
	ProductName       string    `json:"product_name,omitempty"`
	ProductSKU        string    `json:"product_sku,omitempty"`
	Quantity          float64   `json:"quantity"`
	Notes             string    `json:"notes"`
	CreatedAt         time.Time `json:"created_at"`
}

// CreatePRRequest payload saat membuat PR baru
type CreatePRRequest struct {
	Notes string                `json:"notes"`
	Items []CreatePRItemRequest `json:"items"`
}

// CreatePRItemRequest payload item untuk PR baru
type CreatePRItemRequest struct {
	ProductID int64   `json:"product_id"`
	Quantity  float64 `json:"quantity"`
	Notes     string  `json:"notes"`
}

// Validate memeriksa kelayakan input PR baru
func (r *CreatePRRequest) Validate() error {
	r.Notes = strings.TrimSpace(r.Notes)

	if len(r.Items) == 0 {
		return errors.New("daftar item PR tidak boleh kosong (minimal 1 barang)")
	}

	for i, item := range r.Items {
		if item.ProductID <= 0 {
			return errors.New("ID produk pada baris item tidak valid")
		}
		if item.Quantity <= 0 {
			return errors.New("kuantitas barang harus lebih besar dari 0")
		}
		r.Items[i].Notes = strings.TrimSpace(item.Notes)
	}

	return nil
}

// PurchaseOrder merepresentasikan dokumen pesanan resmi ke supplier
type PurchaseOrder struct {
	ID                int64               `json:"id"`
	PONumber          string              `json:"po_number"`
	PurchaseRequestID *int64              `json:"purchase_request_id,omitempty"`
	SupplierID        int64               `json:"supplier_id"`
	SupplierName      string              `json:"supplier_name,omitempty"`
	Status            string              `json:"status"`
	TotalAmount       float64             `json:"total_amount"`
	Notes             string              `json:"notes"`
	Items             []PurchaseOrderItem `json:"items,omitempty"`
	CreatedAt         time.Time           `json:"created_at"`
	UpdatedAt         time.Time           `json:"updated_at"`
}

// PurchaseOrderItem merepresentasikan rincian produk dan harga dalam PO
type PurchaseOrderItem struct {
	ID              int64     `json:"id"`
	PurchaseOrderID int64     `json:"purchase_order_id"`
	ProductID       int64     `json:"product_id"`
	ProductName     string    `json:"product_name,omitempty"`
	ProductSKU      string    `json:"product_sku,omitempty"`
	Quantity        float64   `json:"quantity"`
	UnitPrice       float64   `json:"unit_price"`
	Subtotal        float64   `json:"subtotal"`
	CreatedAt       time.Time `json:"created_at"`
}

// CreatePORequest payload saat membuat PO baru
type CreatePORequest struct {
	PurchaseRequestID *int64              `json:"purchase_request_id"`
	SupplierID        int64               `json:"supplier_id"`
	Notes             string              `json:"notes"`
	Items             []CreatePOItemRequest `json:"items"`
}

// CreatePOItemRequest payload item untuk PO baru
type CreatePOItemRequest struct {
	ProductID int64   `json:"product_id"`
	Quantity  float64 `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
}

// Validate memeriksa kelayakan input PO baru
func (r *CreatePORequest) Validate() error {
	r.Notes = strings.TrimSpace(r.Notes)

	if r.SupplierID <= 0 {
		return errors.New("ID supplier wajib diisi dan harus valid")
	}

	if len(r.Items) == 0 {
		return errors.New("daftar item PO tidak boleh kosong (minimal 1 barang)")
	}

	for _, item := range r.Items {
		if item.ProductID <= 0 {
			return errors.New("ID produk pada baris item tidak valid")
		}
		if item.Quantity <= 0 {
			return errors.New("kuantitas pesanan harus lebih besar dari 0")
		}
		if item.UnitPrice < 0 {
			return errors.New("harga satuan tidak boleh negatif")
		}
	}

	return nil
}
