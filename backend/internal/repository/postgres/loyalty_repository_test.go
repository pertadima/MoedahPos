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
