package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/moedahpos/backend/internal/domain"
)

func TestPOPaymentRepo_Basic(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer func() { _ = db.Close() }()

	repo := NewPOPaymentRepository(db)

	t.Run("Create", func(t *testing.T) {
		p := domain.POPayment{
			POID:    "po1",
			StoreID: "s1",
			Amount:  100000,
			Note:    stringPtr("Partial"),
			PaidBy:  "u1",
		}

		rows := sqlmock.NewRows([]string{"id", "po_id", "store_id", "amount", "note", "paid_by", "paid_at"}).
			AddRow("pp1", p.POID, p.StoreID, p.Amount, p.Note, p.PaidBy, time.Now())
		mock.ExpectQuery(`(?is)INSERT INTO po_payments`).
			WithArgs(p.POID, p.StoreID, p.Amount, p.Note, p.PaidBy).
			WillReturnRows(rows)

		// PaidByName mock
		mock.ExpectQuery(`(?is)SELECT name FROM users WHERE id=\$1`).
			WithArgs(p.PaidBy).
			WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("John"))

		res, err := repo.Create(context.Background(), p)
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("FindByPO", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "po_id", "store_id", "amount", "note", "paid_by", "paid_by_name", "paid_at"}).
			AddRow("pp1", "po1", "s1", 50000, "", "u1", "John", time.Now())
		mock.ExpectQuery(`(?is)SELECT.*FROM po_payments.*WHERE.*SELECT.*FROM payment_records.*WHERE`).
			WithArgs("po1").
			WillReturnRows(rows)

		res, err := repo.FindByPO(context.Background(), "po1")
		assert.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run("AggregateByPO", func(t *testing.T) {
		mock.ExpectQuery(`(?is)SELECT.*COALESCE.*SUM\(amount\).*COALESCE.*SUM\(pr.amount_paid\)`).
			WithArgs("po1").
			WillReturnRows(sqlmock.NewRows([]string{"paid"}).AddRow(150000.0))

		paid, status, err := repo.AggregateByPO(context.Background(), "po1", 200000.0)
		assert.NoError(t, err)
		assert.Equal(t, 150000.0, paid)
		assert.Equal(t, "partial", status)

		// Unpaid
		mock.ExpectQuery(`(?is)SELECT.*COALESCE.*SUM\(amount\).*COALESCE.*SUM\(pr.amount_paid\)`).
			WithArgs("po1").
			WillReturnRows(sqlmock.NewRows([]string{"paid"}).AddRow(0.0))
		_, status, _ = repo.AggregateByPO(context.Background(), "po1", 200000.0)
		assert.Equal(t, "unpaid", status)

		// Paid
		mock.ExpectQuery(`(?is)SELECT.*COALESCE.*SUM\(amount\).*COALESCE.*SUM\(pr.amount_paid\)`).
			WithArgs("po1").
			WillReturnRows(sqlmock.NewRows([]string{"paid"}).AddRow(200000.0))
		_, status, _ = repo.AggregateByPO(context.Background(), "po1", 200000.0)
		assert.Equal(t, "paid", status)
	})
}

func TestPOPaymentRepo_Extended(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer func() { _ = db.Close() }()

	repo := NewPOPaymentRepository(db)

	t.Run("PayableSummary", func(t *testing.T) {
		mock.ExpectQuery(`(?is)WITH combined_payments AS.*SELECT.*FROM purchase_orders`).
			WithArgs("s1").
			WillReturnRows(sqlmock.NewRows([]string{"total_debt", "total_paid", "total_outstanding", "unpaid_count", "partial_count", "overdue_debt", "due_soon_debt", "future_debt"}).
				AddRow(1000000.0, 400000.0, 600000.0, 2, 1, 100000.0, 200000.0, 300000.0))

		res, err := repo.PayableSummary(context.Background(), "s1")
		require.NoError(t, err)
		require.NotNil(t, res)
	})

	t.Run("PopulatePOPayments", func(t *testing.T) {
		pos := []*domain.PurchaseOrder{
			{ID: "po1", TotalAmount: 100000},
		}

		rows := sqlmock.NewRows([]string{"po_id", "amount_paid"}).
			AddRow("po1", 50000.0)

		// Use a more flexible regex for sqlx.In queries
		mock.ExpectQuery(`(?is)WITH combined AS.*po_payments.*payment_records`).
			WithArgs("po1", "po1").
			WillReturnRows(rows)

		repo.PopulatePOPayments(context.Background(), pos)
		assert.Equal(t, 50000.0, pos[0].AmountPaid)
		assert.Equal(t, "partial", pos[0].PaymentStatus)
	})
}
