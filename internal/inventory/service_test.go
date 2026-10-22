package inventory

import (
	"context"
	"fmt"
	"testing"

	"agro-supply-erp/internal/procurement"
)

// Mock Procurement Repo for testing inventory
type mockProcurementRepo struct {
	pos map[int64]*procurement.PurchaseOrder
}

func (m *mockProcurementRepo) CreatePR(ctx context.Context, pr *procurement.PurchaseRequest) error { return nil }
func (m *mockProcurementRepo) GetPRByID(ctx context.Context, id int64) (*procurement.PurchaseRequest, error) { return nil, nil }
func (m *mockProcurementRepo) ListPR(ctx context.Context) ([]*procurement.PurchaseRequest, error) { return nil, nil }
func (m *mockProcurementRepo) UpdatePRStatus(ctx context.Context, id int64, status string) error { return nil }
func (m *mockProcurementRepo) NextPRNumber(ctx context.Context) (string, error) { return "", nil }

func (m *mockProcurementRepo) CreatePO(ctx context.Context, po *procurement.PurchaseOrder) error { return nil }
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
func (m *mockProcurementRepo) NextPONumber(ctx context.Context) (string, error) { return "", nil }

// Mock Inventory Repo
type mockInventoryRepo struct {
	grs          map[int64]*GoodsReceipt
	transactions []*InventoryTransaction
	lastGRID     int64
	procRepo     *mockProcurementRepo
}

func newMockInventoryRepo(procRepo *mockProcurementRepo) *mockInventoryRepo {
	return &mockInventoryRepo{
		grs:      make(map[int64]*GoodsReceipt),
		procRepo: procRepo,
	}
}

func (m *mockInventoryRepo) NextGRNumber(ctx context.Context) (string, error) {
	m.lastGRID++
	return fmt.Sprintf("GR-202610-%04d", m.lastGRID), nil
}

func (m *mockInventoryRepo) GetCumulativeReceivedForPO(ctx context.Context, poID int64) (map[int64]float64, error) {
	result := make(map[int64]float64)
	for _, gr := range m.grs {
		if gr.PurchaseOrderID == poID {
			for _, item := range gr.Items {
				result[item.PurchaseOrderItemID] += item.AcceptedQuantity
			}
		}
	}
	return result, nil
}

func (m *mockInventoryRepo) CreateGoodsReceipt(ctx context.Context, gr *GoodsReceipt, newPOStatus string) error {
	gr.ID = m.lastGRID
	m.grs[gr.ID] = gr

	// Catat ke ledger
	for _, item := range gr.Items {
		if item.AcceptedQuantity > 0 {
			m.transactions = append(m.transactions, &InventoryTransaction{
				ID:              int64(len(m.transactions) + 1),
				ProductID:       item.ProductID,
				TransactionType: TxTypeIn,
				ReferenceType:   RefGoodsReceipt,
				ReferenceID:     gr.ID,
				ReferenceNumber: gr.GRNumber,
				Quantity:        item.AcceptedQuantity,
			})
		}
	}

	// Update status PO
	if newPOStatus != "" {
		_ = m.procRepo.UpdatePOStatus(ctx, gr.PurchaseOrderID, newPOStatus)
	}

	return nil
}

func (m *mockInventoryRepo) GetGoodsReceiptByID(ctx context.Context, id int64) (*GoodsReceipt, error) {
	gr, ok := m.grs[id]
	if !ok {
		return nil, ErrNotFound
	}
	return gr, nil
}

func (m *mockInventoryRepo) ListGoodsReceipts(ctx context.Context) ([]*GoodsReceipt, error) {
	var list []*GoodsReceipt
	for _, gr := range m.grs {
		list = append(list, gr)
	}
	return list, nil
}

func (m *mockInventoryRepo) GetStockSummary(ctx context.Context) ([]*StockSummary, error) {
	stockMap := make(map[int64]float64)
	for _, tx := range m.transactions {
		if tx.TransactionType == TxTypeIn {
			stockMap[tx.ProductID] += tx.Quantity
		} else if tx.TransactionType == TxTypeOut {
			stockMap[tx.ProductID] -= tx.Quantity
		}
	}

	var summaries []*StockSummary
	for prodID, stock := range stockMap {
		summaries = append(summaries, &StockSummary{
			ProductID:    prodID,
			CurrentStock: stock,
		})
	}
	return summaries, nil
}

func (m *mockInventoryRepo) ListTransactions(ctx context.Context, productID *int64) ([]*InventoryTransaction, error) {
	return m.transactions, nil
}

