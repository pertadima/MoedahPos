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

	"github.com/moedahpos/backend/pkg/rbac"
)

func TestStoreContext(t *testing.T) {
	db, mock, _ := sqlmock.New()
	sqlxDB := sqlx.NewDb(db, "postgres")

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sid := StoreIDFromContext(r.Context())
		assert.Equal(t, "s1", sid)
		role := StoreRoleFromContext(r.Context())
		assert.Equal(t, "admin", role)
		w.WriteHeader(http.StatusOK)
	})

	mw := StoreContext(sqlxDB)
	
	t.Run("Extract StoreID", func(t *testing.T) {
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
		
		w := httptest.NewRecorder()
		mw(h).ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestRequirePermission(t *testing.T) {
	db, mock, _ := sqlmock.New()
	sqlxDB := sqlx.NewDb(db, "postgres")
	
	mock.ExpectQuery("SELECT r.name AS role_name, p.name AS permission_name").
		WillReturnRows(sqlmock.NewRows([]string{"role_name", "permission_name"}).
			AddRow("admin", "test:perm"))
	
	rs, _ := rbac.New(sqlxDB)

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("Forbidden", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		
		mw := RequirePermission(rs, "other:perm")
		mw(h).ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Success", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
		// Inject role into context
		ctx := context.WithValue(req.Context(), storeRoleContextKey, "admin")
		req = req.WithContext(ctx)
		
		w := httptest.NewRecorder()
		
		mw := RequirePermission(rs, "test:perm")
		mw(h).ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
