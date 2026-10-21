package procurement

import (
	"context"
	"fmt"
	"testing"

	"agro-supply-erp/internal/product"
	"agro-supply-erp/internal/supplier"
)

// mockProcurementRepo implementasi mock Repository
type mockProcurementRepo struct {
	prs     map[int64]*PurchaseRequest
	pos     map[int64]*PurchaseOrder
	lastPR  int64
	lastPO  int64
}

func newMockProcurementRepo() *mockProcurementRepo {
	return &mockProcurementRepo{
		prs: make(map[int64]*PurchaseRequest),
		pos: make(map[int64]*PurchaseOrder),
	}
}

func (m *mockProcurementRepo) NextPRNumber(ctx context.Context) (string, error) {
	m.lastPR++
	return fmt.Sprintf("PR-202610-%04d", m.lastPR), nil
}

func (m *mockProcurementRepo) NextPONumber(ctx context.Context) (string, error) {
	m.lastPO++
	return fmt.Sprintf("PO-202610-%04d", m.lastPO), nil
}

func (m *mockProcurementRepo) CreatePR(ctx context.Context, pr *PurchaseRequest) error {
	pr.ID = m.lastPR
	m.prs[pr.ID] = pr
	return nil
}

func (m *mockProcurementRepo) GetPRByID(ctx context.Context, id int64) (*PurchaseRequest, error) {
	pr, ok := m.prs[id]
	if !ok {
		return nil, ErrNotFound
	}
	return pr, nil
}

func (m *mockProcurementRepo) ListPR(ctx context.Context) ([]*PurchaseRequest, error) {
	var list []*PurchaseRequest
	for _, pr := range m.prs {
		list = append(list, pr)
	}
	return list, nil
}

func (m *mockProcurementRepo) UpdatePRStatus(ctx context.Context, id int64, status string) error {
	pr, ok := m.prs[id]
	if !ok {
		return ErrNotFound
	}
	pr.Status = status
	return nil
}

func (m *mockProcurementRepo) CreatePO(ctx context.Context, po *PurchaseOrder) error {
	po.ID = m.lastPO
	m.pos[po.ID] = po
	return nil
}

func (m *mockProcurementRepo) GetPOByID(ctx context.Context, id int64) (*PurchaseOrder, error) {
	po, ok := m.pos[id]
	if !ok {
		return nil, ErrNotFound
	}
	return po, nil
}

func (m *mockProcurementRepo) ListPO(ctx context.Context) ([]*PurchaseOrder, error) {
	var list []*PurchaseOrder
	for _, po := range m.pos {
		list = append(list, po)
	}
	return list, nil
}

func (m *mockProcurementRepo) UpdatePOStatus(ctx context.Context, id int64, status string) error {
	po, ok := m.pos[id]
	if !ok {
		return ErrNotFound
	}
	po.Status = status
	return nil
}

// Mock Product Repo
type mockProductRepo struct {
	products map[int64]*product.Product
}

func (m *mockProductRepo) Create(ctx context.Context, p *product.Product) error { return nil }
func (m *mockProductRepo) GetByID(ctx context.Context, id int64) (*product.Product, error) {
	p, ok := m.products[id]
	if !ok {
		return nil, product.ErrNotFound
	}
	return p, nil
}
func (m *mockProductRepo) GetBySKU(ctx context.Context, sku string) (*product.Product, error) { return nil, nil }
func (m *mockProductRepo) List(ctx context.Context) ([]*product.Product, error)              { return nil, nil }
func (m *mockProductRepo) Update(ctx context.Context, p *product.Product) error             { return nil }
func (m *mockProductRepo) Delete(ctx context.Context, id int64) error                       { return nil }

// Mock Supplier Repo
type mockSupplierRepo struct {
	suppliers map[int64]*supplier.Supplier
}

func (m *mockSupplierRepo) Create(ctx context.Context, s *supplier.Supplier) error { return nil }
func (m *mockSupplierRepo) GetByID(ctx context.Context, id int64) (*supplier.Supplier, error) {
	s, ok := m.suppliers[id]
	if !ok {
		return nil, supplier.ErrNotFound
	}
	return s, nil
}
func (m *mockSupplierRepo) List(ctx context.Context) ([]*supplier.Supplier, error)       { return nil, nil }
func (m *mockSupplierRepo) ListActive(ctx context.Context) ([]*supplier.Supplier, error) { return nil, nil }
func (m *mockSupplierRepo) Update(ctx context.Context, s *supplier.Supplier) error      { return nil }
func (m *mockSupplierRepo) Delete(ctx context.Context, id int64) error                  { return nil }

