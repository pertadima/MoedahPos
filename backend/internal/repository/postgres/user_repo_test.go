package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/moedahpos/backend/internal/domain"
)

func TestUserRepo_Remaining(t *testing.T) {
	t.Run("Create", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		repo := NewUserRepo(sqlx.NewDb(db, "postgres"))

		rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "name", "is_active", "created_at", "updated_at", "deleted_at"}).
			AddRow("u1", "test@example.com", "hash", "John", true, time.Now(), time.Now(), nil)
		mock.ExpectQuery(`(?is)INSERT INTO users \(name, email, password_hash\)`).WithArgs("John", "test@example.com", "hash").WillReturnRows(rows)

		res, err := repo.Create(context.Background(), &domain.User{Email: "test@example.com", PasswordHash: "hash", Name: "John"})
		require.NoError(t, err)
		require.NotNil(t, res)
		assert.Equal(t, "u1", res.ID)
	})

	t.Run("FindByID", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		repo := NewUserRepo(sqlx.NewDb(db, "postgres"))

		rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "name", "is_active", "created_at", "updated_at", "deleted_at"}).
			AddRow("u1", "test@example.com", "hash", "John", true, time.Now(), time.Now(), nil)
		mock.ExpectQuery(`(?is)SELECT .* FROM users WHERE id = \$1`).WithArgs("u1").WillReturnRows(rows)

		res, err := repo.FindByID(context.Background(), "u1")
		require.NoError(t, err)
		assert.NotNil(t, res)

		// Not Found
		mock.ExpectQuery(`(?is)SELECT .* FROM users WHERE id = \$1`).WithArgs("unknown").WillReturnError(sql.ErrNoRows)
		res, err = repo.FindByID(context.Background(), "unknown")
		assert.NoError(t, err)
		assert.Nil(t, res)
	})

	t.Run("FindByEmail", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		repo := NewUserRepo(sqlx.NewDb(db, "postgres"))

		rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "name", "is_active", "created_at", "updated_at", "deleted_at"}).
			AddRow("u1", "test@example.com", "hash", "John", true, time.Now(), time.Now(), nil)
		mock.ExpectQuery(`(?is)SELECT .* FROM users WHERE email = \$1`).WithArgs("test@example.com").WillReturnRows(rows)

		res, err := repo.FindByEmail(context.Background(), "test@example.com")
		require.NoError(t, err)
		assert.NotNil(t, res)

		// Not Found
		mock.ExpectQuery(`(?is)SELECT .* FROM users WHERE email = \$1`).WithArgs("unknown@example.com").WillReturnError(sql.ErrNoRows)
		res, err = repo.FindByEmail(context.Background(), "unknown@example.com")
		assert.NoError(t, err)
		assert.Nil(t, res)
	})

	t.Run("ExistsByEmail", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		repo := NewUserRepo(sqlx.NewDb(db, "postgres"))

		mock.ExpectQuery(`(?is)SELECT EXISTS\(SELECT 1 FROM users WHERE email = \$1.*NULL\)`).
			WithArgs("test@example.com").
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		exists, err := repo.ExistsByEmail(context.Background(), "test@example.com")
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("FindStoresByUserID", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		repo := NewUserRepo(sqlx.NewDb(db, "postgres"))

		rows := sqlmock.NewRows([]string{"id", "user_id", "store_id", "role_id", "is_active", "created_at", "store_name", "store_type", "role_name"}).
			AddRow("us1", "u1", "s1", "r1", true, time.Now(), "Store 1", "retail", "Admin")
		mock.ExpectQuery(`(?is)SELECT .* FROM user_stores us .* WHERE us.user_id = \$1`).WithArgs("u1").WillReturnRows(rows)

		res, err := repo.FindStoresByUserID(context.Background(), "u1")
		require.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run("Update", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		repo := NewUserRepo(sqlx.NewDb(db, "postgres"))

		rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "name", "is_active", "created_at", "updated_at", "deleted_at"}).
			AddRow("u1", "test@example.com", "hash", "New John", true, time.Now(), time.Now(), nil)
		mock.ExpectQuery(`(?is)UPDATE users SET name=\$2, email=\$3.*WHERE id=\$1`).WithArgs("u1", "New John", "test@example.com").WillReturnRows(rows)

		res, err := repo.Update(context.Background(), "u1", "New John", "test@example.com")
		require.NoError(t, err)
		require.NotNil(t, res)
		assert.Equal(t, "New John", res.Name)

		// Not Found
		mock.ExpectQuery(`(?is)UPDATE users SET`).WillReturnError(sql.ErrNoRows)
		res, err = repo.Update(context.Background(), "unknown", "name", "email")
		assert.Error(t, err)
		assert.Nil(t, res)
	})

	t.Run("SoftDelete", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		repo := NewUserRepo(sqlx.NewDb(db, "postgres"))

		mock.ExpectExec(`(?is)UPDATE users SET deleted_at=NOW\(\), is_active=false.*WHERE id=\$1`).WithArgs("u1").WillReturnResult(sqlmock.NewResult(1, 1))
		err := repo.SoftDelete(context.Background(), "u1")
		require.NoError(t, err)

		// Not Found
		mock.ExpectExec(`(?is)UPDATE users SET deleted_at=NOW\(\)`).WillReturnResult(sqlmock.NewResult(0, 0))
		err = repo.SoftDelete(context.Background(), "unknown")
		assert.Error(t, err)
	})

	t.Run("ResetPassword", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		repo := NewUserRepo(sqlx.NewDb(db, "postgres"))

		mock.ExpectExec(`(?is)UPDATE users SET password_hash=\$2.*WHERE id=\$1`).WithArgs("u1", "new_hash").WillReturnResult(sqlmock.NewResult(1, 1))
		err := repo.ResetPassword(context.Background(), "u1", "new_hash")
		require.NoError(t, err)

		// Not Found
		mock.ExpectExec(`(?is)UPDATE users SET password_hash=\$2`).WillReturnResult(sqlmock.NewResult(0, 0))
		err = repo.ResetPassword(context.Background(), "unknown", "hash")
		assert.Error(t, err)
	})

	t.Run("Deactivate", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		repo := NewUserRepo(sqlx.NewDb(db, "postgres"))

		mock.ExpectExec(`(?is)UPDATE users SET is_active=false.*WHERE id=\$1`).WithArgs("u1").WillReturnResult(sqlmock.NewResult(1, 1))
		err := repo.Deactivate(context.Background(), "u1")
		require.NoError(t, err)

		// Not Found
		mock.ExpectExec(`(?is)UPDATE users SET is_active=false`).WillReturnResult(sqlmock.NewResult(0, 0))
		err = repo.Deactivate(context.Background(), "unknown")
		assert.Error(t, err)
	})

	t.Run("ListAll", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		repo := NewUserRepo(sqlx.NewDb(db, "postgres"))

		rows := sqlmock.NewRows([]string{"id", "name", "email", "password_hash", "is_active", "created_at", "updated_at", "deleted_at"}).
			AddRow("u1", "A", "a@a.com", "h", true, time.Now(), time.Now(), nil)

		mock.ExpectQuery(`(?is)SELECT COUNT\(\*\) FROM users`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		mock.ExpectQuery(`(?is)SELECT .* FROM users`).WithArgs(10, 0).WillReturnRows(rows)

		res, total, err := repo.ListAll(context.Background(), "", false, 1, 10)
		require.NoError(t, err)
		assert.Len(t, res, 1)
		assert.Equal(t, 1, total)
	})
}

