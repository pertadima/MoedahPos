package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog"

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/dto"
	"github.com/moedahpos/backend/internal/repository"
)

// Sentinel errors ──────────────────────────────────────────────────────────

var (
	ErrTerminNotFound    = errors.New("termin not found")
	ErrTerminOverpayment = errors.New("payment would exceed termin amount")
	ErrPONotReceived     = errors.New("termins can only be created for received purchase orders")
)

// poStatusReceived is used for PO status validation; defined as a constant to
// avoid repeated string literals flagged by goconst linter.
const poStatusReceived = "received"

// TerminService orchestrates the installment payment lifecycle for purchase orders.
//
// Responsibilities:
//   - CreateTerminSchedule: define the payment schedule for a PO.
//   - RecordPayment: apply a payment against one termin (with overpayment guard).
//   - CalculatePODebt: aggregate remaining balance.
//   - GetTerminSchedule: return termins with full payment history.
//   - GenerateDocumentData: assemble all data needed to render a printable document.
type TerminService struct {
	terminRepo  repository.TerminRepository
	paymentRepo repository.PaymentRecordRepository
	poRepo      repository.PurchaseOrderRepository
	storeRepo   repository.StoreRepository
	activitySvc ActivityLogServiceInterface
	log         zerolog.Logger
}

// NewTerminService creates a TerminService with injected repositories.
func NewTerminService(
	terminRepo repository.TerminRepository,
	paymentRepo repository.PaymentRecordRepository,
	poRepo repository.PurchaseOrderRepository,
	storeRepo repository.StoreRepository,
	activitySvc ActivityLogServiceInterface,
	log zerolog.Logger,
) *TerminService {
	return &TerminService{
		terminRepo:  terminRepo,
		paymentRepo: paymentRepo,
		poRepo:      poRepo,
		storeRepo:   storeRepo,
		activitySvc: activitySvc,
		log:         log,
	}
}

// ─── Termin Schedule ─────────────────────────────────────────────────────────

// CreateTerminSchedule replaces the termin schedule for a PO.
// The PO must be in 'received' status (debt only occurs after goods are received).
// Warns (but does not block) if total termin amount != PO total_amount.
func (s *TerminService) CreateTerminSchedule(ctx context.Context, poID string, req dto.CreateTerminScheduleRequest) ([]dto.TerminResponse, error) {
	po, err := s.poRepo.FindByID(ctx, poID)
	if err != nil {
		return nil, fmt.Errorf("TerminService.CreateTerminSchedule find PO: %w", err)
	}
	if po == nil {
		return nil, ErrPONotFound
	}
	if po.Status != poStatusReceived {
		return nil, ErrPONotReceived
	}

	// Build domain termins, parsing due_date strings.
	termins := make([]domain.POTermin, 0, len(req.Termins))
	var totalTermin float64
	for _, t := range req.Termins {
		due, err := time.ParseInLocation("2006-01-02", t.DueDate, time.Local)
		if err != nil {
			return nil, fmt.Errorf("invalid due_date %q: %w", t.DueDate, err)
		}
		termins = append(termins, domain.POTermin{
			POID:         poID,
			TerminNumber: t.TerminNumber,
			Amount:       t.Amount,
			DueDate:      due,
			Notes:        t.Notes,
		})
		totalTermin += t.Amount
	}

	// Warn (log) if total doesn't match PO total — not a hard error.
	if totalTermin != po.TotalAmount {
		s.log.Warn().
			Str("po_id", poID).
			Float64("po_total", po.TotalAmount).
			Float64("termin_total", totalTermin).
			Msg("termin total does not equal PO total_amount")
	}

	if err := s.terminRepo.CreateSchedule(ctx, poID, termins); err != nil {
		return nil, fmt.Errorf("TerminService.CreateTerminSchedule save: %w", err)
	}
	s.log.Info().Str("po_id", poID).Int("count", len(termins)).Msg("termin schedule created")

	return s.GetTerminSchedule(ctx, poID)
}

