package supplier

import (
	"context"
	"testing"
)

// mockRepository adalah tiruan database untuk unit testing supplier service
type mockRepository struct {
	suppliers map[int64]*Supplier
	lastID    int64
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		suppliers: make(map[int64]*Supplier),
	}
}

func (m *mockRepository) Create(ctx context.Context, s *Supplier) error {
	m.lastID++
	s.ID = m.lastID
	m.suppliers[s.ID] = s
	return nil
}

func (m *mockRepository) GetByID(ctx context.Context, id int64) (*Supplier, error) {
	s, ok := m.suppliers[id]
	if !ok {
		return nil, ErrNotFound
	}
	return s, nil
}

func (m *mockRepository) List(ctx context.Context) ([]*Supplier, error) {
	var list []*Supplier
	for _, s := range m.suppliers {
		list = append(list, s)
	}
	return list, nil
}

func (m *mockRepository) ListActive(ctx context.Context) ([]*Supplier, error) {
	var list []*Supplier
	for _, s := range m.suppliers {
		if s.IsActive {
			list = append(list, s)
		}
	}
	return list, nil
}

func (m *mockRepository) Update(ctx context.Context, s *Supplier) error {
	if _, ok := m.suppliers[s.ID]; !ok {
		return ErrNotFound
	}
	m.suppliers[s.ID] = s
	return nil
}

func (m *mockRepository) Delete(ctx context.Context, id int64) error {
	s, ok := m.suppliers[id]
	if !ok {
		return ErrNotFound
	}
	s.IsActive = false
	return nil
}

func TestSupplierService_Create(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo)
	ctx := context.Background()

	// 1. Sukses buat supplier
	req := CreateSupplierRequest{
		Name:        "CV Agro Makmur",
		ContactName: "Budi Santoso",
		Email:       "budi@agromakmur.com",
		Phone:       "08123456789",
		Address:     "Lampung Selatan",
	}

	s, err := service.Create(ctx, req)
	if err != nil {
		t.Fatalf("seharusnya tidak ada error, didapat: %v", err)
	}

	if s.ID != 1 {
		t.Errorf("expected ID 1, got %d", s.ID)
	}
	if s.Name != "CV Agro Makmur" {
		t.Errorf("expected nama 'CV Agro Makmur', got %s", s.Name)
	}
	if !s.IsActive {
		t.Errorf("expected IsActive true, got false")
	}

	// 2. Gagal jika nama kosong
	invalidReq := CreateSupplierRequest{
		Name: "",
	}
	_, err = service.Create(ctx, invalidReq)
	if err == nil {
		t.Errorf("seharusnya error ketika nama supplier kosong")
	}
}

func TestSupplierService_ListActive(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo)
	ctx := context.Background()

	s1, _ := service.Create(ctx, CreateSupplierRequest{Name: "Supplier Aktif 1"})
	s2, _ := service.Create(ctx, CreateSupplierRequest{Name: "Supplier Nonaktif"})

	// Nonaktifkan s2
	_ = service.Delete(ctx, s2.ID)

	// List semua
	all, err := service.List(ctx, false)
	if err != nil || len(all) != 2 {
		t.Errorf("expected 2 suppliers, got %d", len(all))
	}

	// List hanya yang aktif
	activeOnly, err := service.List(ctx, true)
	if err != nil || len(activeOnly) != 1 {
		t.Errorf("expected 1 active supplier, got %d", len(activeOnly))
	}
	if activeOnly[0].ID != s1.ID {
		t.Errorf("expected active supplier ID %d, got %d", s1.ID, activeOnly[0].ID)
	}
}
