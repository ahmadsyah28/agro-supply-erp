package inventory

import (
	"context"
	"errors"
	"fmt"
	"time"

	"agro-supply-erp/internal/database"

	"github.com/jackc/pgx/v5"
)

var (
	ErrNotFound = errors.New("dokumen penerimaan tidak ditemukan")
)

// Repository kontrak akses database untuk modul inventory & ledger
type Repository interface {
	NextGRNumber(ctx context.Context) (string, error)
	CreateGoodsReceipt(ctx context.Context, gr *GoodsReceipt, newPOStatus string) error
	GetGoodsReceiptByID(ctx context.Context, id int64) (*GoodsReceipt, error)
	ListGoodsReceipts(ctx context.Context) ([]*GoodsReceipt, error)
	GetCumulativeReceivedForPO(ctx context.Context, poID int64) (map[int64]float64, error)
	GetStockSummary(ctx context.Context) ([]*StockSummary, error)
	ListTransactions(ctx context.Context, productID *int64) ([]*InventoryTransaction, error)
}

type pgRepository struct {
	db *database.DB
}

func NewRepository(db *database.DB) Repository {
	return &pgRepository{db: db}
}

// NextGRNumber menghasilkan nomor urut GR, misal GR-202610-0001
func (r *pgRepository) NextGRNumber(ctx context.Context) (string, error) {
	var count int64
	err := r.db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM goods_receipts").Scan(&count)
	if err != nil {
		return "", err
	}
	now := time.Now().UTC()
	return fmt.Sprintf("GR-%s-%04d", now.Format("200601"), count+1), nil
}

// GetCumulativeReceivedForPO mengambil total kuantitas yang sudah pernah diterima per baris item PO
func (r *pgRepository) GetCumulativeReceivedForPO(ctx context.Context, poID int64) (map[int64]float64, error) {
	query := `
		SELECT gri.purchase_order_item_id, COALESCE(SUM(gri.accepted_quantity), 0)
		FROM goods_receipt_items gri
		JOIN goods_receipts gr ON gri.goods_receipt_id = gr.id
		WHERE gr.purchase_order_id = $1
		GROUP BY gri.purchase_order_item_id
	`
	rows, err := r.db.Pool.Query(ctx, query, poID)
	if err != nil {
		return nil, fmt.Errorf("gagal menghitung akumulasi penerimaan PO: %w", err)
	}
	defer rows.Close()

	result := make(map[int64]float64)
	for rows.Next() {
		var poItemID int64
		var qty float64
		if err := rows.Scan(&poItemID, &qty); err != nil {
			return nil, err
		}
		result[poItemID] = qty
	}
	return result, nil
}

