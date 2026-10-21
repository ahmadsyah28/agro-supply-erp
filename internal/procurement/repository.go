package procurement

import (
	"context"
	"errors"
	"fmt"
	"time"

	"agro-supply-erp/internal/database"

	"github.com/jackc/pgx/v5"
)

var (
	ErrNotFound = errors.New("dokumen tidak ditemukan")
)

// Repository adalah kontrak interface akses database untuk modul procurement
type Repository interface {
	// Purchase Request
	CreatePR(ctx context.Context, pr *PurchaseRequest) error
	GetPRByID(ctx context.Context, id int64) (*PurchaseRequest, error)
	ListPR(ctx context.Context) ([]*PurchaseRequest, error)
	UpdatePRStatus(ctx context.Context, id int64, status string) error

	// Purchase Order
	CreatePO(ctx context.Context, po *PurchaseOrder) error
	GetPOByID(ctx context.Context, id int64) (*PurchaseOrder, error)
	ListPO(ctx context.Context) ([]*PurchaseOrder, error)
	UpdatePOStatus(ctx context.Context, id int64, status string) error

	// Counter generator untuk nomor unik dokumen
	NextPRNumber(ctx context.Context) (string, error)
	NextPONumber(ctx context.Context) (string, error)
}

type pgRepository struct {
	db *database.DB
}

func NewRepository(db *database.DB) Repository {
	return &pgRepository{db: db}
}

// NextPRNumber menghasilkan nomor PR berurutan, misal PR-202610-0001
func (r *pgRepository) NextPRNumber(ctx context.Context) (string, error) {
	var count int64
	err := r.db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM purchase_requests").Scan(&count)
	if err != nil {
		return "", err
	}
	now := time.Now().UTC()
	return fmt.Sprintf("PR-%s-%04d", now.Format("200601"), count+1), nil
}

// NextPONumber menghasilkan nomor PO berurutan, misal PO-202610-0001
func (r *pgRepository) NextPONumber(ctx context.Context) (string, error) {
	var count int64
	err := r.db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM purchase_orders").Scan(&count)
	if err != nil {
		return "", err
	}
	now := time.Now().UTC()
	return fmt.Sprintf("PO-%s-%04d", now.Format("200601"), count+1), nil
}

