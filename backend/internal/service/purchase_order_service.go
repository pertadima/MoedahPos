package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/dto"
	"github.com/moedahpos/backend/internal/repository"
	"github.com/rs/zerolog"
)

// Purchase Order sentinel errors.
var (
	ErrPONotFound          = errors.New("purchase order not found")
	ErrPONotEditable       = errors.New("purchase order cannot be edited (not in draft status)")
	ErrPOCannotSubmit      = errors.New("purchase order cannot be submitted")
	ErrPOCannotReceive     = errors.New("purchase order cannot be received (must be in ordered status)")
	ErrPOCannotCancel      = errors.New("purchase order cannot be cancelled")
)

// PurchaseOrderService implements PO lifecycle business logic.
type PurchaseOrderService struct {
	poRepo          repository.PurchaseOrderRepository
	productRepo     repository.ProductRepository
	paymentRepo     repository.POPaymentRepository
	priceHistorySvc *PriceHistoryService
	log             zerolog.Logger
}

func NewPurchaseOrderService(
	poRepo repository.PurchaseOrderRepository,
	productRepo repository.ProductRepository,
	paymentRepo repository.POPaymentRepository,
	priceHistorySvc *PriceHistoryService,
	log zerolog.Logger,
) *PurchaseOrderService {
	return &PurchaseOrderService{poRepo: poRepo, productRepo: productRepo, paymentRepo: paymentRepo, priceHistorySvc: priceHistorySvc, log: log}
}

func (s *PurchaseOrderService) ListPOs(ctx context.Context, filter dto.POListFilter) ([]*dto.POResponse, dto.PaginationMeta, error) {
	filter.Defaults()
	pos, total, err := s.poRepo.FindAll(ctx, filter)
	if err != nil {
		return nil, dto.PaginationMeta{}, fmt.Errorf("listing POs: %w", err)
	}
	// Enrich with payment aggregation
	s.paymentRepo.PopulatePOPayments(ctx, pos)
	resp := make([]*dto.POResponse, 0, len(pos))
	for _, po := range pos {
		resp = append(resp, toPOResponse(po))
	}
	return resp, dto.NewMeta(filter.PaginationQuery, total), nil
}

func (s *PurchaseOrderService) GetPO(ctx context.Context, id string) (*dto.POResponse, error) {
	po, err := s.poRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("finding PO: %w", err)
	}
	if po == nil {
		return nil, ErrPONotFound
	}
	// Enrich with payment aggregation
	paid, status, err := s.paymentRepo.AggregateByPO(ctx, po.ID, po.TotalAmount)
	if err == nil {
		po.AmountPaid = paid
		po.PaymentStatus = status
	}
	return toPOResponse(po), nil
}

func (s *PurchaseOrderService) CreatePO(ctx context.Context, storeID string, req *dto.CreatePORequest, userID string) (*dto.POResponse, error) {
	items, totalAmt, err := s.buildItems(ctx, req.Items, storeID)
	if err != nil {
		return nil, err
	}

	poNumber := fmt.Sprintf("PO-%s-%04d", time.Now().Format("20060102"), rand.Intn(9000)+1000)
	po, err := s.poRepo.Create(ctx, &domain.PurchaseOrder{
		StoreID:     storeID,
		SupplierID:  req.SupplierID,
		PONumber:    poNumber,
		TotalAmount: totalAmt,
		OrderedBy:   userID,
		Notes:       req.Notes,
	}, items)
	if err != nil {
		return nil, fmt.Errorf("creating PO: %w", err)
	}
	s.log.Info().Str("po_id", po.ID).Str("po_number", po.PONumber).Msg("PO created")
	return toPOResponse(po), nil
}

