package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/dto"
	"github.com/moedahpos/backend/internal/repository"
)

// POPaymentRepo handles the po_payments table.
type POPaymentRepo struct{ db *sqlx.DB }

// NewPOPaymentRepository creates a new POPaymentRepository.
func NewPOPaymentRepository(db *sql.DB) repository.POPaymentRepository {
	return &POPaymentRepo{db: sqlx.NewDb(db, "postgres")}
}

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
		
		UNION ALL
		
		SELECT pr.id, t.po_id, po.store_id, pr.amount_paid as amount, pr.notes as note, pr.recorded_by as paid_by,
		       COALESCE(u.name,'') AS paid_by_name, pr.created_at as paid_at
		FROM payment_records pr
		JOIN purchase_order_termins t ON t.id = pr.termin_id
		JOIN purchase_orders po ON po.id = t.po_id
		LEFT JOIN users u ON u.id = pr.recorded_by
		WHERE t.po_id = $1
		
		ORDER BY paid_at DESC`
	var rows []*domain.POPayment
	if err := r.db.SelectContext(ctx, &rows, q, poID); err != nil {
		return nil, fmt.Errorf("POPaymentRepo.FindByPO: %w", err)
	}
	return rows, nil
}

// AggregateByPO returns amount_paid and payment_status for a PO.
func (r *POPaymentRepo) AggregateByPO(ctx context.Context, poID string, totalAmount float64) (float64, string, error) {
	var paid float64
	const q = `
		SELECT 
			COALESCE((SELECT SUM(amount) FROM po_payments WHERE po_id=$1), 0) +
			COALESCE((SELECT SUM(pr.amount_paid) FROM payment_records pr JOIN purchase_order_termins t ON t.id = pr.termin_id WHERE t.po_id=$1), 0)
	`
	if err := r.db.QueryRowxContext(ctx, q, poID).Scan(&paid); err != nil {
		return 0, "", fmt.Errorf("POPaymentRepo.AggregateByPO: %w", err)
	}
	status := paymentStatus(paid, totalAmount)
	return paid, status, nil
}

// PayableSummary returns outstanding debt totals for a store.
func (r *POPaymentRepo) PayableSummary(ctx context.Context, storeID string) (*dto.PayableSummary, error) {
	const q = `
		WITH combined_payments AS (
			SELECT po_id, SUM(amount) AS amount_paid FROM po_payments GROUP BY po_id
			UNION ALL
			SELECT t.po_id, SUM(pr.amount_paid) AS amount_paid FROM payment_records pr JOIN purchase_order_termins t ON t.id = pr.termin_id GROUP BY t.po_id
		),
		termin_debt AS (
			SELECT
			  t.id,
			  t.due_date,
			  t.amount - COALESCE(SUM(pr.amount_paid), 0) AS remaining
			FROM purchase_order_termins t
			JOIN purchase_orders po ON po.id = t.po_id
			LEFT JOIN payment_records pr ON pr.termin_id = t.id
			WHERE po.store_id = $1 AND po.status = 'received'
			GROUP BY t.id
			HAVING (t.amount - COALESCE(SUM(pr.amount_paid), 0)) > 0
		)
		SELECT
			COALESCE(SUM(po.total_amount), 0)                    AS total_debt,
			COALESCE(SUM(pp.amount_paid), 0)                     AS total_paid,
			COALESCE(SUM(po.total_amount - pp.amount_paid), 0)   AS total_outstanding,
			COUNT(*) FILTER (WHERE COALESCE(pp.amount_paid,0) = 0) AS unpaid_count,
			COUNT(*) FILTER (WHERE COALESCE(pp.amount_paid,0) > 0 AND COALESCE(pp.amount_paid,0) < po.total_amount) AS partial_count,
			(SELECT COALESCE(SUM(remaining),0) FROM termin_debt WHERE due_date < CURRENT_DATE) AS overdue_debt,
			(SELECT COALESCE(SUM(remaining),0) FROM termin_debt WHERE due_date >= CURRENT_DATE AND due_date <= CURRENT_DATE + 7) AS due_soon_debt,
			(SELECT COALESCE(SUM(remaining),0) FROM termin_debt WHERE due_date > CURRENT_DATE + 7) AS future_debt
		FROM purchase_orders po
		LEFT JOIN (
			SELECT po_id, SUM(amount_paid) AS amount_paid FROM combined_payments GROUP BY po_id
		) pp ON pp.po_id = po.id
		WHERE po.store_id = $1 AND po.status = 'received'`
	var s dto.PayableSummary
	if err := r.db.QueryRowxContext(ctx, q, storeID).Scan(
		&s.TotalDebt, &s.TotalPaid, &s.TotalOutstanding, &s.UnpaidCount, &s.PartialCount,
		&s.OverdueDebt, &s.DueSoonDebt, &s.FutureDebt,
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
	query, args, _ := sqlx.In(`
		WITH combined AS (
			SELECT po_id, amount FROM po_payments WHERE po_id IN (?)
			UNION ALL
			SELECT t.po_id, pr.amount_paid as amount FROM payment_records pr JOIN purchase_order_termins t ON t.id = pr.termin_id WHERE t.po_id IN (?)
		)
		SELECT po_id, COALESCE(SUM(amount),0) AS amount_paid FROM combined GROUP BY po_id`, ids, ids)
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
