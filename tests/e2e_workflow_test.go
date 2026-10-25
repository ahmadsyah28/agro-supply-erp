package tests

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"agro-supply-erp/internal/auth"
	"agro-supply-erp/internal/inventory"
	"agro-supply-erp/internal/middleware"
	"agro-supply-erp/internal/procurement"
	"agro-supply-erp/internal/product"
	"agro-supply-erp/internal/supplier"
)

// --- Mocks untuk Pengujian Alur Penuh End-to-End ---

type mockProductRepo struct {
	items  map[int64]*product.Product
	lastID int64
}

func (m *mockProductRepo) Create(ctx context.Context, p *product.Product) error {
	m.lastID++
	p.ID = m.lastID
	m.items[p.ID] = p
	return nil
}
func (m *mockProductRepo) GetByID(ctx context.Context, id int64) (*product.Product, error) {
	p, ok := m.items[id]
	if !ok {
		return nil, product.ErrNotFound
	}
	return p, nil
}
func (m *mockProductRepo) GetBySKU(ctx context.Context, sku string) (*product.Product, error) { return nil, nil }
func (m *mockProductRepo) List(ctx context.Context) ([]*product.Product, error)              { return nil, nil }
func (m *mockProductRepo) Update(ctx context.Context, p *product.Product) error             { return nil }
func (m *mockProductRepo) Delete(ctx context.Context, id int64) error                       { return nil }

type mockSupplierRepo struct {
	items  map[int64]*supplier.Supplier
	lastID int64
}

func (m *mockSupplierRepo) Create(ctx context.Context, s *supplier.Supplier) error {
	m.lastID++
	s.ID = m.lastID
	m.items[s.ID] = s
	return nil
}
func (m *mockSupplierRepo) GetByID(ctx context.Context, id int64) (*supplier.Supplier, error) {
	s, ok := m.items[id]
	if !ok {
		return nil, supplier.ErrNotFound
	}
	return s, nil
}
func (m *mockSupplierRepo) List(ctx context.Context) ([]*supplier.Supplier, error)       { return nil, nil }
func (m *mockSupplierRepo) ListActive(ctx context.Context) ([]*supplier.Supplier, error) { return nil, nil }
func (m *mockSupplierRepo) Update(ctx context.Context, s *supplier.Supplier) error      { return nil }
func (m *mockSupplierRepo) Delete(ctx context.Context, id int64) error                  { return nil }

type mockProcurementRepo struct {
	prs    map[int64]*procurement.PurchaseRequest
	pos    map[int64]*procurement.PurchaseOrder
	lastPR int64
	lastPO int64
}

func (m *mockProcurementRepo) NextPRNumber(ctx context.Context) (string, error) {
	m.lastPR++
	return fmt.Sprintf("PR-202610-%04d", m.lastPR), nil
}
func (m *mockProcurementRepo) NextPONumber(ctx context.Context) (string, error) {
	m.lastPO++
	return fmt.Sprintf("PO-202610-%04d", m.lastPO), nil
}
func (m *mockProcurementRepo) CreatePR(ctx context.Context, pr *procurement.PurchaseRequest) error {
	pr.ID = m.lastPR
	m.prs[pr.ID] = pr
	return nil
}
func (m *mockProcurementRepo) GetPRByID(ctx context.Context, id int64) (*procurement.PurchaseRequest, error) {
	pr, ok := m.prs[id]
	if !ok {
		return nil, procurement.ErrNotFound
	}
	return pr, nil
}
func (m *mockProcurementRepo) ListPR(ctx context.Context) ([]*procurement.PurchaseRequest, error) { return nil, nil }
func (m *mockProcurementRepo) UpdatePRStatus(ctx context.Context, id int64, status string) error {
	if pr, ok := m.prs[id]; ok {
		pr.Status = status
	}
	return nil
}
func (m *mockProcurementRepo) CreatePO(ctx context.Context, po *procurement.PurchaseOrder) error {
	po.ID = m.lastPO
	for i := range po.Items {
		po.Items[i].ID = int64(100 + i + 1)
	}
	m.pos[po.ID] = po
	return nil
}
func (m *mockProcurementRepo) GetPOByID(ctx context.Context, id int64) (*procurement.PurchaseOrder, error) {
	po, ok := m.pos[id]
	if !ok {
		return nil, procurement.ErrNotFound
	}
	return po, nil
}
func (m *mockProcurementRepo) ListPO(ctx context.Context) ([]*procurement.PurchaseOrder, error) { return nil, nil }
func (m *mockProcurementRepo) UpdatePOStatus(ctx context.Context, id int64, status string) error {
	if po, ok := m.pos[id]; ok {
		po.Status = status
	}
	return nil
}