// CreatePR menyimpan header PR dan rincian item dalam 1 TRANSAKSI ATOMIK
func (r *pgRepository) CreatePR(ctx context.Context, pr *PurchaseRequest) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("gagal memulai transaksi database: %w", err)
	}
	defer tx.Rollback(ctx)

	now := time.Now().UTC()
	pr.CreatedAt = now
	pr.UpdatedAt = now
	pr.Status = PRStatusDraft

	// 1. Simpan Header PR
	queryHeader := `
		INSERT INTO purchase_requests (pr_number, status, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	if err := tx.QueryRow(ctx, queryHeader, pr.PRNumber, pr.Status, pr.Notes, pr.CreatedAt, pr.UpdatedAt).Scan(&pr.ID); err != nil {
		return fmt.Errorf("gagal menyimpan header PR: %w", err)
	}

	// 2. Simpan Item-item PR
	queryItem := `
		INSERT INTO purchase_request_items (purchase_request_id, product_id, quantity, notes, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	for i := range pr.Items {
		pr.Items[i].PurchaseRequestID = pr.ID
		pr.Items[i].CreatedAt = now
		if err := tx.QueryRow(ctx, queryItem,
			pr.ID,
			pr.Items[i].ProductID,
			pr.Items[i].Quantity,
			pr.Items[i].Notes,
			pr.Items[i].CreatedAt,
		).Scan(&pr.Items[i].ID); err != nil {
			return fmt.Errorf("gagal menyimpan item PR baris %d: %w", i+1, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("gagal commit transaksi PR: %w", err)
	}

	return nil
}

// GetPRByID mengambil header PR beserta seluruh item barangnya (JOIN)
func (r *pgRepository) GetPRByID(ctx context.Context, id int64) (*PurchaseRequest, error) {
	queryHeader := `
		SELECT id, pr_number, status, notes, created_at, updated_at
		FROM purchase_requests
		WHERE id = $1
	`
	pr := &PurchaseRequest{}
	err := r.db.Pool.QueryRow(ctx, queryHeader, id).Scan(
		&pr.ID,
		&pr.PRNumber,
		&pr.Status,
		&pr.Notes,
		&pr.CreatedAt,
		&pr.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("gagal mengambil PR id %d: %w", id, err)
	}

	// Ambil rincian item dengan info nama & SKU produk
	queryItems := `
		SELECT i.id, i.purchase_request_id, i.product_id, p.name, p.sku, i.quantity, i.notes, i.created_at
		FROM purchase_request_items i
		JOIN products p ON i.product_id = p.id
		WHERE i.purchase_request_id = $1
		ORDER BY i.id ASC
	`
	rows, err := r.db.Pool.Query(ctx, queryItems, id)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil item PR: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item PurchaseRequestItem
		if err := rows.Scan(
			&item.ID,
			&item.PurchaseRequestID,
			&item.ProductID,
			&item.ProductName,
			&item.ProductSKU,
			&item.Quantity,
			&item.Notes,
			&item.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("gagal memindai item PR: %w", err)
		}
		pr.Items = append(pr.Items, item)
	}

	return pr, nil
}

// ListPR mengambil seluruh daftar dokumen PR
func (r *pgRepository) ListPR(ctx context.Context) ([]*PurchaseRequest, error) {
	query := `
		SELECT id, pr_number, status, notes, created_at, updated_at
		FROM purchase_requests
		ORDER BY id DESC
	`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar PR: %w", err)
	}
	defer rows.Close()

	var list []*PurchaseRequest
	for rows.Next() {
		pr := &PurchaseRequest{}
		if err := rows.Scan(
			&pr.ID,
			&pr.PRNumber,
			&pr.Status,
			&pr.Notes,
			&pr.CreatedAt,
			&pr.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("gagal memindai PR: %w", err)
		}
		list = append(list, pr)
	}
	return list, nil
}