func TestGoodsReceipt_PartialAndCompleteReceipt(t *testing.T) {
	// Skenario dari ARCHITECTURE.md:
	// PO: 7.000 kg Batok Kelapa
	// Truk 1: Dikirim 6.000 kg, Lolos 5.800 kg, Ditolak 200 kg -> PO partially_received, Ledger +5.800
	// Truk 2: Dikirim 1.200 kg, Lolos 1.200 kg, Ditolak 0 kg   -> PO completed, Ledger +1.200 -> Total Stok 7.000 kg
	procRepo := &mockProcurementRepo{
		pos: map[int64]*procurement.PurchaseOrder{
			1: {
				ID:       1,
				PONumber: "PO-202610-0001",
				Status:   procurement.POStatusApproved,
				Items: []procurement.PurchaseOrderItem{
					{ID: 101, ProductID: 1, ProductName: "Batok Kelapa", Quantity: 7000, UnitPrice: 2000},
				},
			},
		},
	}

	invRepo := newMockInventoryRepo(procRepo)
	service := NewService(invRepo, procRepo)
	ctx := context.Background()

	// 1. Truk 1 Tiba (Penerimaan Parsial)
	gr1, err := service.CreateGoodsReceipt(ctx, CreateGRRequest{
		PurchaseOrderID:     1,
		DeliveryOrderNumber: "SJ-SUPPLIER-001",
		Notes:               "Truk 1 - Pengiriman sebagian",
		Items: []CreateGRItemRequest{
			{
				PurchaseOrderItemID: 101,
				ProductID:           1,
				ReceivedQuantity:    6000,
				AcceptedQuantity:    5800,
				RejectedQuantity:    200,
				RejectionReason:     "Basah & bercampur pasir",
			},
		},
	})
	if err != nil {
		t.Fatalf("seharusnya penerimaan parsial 1 berhasil: %v", err)
	}

	// Status PO harus berubah menjadi PARTIALLY_RECEIVED
	po := procRepo.pos[1]
	if po.Status != procurement.POStatusPartiallyReceived {
		t.Errorf("expected PO status PARTIALLY_RECEIVED, got: %s", po.Status)
	}

	// Stok di ledger hanya boleh bertambah 5.800 (yang ditolak tidak boleh masuk stok)
	stockSummary, _ := invRepo.GetStockSummary(ctx)
	if len(stockSummary) != 1 || stockSummary[0].CurrentStock != 5800 {
		t.Errorf("expected current stock 5800, got: %v", stockSummary[0].CurrentStock)
	}
	if gr1.Items[0].AcceptedQuantity != 5800 {
		t.Errorf("expected accepted quantity 5800, got: %v", gr1.Items[0].AcceptedQuantity)
	}

	// 2. Truk 2 Tiba (Penerimaan Sisa 1.200 kg)
	_, err = service.CreateGoodsReceipt(ctx, CreateGRRequest{
		PurchaseOrderID:     1,
		DeliveryOrderNumber: "SJ-SUPPLIER-002",
		Notes:               "Truk 2 - Pelunasan sisa pesanan",
		Items: []CreateGRItemRequest{
			{
				PurchaseOrderItemID: 101,
				ProductID:           1,
				ReceivedQuantity:    1200,
				AcceptedQuantity:    1200,
				RejectedQuantity:    0,
			},
		},
	})
	if err != nil {
		t.Fatalf("seharusnya penerimaan sisa berhasil: %v", err)
	}

	// Status PO sekarang harus COMPLETED karena 5.800 + 1.200 = 7.000 kg (lunas)
	po = procRepo.pos[1]
	if po.Status != procurement.POStatusCompleted {
		t.Errorf("expected PO status COMPLETED, got: %s", po.Status)
	}

	// Total stok di buku besar (ledger) sekarang harus genap 7.000 kg
	stockSummary, _ = invRepo.GetStockSummary(ctx)
	if stockSummary[0].CurrentStock != 7000 {
		t.Errorf("expected total stock 7000, got: %v", stockSummary[0].CurrentStock)
	}
}

func TestGoodsReceipt_ValidationRules(t *testing.T) {
	procRepo := &mockProcurementRepo{
		pos: map[int64]*procurement.PurchaseOrder{
			1: {
				ID:     1,
				Status: procurement.POStatusDraft, // Masih DRAFT
				Items: []procurement.PurchaseOrderItem{
					{ID: 101, ProductID: 1, Quantity: 1000},
				},
			},
			2: {
				ID:     2,
				Status: procurement.POStatusApproved,
				Items: []procurement.PurchaseOrderItem{
					{ID: 201, ProductID: 1, Quantity: 500},
				},
			},
		},
	}

	invRepo := newMockInventoryRepo(procRepo)
	service := NewService(invRepo, procRepo)
	ctx := context.Background()

	// 1. Terima barang untuk PO yang masih DRAFT -> HARUS DITOLAK
	_, err := service.CreateGoodsReceipt(ctx, CreateGRRequest{
		PurchaseOrderID: 1, // status DRAFT
		Items: []CreateGRItemRequest{
			{PurchaseOrderItemID: 101, ProductID: 1, ReceivedQuantity: 1000, AcceptedQuantity: 1000},
		},
	})
	if err == nil {
		t.Errorf("seharusnya error ketika menerima barang untuk PO DRAFT")
	}

	// 2. Terima barang melebihi kuantitas PO (Over-receipt) -> HARUS DITOLAK
	_, err = service.CreateGoodsReceipt(ctx, CreateGRRequest{
		PurchaseOrderID: 2, // PO quantity 500
		Items: []CreateGRItemRequest{
			{PurchaseOrderItemID: 201, ProductID: 1, ReceivedQuantity: 600, AcceptedQuantity: 600}, // melebihi 500
		},
	})
	if err == nil {
		t.Errorf("seharusnya error ketika kuantitas barang melebihi pesanan PO")
	}
}