type mockInventoryRepo struct {
	grs      map[int64]*inventory.GoodsReceipt
	ledger   []*inventory.InventoryTransaction
	lastGR   int64
	procRepo *mockProcurementRepo
}

func (m *mockInventoryRepo) NextGRNumber(ctx context.Context) (string, error) {
	m.lastGR++
	return fmt.Sprintf("GR-202610-%04d", m.lastGR), nil
}
func (m *mockInventoryRepo) GetCumulativeReceivedForPO(ctx context.Context, poID int64) (map[int64]float64, error) {
	res := make(map[int64]float64)
	for _, gr := range m.grs {
		if gr.PurchaseOrderID == poID {
			for _, item := range gr.Items {
				res[item.PurchaseOrderItemID] += item.AcceptedQuantity
			}
		}
	}
	return res, nil
}
func (m *mockInventoryRepo) CreateGoodsReceipt(ctx context.Context, gr *inventory.GoodsReceipt, newPOStatus string) error {
	gr.ID = m.lastGR
	m.grs[gr.ID] = gr
	for _, item := range gr.Items {
		if item.AcceptedQuantity > 0 {
			m.ledger = append(m.ledger, &inventory.InventoryTransaction{
				ID:              int64(len(m.ledger) + 1),
				ProductID:       item.ProductID,
				TransactionType: inventory.TxTypeIn,
				Quantity:        item.AcceptedQuantity,
			})
		}
	}
	if newPOStatus != "" {
		_ = m.procRepo.UpdatePOStatus(ctx, gr.PurchaseOrderID, newPOStatus)
	}
	return nil
}
func (m *mockInventoryRepo) GetGoodsReceiptByID(ctx context.Context, id int64) (*inventory.GoodsReceipt, error) {
	return m.grs[id], nil
}
func (m *mockInventoryRepo) ListGoodsReceipts(ctx context.Context) ([]*inventory.GoodsReceipt, error) { return nil, nil }
func (m *mockInventoryRepo) GetStockSummary(ctx context.Context) ([]*inventory.StockSummary, error) {
	stockMap := make(map[int64]float64)
	for _, tx := range m.ledger {
		if tx.TransactionType == inventory.TxTypeIn {
			stockMap[tx.ProductID] += tx.Quantity
		}
	}
	var res []*inventory.StockSummary
	for id, qty := range stockMap {
		res = append(res, &inventory.StockSummary{ProductID: id, CurrentStock: qty})
	}
	return res, nil
}
func (m *mockInventoryRepo) ListTransactions(ctx context.Context, productID *int64) ([]*inventory.InventoryTransaction, error) {
	return m.ledger, nil
}

