package procurement

import (
	"context"
	"errors"
	"fmt"

	"agro-supply-erp/internal/product"
	"agro-supply-erp/internal/supplier"
)

var (
	ErrInvalidID          = errors.New("ID dokumen tidak valid")
	ErrInvalidStatusFlow  = errors.New("transisi status dokumen tidak sah")
	ErrSupplierNotActive  = errors.New("supplier tidak aktif dan tidak dapat menerima Purchase Order")
	ErrPRNotApproved      = errors.New("Purchase Order hanya dapat dibuat dari Purchase Request yang telah disetujui (APPROVED)")
	ErrProductNotActive   = errors.New("salah satu produk tidak aktif atau tidak ditemukan")
)

// Service kontrak logika bisnis untuk procurement (PR & PO)
type Service interface {
	// Purchase Request
	CreatePR(ctx context.Context, req CreatePRRequest) (*PurchaseRequest, error)
	GetPRByID(ctx context.Context, id int64) (*PurchaseRequest, error)
	ListPR(ctx context.Context) ([]*PurchaseRequest, error)
	SubmitPR(ctx context.Context, id int64) (*PurchaseRequest, error)
	ApprovePR(ctx context.Context, id int64) (*PurchaseRequest, error)
	RejectPR(ctx context.Context, id int64) (*PurchaseRequest, error)

	// Purchase Order
	CreatePO(ctx context.Context, req CreatePORequest) (*PurchaseOrder, error)
	GetPOByID(ctx context.Context, id int64) (*PurchaseOrder, error)
	ListPO(ctx context.Context) ([]*PurchaseOrder, error)
	ApprovePO(ctx context.Context, id int64) (*PurchaseOrder, error)
	CancelPO(ctx context.Context, id int64) (*PurchaseOrder, error)
}

type procurementService struct {
	repo         Repository
	productRepo  product.Repository
	supplierRepo supplier.Repository
}

// NewService membuat instance Service procurement dengan Dependency Injection
func NewService(repo Repository, productRepo product.Repository, supplierRepo supplier.Repository) Service {
	return &procurementService{
		repo:         repo,
		productRepo:  productRepo,
		supplierRepo: supplierRepo,
	}
}

// CreatePR memvalidasi item dan membuat dokumen PR berstatus DRAFT
func (s *procurementService) CreatePR(ctx context.Context, req CreatePRRequest) (*PurchaseRequest, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Validasi semua produk yang diminta benar-benar ada dan aktif
	for _, item := range req.Items {
		prod, err := s.productRepo.GetByID(ctx, item.ProductID)
		if err != nil || !prod.IsActive {
			return nil, fmt.Errorf("%w (ID: %d)", ErrProductNotActive, item.ProductID)
		}
	}

	prNumber, err := s.repo.NextPRNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal menghasilkan nomor PR: %w", err)
	}

	pr := &PurchaseRequest{
		PRNumber: prNumber,
		Notes:    req.Notes,
		Status:   PRStatusDraft,
	}

	for _, item := range req.Items {
		pr.Items = append(pr.Items, PurchaseRequestItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Notes:     item.Notes,
		})
	}

	if err := s.repo.CreatePR(ctx, pr); err != nil {
		return nil, err
	}

	return s.repo.GetPRByID(ctx, pr.ID)
}

func (s *procurementService) GetPRByID(ctx context.Context, id int64) (*PurchaseRequest, error) {
	if id <= 0 {
		return nil, ErrInvalidID
	}
	return s.repo.GetPRByID(ctx, id)
}

func (s *procurementService) ListPR(ctx context.Context) ([]*PurchaseRequest, error) {
	return s.repo.ListPR(ctx)
}

// SubmitPR mengajukan PR dari DRAFT -> SUBMITTED
func (s *procurementService) SubmitPR(ctx context.Context, id int64) (*PurchaseRequest, error) {
	if id <= 0 {
		return nil, ErrInvalidID
	}

	pr, err := s.repo.GetPRByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if pr.Status != PRStatusDraft {
		return nil, fmt.Errorf("%w: hanya PR berstatus DRAFT yang dapat diajukan", ErrInvalidStatusFlow)
	}

	if err := s.repo.UpdatePRStatus(ctx, id, PRStatusSubmitted); err != nil {
		return nil, err
	}

	return s.repo.GetPRByID(ctx, id)
}

// ApprovePR menyetujui PR dari SUBMITTED -> APPROVED
func (s *procurementService) ApprovePR(ctx context.Context, id int64) (*PurchaseRequest, error) {
	if id <= 0 {
		return nil, ErrInvalidID
	}

	pr, err := s.repo.GetPRByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if pr.Status != PRStatusSubmitted {
		return nil, fmt.Errorf("%w: hanya PR berstatus SUBMITTED yang dapat disetujui", ErrInvalidStatusFlow)
	}

	if err := s.repo.UpdatePRStatus(ctx, id, PRStatusApproved); err != nil {
		return nil, err
	}

	return s.repo.GetPRByID(ctx, id)
}

