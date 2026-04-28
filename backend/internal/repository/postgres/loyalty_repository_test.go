package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestLoyaltyRepo_GetBalance(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer func() { _ = db.Close() }()

	repo := NewLoyaltyRepository(db)

	ctx := context.Background()
	cid := "c1"

	mock.ExpectQuery(`SELECT COALESCE\(SUM\(points_delta\), 0\) FROM loyalty_ledger WHERE customer_id = \$1`).
		WithArgs(cid).
		WillReturnRows(sqlmock.NewRows([]string{"coalesce"}).AddRow(150.5))

	bal, err := repo.GetBalance(ctx, cid)
	assert.NoError(t, err)
	assert.Equal(t, 150.5, bal)
}

func TestLoyaltyRepo_EarnPoints(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer func() { _ = db.Close() }()

	repo := NewLoyaltyRepository(db)

	ctx := context.Background()
	cid := "c1"
	tid := "t1"

	rows := sqlmock.NewRows([]string{"id", "customer_id", "points_delta", "transaction_id", "type", "created_at"}).
		AddRow("l1", cid, 10.0, tid, "EARN", time.Now())

	mock.ExpectQuery(`INSERT INTO loyalty_ledger`).
		WithArgs(cid, 10.0, &tid).
		WillReturnRows(rows)

	entry, err := repo.EarnPoints(ctx, cid, &tid, 10.0)
	assert.NoError(t, err)
	assert.NotNil(t, entry)
	assert.Equal(t, 10.0, entry.PointsDelta)
}

func TestMembershipTierRepo_FindAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer func() { _ = db.Close() }()

	repo := NewMembershipTierRepository(db)

	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "name", "multiplier", "created_at", "updated_at"}).
		AddRow("t1", "Gold", 1.2, time.Now(), time.Now())

	mock.ExpectQuery(`SELECT id, name, multiplier, created_at, updated_at FROM membership_tiers`).
		WillReturnRows(rows)

	tiers, err := repo.FindAll(ctx)
	assert.NoError(t, err)
	assert.Len(t, tiers, 1)
	assert.Equal(t, "Gold", tiers[0].Name)
}

func TestLoyaltyRepo_SpendPoints(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer func() { _ = db.Close() }()
	repo := NewLoyaltyRepository(db)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "customer_id", "points_delta", "transaction_id", "type", "created_at"}).
		AddRow("l2", "c1", -5.0, "t1", "SPEND", time.Now())

	mock.ExpectQuery(`INSERT INTO loyalty_ledger`).WithArgs("c1", -5.0, sqlmock.AnyArg()).WillReturnRows(rows)

	entry, err := repo.SpendPoints(ctx, "c1", nil, 5.0)
	assert.NoError(t, err)
	assert.Equal(t, -5.0, entry.PointsDelta)
}

func TestLoyaltyRepo_GetHistory(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer func() { _ = db.Close() }()
	repo := NewLoyaltyRepository(db)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "customer_id", "points_delta", "transaction_id", "type", "created_at"}).
		AddRow("l1", "c1", 10.0, "t1", "EARN", time.Now())

	mock.ExpectQuery(`SELECT .* FROM loyalty_ledger`).WithArgs("c1").WillReturnRows(rows)

	history, err := repo.GetHistory(ctx, "c1")
	assert.NoError(t, err)
	assert.Len(t, history, 1)
}

func TestLoyaltyRepo_AssignTier(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer func() { _ = db.Close() }()
	repo := NewLoyaltyRepository(db)
	ctx := context.Background()

	mock.ExpectExec(`UPDATE customers SET loyalty_tier_id = .* WHERE id = .*`).WithArgs("t1", "c1").WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.AssignTier(ctx, "c1", "t1")
	assert.NoError(t, err)
}

func TestLoyaltyRepo_GetCustomerTier(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer func() { _ = db.Close() }()
	repo := NewLoyaltyRepository(db)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "name", "multiplier", "created_at", "updated_at"}).
		AddRow("t1", "Gold", 1.2, time.Now(), time.Now())

	mock.ExpectQuery(`SELECT .* FROM membership_tiers .* JOIN customers .*`).WithArgs("c1").WillReturnRows(rows)

	tier, err := repo.GetCustomerTier(ctx, "c1")
	assert.NoError(t, err)
	if assert.NotNil(t, tier) {
		assert.Equal(t, "Gold", tier.Name)
	}
}
