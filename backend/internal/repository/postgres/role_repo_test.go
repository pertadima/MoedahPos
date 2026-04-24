package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestRoleRepo(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer func() { _ = db.Close() }()

	repo := NewRoleRepository(db)

	t.Run("ListRoles", func(t *testing.T) {
		roleRows := sqlmock.NewRows([]string{"id", "name", "description", "created_at", "updated_at"}).
			AddRow("r1", "Admin", "All access", time.Now(), time.Now()).
			AddRow("r2", "Staff", "Retail staff", time.Now(), time.Now())
		mock.ExpectQuery(`SELECT .* FROM roles ORDER BY name`).WillReturnRows(roleRows)

		permRows := sqlmock.NewRows([]string{"role_id", "perm_name"}).
			AddRow("r1", "product.view").
			AddRow("r1", "product.create").
			AddRow("r2", "product.view")
		mock.ExpectQuery(`SELECT .* FROM role_permissions rp`).WillReturnRows(permRows)

		roles, err := repo.ListRoles(context.Background())
		assert.NoError(t, err)
		assert.Len(t, roles, 2)
		assert.Equal(t, "Admin", roles[0].Name)
		assert.Len(t, roles[0].Permissions, 2)
		assert.Len(t, roles[1].Permissions, 1)
	})
}