// CreateGoodsReceipt mengeksekusi penyimpanan GR, Item, Ledger Stok, dan status PO dalam 1 TRANSAKSI ATOMIK
func (r *pgRepository) CreateGoodsReceipt(ctx context.Context, gr *GoodsReceipt, newPOStatus string) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("gagal memulai transaksi database: %w", err)
	}
	defer tx.Rollback(ctx)

	now := time.Now().UTC()
	gr.CreatedAt = now
	gr.ReceiptDate = now

	// 1. Simpan Header Goods Receipt
	queryHeader := `
		INSERT INTO goods_receipts (gr_number, purchase_order_id, delivery_order_number, receipt_date, notes, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	if err := tx.QueryRow(ctx, queryHeader,
		gr.GRNumber,
		gr.PurchaseOrderID,
		gr.DeliveryOrderNumber,
		gr.ReceiptDate,
		gr.Notes,
		gr.CreatedAt,
	).Scan(&gr.ID); err != nil {
		return fmt.Errorf("gagal menyimpan header GR: %w", err)
	}

	// 2. Simpan Item-item Goods Receipt & Catat ke Buku Besar (Inventory Ledger)
	queryItem := `
		INSERT INTO goods_receipt_items (
			goods_receipt_id, purchase_order_item_id, product_id,
			received_quantity, accepted_quantity, rejected_quantity,
			rejection_reason, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	queryLedger := `
		INSERT INTO inventory_transactions (
			product_id, transaction_type, reference_type, reference_id, reference_number, quantity, notes, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	for i := range gr.Items {
		gr.Items[i].GoodsReceiptID = gr.ID
		gr.Items[i].CreatedAt = now

		if err := tx.QueryRow(ctx, queryItem,
			gr.ID,
			gr.Items[i].PurchaseOrderItemID,
			gr.Items[i].ProductID,
			gr.Items[i].ReceivedQuantity,
			gr.Items[i].AcceptedQuantity,
			gr.Items[i].RejectedQuantity,
			gr.Items[i].RejectionReason,
			gr.Items[i].CreatedAt,
		).Scan(&gr.Items[i].ID); err != nil {
			return fmt.Errorf("gagal menyimpan baris item GR ke-%d: %w", i+1, err)
		}

		// ATURAN PENTING: Hanya kuantitas yang lolos (accepted_quantity) yang masuk ke buku besar stok!
		if gr.Items[i].AcceptedQuantity > 0 {
			notes := fmt.Sprintf("Penerimaan barang dari PO #%d (Surat Jalan: %s)", gr.PurchaseOrderID, gr.DeliveryOrderNumber)
			if _, err := tx.Exec(ctx, queryLedger,
				gr.Items[i].ProductID,
				TxTypeIn,
				RefGoodsReceipt,
				gr.ID,
				gr.GRNumber,
				gr.Items[i].AcceptedQuantity,
				notes,
				now,
			); err != nil {
				return fmt.Errorf("gagal mencatat ledger stok item ke-%d: %w", i+1, err)
			}
		}
	}

	// 3. Update Status Dokumen Purchase Order (PO)
	if newPOStatus != "" {
		updatePOQuery := `UPDATE purchase_orders SET status = $1, updated_at = $2 WHERE id = $3`
		if _, err := tx.Exec(ctx, updatePOQuery, newPOStatus, now, gr.PurchaseOrderID); err != nil {
			return fmt.Errorf("gagal mengupdate status PO ke %s: %w", newPOStatus, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("gagal commit transaksi goods receipt: %w", err)
	}

	return nil
}

// GetGoodsReceiptByID mengambil detail dokumen GR beserta rincian itemnya
func (r *pgRepository) GetGoodsReceiptByID(ctx context.Context, id int64) (*GoodsReceipt, error) {
	queryHeader := `
		SELECT gr.id, gr.gr_number, gr.purchase_order_id, po.po_number, gr.delivery_order_number, gr.receipt_date, gr.notes, gr.created_at
		FROM goods_receipts gr
		JOIN purchase_orders po ON gr.purchase_order_id = po.id
		WHERE gr.id = $1
	`
	gr := &GoodsReceipt{}
	err := r.db.Pool.QueryRow(ctx, queryHeader, id).Scan(
		&gr.ID,
		&gr.GRNumber,
		&gr.PurchaseOrderID,
		&gr.PONumber,
		&gr.DeliveryOrderNumber,
		&gr.ReceiptDate,
		&gr.Notes,
		&gr.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("gagal mengambil GR id %d: %w", id, err)
	}

	queryItems := `
		SELECT gri.id, gri.goods_receipt_id, gri.purchase_order_item_id, gri.product_id,
		       p.name, p.sku, p.unit,
		       gri.received_quantity, gri.accepted_quantity, gri.rejected_quantity,
		       COALESCE(gri.rejection_reason, ''), gri.created_at
		FROM goods_receipt_items gri
		JOIN products p ON gri.product_id = p.id
		WHERE gri.goods_receipt_id = $1
		ORDER BY gri.id ASC
	`
	rows, err := r.db.Pool.Query(ctx, queryItems, id)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil item GR: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item GoodsReceiptItem
		if err := rows.Scan(
			&item.ID,
			&item.GoodsReceiptID,
			&item.PurchaseOrderItemID,
			&item.ProductID,
			&item.ProductName,
			&item.ProductSKU,
			&item.ProductUnit,
			&item.ReceivedQuantity,
			&item.AcceptedQuantity,
			&item.RejectedQuantity,
			&item.RejectionReason,
			&item.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("gagal memindai item GR: %w", err)
		}
		gr.Items = append(gr.Items, item)
	}

	return gr, nil
}

// ListGoodsReceipts mengambil seluruh daftar dokumen GR
func (r *pgRepository) ListGoodsReceipts(ctx context.Context) ([]*GoodsReceipt, error) {
	query := `
		SELECT gr.id, gr.gr_number, gr.purchase_order_id, po.po_number, gr.delivery_order_number, gr.receipt_date, gr.notes, gr.created_at
		FROM goods_receipts gr
		JOIN purchase_orders po ON gr.purchase_order_id = po.id
		ORDER BY gr.id DESC
	`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar GR: %w", err)
	}
	defer rows.Close()

	var list []*GoodsReceipt
	for rows.Next() {
		gr := &GoodsReceipt{}
		if err := rows.Scan(
			&gr.ID,
			&gr.GRNumber,
			&gr.PurchaseOrderID,
			&gr.PONumber,
			&gr.DeliveryOrderNumber,
			&gr.ReceiptDate,
			&gr.Notes,
			&gr.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("gagal memindai GR: %w", err)
		}
		list = append(list, gr)
	}
	return list, nil
}

// GetStockSummary menghitung saldo stok saat ini dari akumulasi buku besar (Inventory Ledger)
func (r *pgRepository) GetStockSummary(ctx context.Context) ([]*StockSummary, error) {
	query := `
		SELECT p.id, p.name, p.sku, p.unit,
		       COALESCE(SUM(CASE WHEN it.transaction_type = 'IN' THEN it.quantity ELSE 0 END), 0) AS total_in,
		       COALESCE(SUM(CASE WHEN it.transaction_type = 'OUT' THEN it.quantity ELSE 0 END), 0) AS total_out
		FROM products p
		LEFT JOIN inventory_transactions it ON p.id = it.product_id
		GROUP BY p.id, p.name, p.sku, p.unit
		ORDER BY p.name ASC
	`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil saldo stok: %w", err)
	}
	defer rows.Close()

	var summaries []*StockSummary
	for rows.Next() {
		s := &StockSummary{}
		if err := rows.Scan(
			&s.ProductID,
			&s.ProductName,
			&s.ProductSKU,
			&s.ProductUnit,
			&s.TotalIn,
			&s.TotalOut,
		); err != nil {
			return nil, fmt.Errorf("gagal memindai saldo stok: %w", err)
		}
		s.CurrentStock = s.TotalIn - s.TotalOut
		summaries = append(summaries, s)
	}
	return summaries, nil
}

// ListTransactions mengambil catatan buku besar (Stock Ledger)
func (r *pgRepository) ListTransactions(ctx context.Context, productID *int64) ([]*InventoryTransaction, error) {
	query := `
		SELECT it.id, it.product_id, p.name, p.sku, p.unit,
		       it.transaction_type, it.reference_type, it.reference_id, it.reference_number,
		       it.quantity, COALESCE(it.notes, ''), it.created_at
		FROM inventory_transactions it
		JOIN products p ON it.product_id = p.id
	`
	var rows pgx.Rows
	var err error

	if productID != nil && *productID > 0 {
		query += " WHERE it.product_id = $1 ORDER BY it.id DESC"
		rows, err = r.db.Pool.Query(ctx, query, *productID)
	} else {
		query += " ORDER BY it.id DESC"
		rows, err = r.db.Pool.Query(ctx, query)
	}

	if err != nil {
		return nil, fmt.Errorf("gagal mengambil riwayat transaksi stok: %w", err)
	}
	defer rows.Close()

	var list []*InventoryTransaction
	for rows.Next() {
		tx := &InventoryTransaction{}
		if err := rows.Scan(
			&tx.ID,
			&tx.ProductID,
			&tx.ProductName,
			&tx.ProductSKU,
			&tx.ProductUnit,
			&tx.TransactionType,
			&tx.ReferenceType,
			&tx.ReferenceID,
			&tx.ReferenceNumber,
			&tx.Quantity,
			&tx.Notes,
			&tx.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("gagal memindai transaksi stok: %w", err)
		}
		list = append(list, tx)
	}
	return list, nil
}
