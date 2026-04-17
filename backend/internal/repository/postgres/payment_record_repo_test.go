package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	"github.com/moedahpos/backend/internal/domain"
)

func TestPaymentRecordRepo_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewPaymentRecordRepo(sqlxDB)

	ctx := context.Background()
	uid := "u1"
	rec := domain.PaymentRecord{
		TerminID:      "t1",
		AmountPaid:    100,
		PaymentDate:   time.Now(),
		PaymentMethod: "cash",
		Notes:         "notes",
		RecordedBy:    &uid,
	}

	rows := sqlmock.NewRows([]string{"id", "termin_id", "amount_paid", "payment_date", "payment_method", "notes", "recorded_by", "created_at", "recorded_by_name"}).
		AddRow("r1", "t1", 100, rec.PaymentDate, "cash", "notes", uid, time.Now(), "User 1")

	mock.ExpectQuery("(?s)WITH ins AS.*INSERT INTO payment_records.*SELECT.*FROM ins i.*LEFT JOIN users u").
		WithArgs(rec.TerminID, rec.AmountPaid, rec.PaymentDate.Format("2006-01-02"), rec.PaymentMethod, rec.Notes, rec.RecordedBy).
		WillReturnRows(rows)

	res, err := repo.Create(ctx, rec)
	if err != nil {
		t.Fatalf("failed to create payment record: %v", err)
	}
	assert.NotNil(t, res)
	assert.Equal(t, "User 1", res.RecordedByName)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPaymentRecordRepo_FindByTermin(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewPaymentRecordRepo(sqlxDB)

	ctx := context.Background()
	tid := "t1"

	rows := sqlmock.NewRows([]string{"id", "termin_id", "amount_paid", "payment_date", "payment_method", "notes", "recorded_by", "created_at", "recorded_by_name"}).
		AddRow("r1", tid, 100, time.Now(), "cash", "", "u1", time.Now(), "User 1").
		AddRow("r2", tid, 50, time.Now().Add(-24*time.Hour), "transfer", "", "u1", time.Now(), "User 1")

	mock.ExpectQuery(`(?s)SELECT .* FROM payment_records p .* WHERE p.termin_id = \$1`).
		WithArgs(tid).
		WillReturnRows(rows)

	res, err := repo.FindByTermin(ctx, tid)
	assert.NoError(t, err)
	assert.Len(t, res, 2)
	assert.Equal(t, "r1", res[0].ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}
