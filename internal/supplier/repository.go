package supplier

import (
	"context"
	"errors"
	"fmt"
	"time"

	"agro-supply-erp/internal/database"

	"github.com/jackc/pgx/v5"
)

var (
	ErrNotFound = errors.New("supplier tidak ditemukan")
)

// Repository adalah kontrak interface akses database untuk modul supplier
type Repository interface {
	Create(ctx context.Context, s *Supplier) error
	GetByID(ctx context.Context, id int64) (*Supplier, error)
	List(ctx context.Context) ([]*Supplier, error)
	ListActive(ctx context.Context) ([]*Supplier, error)
	Update(ctx context.Context, s *Supplier) error
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

// Create menyimpan supplier baru ke database
func (r *pgRepository) Create(ctx context.Context, s *Supplier) error {
	query := `
		INSERT INTO suppliers (name, contact_name, email, phone, address, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	now := time.Now().UTC()
	s.CreatedAt = now
	s.UpdatedAt = now
	if !s.IsActive {
		s.IsActive = true // default supplier baru berstatus aktif
	}

	err := r.db.Pool.QueryRow(ctx, query,
		s.Name,
		s.ContactName,
		s.Email,
		s.Phone,
		s.Address,
		s.IsActive,
		s.CreatedAt,
		s.UpdatedAt,
	).Scan(&s.ID)

	if err != nil {
		return fmt.Errorf("gagal membuat supplier: %w", err)
	}

	return nil
}

// GetByID mengambil data supplier berdasarkan ID
func (r *pgRepository) GetByID(ctx context.Context, id int64) (*Supplier, error) {
	query := `
		SELECT id, name, COALESCE(contact_name, ''), COALESCE(email, ''), COALESCE(phone, ''), COALESCE(address, ''), is_active, COALESCE(created_at, CURRENT_TIMESTAMP), COALESCE(updated_at, CURRENT_TIMESTAMP)
		FROM suppliers
		WHERE id = $1
	`

	s := &Supplier{}
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&s.ID,
		&s.Name,
		&s.ContactName,
		&s.Email,
		&s.Phone,
		&s.Address,
		&s.IsActive,
		&s.CreatedAt,
		&s.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("gagal mengambil supplier id %d: %w", id, err)
	}

	return s, nil
}

// List mengambil seluruh daftar supplier
func (r *pgRepository) List(ctx context.Context) ([]*Supplier, error) {
	query := `
		SELECT id, name, COALESCE(contact_name, ''), COALESCE(email, ''), COALESCE(phone, ''), COALESCE(address, ''), is_active, COALESCE(created_at, CURRENT_TIMESTAMP), COALESCE(updated_at, CURRENT_TIMESTAMP)
		FROM suppliers
		ORDER BY id DESC
	`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar supplier: %w", err)
	}
	defer rows.Close()

	var suppliers []*Supplier
	for rows.Next() {
		s := &Supplier{}
		if err := rows.Scan(
			&s.ID,
			&s.Name,
			&s.ContactName,
			&s.Email,
			&s.Phone,
			&s.Address,
			&s.IsActive,
			&s.CreatedAt,
			&s.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("gagal memindai baris supplier: %w", err)
		}
		suppliers = append(suppliers, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error pada iterasi baris supplier: %w", err)
	}

	return suppliers, nil
}

// ListActive mengambil hanya supplier yang aktif (untuk kebutuhan PO/Procurement)
func (r *pgRepository) ListActive(ctx context.Context) ([]*Supplier, error) {
	query := `
		SELECT id, name, COALESCE(contact_name, ''), COALESCE(email, ''), COALESCE(phone, ''), COALESCE(address, ''), is_active, COALESCE(created_at, CURRENT_TIMESTAMP), COALESCE(updated_at, CURRENT_TIMESTAMP)
		FROM suppliers
		WHERE is_active = TRUE
		ORDER BY name ASC
	`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar supplier aktif: %w", err)
	}
	defer rows.Close()

	var suppliers []*Supplier
	for rows.Next() {
		s := &Supplier{}
		if err := rows.Scan(
			&s.ID,
			&s.Name,
			&s.ContactName,
			&s.Email,
			&s.Phone,
			&s.Address,
			&s.IsActive,
			&s.CreatedAt,
			&s.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("gagal memindai baris supplier: %w", err)
		}
		suppliers = append(suppliers, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error pada iterasi baris supplier aktif: %w", err)
	}

	return suppliers, nil
}

// Update memperbarui data profil supplier
func (r *pgRepository) Update(ctx context.Context, s *Supplier) error {
	query := `
		UPDATE suppliers
		SET name = $1, contact_name = $2, email = $3, phone = $4, address = $5, is_active = $6, updated_at = $7
		WHERE id = $8
	`

	s.UpdatedAt = time.Now().UTC()
	cmdTag, err := r.db.Pool.Exec(ctx, query,
		s.Name,
		s.ContactName,
		s.Email,
		s.Phone,
		s.Address,
		s.IsActive,
		s.UpdatedAt,
		s.ID,
	)

	if err != nil {
		return fmt.Errorf("gagal mengupdate supplier: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// Delete menonaktifkan supplier (Soft Delete)
func (r *pgRepository) Delete(ctx context.Context, id int64) error {
	query := `
		UPDATE suppliers
		SET is_active = FALSE, updated_at = $1
		WHERE id = $2
	`

	cmdTag, err := r.db.Pool.Exec(ctx, query, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("gagal menonaktifkan supplier: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