// TestEndToEndProcurementToInventoryWorkflow menguji seluruh alur bisnis dari hulu ke hilir
func TestEndToEndProcurementToInventoryWorkflow(t *testing.T) {
	ctx := context.Background()

	// Inisialisasi Repositori
	prodRepo := &mockProductRepo{items: make(map[int64]*product.Product)}
	suppRepo := &mockSupplierRepo{items: make(map[int64]*supplier.Supplier)}
	procRepo := &mockProcurementRepo{
		prs: make(map[int64]*procurement.PurchaseRequest),
		pos: make(map[int64]*procurement.PurchaseOrder),
	}
	invRepo := &mockInventoryRepo{
		grs:      make(map[int64]*inventory.GoodsReceipt),
		procRepo: procRepo,
	}

	// Inisialisasi Service
	prodService := product.NewService(prodRepo)
	suppService := supplier.NewService(suppRepo)
	procService := procurement.NewService(procRepo, prodRepo, suppRepo)
	invService := inventory.NewService(invRepo, procRepo)

	// ==========================================
	// LANGKAH 1: Master Data Produk & Supplier
	// ==========================================
	prod, err := prodService.Create(ctx, product.CreateProductRequest{
		SKU:  "PROD-BATOK-001",
		Name: "Batok Kelapa Kering Grade A",
		Unit: "kg",
	})
	if err != nil {
		t.Fatalf("Langkah 1 Gagal: buat produk: %v", err)
	}

	supp, err := suppService.Create(ctx, supplier.CreateSupplierRequest{
		Name:        "CV Agro Lestari Mandiri",
		ContactName: "Pak Hendra",
		Phone:       "081298765432",
	})
	if err != nil {
		t.Fatalf("Langkah 1 Gagal: buat supplier: %v", err)
	}

	// ==========================================
	// LANGKAH 2: Purchase Request (PR) Workflow
	// ==========================================
	pr, err := procService.CreatePR(ctx, procurement.CreatePRRequest{
		Notes: "Kebutuhan bahan baku produksi arang briket",
		Items: []procurement.CreatePRItemRequest{
			{ProductID: prod.ID, Quantity: 5000, Notes: "Kadar air maks 12%"},
		},
	})
	if err != nil || pr.Status != procurement.PRStatusDraft {
		t.Fatalf("Langkah 2 Gagal: buat PR status DRAFT: %v", err)
	}

	// Submit PR
	pr, err = procService.SubmitPR(ctx, pr.ID)
	if err != nil || pr.Status != procurement.PRStatusSubmitted {
		t.Fatalf("Langkah 2 Gagal: submit PR: %v", err)
	}

	// Approve PR
	pr, err = procService.ApprovePR(ctx, pr.ID)
	if err != nil || pr.Status != procurement.PRStatusApproved {
		t.Fatalf("Langkah 2 Gagal: approve PR: %v", err)
	}

	// ==========================================
	// LANGKAH 3: Purchase Order (PO) Workflow
	// ==========================================
	po, err := procService.CreatePO(ctx, procurement.CreatePORequest{
		PurchaseRequestID: &pr.ID,
		SupplierID:        supp.ID,
		Notes:             "Pengiriman Franco Pabrik",
		Items: []procurement.CreatePOItemRequest{
			{ProductID: prod.ID, Quantity: 5000, UnitPrice: 2500},
		},
	})
	if err != nil || po.Status != procurement.POStatusDraft {
		t.Fatalf("Langkah 3 Gagal: buat PO: %v", err)
	}

	// Approve PO
	po, err = procService.ApprovePO(ctx, po.ID)
	if err != nil || po.Status != procurement.POStatusApproved {
		t.Fatalf("Langkah 3 Gagal: approve PO: %v", err)
	}

	poItemID := po.Items[0].ID

	// =======================================================
	// LANGKAH 4: Truk 1 Tiba (Penerimaan Parsial 3.000 kg)
	// =======================================================
	_, err = invService.CreateGoodsReceipt(ctx, inventory.CreateGRRequest{
		PurchaseOrderID:     po.ID,
		DeliveryOrderNumber: "SJ-001/ALM/2026",
		Notes:               "Truk 1 - Muatan 3.000 kg",
		Items: []inventory.CreateGRItemRequest{
			{
				PurchaseOrderItemID: poItemID,
				ProductID:           prod.ID,
				ReceivedQuantity:    3000,
				AcceptedQuantity:    2900,
				RejectedQuantity:    100,
				RejectionReason:     "100 kg basah kehujanan",
			},
		},
	})
	if err != nil {
		t.Fatalf("Langkah 4 Gagal: penerimaan parsial 1: %v", err)
	}

	// Cek Status PO harus PARTIALLY_RECEIVED
	poAfterTruk1 := procRepo.pos[po.ID]
	if poAfterTruk1.Status != procurement.POStatusPartiallyReceived {
		t.Errorf("expected PO status PARTIALLY_RECEIVED, got: %s", poAfterTruk1.Status)
	}

	// Cek Stok di Buku Besar harus 2.900 kg (hanya yang lolos)
	stock, _ := invService.GetStockSummary(ctx)
	if stock[0].CurrentStock != 2900 {
		t.Errorf("expected stok setelah truk 1 adalah 2900, got: %v", stock[0].CurrentStock)
	}

	// =======================================================
	// LANGKAH 5: Truk 2 Tiba (Penerimaan Sisa 2.100 kg Lunas)
	// =======================================================
	_, err = invService.CreateGoodsReceipt(ctx, inventory.CreateGRRequest{
		PurchaseOrderID:     po.ID,
		DeliveryOrderNumber: "SJ-002/ALM/2026",
		Notes:               "Truk 2 - Pelunasan sisa pesanan",
		Items: []inventory.CreateGRItemRequest{
			{
				PurchaseOrderItemID: poItemID,
				ProductID:           prod.ID,
				ReceivedQuantity:    2100,
				AcceptedQuantity:    2100,
				RejectedQuantity:    0,
			},
		},
	})
	if err != nil {
		t.Fatalf("Langkah 5 Gagal: penerimaan sisa 2: %v", err)
	}

	// Cek Status PO sekarang harus COMPLETED (2.900 + 2.100 = 5.000 kg)
	poAfterTruk2 := procRepo.pos[po.ID]
	if poAfterTruk2.Status != procurement.POStatusCompleted {
		t.Errorf("expected PO status COMPLETED, got: %s", poAfterTruk2.Status)
	}

	// Cek Stok Akhir di Gudang genap 5.000 kg
	stockFinal, _ := invService.GetStockSummary(ctx)
	if stockFinal[0].CurrentStock != 5000 {
		t.Errorf("expected saldo akhir stok 5000, got: %v", stockFinal[0].CurrentStock)
	}

	// Cek Buku Besar (Ledger) memiliki tepat 2 transaksi mutasi masuk
	txs, _ := invService.ListTransactions(ctx, nil)
	if len(txs) != 2 {
		t.Errorf("expected 2 transaksi ledger, got: %d", len(txs))
	}
}

