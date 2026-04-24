package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/repository"
)

// TerminRepo is the PostgreSQL implementation of repository.TerminRepository.
type TerminRepo struct{ db *sqlx.DB }

// NewTerminRepository creates a new TerminRepository.
func NewTerminRepository(db *sql.DB) repository.TerminRepository {
	return &TerminRepo{db: sqlx.NewDb(db, "postgres")}
}

// CreateSchedule atomically replaces the termin schedule for a PO.
// It deletes any existing termins (and their payment_records via CASCADE)
// then inserts the new schedule — all within a single DB transaction.
func (r *TerminRepo) CreateSchedule(ctx context.Context, poID string, termins []domain.POTermin) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("TerminRepo.CreateSchedule begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Delete existing schedule for this PO (CASCADE removes payment_records too).
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM purchase_order_termins WHERE po_id = $1`, poID,
	); err != nil {
		return fmt.Errorf("TerminRepo.CreateSchedule delete old: %w", err)
	}

	const insertQ = `
		INSERT INTO purchase_order_termins
			(po_id, termin_number, amount, due_date, status, notes)
		VALUES ($1, $2, $3, $4, 'unpaid', $5)`

	for _, t := range termins {
		if _, err := tx.ExecContext(ctx, insertQ,
			poID, t.TerminNumber, t.Amount, t.DueDate.Format("2006-01-02"), t.Notes,
		); err != nil {
			return fmt.Errorf("TerminRepo.CreateSchedule insert termin %d: %w", t.TerminNumber, err)
		}
	}

	return tx.Commit()
}

// FindByPO returns all termins for a PO with aggregated payment totals, ordered by termin_number.
func (r *TerminRepo) FindByPO(ctx context.Context, poID string) ([]*domain.POTermin, error) {
	const q = `
		SELECT t.id, t.po_id, t.termin_number, t.amount, t.due_date, t.status, t.notes,
		       t.created_at, t.updated_at,
		       COALESCE(SUM(p.amount_paid), 0)          AS amount_paid,
		       t.amount - COALESCE(SUM(p.amount_paid),0) AS amount_due
		FROM purchase_order_termins t
		LEFT JOIN payment_records p ON p.termin_id = t.id
		WHERE t.po_id = $1
		GROUP BY t.id
		ORDER BY t.termin_number`
	var termins []*domain.POTermin
	if err := r.db.SelectContext(ctx, &termins, q, poID); err != nil {
		return nil, fmt.Errorf("TerminRepo.FindByPO: %w", err)
	}
	return termins, nil
}

// FindByID returns a single termin with aggregated payment totals.
func (r *TerminRepo) FindByID(ctx context.Context, terminID string) (*domain.POTermin, error) {
	const q = `
		SELECT t.id, t.po_id, t.termin_number, t.amount, t.due_date, t.status, t.notes,
		       t.created_at, t.updated_at,
		       COALESCE(SUM(p.amount_paid), 0)           AS amount_paid,
		       t.amount - COALESCE(SUM(p.amount_paid),0) AS amount_due
		FROM purchase_order_termins t
		LEFT JOIN payment_records p ON p.termin_id = t.id
		WHERE t.id = $1
		GROUP BY t.id`
	var t domain.POTermin
	if err := r.db.GetContext(ctx, &t, q, terminID); err != nil {
		return nil, fmt.Errorf("TerminRepo.FindByID: %w", err)
	}
	return &t, nil
}

// UpdateStatus recalculates and persists a termin's payment status.
//
// Status derivation rules:
//   - paid:    amount_paid >= amount
//   - partial: 0 < amount_paid < amount  AND due_date >= today
//   - overdue: amount_paid < amount      AND due_date < today
//   - unpaid:  amount_paid == 0          AND due_date >= today
func (r *TerminRepo) UpdateStatus(ctx context.Context, terminID string) error {
	const q = `
		UPDATE purchase_order_termins
		SET status = CASE
			WHEN COALESCE((SELECT SUM(amount_paid) FROM payment_records WHERE termin_id = $1), 0) >= amount
				THEN 'paid'
			WHEN COALESCE((SELECT SUM(amount_paid) FROM payment_records WHERE termin_id = $1), 0) > 0
			     AND due_date < CURRENT_DATE
				THEN 'overdue'
			WHEN COALESCE((SELECT SUM(amount_paid) FROM payment_records WHERE termin_id = $1), 0) > 0
				THEN 'partial'
			WHEN due_date < CURRENT_DATE
				THEN 'overdue'
			ELSE 'unpaid'
		END,
		updated_at = NOW()
		WHERE id = $1`
	if _, err := r.db.ExecContext(ctx, q, terminID); err != nil {
		return fmt.Errorf("TerminRepo.UpdateStatus: %w", err)
	}
	return nil
}

// DebtSummary aggregates payment totals across all termins for a PO.
func (r *TerminRepo) DebtSummary(ctx context.Context, poID string, totalAmount float64) (*domain.PODebtSummary, error) {
	const q = `
		SELECT
		    $1::uuid                                                       AS po_id,
		    $2::numeric                                                    AS total_amount,
		    COALESCE(SUM(t.amount), 0)                                     AS total_termin,
		    COALESCE(SUM(COALESCE(pr.paid, 0)), 0)                        AS total_paid,
		    COALESCE(SUM(t.amount), 0) - COALESCE(SUM(COALESCE(pr.paid,0)),0) AS remaining_debt,
		    COUNT(t.id)::int                                               AS termin_count,
		    COUNT(t.id) FILTER (WHERE t.status = 'overdue')::int          AS overdue_count
		FROM purchase_order_termins t
		LEFT JOIN (
		    SELECT termin_id, SUM(amount_paid) AS paid
		    FROM payment_records
		    GROUP BY termin_id
		) pr ON pr.termin_id = t.id
		WHERE t.po_id = $1`

	ds := &domain.PODebtSummary{}
	if err := r.db.GetContext(ctx, ds, q, poID, totalAmount); err != nil {
		return nil, fmt.Errorf("TerminRepo.DebtSummary: %w", err)
	}

	// Derive human-readable status from aggregate numbers.
	switch {
	case ds.TotalPaid >= ds.TotalTermin && ds.TotalTermin > 0:
		ds.Status = "paid"
	case ds.TotalPaid > 0:
		ds.Status = "partial"
	default:
		ds.Status = "unpaid"
	}

	return ds, nil
}

// ─── Compile-time type assertion ──────────────────────────────────────────────

var _ interface {
	CreateSchedule(context.Context, string, []domain.POTermin) error
	FindByPO(context.Context, string) ([]*domain.POTermin, error)
	FindByID(context.Context, string) (*domain.POTermin, error)
	UpdateStatus(context.Context, string) error
	DebtSummary(context.Context, string, float64) (*domain.PODebtSummary, error)
} = (*TerminRepo)(nil)

// ensure time import is used (DueDate formatting in CreateSchedule).
var _ = time.Now