// GetTerminSchedule returns all termins for a PO with full payment history.
func (s *TerminService) GetTerminSchedule(ctx context.Context, poID string) ([]dto.TerminResponse, error) {
	termins, err := s.terminRepo.FindByPO(ctx, poID)
	if err != nil {
		return nil, fmt.Errorf("TerminService.GetTerminSchedule: %w", err)
	}

	now := time.Now().Truncate(24 * time.Hour)
	out := make([]dto.TerminResponse, 0, len(termins))
	for _, t := range termins {
		// Load individual payment records for this termin.
		records, err := s.paymentRepo.FindByTermin(ctx, t.ID)
		if err != nil {
			s.log.Warn().Err(err).Str("termin_id", t.ID).Msg("could not load payment records")
		}
		isOverdue := t.DueDate.Before(now) && t.Status != "paid"
		out = append(out, terminToResponse(t, records, isOverdue))
	}
	return out, nil
}

// ─── Payment Recording ────────────────────────────────────────────────────────

// RecordPayment applies a payment against a single termin.
//
// Guards:
//   - Payment amount must not exceed the termin's remaining balance.
//   - After insert, the termin status is recalculated (unpaid→partial→paid).
func (s *TerminService) RecordPayment(ctx context.Context, terminID, userID string, req dto.RecordPaymentRequest) (*dto.PaymentRecordResponse, error) {
	termin, err := s.terminRepo.FindByID(ctx, terminID)
	if err != nil {
		return nil, fmt.Errorf("TerminService.RecordPayment find termin: %w", err)
	}
	if termin == nil {
		return nil, ErrTerminNotFound
	}

	// Guard: overpayment prevention.
	if req.AmountPaid > termin.AmountDue {
		return nil, fmt.Errorf("%w: requested %.2f, remaining %.2f",
			ErrTerminOverpayment, req.AmountPaid, termin.AmountDue)
	}

	payDate, err := time.ParseInLocation("2006-01-02", req.PaymentDate, time.Local)
	if err != nil {
		return nil, fmt.Errorf("invalid payment_date %q: %w", req.PaymentDate, err)
	}

	record, err := s.paymentRepo.Create(ctx, domain.PaymentRecord{
		TerminID:      terminID,
		AmountPaid:    req.AmountPaid,
		PaymentDate:   payDate,
		PaymentMethod: req.PaymentMethod,
		Notes:         req.Notes,
		RecordedBy:    &userID,
	})
	if err != nil {
		return nil, fmt.Errorf("TerminService.RecordPayment create: %w", err)
	}

	// Recalculate termin status after the payment is committed.
	if err := s.terminRepo.UpdateStatus(ctx, terminID); err != nil {
		s.log.Warn().Err(err).Str("termin_id", terminID).Msg("failed to update termin status after payment")
	}

	po, poErr := s.poRepo.FindByID(ctx, termin.POID)
	poNumber := "Unknown PO"
	storeID := ""
	if poErr == nil && po != nil {
		poNumber = po.PONumber
		storeID = po.StoreID
	}

	s.activitySvc.LogActivity(ctx, userID, storeID, domain.ActionPurchaseOrderPayment, domain.ModulePurchase, termin.POID, map[string]interface{}{
		"po_number":      poNumber,
		"payment_amount": req.AmountPaid,
		"payment_method": req.PaymentMethod,
	})

	s.log.Info().
		Str("termin_id", terminID).
		Float64("amount", req.AmountPaid).
		Msg("payment recorded")

	return paymentRecordToResponse(record), nil
}

// ─── Debt Calculation ─────────────────────────────────────────────────────────

// CalculatePODebt returns aggregated debt metrics for a PO.
func (s *TerminService) CalculatePODebt(ctx context.Context, poID string) (*dto.PODebtSummaryResponse, error) {
	po, err := s.poRepo.FindByID(ctx, poID)
	if err != nil || po == nil {
		return nil, fmt.Errorf("TerminService.CalculatePODebt find PO: %w", err)
	}
	ds, err := s.terminRepo.DebtSummary(ctx, poID, po.TotalAmount)
	if err != nil {
		return nil, fmt.Errorf("TerminService.CalculatePODebt: %w", err)
	}
	return &dto.PODebtSummaryResponse{
		POID:          poID,
		PONumber:      po.PONumber,
		TotalAmount:   ds.TotalAmount,
		TotalTermin:   ds.TotalTermin,
		TotalPaid:     ds.TotalPaid,
		RemainingDebt: ds.RemainingDebt,
		Status:        ds.Status,
		TerminCount:   ds.TerminCount,
		OverdueCount:  ds.OverdueCount,
	}, nil
}

