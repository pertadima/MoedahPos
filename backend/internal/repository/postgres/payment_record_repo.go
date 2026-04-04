package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/moedahpos/backend/internal/domain"
)

// PaymentRecordRepo is the PostgreSQL implementation of repository.PaymentRecordRepository.
type PaymentRecordRepo struct{ db *sqlx.DB }

// NewPaymentRecordRepo creates a PaymentRecordRepo backed by the given database.
func NewPaymentRecordRepo(db *sqlx.DB) *PaymentRecordRepo { return &PaymentRecordRepo{db: db} }

// Create inserts one payment record and returns the persisted row with recorded_by_name joined.
func (r *PaymentRecordRepo) Create(ctx context.Context, rec domain.PaymentRecord) (*domain.PaymentRecord, error) {
	const q = `
		WITH ins AS (
			INSERT INTO payment_records
				(termin_id, amount_paid, payment_date, payment_method, notes, recorded_by)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING *
		)
		SELECT i.*, COALESCE(u.name, '') AS recorded_by_name
		FROM ins i
		LEFT JOIN users u ON u.id = i.recorded_by`

	out := &domain.PaymentRecord{}
	if err := r.db.GetContext(ctx, out, q,
		rec.TerminID, rec.AmountPaid, rec.PaymentDate.Format("2006-01-02"),
		rec.PaymentMethod, rec.Notes, rec.RecordedBy,
	); err != nil {
		return nil, fmt.Errorf("PaymentRecordRepo.Create: %w", err)
	}
	return out, nil
}

// FindByTermin returns all payment records for a termin ordered newest-first.
func (r *PaymentRecordRepo) FindByTermin(ctx context.Context, terminID string) ([]*domain.PaymentRecord, error) {
	const q = `
		SELECT p.*, COALESCE(u.name, '') AS recorded_by_name
		FROM payment_records p
		LEFT JOIN users u ON u.id = p.recorded_by
		WHERE p.termin_id = $1
		ORDER BY p.payment_date DESC, p.created_at DESC`

	var records []*domain.PaymentRecord
	if err := r.db.SelectContext(ctx, &records, q, terminID); err != nil {
		return nil, fmt.Errorf("PaymentRecordRepo.FindByTermin: %w", err)
	}
	return records, nil
}
