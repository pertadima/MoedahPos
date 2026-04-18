package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	"github.com/moedahpos/backend/internal/domain"
)

func TestTerminRepo_CreateSchedule(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer func() { _ = db.Close() }()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewTerminRepo(sqlxDB)

	ctx := context.Background()
	poID := "po1"
	termins := []domain.POTermin{
		{TerminNumber: 1, Amount: 100, DueDate: time.Now(), Notes: "Note 1"},
	}

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM purchase_order_termins WHERE po_id = \$1`).
		WithArgs(poID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec(`INSERT INTO purchase_order_termins`).
		WithArgs(poID, 1, 100.0, termins[0].DueDate.Format("2006-01-02"), "Note 1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err = repo.CreateSchedule(ctx, poID, termins)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTerminRepo_FindByPO(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer func() { _ = db.Close() }()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewTerminRepo(sqlxDB)

	ctx := context.Background()
	poID := "po1"

	rows := sqlmock.NewRows([]string{"id", "po_id", "termin_number", "amount", "due_date", "status", "notes", "created_at", "updated_at", "amount_paid", "amount_due"}).
		AddRow("t1", poID, 1, 100, time.Now(), "unpaid", "", time.Now(), time.Now(), 0, 100)

	mock.ExpectQuery(`SELECT .* FROM purchase_order_termins t.* WHERE t.po_id = \$1`).
		WithArgs(poID).
		WillReturnRows(rows)

	res, err := repo.FindByPO(ctx, poID)
	assert.NoError(t, err)
	assert.Len(t, res, 1)
	assert.Equal(t, "t1", res[0].ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTerminRepo_FindByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer func() { _ = db.Close() }()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewTerminRepo(sqlxDB)

	ctx := context.Background()
	tid := "t1"

	rows := sqlmock.NewRows([]string{"id", "po_id", "termin_number", "amount", "due_date", "status", "notes", "created_at", "updated_at", "amount_paid", "amount_due"}).
		AddRow(tid, "po1", 1, 100, time.Now(), "partial", "", time.Now(), time.Now(), 40, 60)

	mock.ExpectQuery(`SELECT .* FROM purchase_order_termins t.* WHERE t.id = \$1`).
		WithArgs(tid).
		WillReturnRows(rows)

	res, err := repo.FindByID(ctx, tid)
	assert.NoError(t, err)
	assert.Equal(t, tid, res.ID)
	assert.Equal(t, 40.0, res.AmountPaid)

	// Not Found
	mock.ExpectQuery(`SELECT .* FROM purchase_order_termins t.* WHERE t.id = \$1`).WithArgs("unknown").WillReturnError(sql.ErrNoRows)
	res, err = repo.FindByID(ctx, "unknown")
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTerminRepo_UpdateStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer func() { _ = db.Close() }()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewTerminRepo(sqlxDB)

	ctx := context.Background()
	tid := "t1"

	mock.ExpectExec(`UPDATE purchase_order_termins SET status = CASE .* WHERE id = \$1`).
		WithArgs(tid).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.UpdateStatus(ctx, tid)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTerminRepo_DebtSummary(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer func() { _ = db.Close() }()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewTerminRepo(sqlxDB)

	ctx := context.Background()
	poID := "po1"
	total := 1000.0

	rows := sqlmock.NewRows([]string{"po_id", "total_amount", "total_termin", "total_paid", "remaining_debt", "termin_count", "overdue_count"}).
		AddRow(poID, total, 1000, 500, 500, 2, 0)

	mock.ExpectQuery(`SELECT .* FROM purchase_order_termins t.* WHERE t.po_id = \$1`).
		WithArgs(poID, total).
		WillReturnRows(rows)

	res, err := repo.DebtSummary(ctx, poID, total)
	assert.NoError(t, err)
	assert.Equal(t, "partial", res.Status)
	assert.Equal(t, 500.0, res.TotalPaid)

	// Case: Paid
	mock.ExpectQuery(`SELECT .* FROM purchase_order_termins t.* WHERE t.po_id = \$1`).WithArgs(poID, total).
		WillReturnRows(sqlmock.NewRows([]string{"po_id", "total_amount", "total_termin", "total_paid", "remaining_debt", "termin_count", "overdue_count"}).
			AddRow(poID, total, 1000, 1000, 0, 2, 0))
	res, err = repo.DebtSummary(ctx, poID, total)
	assert.NoError(t, err)
	assert.Equal(t, "paid", res.Status)

	// Case: Unpaid
	mock.ExpectQuery(`SELECT .* FROM purchase_order_termins t.* WHERE t.po_id = \$1`).WithArgs(poID, total).
		WillReturnRows(sqlmock.NewRows([]string{"po_id", "total_amount", "total_termin", "total_paid", "remaining_debt", "termin_count", "overdue_count"}).
			AddRow(poID, total, 1000, 0, 1000, 2, 0))
	res, err = repo.DebtSummary(ctx, poID, total)
	assert.NoError(t, err)
	assert.Equal(t, "unpaid", res.Status)
	assert.NoError(t, mock.ExpectationsWereMet())
}
