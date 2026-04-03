package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/dto"
)

// POPaymentRepo handles the po_payments table.
type POPaymentRepo struct{ db *sqlx.DB }

func NewPOPaymentRepo(db *sqlx.DB) *POPaymentRepo { return &POPaymentRepo{db: db} }

// Create inserts one payment record.
func (r *POPaymentRepo) Create(ctx context.Context, p domain.POPayment) (*domain.POPayment, error) {
	const q = `
		INSERT INTO po_payments (po_id, store_id, amount, note, paid_by)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING id, po_id, store_id, amount, note, paid_by, paid_at`
	var out domain.POPayment
	if err := r.db.QueryRowxContext(ctx, q,
		p.POID, p.StoreID, p.Amount, p.Note, p.PaidBy,
	).StructScan(&out); err != nil {
		return nil, fmt.Errorf("POPaymentRepo.Create: %w", err)
	}
	// Populate paid_by_name
	_ = r.db.QueryRowxContext(ctx, `SELECT name FROM users WHERE id=$1`, out.PaidBy).Scan(&out.PaidByName)
	return &out, nil
}

// FindByPO returns all payments for a PO, newest first.
func (r *POPaymentRepo) FindByPO(ctx context.Context, poID string) ([]*domain.POPayment, error) {
	const q = `
		SELECT pp.id, pp.po_id, pp.store_id, pp.amount, pp.note, pp.paid_by,
		       COALESCE(u.name,'') AS paid_by_name, pp.paid_at
		FROM po_payments pp
		LEFT JOIN users u ON u.id = pp.paid_by
		WHERE pp.po_id = $1
		ORDER BY pp.paid_at DESC`
	var rows []*domain.POPayment
	if err := r.db.SelectContext(ctx, &rows, q, poID); err != nil {
		return nil, fmt.Errorf("POPaymentRepo.FindByPO: %w", err)
	}
	return rows, nil
}

// AggregateByPO returns amount_paid and payment_status for a PO.
func (r *POPaymentRepo) AggregateByPO(ctx context.Context, poID string, totalAmount float64) (float64, string, error) {
	var paid float64
	if err := r.db.QueryRowxContext(ctx,
		`SELECT COALESCE(SUM(amount),0) FROM po_payments WHERE po_id=$1`, poID,
	).Scan(&paid); err != nil {
		return 0, "", fmt.Errorf("POPaymentRepo.AggregateByPO: %w", err)
	}
	status := paymentStatus(paid, totalAmount)
	return paid, status, nil
}

// PayableSummary returns outstanding debt totals for a store.
func (r *POPaymentRepo) PayableSummary(ctx context.Context, storeID string) (*dto.PayableSummary, error) {
	const q = `
		SELECT
			COALESCE(SUM(po.total_amount), 0)                    AS total_debt,
			COALESCE(SUM(pp.amount_paid), 0)                     AS total_paid,
			COALESCE(SUM(po.total_amount - pp.amount_paid), 0)   AS total_outstanding,
			COUNT(*) FILTER (WHERE pp.amount_paid = 0)           AS unpaid_count,
			COUNT(*) FILTER (WHERE pp.amount_paid > 0 AND pp.amount_paid < po.total_amount) AS partial_count
		FROM purchase_orders po
		LEFT JOIN (
			SELECT po_id, SUM(amount) AS amount_paid
			FROM po_payments GROUP BY po_id
		) pp ON pp.po_id = po.id
		WHERE po.store_id = $1 AND po.status = 'received'`
	var s dto.PayableSummary
	if err := r.db.QueryRowxContext(ctx, q, storeID).Scan(
		&s.TotalDebt, &s.TotalPaid, &s.TotalOutstanding, &s.UnpaidCount, &s.PartialCount,
	); err != nil {
		return nil, fmt.Errorf("POPaymentRepo.PayableSummary: %w", err)
	}
	return &s, nil
}

func paymentStatus(paid, total float64) string {
	switch {
	case paid <= 0:
		return "unpaid"
	case paid >= total:
		return "paid"
	default:
		return "partial"
	}
}

// PopulatePOPayments enriches a slice of POs with payment aggregation.
func (r *POPaymentRepo) PopulatePOPayments(ctx context.Context, pos []*domain.PurchaseOrder) {
	if len(pos) == 0 {
		return
	}
	ids := make([]string, len(pos))
	for i, po := range pos {
		ids[i] = po.ID
	}
	type agg struct {
		POID   string  `db:"po_id"`
		Amount float64 `db:"amount_paid"`
	}
	query, args, _ := sqlx.In(`SELECT po_id, COALESCE(SUM(amount),0) AS amount_paid FROM po_payments WHERE po_id IN (?) GROUP BY po_id`, ids)
	query = r.db.Rebind(query)
	var aggs []agg
	_ = r.db.SelectContext(ctx, &aggs, query, args...)

	m := make(map[string]float64, len(aggs))
	for _, a := range aggs {
		m[a.POID] = a.Amount
	}
	for _, po := range pos {
		paid := m[po.ID]
		po.AmountPaid = paid
		po.PaymentStatus = paymentStatus(paid, po.TotalAmount)
	}
	_ = time.Now() // keep time import used
}