func (s *PurchaseOrderService) UpdatePO(ctx context.Context, id string, req *dto.UpdatePORequest, storeID string) (*dto.POResponse, error) {
	existing, err := s.poRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("finding PO: %w", err)
	}
	if existing == nil {
		return nil, ErrPONotFound
	}
	if existing.Status != "draft" {
		return nil, ErrPONotEditable
	}

	items, totalAmt, err := s.buildItems(ctx, req.Items, storeID)
	if err != nil {
		return nil, err
	}

	existing.SupplierID = req.SupplierID
	existing.Notes = req.Notes
	existing.TotalAmount = totalAmt

	po, err := s.poRepo.Update(ctx, existing, items)
	if err != nil {
		return nil, fmt.Errorf("updating PO: %w", err)
	}
	if po == nil {
		return nil, ErrPONotEditable
	}
	return toPOResponse(po), nil
}

// SubmitPO changes PO status from draft → ordered.
func (s *PurchaseOrderService) SubmitPO(ctx context.Context, id, userID string) error {
	po, err := s.poRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("finding PO: %w", err)
	}
	if po == nil {
		return ErrPONotFound
	}
	if po.Status != "draft" {
		return ErrPOCannotSubmit
	}
	if err := s.poRepo.Submit(ctx, id, userID); err != nil {
		return fmt.Errorf("submitting PO: %w", err)
	}
	s.log.Info().Str("po_id", id).Msg("PO submitted")
	return nil
}

// ReceivePO changes PO status to received and updates stock levels.
func (s *PurchaseOrderService) ReceivePO(ctx context.Context, id, userID string) error {
	po, err := s.poRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("finding PO: %w", err)
	}
	if po == nil {
		return ErrPONotFound
	}
	if po.Status != "ordered" {
		return ErrPOCannotReceive
	}
	if err := s.poRepo.Receive(ctx, id, userID); err != nil {
		return fmt.Errorf("receiving PO: %w", err)
	}

	// Record cost price changes from PO items
	if s.priceHistorySvc != nil {
		poID := id
		for _, item := range po.Items {
			product, err := s.productRepo.FindByID(ctx, item.ProductID)
			if err != nil || product == nil {
				continue
			}
			if product.CostPrice == item.UnitCost {
				continue // no cost change
			}
			notes := fmt.Sprintf("PO %s received — qty %.2f @ %.2f", po.PONumber, item.ReceivedQty, item.UnitCost)
			_ = s.priceHistorySvc.RecordChange(ctx,
				item.ProductID, po.StoreID, userID,
				product.CostPrice, item.UnitCost,
				product.SellPrice, product.SellPrice, // sell price unchanged
				"purchase_order", &poID, &notes,
			)
		}
	}

	s.log.Info().Str("po_id", id).Str("received_by", userID).Msg("PO received — stock updated")
	return nil
}

// CancelPO cancels a PO (soft — sets status to 'cancelled').
func (s *PurchaseOrderService) CancelPO(ctx context.Context, id string) error {
	po, err := s.poRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("finding PO: %w", err)
	}
	if po == nil {
		return ErrPONotFound
	}
	if po.Status == "received" || po.Status == "cancelled" {
		return ErrPOCannotCancel
	}
	return s.poRepo.Cancel(ctx, id)
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func (s *PurchaseOrderService) buildItems(ctx context.Context, inputs []dto.POItemInput, storeID string) ([]domain.POItem, float64, error) {
	var items []domain.POItem
	var totalAmt float64
	for _, input := range inputs {
		product, err := s.productRepo.FindByID(ctx, input.ProductID)
		if err != nil {
			return nil, 0, fmt.Errorf("finding product: %w", err)
		}
		if product == nil || product.StoreID != storeID {
			return nil, 0, fmt.Errorf("%w: product %s", ErrProductNotFound, input.ProductID)
		}
		subtotal := input.Quantity * input.UnitCost
		totalAmt += subtotal
		items = append(items, domain.POItem{
			ProductID: input.ProductID,
			Quantity:  input.Quantity,
			UnitCost:  input.UnitCost,
			Subtotal:  subtotal,
		})
	}
	return items, totalAmt, nil
}

