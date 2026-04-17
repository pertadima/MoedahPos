package postgres

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/dto"
)

func TestSupplierRepo_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewSupplierRepo(sqlxDB)

	s := &domain.Supplier{
		Name:        "Supplier A",
		ContactName: "John Doe",
		Phone:       "123456",
		Email:       "supplier@example.com",
		Address:     "Street 1",
	}

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "contact_name", "phone", "email", "address", "is_active", "created_at", "updated_at", "deleted_at"}).
			AddRow("s1", s.Name, s.ContactName, s.Phone, s.Email, s.Address, true, time.Now(), time.Now(), nil)

		mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO suppliers (name, contact_name, phone, email, address) VALUES ($1,$2,$3,$4,$5) RETURNING id, name, contact_name, phone, email, address, is_active, created_at, updated_at, deleted_at")).
			WithArgs(s.Name, s.ContactName, s.Phone, s.Email, s.Address).
			WillReturnRows(rows)

		result, err := repo.Create(context.Background(), s)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "s1", result.ID)
	})
}

func TestSupplierRepo_FindAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewSupplierRepo(sqlxDB)

	filter := dto.SupplierListFilter{
		Search:   "test",
		IsActive: boolPtr(true),
	}
	filter.PerPage = 10
	filter.Page = 1

	t.Run("success", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM suppliers WHERE deleted_at IS NULL AND name ILIKE $1 AND is_active = $2")).
			WithArgs("%test%", true).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, contact_name, phone, email, address, is_active, created_at, updated_at, deleted_at FROM suppliers WHERE deleted_at IS NULL AND name ILIKE $1 AND is_active = $2 ORDER BY name ASC LIMIT $3 OFFSET $4")).
			WithArgs("%test%", true, 10, 0).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "contact_name", "phone", "email", "address", "is_active", "created_at", "updated_at", "deleted_at"}).
				AddRow("s1", "Supplier A", "John", "123", "a@a.com", "Addr", true, time.Now(), time.Now(), nil))

		suppliers, total, err := repo.FindAll(context.Background(), filter)
		assert.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, suppliers, 1)
	})
}

func TestSupplierRepo_FindByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewSupplierRepo(sqlxDB)

	id := "s1"

	t.Run("success", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, contact_name, phone, email, address, is_active, created_at, updated_at, deleted_at FROM suppliers WHERE id = $1 AND deleted_at IS NULL")).
			WithArgs(id).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "contact_name", "phone", "email", "address", "is_active", "created_at", "updated_at", "deleted_at"}).
				AddRow(id, "S1", "C1", "123", "a@a.com", "Addr", true, time.Now(), time.Now(), nil))

		res, err := repo.FindByID(context.Background(), id)
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, contact_name, phone, email, address, is_active, created_at, updated_at, deleted_at FROM suppliers WHERE id = $1 AND deleted_at IS NULL")).
			WithArgs(id).
			WillReturnError(sql.ErrNoRows)

		res, err := repo.FindByID(context.Background(), id)
		assert.NoError(t, err)
		assert.Nil(t, res)
	})
}

func TestSupplierRepo_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewSupplierRepo(sqlxDB)

	s := &domain.Supplier{
		ID:          "s1",
		Name:        "S1 Updated",
		ContactName: "C1",
		Phone:       "123",
		Email:       "a@a.com",
		Address:     "Addr",
		IsActive:    true,
	}

	t.Run("success", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta("UPDATE suppliers SET name=$1, contact_name=$2, phone=$3, email=$4, address=$5, is_active=$6, updated_at=NOW() WHERE id=$7 AND deleted_at IS NULL RETURNING id, name, contact_name, phone, email, address, is_active, created_at, updated_at, deleted_at")).
			WithArgs(s.Name, s.ContactName, s.Phone, s.Email, s.Address, s.IsActive, s.ID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "contact_name", "phone", "email", "address", "is_active", "created_at", "updated_at", "deleted_at"}).
				AddRow(s.ID, s.Name, s.ContactName, s.Phone, s.Email, s.Address, s.IsActive, time.Now(), time.Now(), nil))

		res, err := repo.Update(context.Background(), s)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, s.Name, res.Name)
	})
}

func TestSupplierRepo_SoftDelete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewSupplierRepo(sqlxDB)

	id := "s1"

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta("UPDATE suppliers SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL")).
			WithArgs(id).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.SoftDelete(context.Background(), id)
		assert.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta("UPDATE suppliers SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL")).
			WithArgs(id).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.SoftDelete(context.Background(), id)
		assert.Error(t, err)
		assert.Equal(t, "supplier not found", err.Error())
	})
}

func boolPtr(b bool) *bool { return &b }
