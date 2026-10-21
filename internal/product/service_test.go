package product

import (
	"context"
	"testing"
)

// mockRepository adalah tiruan (mock) dari interface Repository untuk unit testing
type mockRepository struct {
	products map[int64]*Product
	lastID   int64
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		products: make(map[int64]*Product),
	}
}

func (m *mockRepository) Create(ctx context.Context, p *Product) error {
	for _, existing := range m.products {
		if existing.SKU == p.SKU {
			return ErrSKUExists
		}
	}
	m.lastID++
	p.ID = m.lastID
	m.products[p.ID] = p
	return nil
}

func (m *mockRepository) GetByID(ctx context.Context, id int64) (*Product, error) {
	p, ok := m.products[id]
	if !ok {
		return nil, ErrNotFound
	}
	return p, nil
}

func (m *mockRepository) GetBySKU(ctx context.Context, sku string) (*Product, error) {
	for _, p := range m.products {
		if p.SKU == sku {
			return p, nil
		}
	}
	return nil, ErrNotFound
}

func (m *mockRepository) List(ctx context.Context) ([]*Product, error) {
	var list []*Product
	for _, p := range m.products {
		list = append(list, p)
	}
	return list, nil
}

func (m *mockRepository) Update(ctx context.Context, p *Product) error {
	if _, ok := m.products[p.ID]; !ok {
		return ErrNotFound
	}
	m.products[p.ID] = p
	return nil
}

func (m *mockRepository) Delete(ctx context.Context, id int64) error {
	p, ok := m.products[id]
	if !ok {
		return ErrNotFound
	}
	p.IsActive = false
	return nil
}

func TestProductService_Create(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo)
	ctx := context.Background()

	// 1. Sukses membuat produk
	req := CreateProductRequest{
		SKU:  "PROD-001",
		Name: "Batok Kelapa Grade A",
		Unit: "kg",
	}

	p, err := service.Create(ctx, req)
	if err != nil {
		t.Fatalf("seharusnya tidak ada error, didapat: %v", err)
	}

	if p.ID != 1 {
		t.Errorf("expected ID 1, got %d", p.ID)
	}
	if p.SKU != "PROD-001" {
		t.Errorf("expected SKU PROD-001, got %s", p.SKU)
	}
	if !p.IsActive {
		t.Errorf("expected IsActive true, got false")
	}

	// 2. Gagal jika SKU kosong (Validasi)
	invalidReq := CreateProductRequest{
		SKU:  "",
		Name: "Batok Kelapa",
		Unit: "kg",
	}
	_, err = service.Create(ctx, invalidReq)
	if err == nil {
		t.Errorf("seharusnya error ketika SKU kosong")
	}

	// 3. Gagal jika SKU duplikat
	_, err = service.Create(ctx, req)
	if err != ErrSKUExists {
		t.Errorf("expected ErrSKUExists, got %v", err)
	}
}

func TestProductService_GetByID(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo)
	ctx := context.Background()

	_, _ = service.Create(ctx, CreateProductRequest{
		SKU:  "PROD-002",
		Name: "Arang Batok Kelapa",
		Unit: "ton",
	})

	// 1. Sukses ambil ID valid
	p, err := service.GetByID(ctx, 1)
	if err != nil {
		t.Fatalf("seharusnya berhasil mengambil produk, didapat: %v", err)
	}
	if p.Name != "Arang Batok Kelapa" {
		t.Errorf("nama produk tidak sesuai: %s", p.Name)
	}

	// 2. Gagal jika ID tidak ada
	_, err = service.GetByID(ctx, 999)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}

	// 3. Gagal jika ID tidak valid (<= 0)
	_, err = service.GetByID(ctx, 0)
	if err != ErrInvalidID {
		t.Errorf("expected ErrInvalidID, got %v", err)
	}
}