func TestEndToEndRoleBasedAccessControl(t *testing.T) {
	// Setup user claims representing different departments
	purchasingClaims := &auth.JWTClaims{
		UserID:   10,
		Username: "purchasing_officer",
		Role:     auth.RolePurchasing,
	}

	managerClaims := &auth.JWTClaims{
		UserID:   20,
		Username: "procurement_manager",
		Role:     auth.RoleManager,
	}

	warehouseClaims := &auth.JWTClaims{
		UserID:   30,
		Username: "warehouse_officer",
		Role:     auth.RoleWarehouse,
	}

	adminClaims := &auth.JWTClaims{
		UserID:   1,
		Username: "admin",
		Role:     auth.RoleAdmin,
	}

	// Handlers with protection
	approveHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"approved"}`)
	})

	grHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"goods_receipt_created"}`)
	})

	requireManager := middleware.RequireRole(auth.RoleManager)
	requireWarehouse := middleware.RequireRole(auth.RoleWarehouse)

	// 1. Purchasing mencoba menyetujui (approve) PR -> HARUS DITOLAK 403 Forbidden
	req := httptest.NewRequest("POST", "/api/v1/purchase-requests/1/approve", nil)
	req = req.WithContext(auth.SetClaimsToContext(req.Context(), purchasingClaims))
	rr := httptest.NewRecorder()
	requireManager(approveHandler).ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Errorf("Segregation of Duties Violation: Purchasing seharusnya tidak boleh approve PR! Status code: %d", rr.Code)
	}

	// 2. Manager menyetujui (approve) PR -> HARUS BERHASIL 200 OK
	req = httptest.NewRequest("POST", "/api/v1/purchase-requests/1/approve", nil)
	req = req.WithContext(auth.SetClaimsToContext(req.Context(), managerClaims))
	rr = httptest.NewRecorder()
	requireManager(approveHandler).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("Manager seharusnya diizinkan approve PR! Status code: %d", rr.Code)
	}

	// 3. Purchasing mencoba membuat Goods Receipt -> HARUS DITOLAK 403 Forbidden
	req = httptest.NewRequest("POST", "/api/v1/goods-receipts", nil)
	req = req.WithContext(auth.SetClaimsToContext(req.Context(), purchasingClaims))
	rr = httptest.NewRecorder()
	requireWarehouse(grHandler).ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Errorf("Purchasing tidak boleh membuat Goods Receipt! Status code: %d", rr.Code)
	}

	// 4. Warehouse mencatat Goods Receipt fisik & QC -> HARUS BERHASIL 200 OK
	req = httptest.NewRequest("POST", "/api/v1/goods-receipts", nil)
	req = req.WithContext(auth.SetClaimsToContext(req.Context(), warehouseClaims))
	rr = httptest.NewRecorder()
	requireWarehouse(grHandler).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("Warehouse seharusnya diizinkan membuat Goods Receipt! Status code: %d", rr.Code)
	}

	// 5. Admin memiliki akses penuh (Bypass) ke semua aksi
	req = httptest.NewRequest("POST", "/api/v1/purchase-requests/1/approve", nil)
	req = req.WithContext(auth.SetClaimsToContext(req.Context(), adminClaims))
	rr = httptest.NewRecorder()
	requireManager(approveHandler).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("Admin seharusnya diizinkan bypass approve PR! Status code: %d", rr.Code)
	}

	req = httptest.NewRequest("POST", "/api/v1/goods-receipts", nil)
	req = req.WithContext(auth.SetClaimsToContext(req.Context(), adminClaims))
	rr = httptest.NewRecorder()
	requireWarehouse(grHandler).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("Admin seharusnya diizinkan bypass Goods Receipt! Status code: %d", rr.Code)
	}
}

