package postgres

import (
	"context"
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
	defer db.Close()

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
	defer db.Close()

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
	defer db.Close()

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
	})
}

func TestCustomerRepo_SoftDelete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewCustomerRepo(sqlxDB)

	id := "c1"

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta("UPDATE customers SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL")).
			WithArgs(id).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.SoftDelete(context.Background(), id)
		assert.NoError(t, err)
	})
}