// ─── Document Generation ─────────────────────────────────────────────────────

// GenerateDocumentData assembles everything the frontend needs to render
// a printable invoice, payment receipt, or termin agreement.
// docType must be one of: "invoice", "receipt", "termin_agreement".
func (s *TerminService) GenerateDocumentData(ctx context.Context, poID, docType string) (*dto.PODocumentData, error) { //nolint:cyclop // data assembly is inherently branchy
	po, err := s.poRepo.FindByID(ctx, poID)
	if err != nil || po == nil {
		return nil, fmt.Errorf("TerminService.GenerateDocumentData find PO: %w", err)
	}

	termins, err := s.GetTerminSchedule(ctx, poID)
	if err != nil {
		return nil, err
	}
	debt, err := s.CalculatePODebt(ctx, poID)
	if err != nil {
		return nil, err
	}

	supplierName := ""
	if po.SupplierName != nil {
		supplierName = *po.SupplierName
	}

	// Map PO to POResponse for consistency with existing API shape.
	poResp := dto.POResponse{
		ID:           po.ID,
		StoreID:      po.StoreID,
		SupplierID:   po.SupplierID,
		SupplierName: po.SupplierName,
		PONumber:     po.PONumber,
		Status:       po.Status,
		TotalAmount:  po.TotalAmount,
		Notes:        ptrToString(po.Notes),
		CreatedAt:    po.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    po.UpdatedAt.Format(time.RFC3339),
	}

	return &dto.PODocumentData{
		DocType:      docType,
		GeneratedAt:  time.Now().Format(time.RFC3339),
		PO:           poResp,
		DebtSummary:  *debt,
		Termins:      termins,
		SupplierName: supplierName,
	}, nil
}

func (s *TerminService) LogDocumentGenerate(ctx context.Context, poID, storeID, userID, docType string) error {
	po, err := s.poRepo.FindByID(ctx, poID)
	if err != nil || po == nil {
		return fmt.Errorf("LogDocumentGenerate find PO: %w", err)
	}
	s.activitySvc.LogActivity(ctx, userID, storeID, domain.ActionPurchaseOrderDocumentGenerate, domain.ModulePurchase, poID, map[string]interface{}{
		"po_number":     po.PONumber,
		"document_type": docType,
	})
	return nil
}

// ─── Mappers ─────────────────────────────────────────────────────────────────

func terminToResponse(t *domain.POTermin, records []*domain.PaymentRecord, isOverdue bool) dto.TerminResponse {
	recs := make([]dto.PaymentRecordResponse, 0, len(records))
	for _, r := range records {
		recs = append(recs, *paymentRecordToResponse(r))
	}
	return dto.TerminResponse{
		ID:           t.ID,
		POID:         t.POID,
		TerminNumber: t.TerminNumber,
		Amount:       t.Amount,
		DueDate:      t.DueDate.Format("2006-01-02"),
		Status:       t.Status,
		Notes:        t.Notes,
		AmountPaid:   t.AmountPaid,
		AmountDue:    t.AmountDue,
		IsOverdue:    isOverdue,
		Payments:     recs,
		CreatedAt:    t.CreatedAt.Format(time.RFC3339),
	}
}

func paymentRecordToResponse(r *domain.PaymentRecord) *dto.PaymentRecordResponse {
	return &dto.PaymentRecordResponse{
		ID:             r.ID,
		TerminID:       r.TerminID,
		AmountPaid:     r.AmountPaid,
		PaymentDate:    r.PaymentDate.Format("2006-01-02"),
		PaymentMethod:  r.PaymentMethod,
		Notes:          r.Notes,
		RecordedByName: r.RecordedByName,
		CreatedAt:      r.CreatedAt.Format(time.RFC3339),
	}
}
