package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/dto"
	"github.com/moedahpos/backend/internal/service"
	"github.com/moedahpos/backend/internal/service/mocks"
	"github.com/moedahpos/backend/internal/validator"
)

func TestUserAdminHandler(t *testing.T) {
	svc := new(mocks.UserAdminServiceInterface)
	v := validator.New()
	log := zerolog.Nop()
	h := NewUserAdminHandler(svc, v, log)

	t.Run("List", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/admin/users?search=test&page=1", nil)
		w := httptest.NewRecorder()

		svc.On("ListUsers", mock.Anything, "test", false, 1, 20).Return([]dto.UserResponse{{ID: "u1"}}, 1, nil).Once()

		h.List(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("Get", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/admin/users/u1", nil)
		w := httptest.NewRecorder()

		// Inject Chi URL Param
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("userId", "u1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		svc.On("GetUser", mock.Anything, "u1").Return(&dto.UserResponse{ID: "u1"}, nil).Once()

		h.Get(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("Get_NotFound", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/admin/users/u2", nil)
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("userId", "u2")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		svc.On("GetUser", mock.Anything, "u2").Return(nil, service.ErrAdminUserNotFound).Once()

		h.Get(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Create", func(t *testing.T) {
		reqBody := dto.CreateUserRequest{
			Name:     "New User",
			Email:    "new@example.com",
			Password: "password123",
		}
		body, _ := json.Marshal(reqBody) // nolint:gosec
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/admin/users", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		svc.On("CreateUser", mock.Anything, mock.Anything).Return(&dto.UserResponse{ID: "u1"}, nil).Once()

		h.Create(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("Create_ValidationError", func(t *testing.T) {
		reqBody := dto.CreateUserRequest{Name: "U"} // missing email, short name
		body, _ := json.Marshal(reqBody)            // nolint:gosec
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/admin/users", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		h.Create(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("Update", func(t *testing.T) {
		reqBody := dto.UpdateUserRequest{Name: "Updated", Email: "up@eg.com"}
		body, _ := json.Marshal(reqBody) // nolint:gosec
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPut, "/admin/users/u1", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("userId", "u1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		svc.On("UpdateUser", mock.Anything, "u1", mock.Anything).Return(&dto.UserResponse{ID: "u1"}, nil).Once()

		h.Update(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Deactivate", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/admin/users/u1/deactivate", nil)
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("userId", "u1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		svc.On("DeactivateUser", mock.Anything, "u1").Return(nil).Once()

		h.Deactivate(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("ResetPassword", func(t *testing.T) {
		reqBody := dto.ResetPasswordRequest{Password: "newpassword"}
		body, _ := json.Marshal(reqBody) // nolint:gosec
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/admin/users/u1/reset-password", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("userId", "u1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		svc.On("ResetPassword", mock.Anything, "u1", mock.Anything).Return(nil).Once()

		h.ResetPassword(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("ListRoles", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/admin/roles", nil)
		w := httptest.NewRecorder()

		svc.On("ListRoles", mock.Anything).Return([]*domain.Role{{ID: "r1"}}, nil).Once()

		h.ListRoles(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
