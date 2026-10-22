package inventory

import (
	"context"
	"errors"
	"fmt"

	"agro-supply-erp/internal/procurement"
)

var (
	ErrInvalidID      = errors.New("ID dokumen tidak valid")
	ErrPOInvalidState = errors.New("barang hanya dapat diterima untuk Purchase Order yang berstatus APPROVED atau PARTIALLY_RECEIVED")
	ErrPOItemNotFound = errors.New("baris item Purchase Order tidak ditemukan dalam PO terkait")
	ErrOverReceipt    = errors.New("kuantitas barang yang diterima melebihi sisa kuantitas terbuka pada PO")
)

// Service kontrak logika bisnis untuk inventory dan penerimaan barang
type Service interface {
	CreateGoodsReceipt(ctx context.Context, req CreateGRRequest) (*GoodsReceipt, error)
	GetGoodsReceiptByID(ctx context.Context, id int64) (*GoodsReceipt, error)
	ListGoodsReceipts(ctx context.Context) ([]*GoodsReceipt, error)
	GetStockSummary(ctx context.Context) ([]*StockSummary, error)
	ListTransactions(ctx context.Context, productID *int64) ([]*InventoryTransaction, error)
}

type inventoryService struct {
	repo            Repository
	procurementRepo procurement.Repository
}

func NewService(repo Repository, procurementRepo procurement.Repository) Service {
	return &inventoryService{
		repo:            repo,
		procurementRepo: procurementRepo,
	}
}

// CreateGoodsReceipt memvalidasi PO, menghitung status parsial/penuh, dan mencatat penerimaan stok
func (s *inventoryService) CreateGoodsReceipt(ctx context.Context, req CreateGRRequest) (*GoodsReceipt, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// 1. Ambil & Validasi Status PO
	po, err := s.procurementRepo.GetPOByID(ctx, req.PurchaseOrderID)
	if err != nil {
		return nil, fmt.Errorf("Purchase Order id %d tidak ditemukan: %w", req.PurchaseOrderID, err)
	}

	if po.Status != procurement.POStatusApproved && po.Status != procurement.POStatusPartiallyReceived {
		return nil, fmt.Errorf("%w (Status PO saat ini: %s)", ErrPOInvalidState, po.Status)
	}

	// Petakan item-item PO untuk pencocokan cepat
	poItemMap := make(map[int64]procurement.PurchaseOrderItem)
	for _, poItem := range po.Items {
		poItemMap[poItem.ID] = poItem
	}

	// 2. Ambil riwayat penerimaan sebelumnya untuk PO ini
	priorReceived, err := s.repo.GetCumulativeReceivedForPO(ctx, po.ID)
	if err != nil {
		return nil, err
	}

	// 3. Validasi setiap baris item penerimaan
	newReceiptTotals := make(map[int64]float64)
	for k, v := range priorReceived {
		newReceiptTotals[k] = v
	}

	for _, reqItem := range req.Items {
		poItem, ok := poItemMap[reqItem.PurchaseOrderItemID]
		if !ok || poItem.ProductID != reqItem.ProductID {
			return nil, fmt.Errorf("%w: item ID %d", ErrPOItemNotFound, reqItem.PurchaseOrderItemID)
		}

		currentCumulative := newReceiptTotals[reqItem.PurchaseOrderItemID]
		if currentCumulative+reqItem.AcceptedQuantity > poItem.Quantity {
			return nil, fmt.Errorf("%w pada produk %s (Dipesan: %.2f, Sudah Diterima: %.2f, Diterima Sekarang: %.2f)",
				ErrOverReceipt, poItem.ProductName, poItem.Quantity, currentCumulative, reqItem.AcceptedQuantity)
		}

		newReceiptTotals[reqItem.PurchaseOrderItemID] += reqItem.AcceptedQuantity
	}

	// 4. Tentukan Status PO Baru (COMPLETED vs PARTIALLY_RECEIVED)
	allCompleted := true
	for _, poItem := range po.Items {
		accepted := newReceiptTotals[poItem.ID]
		// Jika kuantitas yang diterima masih kurang dari kuantitas pesanan, berarti masih parsial
		if accepted < poItem.Quantity {
			allCompleted = false
			break
		}
	}

	newPOStatus := procurement.POStatusPartiallyReceived
	if allCompleted {
		newPOStatus = procurement.POStatusCompleted
	}

	// 5. Generate nomor GR baru
	grNumber, err := s.repo.NextGRNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal menghasilkan nomor Goods Receipt: %w", err)
	}

	gr := &GoodsReceipt{
		GRNumber:            grNumber,
		PurchaseOrderID:     po.ID,
		DeliveryOrderNumber: req.DeliveryOrderNumber,
		Notes:               req.Notes,
	}

	for _, item := range req.Items {
		gr.Items = append(gr.Items, GoodsReceiptItem{
			PurchaseOrderItemID: item.PurchaseOrderItemID,
			ProductID:           item.ProductID,
			ReceivedQuantity:    item.ReceivedQuantity,
			AcceptedQuantity:    item.AcceptedQuantity,
			RejectedQuantity:    item.RejectedQuantity,
			RejectionReason:     item.RejectionReason,
		})
	}

	// 6. Simpan dokumen, catat ledger, dan update status PO dalam 1 transaksi
	if err := s.repo.CreateGoodsReceipt(ctx, gr, newPOStatus); err != nil {
		return nil, err
	}

	return s.repo.GetGoodsReceiptByID(ctx, gr.ID)
}

func (s *inventoryService) GetGoodsReceiptByID(ctx context.Context, id int64) (*GoodsReceipt, error) {
	if id <= 0 {
		return nil, ErrInvalidID
	}
	return s.repo.GetGoodsReceiptByID(ctx, id)
}

func (s *inventoryService) ListGoodsReceipts(ctx context.Context) ([]*GoodsReceipt, error) {
	return s.repo.ListGoodsReceipts(ctx)
}

func (s *inventoryService) GetStockSummary(ctx context.Context) ([]*StockSummary, error) {
	return s.repo.GetStockSummary(ctx)
}

func (s *inventoryService) ListTransactions(ctx context.Context, productID *int64) ([]*InventoryTransaction, error) {
	return s.repo.ListTransactions(ctx, productID)
}
