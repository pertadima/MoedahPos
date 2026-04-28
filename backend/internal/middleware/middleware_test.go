package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestStoreContext(t *testing.T) {
	db, mock, _ := sqlmock.New()
	sqlxDB := sqlx.NewDb(db, "postgres")

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sid := StoreIDFromContext(r.Context())
		assert.Equal(t, "s1", sid)
		roles := StoreRolesFromContext(r.Context())
		assert.ElementsMatch(t, []string{"admin"}, roles)
		perms := StorePermissionsFromContext(r.Context())
		assert.ElementsMatch(t, []string{"inventory:read"}, perms)
		w.WriteHeader(http.StatusOK)
	})

	mw := StoreContext(sqlxDB)

	t.Run("Extract StoreID and Roles/Permissions", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/stores/s1", nil)
		// Inject userID into context
		ctx := context.WithValue(req.Context(), UserIDKey, "u123")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
		req = req.WithContext(ctx)

		mock.ExpectQuery("SELECT r.name").
			WithArgs("u123", "s1").
			WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("admin"))

		mock.ExpectQuery("SELECT DISTINCT p.name").
			WithArgs("u123", "s1").
			WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("inventory:read"))

		w := httptest.NewRecorder()
		mw(h).ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestRequirePermission(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("Forbidden", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
		ctx := context.WithValue(req.Context(), storePermissionsContextKey, []string{"inventory:read"})
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		mw := RequirePermission("keuangan:read")
		mw(h).ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Success via Permission", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
		ctx := context.WithValue(req.Context(), storePermissionsContextKey, []string{"keuangan:read"})
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		mw := RequirePermission("keuangan:read")
		mw(h).ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Success via Superadmin", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
		ctx := context.WithValue(req.Context(), storeRolesContextKey, []string{"superadmin"})
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		mw := RequirePermission("keuangan:read") // shouldn't matter
		mw(h).ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestAdminMiddlewares(t *testing.T) {
	db, mock, _ := sqlmock.New()
	sqlxDB := sqlx.NewDb(db, "postgres")
	h := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("RequireAdminOrSuperAdmin_Success", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
		req = req.WithContext(context.WithValue(req.Context(), UserIDKey, "u1"))
		w := httptest.NewRecorder()

		mock.ExpectQuery("SELECT COUNT").WithArgs("u1").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		mw := RequireAdminOrSuperAdmin(sqlxDB)
		mw(h).ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("RequireSuperAdmin_Forbidden", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
		req = req.WithContext(context.WithValue(req.Context(), UserIDKey, "u1"))
		w := httptest.NewRecorder()

		mock.ExpectQuery("SELECT COUNT").WithArgs("u1").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		mw := RequireSuperAdmin(sqlxDB)
		mw(h).ServeHTTP(w, req)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}