// ─── Mapper ───────────────────────────────────────────────────────────────────

func toPOResponse(po *domain.PurchaseOrder) *dto.POResponse {
	resp := &dto.POResponse{
		ID:             po.ID,
		StoreID:        po.StoreID,
		SupplierID:     po.SupplierID,
		SupplierName:   po.SupplierName,
		PONumber:       po.PONumber,
		Status:         po.Status,
		TotalAmount:    po.TotalAmount,
		AmountPaid:     po.AmountPaid,
		AmountDue:      po.TotalAmount - po.AmountPaid,
		PaymentStatus:  po.PaymentStatus,
		OrderedByName:  po.OrderedByName,
		ReceivedByName: po.ReceivedByName,
		Notes:          po.Notes,
		CreatedAt:      po.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      po.UpdatedAt.Format(time.RFC3339),
	}
	if po.PaymentStatus == "" {
		resp.PaymentStatus = "unpaid"
	}
	if po.OrderedAt != nil {
		t := po.OrderedAt.Format(time.RFC3339)
		resp.OrderedAt = &t
	}
	if po.ReceivedAt != nil {
		t := po.ReceivedAt.Format(time.RFC3339)
		resp.ReceivedAt = &t
	}
	for _, item := range po.Items {
		resp.Items = append(resp.Items, dto.POItemResponse{
			ID:          item.ID,
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			ProductSKU:  item.ProductSKU,
			Unit:        item.Unit,
			Quantity:    item.Quantity,
			UnitCost:    item.UnitCost,
			ReceivedQty: item.ReceivedQty,
			Subtotal:    item.Subtotal,
		})
	}
	return resp
}

// ─── Payment methods ───────────────────────────────────────────────────────────

// CreatePayment records a payment against a received PO.
func (s *PurchaseOrderService) CreatePayment(ctx context.Context, poID, storeID, userID string, req dto.POPaymentRequest) (*dto.POPaymentResponse, error) {
	po, err := s.poRepo.FindByID(ctx, poID)
	if err != nil {
		return nil, fmt.Errorf("finding PO: %w", err)
	}
	if po == nil {
		return nil, ErrPONotFound
	}
	if po.Status != "received" {
		return nil, fmt.Errorf("hutang hanya bisa dicatat untuk PO yang sudah diterima")
	}

	note := req.Note
	var notePtr *string
	if note != "" {
		notePtr = &note
	}

	out, err := s.paymentRepo.Create(ctx, domain.POPayment{
		POID:    poID,
		StoreID: storeID,
		Amount:  req.Amount,
		Note:    notePtr,
		PaidBy:  userID,
	})
	if err != nil {
		return nil, fmt.Errorf("recording payment: %w", err)
	}
	return toPaymentResponse(out), nil
}

// ListPayments returns all payments for a PO.
func (s *PurchaseOrderService) ListPayments(ctx context.Context, poID string) ([]*dto.POPaymentResponse, error) {
	rows, err := s.paymentRepo.FindByPO(ctx, poID)
	if err != nil {
		return nil, fmt.Errorf("listing payments: %w", err)
	}
	out := make([]*dto.POPaymentResponse, 0, len(rows))
	for _, r := range rows {
		out = append(out, toPaymentResponse(r))
	}
	return out, nil
}

// PayableSummary returns aggregate debt metrics for the store.
func (s *PurchaseOrderService) PayableSummary(ctx context.Context, storeID string) (*dto.PayableSummary, error) {
	return s.paymentRepo.PayableSummary(ctx, storeID)
}

func toPaymentResponse(p *domain.POPayment) *dto.POPaymentResponse {
	return &dto.POPaymentResponse{
		ID:         p.ID,
		POID:       p.POID,
		Amount:     p.Amount,
		Note:       p.Note,
		PaidBy:     p.PaidBy,
		PaidByName: p.PaidByName,
		PaidAt:     p.PaidAt.Format(time.RFC3339),
	}
}