// UpdatePRStatus memperbarui status PR (SUBMITTED, APPROVED, REJECTED)
func (r *pgRepository) UpdatePRStatus(ctx context.Context, id int64, status string) error {
	query := `
		UPDATE purchase_requests
		SET status = $1, updated_at = $2
		WHERE id = $3
	`
	cmdTag, err := r.db.Pool.Exec(ctx, query, status, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("gagal mengupdate status PR: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// CreatePO menyimpan PO dan item-itemnya secara atomik dalam 1 transaksi
func (r *pgRepository) CreatePO(ctx context.Context, po *PurchaseOrder) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("gagal memulai transaksi database: %w", err)
	}
	defer tx.Rollback(ctx)

	now := time.Now().UTC()
	po.CreatedAt = now
	po.UpdatedAt = now
	po.Status = POStatusDraft

	// Hitung total amount dari subtotal item
	var total float64
	for i := range po.Items {
		po.Items[i].Subtotal = po.Items[i].Quantity * po.Items[i].UnitPrice
		total += po.Items[i].Subtotal
	}
	po.TotalAmount = total

	// 1. Simpan Header PO
	queryHeader := `
		INSERT INTO purchase_orders (po_number, purchase_request_id, supplier_id, status, total_amount, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`
	if err := tx.QueryRow(ctx, queryHeader,
		po.PONumber,
		po.PurchaseRequestID,
		po.SupplierID,
		po.Status,
		po.TotalAmount,
		po.Notes,
		po.CreatedAt,
		po.UpdatedAt,
	).Scan(&po.ID); err != nil {
		return fmt.Errorf("gagal menyimpan header PO: %w", err)
	}

	// 2. Simpan Item-item PO
	queryItem := `
		INSERT INTO purchase_order_items (purchase_order_id, product_id, quantity, unit_price, subtotal, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	for i := range po.Items {
		po.Items[i].PurchaseOrderID = po.ID
		po.Items[i].CreatedAt = now
		if err := tx.QueryRow(ctx, queryItem,
			po.ID,
			po.Items[i].ProductID,
			po.Items[i].Quantity,
			po.Items[i].UnitPrice,
			po.Items[i].Subtotal,
			po.Items[i].CreatedAt,
		).Scan(&po.Items[i].ID); err != nil {
			return fmt.Errorf("gagal menyimpan item PO baris %d: %w", i+1, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("gagal commit transaksi PO: %w", err)
	}

	return nil
}

// GetPOByID mengambil PO lengkap beserta item dan nama supplier
func (r *pgRepository) GetPOByID(ctx context.Context, id int64) (*PurchaseOrder, error) {
	queryHeader := `
		SELECT po.id, po.po_number, po.purchase_request_id, po.supplier_id, s.name, po.status, po.total_amount, po.notes, po.created_at, po.updated_at
		FROM purchase_orders po
		JOIN suppliers s ON po.supplier_id = s.id
		WHERE po.id = $1
	`
	po := &PurchaseOrder{}
	err := r.db.Pool.QueryRow(ctx, queryHeader, id).Scan(
		&po.ID,
		&po.PONumber,
		&po.PurchaseRequestID,
		&po.SupplierID,
		&po.SupplierName,
		&po.Status,
		&po.TotalAmount,
		&po.Notes,
		&po.CreatedAt,
		&po.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("gagal mengambil PO id %d: %w", id, err)
	}

	queryItems := `
		SELECT i.id, i.purchase_order_id, i.product_id, p.name, p.sku, i.quantity, i.unit_price, i.subtotal, i.created_at
		FROM purchase_order_items i
		JOIN products p ON i.product_id = p.id
		WHERE i.purchase_order_id = $1
		ORDER BY i.id ASC
	`
	rows, err := r.db.Pool.Query(ctx, queryItems, id)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil item PO: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item PurchaseOrderItem
		if err := rows.Scan(
			&item.ID,
			&item.PurchaseOrderID,
			&item.ProductID,
			&item.ProductName,
			&item.ProductSKU,
			&item.Quantity,
			&item.UnitPrice,
			&item.Subtotal,
			&item.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("gagal memindai item PO: %w", err)
		}
		po.Items = append(po.Items, item)
	}

	return po, nil
}

// ListPO mengambil daftar seluruh dokumen PO
func (r *pgRepository) ListPO(ctx context.Context) ([]*PurchaseOrder, error) {
	query := `
		SELECT po.id, po.po_number, po.purchase_request_id, po.supplier_id, s.name, po.status, po.total_amount, po.notes, po.created_at, po.updated_at
		FROM purchase_orders po
		JOIN suppliers s ON po.supplier_id = s.id
		ORDER BY po.id DESC
	`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar PO: %w", err)
	}
	defer rows.Close()

	var list []*PurchaseOrder
	for rows.Next() {
		po := &PurchaseOrder{}
		if err := rows.Scan(
			&po.ID,
			&po.PONumber,
			&po.PurchaseRequestID,
			&po.SupplierID,
			&po.SupplierName,
			&po.Status,
			&po.TotalAmount,
			&po.Notes,
			&po.CreatedAt,
			&po.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("gagal memindai PO: %w", err)
		}
		list = append(list, po)
	}
	return list, nil
}

// UpdatePOStatus memperbarui status PO (APPROVED, CANCELLED, dll)
func (r *pgRepository) UpdatePOStatus(ctx context.Context, id int64, status string) error {
	query := `
		UPDATE purchase_orders
		SET status = $1, updated_at = $2
		WHERE id = $3
	`
	cmdTag, err := r.db.Pool.Exec(ctx, query, status, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("gagal mengupdate status PO: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
