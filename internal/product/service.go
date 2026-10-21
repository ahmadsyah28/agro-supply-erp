package product

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrInvalidID = errors.New("ID produk tidak valid")
)

// Service adalah kontrak fungsi logika bisnis untuk produk
type Service interface {
	Create(ctx context.Context, req CreateProductRequest) (*Product, error)
	GetByID(ctx context.Context, id int64) (*Product, error)
	List(ctx context.Context) ([]*Product, error)
	Update(ctx context.Context, id int64, req UpdateProductRequest) (*Product, error)
	Delete(ctx context.Context, id int64) error
}

// productService adalah implementasi konkret dari Service
type productService struct {
	repo Repository
}

// NewService membuat instance baru dari Service dengan menyuntikkan Repository (Dependency Injection)
func NewService(repo Repository) Service {
	return &productService{
		repo: repo,
	}
}

// Create memproses pembuatan produk baru setelah memvalidasi input
func (s *productService) Create(ctx context.Context, req CreateProductRequest) (*Product, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	p := &Product{
		SKU:      req.SKU,
		Name:     req.Name,
		Unit:     req.Unit,
		IsActive: true,
	}

	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}

	return p, nil
}

// GetByID mengambil produk berdasarkan ID
func (s *productService) GetByID(ctx context.Context, id int64) (*Product, error) {
	if id <= 0 {
		return nil, ErrInvalidID
	}

	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// List mengambil semua produk
func (s *productService) List(ctx context.Context) ([]*Product, error) {
	return s.repo.List(ctx)
}

// Update memproses pembaruan data produk
func (s *productService) Update(ctx context.Context, id int64, req UpdateProductRequest) (*Product, error) {
	if id <= 0 {
		return nil, ErrInvalidID
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	// 1. Pastikan produk yang ingin diubah benar-benar ada
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 2. Terapkan perubahan data
	p.Name = req.Name
	p.Unit = req.Unit
	if req.IsActive != nil {
		p.IsActive = *req.IsActive
	}

	// 3. Simpan perubahan ke database via repository
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, fmt.Errorf("gagal mengupdate produk: %w", err)
	}

	return p, nil
}

// Delete menonaktifkan produk berdasarkan ID
func (s *productService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrInvalidID
	}

	return s.repo.Delete(ctx, id)
}
