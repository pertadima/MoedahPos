package rbac

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestRBAC(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer func() { _ = db.Close() }()

	sqlxDB := sqlx.NewDb(db, "postgres")

	t.Run("New and Has", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"role_name", "permission_name"}).
			AddRow("admin", "view_reports").
			AddRow("admin", "edit_products").
			AddRow("cashier", "view_reports")

		mock.ExpectQuery(`SELECT .* FROM role_permissions`).WillReturnRows(rows)

		store, err := New(sqlxDB)
		assert.NoError(t, err)
		assert.NotNil(t, store)

		assert.True(t, store.Has("admin", "view_reports"))
		assert.True(t, store.Has("admin", "edit_products"))
		assert.True(t, store.Has("cashier", "view_reports"))
		assert.False(t, store.Has("cashier", "edit_products"))
		assert.False(t, store.Has("guest", "view_reports"))
	})

	t.Run("New Error", func(t *testing.T) {
		mock.ExpectQuery(`SELECT .* FROM role_permissions`).WillReturnError(assert.AnError)
		_, err := New(sqlxDB)
		assert.Error(t, err)
	})

	t.Run("IsSuperAdmin", func(t *testing.T) {
		assert.True(t, IsSuperAdmin("superadmin"))
		assert.False(t, IsSuperAdmin("admin"))
	})
}
