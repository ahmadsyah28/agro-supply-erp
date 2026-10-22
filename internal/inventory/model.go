package inventory

import (
	"errors"
	"strings"
	"time"
)

// Jenis transaksi inventory
const (
	TxTypeIn         = "IN"
	TxTypeOut        = "OUT"
	TxTypeAdjustment = "ADJUSTMENT"
)

// Tipe dokumen referensi
const (
	RefGoodsReceipt = "GOODS_RECEIPT"
)

// GoodsReceipt merepresentasikan dokumen bukti penerimaan barang dari supplier
type GoodsReceipt struct {
	ID                  int64              `json:"id"`
	GRNumber            string             `json:"gr_number"`
	PurchaseOrderID     int64              `json:"purchase_order_id"`
	PONumber            string             `json:"po_number,omitempty"`
	DeliveryOrderNumber string             `json:"delivery_order_number"`
	ReceiptDate         time.Time          `json:"receipt_date"`
	Notes               string             `json:"notes"`
	Items               []GoodsReceiptItem `json:"items,omitempty"`
	CreatedAt           time.Time          `json:"created_at"`
}

// GoodsReceiptItem rincian fisik, barang lolos QC, dan barang cacat
type GoodsReceiptItem struct {
	ID                  int64     `json:"id"`
	GoodsReceiptID      int64     `json:"goods_receipt_id"`
	PurchaseOrderItemID int64     `json:"purchase_order_item_id"`
	ProductID           int64     `json:"product_id"`
	ProductName         string    `json:"product_name,omitempty"`
	ProductSKU          string    `json:"product_sku,omitempty"`
	ProductUnit         string    `json:"product_unit,omitempty"`
	ReceivedQuantity    float64   `json:"received_quantity"`
	AcceptedQuantity    float64   `json:"accepted_quantity"`
	RejectedQuantity    float64   `json:"rejected_quantity"`
	RejectionReason     string    `json:"rejection_reason,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
}

// InventoryTransaction buku besar (ledger) pencatat setiap riwayat pergerakan stok
type InventoryTransaction struct {
	ID              int64     `json:"id"`
	ProductID       int64     `json:"product_id"`
	ProductName     string    `json:"product_name,omitempty"`
	ProductSKU      string    `json:"product_sku,omitempty"`
	ProductUnit     string    `json:"product_unit,omitempty"`
	TransactionType string    `json:"transaction_type"` // IN, OUT, ADJUSTMENT
	ReferenceType   string    `json:"reference_type"`   // GOODS_RECEIPT, dll
	ReferenceID     int64     `json:"reference_id"`
	ReferenceNumber string    `json:"reference_number"`
	Quantity        float64   `json:"quantity"`
	Notes           string    `json:"notes"`
	CreatedAt       time.Time `json:"created_at"`
}

// StockSummary saldo stok terkini per produk
type StockSummary struct {
	ProductID    int64   `json:"product_id"`
	ProductName  string  `json:"product_name"`
	ProductSKU   string  `json:"product_sku"`
	ProductUnit  string  `json:"product_unit"`
	TotalIn      float64 `json:"total_in"`
	TotalOut     float64 `json:"total_out"`
	CurrentStock float64 `json:"current_stock"`
}

// CreateGRRequest payload saat membuat penerimaan barang
type CreateGRRequest struct {
	PurchaseOrderID     int64               `json:"purchase_order_id"`
	DeliveryOrderNumber string              `json:"delivery_order_number"`
	Notes               string              `json:"notes"`
	Items               []CreateGRItemRequest `json:"items"`
}

// CreateGRItemRequest payload untuk setiap item penerimaan barang
type CreateGRItemRequest struct {
	PurchaseOrderItemID int64   `json:"purchase_order_item_id"`
	ProductID           int64   `json:"product_id"`
	ReceivedQuantity    float64 `json:"received_quantity"`
	AcceptedQuantity    float64 `json:"accepted_quantity"`
	RejectedQuantity    float64 `json:"rejected_quantity"`
	RejectionReason     string  `json:"rejection_reason"`
}

// Validate memeriksa kelayakan input penerimaan barang
func (r *CreateGRRequest) Validate() error {
	r.DeliveryOrderNumber = strings.TrimSpace(r.DeliveryOrderNumber)
	r.Notes = strings.TrimSpace(r.Notes)

	if r.PurchaseOrderID <= 0 {
		return errors.New("ID Purchase Order wajib diisi dan harus valid")
	}

	if len(r.Items) == 0 {
		return errors.New("daftar item penerimaan barang tidak boleh kosong (minimal 1 barang)")
	}

	for i, item := range r.Items {
		if item.PurchaseOrderItemID <= 0 {
			return errors.New("ID baris item Purchase Order tidak valid")
		}
		if item.ProductID <= 0 {
			return errors.New("ID produk pada baris item tidak valid")
		}
		if item.ReceivedQuantity <= 0 {
			return errors.New("kuantitas barang diterima harus lebih besar dari 0")
		}
		if item.AcceptedQuantity < 0 || item.RejectedQuantity < 0 {
			return errors.New("kuantitas lolos (accepted) atau ditolak (rejected) tidak boleh negatif")
		}

		// Aturan: jumlah lolos + jumlah ditolak harus pas sama dengan kuantitas fisik yang tiba
		totalCheck := item.AcceptedQuantity + item.RejectedQuantity
		// Gunakan toleransi desimal kecil untuk perbandingan float
		diff := totalCheck - item.ReceivedQuantity
		if diff < -0.001 || diff > 0.001 {
			return errors.New("total kuantitas diterima (received) harus sama dengan jumlah lolos (accepted) + ditolak (rejected)")
		}

		r.Items[i].RejectionReason = strings.TrimSpace(item.RejectionReason)
		if item.RejectedQuantity > 0 && r.Items[i].RejectionReason == "" {
			return errors.New("alasan penolakan barang wajib diisi jika terdapat kuantitas yang ditolak")
		}
	}

	return nil
}