// RejectPR menolak PR dari SUBMITTED -> REJECTED
func (s *procurementService) RejectPR(ctx context.Context, id int64) (*PurchaseRequest, error) {
	if id <= 0 {
		return nil, ErrInvalidID
	}

	pr, err := s.repo.GetPRByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if pr.Status != PRStatusSubmitted {
		return nil, fmt.Errorf("%w: hanya PR berstatus SUBMITTED yang dapat ditolak", ErrInvalidStatusFlow)
	}

	if err := s.repo.UpdatePRStatus(ctx, id, PRStatusRejected); err != nil {
		return nil, err
	}

	return s.repo.GetPRByID(ctx, id)
}

// CreatePO membuat PO dengan validasi supplier aktif & status PR yang disetujui
func (s *procurementService) CreatePO(ctx context.Context, req CreatePORequest) (*PurchaseOrder, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// 1. Validasi Supplier Wajib Aktif (Business Rule Penting)
	supp, err := s.supplierRepo.GetByID(ctx, req.SupplierID)
	if err != nil {
		return nil, fmt.Errorf("supplier id %d tidak ditemukan: %w", req.SupplierID, err)
	}
	if !supp.IsActive {
		return nil, ErrSupplierNotActive
	}

	// 2. Jika PO merujuk ke PR, pastikan PR berstatus APPROVED
	if req.PurchaseRequestID != nil {
		pr, err := s.repo.GetPRByID(ctx, *req.PurchaseRequestID)
		if err != nil {
			return nil, fmt.Errorf("PR id %d tidak ditemukan: %w", *req.PurchaseRequestID, err)
		}
		if pr.Status != PRStatusApproved {
			return nil, ErrPRNotApproved
		}
	}

	// 3. Validasi Produk
	for _, item := range req.Items {
		prod, err := s.productRepo.GetByID(ctx, item.ProductID)
		if err != nil || !prod.IsActive {
			return nil, fmt.Errorf("%w (ID: %d)", ErrProductNotActive, item.ProductID)
		}
	}

	poNumber, err := s.repo.NextPONumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal menghasilkan nomor PO: %w", err)
	}

	po := &PurchaseOrder{
		PONumber:          poNumber,
		PurchaseRequestID: req.PurchaseRequestID,
		SupplierID:        req.SupplierID,
		Notes:             req.Notes,
		Status:            POStatusDraft,
	}

	for _, item := range req.Items {
		po.Items = append(po.Items, PurchaseOrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
		})
	}

	if err := s.repo.CreatePO(ctx, po); err != nil {
		return nil, err
	}

	return s.repo.GetPOByID(ctx, po.ID)
}

func (s *procurementService) GetPOByID(ctx context.Context, id int64) (*PurchaseOrder, error) {
	if id <= 0 {
		return nil, ErrInvalidID
	}
	return s.repo.GetPOByID(ctx, id)
}

func (s *procurementService) ListPO(ctx context.Context) ([]*PurchaseOrder, error) {
	return s.repo.ListPO(ctx)
}

// ApprovePO menyetujui PO dari DRAFT -> APPROVED
func (s *procurementService) ApprovePO(ctx context.Context, id int64) (*PurchaseOrder, error) {
	if id <= 0 {
		return nil, ErrInvalidID
	}

	po, err := s.repo.GetPOByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if po.Status != POStatusDraft {
		return nil, fmt.Errorf("%w: hanya PO berstatus DRAFT yang dapat disetujui", ErrInvalidStatusFlow)
	}

	if err := s.repo.UpdatePOStatus(ctx, id, POStatusApproved); err != nil {
		return nil, err
	}

	return s.repo.GetPOByID(ctx, id)
}

// CancelPO membatalkan PO
func (s *procurementService) CancelPO(ctx context.Context, id int64) (*PurchaseOrder, error) {
	if id <= 0 {
		return nil, ErrInvalidID
	}

	po, err := s.repo.GetPOByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if po.Status == POStatusCompleted || po.Status == POStatusCancelled {
		return nil, fmt.Errorf("%w: dokumen PO yang sudah selesai atau dibatalkan tidak bisa diubah", ErrInvalidStatusFlow)
	}

	if err := s.repo.UpdatePOStatus(ctx, id, POStatusCancelled); err != nil {
		return nil, err
	}

	return s.repo.GetPOByID(ctx, id)
}
