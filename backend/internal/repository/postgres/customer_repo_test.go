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

func TestCustomerRepo_FindAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewCustomerRepo(sqlxDB)

	storeID := "store-1"
	filter := dto.CustomerListFilter{
		StoreID: storeID,
		PaginationQuery: dto.PaginationQuery{
			Page:    1,
			PerPage: 10,
		},
	}

	t.Run("success", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM customers WHERE store_id = $1 AND deleted_at IS NULL")).
			WithArgs(storeID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		rows := sqlmock.NewRows([]string{"id", "store_id", "name", "phone", "email", "address", "notes", "created_at", "updated_at"}).
			AddRow("c1", storeID, "John Doe", "12345", "john@example.com", "Address", "Notes", time.Now(), time.Now())

		mock.ExpectQuery(regexp.QuoteMeta("SELECT id,store_id,name,phone,email,address,notes,created_at,updated_at FROM customers WHERE store_id = $1 AND deleted_at IS NULL ORDER BY name ASC LIMIT $2 OFFSET $3")).
			WithArgs(storeID, 10, 0).
			WillReturnRows(rows)

		customers, total, err := repo.FindAll(context.Background(), filter)
		assert.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, customers, 1)
		assert.Equal(t, "John Doe", customers[0].Name)
	})
}

func TestCustomerRepo_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewCustomerRepo(sqlxDB)

	c := &domain.Customer{
		StoreID: "store-1",
		Name:    "John Doe",
	}

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "store_id", "name", "phone", "email", "address", "notes", "created_at", "updated_at"}).
			AddRow("c1", c.StoreID, c.Name, nil, nil, nil, nil, time.Now(), time.Now())

		mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO customers (store_id, name, phone, email, address, notes) VALUES ($1,$2,$3,$4,$5,$6) RETURNING id, store_id, name, phone, email, address, notes, created_at, updated_at")).
			WithArgs(c.StoreID, c.Name, nil, nil, nil, nil).
			WillReturnRows(rows)

		out, err := repo.Create(context.Background(), c)
		assert.NoError(t, err)
		assert.Equal(t, "c1", out.ID)
		assert.Equal(t, "John Doe", out.Name)
	})
}

func TestCustomerRepo_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewCustomerRepo(sqlxDB)

	c := &domain.Customer{
		ID:   "c1",
		Name: "Jane Doe",
	}

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "store_id", "name", "phone", "email", "address", "notes", "created_at", "updated_at"}).
			AddRow("c1", "store-1", "Jane Doe", nil, nil, nil, nil, time.Now(), time.Now())

		mock.ExpectQuery(regexp.QuoteMeta("UPDATE customers SET name=$1,phone=$2,email=$3,address=$4,notes=$5,updated_at=NOW() WHERE id=$6 AND deleted_at IS NULL RETURNING id,store_id,name,phone,email,address,notes,created_at,updated_at")).
			WithArgs(c.Name, nil, nil, nil, nil, c.ID).
			WillReturnRows(rows)

		out, err := repo.Update(context.Background(), c)
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.Equal(t, "Jane Doe", out.Name)

		// Not Found
		mock.ExpectQuery(regexp.QuoteMeta("UPDATE customers SET")).WillReturnError(sql.ErrNoRows)
		out, err = repo.Update(context.Background(), &domain.Customer{ID: "unknown"})
		assert.NoError(t, err)
		assert.Nil(t, out)
	})
}

func TestCustomerRepo_SoftDelete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewCustomerRepo(sqlxDB)

	id := "c1"

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta("UPDATE customers SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL")).
			WithArgs(id).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.SoftDelete(context.Background(), id)
		assert.NoError(t, err)

		// Not Found
		mock.ExpectExec(regexp.QuoteMeta("UPDATE customers SET deleted_at=NOW()")).WillReturnResult(sqlmock.NewResult(0, 0))
		err = repo.SoftDelete(context.Background(), "unknown")
		assert.Error(t, err)
	})
}

func TestCustomerRepo_FindByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := NewCustomerRepo(sqlx.NewDb(db, "postgres"))

	t.Run("success", func(t *testing.T) {
		mock.ExpectQuery(`(?is)SELECT .* FROM customers WHERE id=\$1`).
			WithArgs("c1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("c1", "John"))

		res, err := repo.FindByID(context.Background(), "c1")
		assert.NoError(t, err)
		assert.NotNil(t, res)

		// Not Found
		mock.ExpectQuery(`(?is)SELECT .* FROM customers WHERE id=\$1`).WillReturnError(sql.ErrNoRows)
		res, err = repo.FindByID(context.Background(), "unknown")
		assert.NoError(t, err)
		assert.Nil(t, res)
	})
}

func TestCustomerRepo_SearchByPhone(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := NewCustomerRepo(sqlx.NewDb(db, "postgres"))

	t.Run("success", func(t *testing.T) {
		mock.ExpectQuery(`(?is)SELECT .* FROM customers WHERE store_id=\$1 AND phone ILIKE \$2`).
			WithArgs("s1", "123%").
			WillReturnRows(sqlmock.NewRows([]string{"id", "phone"}).AddRow("c1", "123"))

		res, err := repo.SearchByPhone(context.Background(), "s1", "123")
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})
}
