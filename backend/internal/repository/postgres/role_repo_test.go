package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"github.com/moedahpos/backend/internal/domain"
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

	t.Run("CreateRole", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(`INSERT INTO roles`).WithArgs("Admin", "Desc").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "created_at", "updated_at"}).AddRow("r1", "Admin", "Desc", time.Now(), time.Now()))
		mock.ExpectExec(`INSERT INTO role_permissions`).WithArgs("r1", "p1").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		role, err := repo.CreateRole(context.Background(), &domain.Role{Name: "Admin", Description: "Desc"}, []string{"p1"})
		assert.NoError(t, err)
		assert.Equal(t, "r1", role.ID)
	})

	t.Run("UpdateRole", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(`UPDATE roles`).WithArgs("Admin2", "Desc2", "r1").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "created_at", "updated_at"}).AddRow("r1", "Admin2", "Desc2", time.Now(), time.Now()))
		mock.ExpectExec(`DELETE FROM role_permissions`).WithArgs("r1").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`INSERT INTO role_permissions`).WithArgs("r1", "p2").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		role, err := repo.UpdateRole(context.Background(), &domain.Role{ID: "r1", Name: "Admin2", Description: "Desc2"}, []string{"p2"})
		assert.NoError(t, err)
		assert.NotNil(t, role)
	})

	t.Run("DeleteRole", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM role_permissions`).WithArgs("r1").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`DELETE FROM roles`).WithArgs("r1").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.DeleteRole(context.Background(), "r1")
		assert.NoError(t, err)
	})

	t.Run("ListPermissions", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "description"}).AddRow("p1", "P1", "D1")
		mock.ExpectQuery(`SELECT id, name, description FROM permissions`).WillReturnRows(rows)
		perms, err := repo.ListPermissions(context.Background())
		assert.NoError(t, err)
		assert.Len(t, perms, 1)
	})
}