func TestRefreshTokenRepo_Remaining(t *testing.T) {
	t.Run("Create", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		repo := NewRefreshTokenRepo(sqlx.NewDb(db, "postgres"))

		mock.ExpectExec(`(?is)INSERT INTO refresh_tokens`).WithArgs("u1", "h1", sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Create(context.Background(), &domain.RefreshToken{UserID: "u1", TokenHash: "h1", ExpiresAt: time.Now().Add(time.Hour)})
		require.NoError(t, err)
	})

	t.Run("FindByHash", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		repo := NewRefreshTokenRepo(sqlx.NewDb(db, "postgres"))

		rows := sqlmock.NewRows([]string{"id", "user_id", "token_hash", "expires_at", "revoked", "created_at"}).
			AddRow("t1", "u1", "h1", time.Now().Add(time.Hour), false, time.Now())
		mock.ExpectQuery(`(?is)SELECT .* FROM refresh_tokens WHERE token_hash = \$1`).WithArgs("h1").WillReturnRows(rows)

		res, err := repo.FindByHash(context.Background(), "h1")
		require.NoError(t, err)
		assert.NotNil(t, res)

		// Not Found
		mock.ExpectQuery(`(?is)SELECT .* FROM refresh_tokens WHERE token_hash = \$1`).WithArgs("unknown").WillReturnError(sql.ErrNoRows)
		res, err = repo.FindByHash(context.Background(), "unknown")
		assert.NoError(t, err)
		assert.Nil(t, res)
	})

	t.Run("RevokeByID", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		repo := NewRefreshTokenRepo(sqlx.NewDb(db, "postgres"))

		mock.ExpectExec(`(?is)UPDATE refresh_tokens SET revoked = true WHERE id = \$1`).WithArgs("t1").WillReturnResult(sqlmock.NewResult(1, 1))
		err := repo.RevokeByID(context.Background(), "t1")
		assert.NoError(t, err)
	})

	t.Run("RevokeAllByUserID", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		repo := NewRefreshTokenRepo(sqlx.NewDb(db, "postgres"))

		mock.ExpectExec(`(?is)UPDATE refresh_tokens SET revoked = true WHERE user_id = \$1`).WithArgs("u1").WillReturnResult(sqlmock.NewResult(1, 1))
		err := repo.RevokeAllByUserID(context.Background(), "u1")
		assert.NoError(t, err)
	})

	t.Run("DeleteExpired", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		repo := NewRefreshTokenRepo(sqlx.NewDb(db, "postgres"))

		before := time.Now()
		mock.ExpectExec(`(?is)DELETE FROM refresh_tokens WHERE expires_at < \$1`).WithArgs(before).WillReturnResult(sqlmock.NewResult(1, 1))
		err := repo.DeleteExpired(context.Background(), before)
		assert.NoError(t, err)
	})
}

func TestUserRepo_SetStores(t *testing.T) {
	t.Run("SetStores", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		repo := NewUserRepo(sqlx.NewDb(db, "postgres"))

		mock.ExpectBegin()
		mock.ExpectExec(`(?is)UPDATE user_stores SET is_active=false WHERE user_id=\$1`).WithArgs("u1").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`(?is)INSERT INTO user_stores`).WithArgs("u1", "s1", "r1").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.SetStores(context.Background(), "u1", []domain.StoreAssignment{
			{StoreID: "s1", RoleID: "r1"},
		})
		assert.NoError(t, err)
	})
}
