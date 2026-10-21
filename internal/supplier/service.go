package supplier

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrInvalidID = errors.New("ID supplier tidak valid")
)

// Service adalah kontrak fungsi logika bisnis untuk modul supplier
type Service interface {
	Create(ctx context.Context, req CreateSupplierRequest) (*Supplier, error)
	GetByID(ctx context.Context, id int64) (*Supplier, error)
	List(ctx context.Context, onlyActive bool) ([]*Supplier, error)
	Update(ctx context.Context, id int64, req UpdateSupplierRequest) (*Supplier, error)
	Delete(ctx context.Context, id int64) error
}

// supplierService adalah implementasi konkret dari Service
type supplierService struct {
	repo Repository
}

// NewService membuat instance Service baru dengan menyuntikkan Repository
func NewService(repo Repository) Service {
	return &supplierService{
		repo: repo,
	}
}

// Create memproses pendaftaran supplier baru setelah memvalidasi input
func (s *supplierService) Create(ctx context.Context, req CreateSupplierRequest) (*Supplier, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	supplier := &Supplier{
		Name:        req.Name,
		ContactName: req.ContactName,
		Email:       req.Email,
		Phone:       req.Phone,
		Address:     req.Address,
		IsActive:    true,
	}

	if err := s.repo.Create(ctx, supplier); err != nil {
		return nil, err
	}

	return supplier, nil
}

// GetByID mengambil supplier berdasarkan ID
func (s *supplierService) GetByID(ctx context.Context, id int64) (*Supplier, error) {
	if id <= 0 {
		return nil, ErrInvalidID
	}

	return s.repo.GetByID(ctx, id)
}

// List mengambil daftar supplier (dengan opsi filter hanya yang aktif)
func (s *supplierService) List(ctx context.Context, onlyActive bool) ([]*Supplier, error) {
	if onlyActive {
		return s.repo.ListActive(ctx)
	}
	return s.repo.List(ctx)
}

// Update memproses perubahan data profil supplier
func (s *supplierService) Update(ctx context.Context, id int64, req UpdateSupplierRequest) (*Supplier, error) {
	if id <= 0 {
		return nil, ErrInvalidID
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	// 1. Pastikan supplier yang ingin diubah terdaftar
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 2. Terapkan perubahan
	existing.Name = req.Name
	existing.ContactName = req.ContactName
	existing.Email = req.Email
	existing.Phone = req.Phone
	existing.Address = req.Address
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

	// 3. Simpan perubahan ke database via repository
	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("gagal mengupdate supplier: %w", err)
	}

	return existing, nil
}

// Delete menonaktifkan supplier (soft-delete)
func (s *supplierService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrInvalidID
	}

	return s.repo.Delete(ctx, id)
}
