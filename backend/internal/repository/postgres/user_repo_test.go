package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	"github.com/moedahpos/backend/internal/domain"
)

func TestUserRepo_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewUserRepo(sqlxDB)

	user := &domain.User{
		Name:         "Admin",
		Email:        "admin@example.com",
		PasswordHash: "hash",
	}

	rows := sqlmock.NewRows([]string{"id", "name", "email", "password_hash", "is_active", "created_at", "updated_at", "deleted_at"}).
		AddRow("u1", user.Name, user.Email, user.PasswordHash, true, time.Now(), time.Now(), nil)

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs(user.Name, user.Email, user.PasswordHash).
		WillReturnRows(rows)

	res, err := repo.Create(context.Background(), user)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "u1", res.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepo_FindByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewUserRepo(sqlxDB)

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "email", "password_hash", "is_active", "created_at", "updated_at", "deleted_at"}).
			AddRow("u1", "Name", "email@test.com", "hash", true, time.Now(), time.Now(), nil)

		mock.ExpectQuery(`SELECT .* FROM users WHERE id = \$1 AND deleted_at IS NULL`).
			WithArgs("u1").
			WillReturnRows(rows)

		user, err := repo.FindByID(context.Background(), "u1")

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "u1", user.ID)
	})

	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectQuery(`SELECT .* FROM users WHERE id = \$1 AND deleted_at IS NULL`).
			WithArgs("u2").
			WillReturnError(sql.ErrNoRows)

		user, err := repo.FindByID(context.Background(), "u2")

		assert.NoError(t, err)
		assert.Nil(t, user)
	})
}

func TestRefreshTokenRepo_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewRefreshTokenRepo(sqlxDB)

	token := &domain.RefreshToken{
		UserID:    "u1",
		TokenHash: "hash",
		ExpiresAt: time.Now().Add(time.Hour),
	}

	mock.ExpectExec(`INSERT INTO refresh_tokens`).
		WithArgs(token.UserID, token.TokenHash, token.ExpiresAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Create(context.Background(), token)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRefreshTokenRepo_FindByHash(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewRefreshTokenRepo(sqlxDB)

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "user_id", "token_hash", "expires_at", "revoked", "created_at"}).
			AddRow("t1", "u1", "hash", time.Now().Add(time.Hour), false, time.Now())

		mock.ExpectQuery(`SELECT .* FROM refresh_tokens WHERE token_hash = \$1`).
			WithArgs("hash").
			WillReturnRows(rows)

		res, err := repo.FindByHash(context.Background(), "hash")

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "t1", res.ID)
	})
}