func TestPR_Workflow(t *testing.T) {
	procRepo := newMockProcurementRepo()
	prodRepo := &mockProductRepo{
		products: map[int64]*product.Product{
			1: {ID: 1, Name: "Batok Kelapa", IsActive: true},
		},
	}
	suppRepo := &mockSupplierRepo{}
	service := NewService(procRepo, prodRepo, suppRepo)
	ctx := context.Background()

	// 1. Buat PR baru
	pr, err := service.CreatePR(ctx, CreatePRRequest{
		Notes: "Kebutuhan bahan baku mingguan",
		Items: []CreatePRItemRequest{
			{ProductID: 1, Quantity: 5000, Notes: "Kadar air maks 15%"},
		},
	})
	if err != nil {
		t.Fatalf("seharusnya berhasil membuat PR, didapat error: %v", err)
	}
	if pr.Status != PRStatusDraft {
		t.Errorf("expected status DRAFT, got %s", pr.Status)
	}

	// 2. Submit PR
	pr, err = service.SubmitPR(ctx, pr.ID)
	if err != nil {
		t.Fatalf("seharusnya berhasil submit PR: %v", err)
	}
	if pr.Status != PRStatusSubmitted {
		t.Errorf("expected status SUBMITTED, got %s", pr.Status)
	}

	// 3. Approve PR
	pr, err = service.ApprovePR(ctx, pr.ID)
	if err != nil {
		t.Fatalf("seharusnya berhasil approve PR: %v", err)
	}
	if pr.Status != PRStatusApproved {
		t.Errorf("expected status APPROVED, got %s", pr.Status)
	}
}

func TestPO_ValidationRules(t *testing.T) {
	procRepo := newMockProcurementRepo()
	prodRepo := &mockProductRepo{
		products: map[int64]*product.Product{
			1: {ID: 1, Name: "Batok Kelapa", IsActive: true},
		},
	}
	suppRepo := &mockSupplierRepo{
		suppliers: map[int64]*supplier.Supplier{
			1: {ID: 1, Name: "Supplier Aktif", IsActive: true},
			2: {ID: 2, Name: "Supplier Nonaktif", IsActive: false},
		},
	}
	service := NewService(procRepo, prodRepo, suppRepo)
	ctx := context.Background()

	// Siapkan 1 PR berstatus DRAFT dan 1 PR berstatus APPROVED
	draftPR, _ := service.CreatePR(ctx, CreatePRRequest{
		Notes: "Draft PR",
		Items: []CreatePRItemRequest{{ProductID: 1, Quantity: 1000}},
	})

	approvedPR, _ := service.CreatePR(ctx, CreatePRRequest{
		Notes: "Approved PR",
		Items: []CreatePRItemRequest{{ProductID: 1, Quantity: 2000}},
	})
	_, _ = service.SubmitPR(ctx, approvedPR.ID)
	_, _ = service.ApprovePR(ctx, approvedPR.ID)

	// Uji Coba 1: Buat PO ke Supplier Nonaktif -> HARUS DITOLAK
	_, err := service.CreatePO(ctx, CreatePORequest{
		SupplierID: 2, // nonaktif
		Items:      []CreatePOItemRequest{{ProductID: 1, Quantity: 500, UnitPrice: 2000}},
	})
	if err != ErrSupplierNotActive {
		t.Errorf("expected ErrSupplierNotActive, got: %v", err)
	}

	// Uji Coba 2: Buat PO dari PR yang belum diapprove (masih DRAFT) -> HARUS DITOLAK
	_, err = service.CreatePO(ctx, CreatePORequest{
		SupplierID:        1, // aktif
		PurchaseRequestID: &draftPR.ID,
		Items:             []CreatePOItemRequest{{ProductID: 1, Quantity: 500, UnitPrice: 2000}},
	})
	if err != ErrPRNotApproved {
		t.Errorf("expected ErrPRNotApproved, got: %v", err)
	}

	// Uji Coba 3: Buat PO dengan Supplier Aktif & PR Approved -> HARUS SUKSES
	po, err := service.CreatePO(ctx, CreatePORequest{
		SupplierID:        1,
		PurchaseRequestID: &approvedPR.ID,
		Items:             []CreatePOItemRequest{{ProductID: 1, Quantity: 2000, UnitPrice: 2500}},
	})
	if err != nil {
		t.Fatalf("seharusnya berhasil membuat PO: %v", err)
	}
	if po.Status != POStatusDraft {
		t.Errorf("expected PO status DRAFT, got %s", po.Status)
	}

	// Uji Coba 4: Approve PO
	po, err = service.ApprovePO(ctx, po.ID)
	if err != nil {
		t.Fatalf("seharusnya berhasil approve PO: %v", err)
	}
	if po.Status != POStatusApproved {
		t.Errorf("expected PO status APPROVED, got %s", po.Status)
	}
}
