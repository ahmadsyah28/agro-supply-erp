package product

import(
	"context"
	"errors"
	"fmt"
	"time"
	"agro-supply-erp/internal/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrNotFound  = errors.New("produk tidak ditemukan")
	ErrSKUExists = errors.New("SKU produk sudah terdaftar")
)

type Repository interface{
	Create(ctx context.Context, p *Product) error
	GetByID(ctx context.Context, id int64) (*Product, error)
	GetBySKU(ctx context.Context, sku string) (*Product, error)
	List(ctx context.Context) ([]*Product, error)
	Update(ctx context.Context, p *Product) error
	Delete(ctx context.Context, id int64) error
}

// pgRepository adalah implementasi Repository menggunakan PostgreSQL
type pgRepository struct {
	db *database.DB
}

// NewRepository membuat instance Repository baru
func NewRepository(db *database.DB) Repository {
	return &pgRepository{db: db}
}

// Create menyimpan data produk baru ke tabel products
func (r *pgRepository) Create(ctx context.Context, p *Product) error {
	query := `
		INSERT INTO products (sku, name, unit, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	now := time.Now().UTC()
	p.CreatedAt = now
	p.UpdatedAt = now
	if !p.IsActive {
		p.IsActive = true // default produk baru aktif
	}
	err := r.db.Pool.QueryRow(ctx, query,
		p.SKU,
		p.Name,
		p.Unit,
		p.IsActive,
		p.CreatedAt,
		p.UpdatedAt,
	).Scan(&p.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		// Kode PostgreSQL 23505 adalah unique_violation (SKU duplikat)
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrSKUExists
		}
		return fmt.Errorf("gagal membuat produk: %w", err)
	}
	return nil
}

// GetByID mengambil produk berdasarkan ID
func (r *pgRepository) GetByID(ctx context.Context, id int64) (*Product, error) {
	query := `
		SELECT id, sku, name, unit, is_active, created_at, updated_at
		FROM products
		WHERE id = $1
	`
	product := &Product{}
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&product.ID,
		&product.SKU,
		&product.Name,
		&product.Unit,
		&product.IsActive,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("gagal mengambil produk: %w", err)
	}
	return product, nil
}

// GetBySKU mengambil produk berdasarkan SKU (kode unik)
func (r *pgRepository) GetBySKU(ctx context.Context, sku string) (*Product, error) {
	query := `
		SELECT id, sku, name, unit, is_active, created_at, updated_at
		FROM products
		WHERE sku = $1
	`
	p := &Product{}
	err := r.db.Pool.QueryRow(ctx, query, sku).Scan(
		&p.ID,
		&p.SKU,
		&p.Name,
		&p.Unit,
		&p.IsActive,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("gagal mengambil produk sku %s: %w", sku, err)
	}
	return p, nil
}

// List mengambil seluruh daftar produk
func (r *pgRepository) List(ctx context.Context) ([]*Product, error) {
	query := `
		SELECT id, sku, name, unit, is_active, created_at, updated_at
		FROM products
		ORDER BY id DESC
	`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar produk: %w", err)
	}
	defer rows.Close()
	var products []*Product
	for rows.Next() {
		p := &Product{}
		if err := rows.Scan(
			&p.ID,
			&p.SKU,
			&p.Name,
			&p.Unit,
			&p.IsActive,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("gagal memindai baris produk: %w", err)
		}
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error pada iterasi baris produk: %w", err)
	}
	return products, nil
}

// Update memperbarui data produk
func (r *pgRepository) Update(ctx context.Context, p *Product) error {
	query := `
		UPDATE products
		SET name = $1, unit = $2, is_active = $3, updated_at = $4
		WHERE id = $5
	`

	p.UpdatedAt = time.Now().UTC()
	cmdTag, err := r.db.Pool.Exec(ctx, query,
		p.Name,
		p.Unit,
		p.IsActive,
		p.UpdatedAt,
		p.ID,
	)

	if err != nil {
		return fmt.Errorf("gagal mengupdate produk: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// Delete menonaktifkan produk (Soft Delete sesuai prinsip arsitektur)
func (r *pgRepository) Delete(ctx context.Context, id int64) error {
	query := `
		UPDATE products
		SET is_active = FALSE, updated_at = $1
		WHERE id = $2
	`
	cmdTag, err := r.db.Pool.Exec(ctx, query, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("gagal menonaktifkan produk: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
