package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/moedahpos/backend/internal/dto"
	"github.com/moedahpos/backend/internal/middleware"
	"github.com/moedahpos/backend/internal/service"
	"github.com/moedahpos/backend/internal/service/mocks"
	"github.com/moedahpos/backend/internal/validator"
	"github.com/moedahpos/backend/pkg/jwt"
)

func TestAuthHandler_Register(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		aSvc := new(mocks.AuthServiceInterface)
		v := validator.New()
		log := zerolog.Nop()
		h := NewAuthHandler(aSvc, v, log)

		reqBody := dto.RegisterRequest{
			Name:     "Test",
			Email:    "test@example.com",
			Password: "password",
		}
		body, _ := json.Marshal(reqBody) // nolint:gosec
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/auth/register", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		aSvc.On("Register", mock.Anything, &reqBody).Return(&dto.RegisterResponse{ID: "1"}, nil)

		h.Register(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		aSvc.AssertExpectations(t)
	})

	t.Run("Email Taken", func(t *testing.T) {
		aSvc := new(mocks.AuthServiceInterface)
		v := validator.New()
		h := NewAuthHandler(aSvc, v, zerolog.Nop())

		reqBody := dto.RegisterRequest{Email: "taken@example.com", Name: "N", Password: "P"}
		body, _ := json.Marshal(reqBody) // nolint:gosec
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/auth/register", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		aSvc.On("Register", mock.Anything, &reqBody).Return(nil, service.ErrEmailTaken)

		h.Register(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		aSvc := new(mocks.AuthServiceInterface)
		v := validator.New()
		h := NewAuthHandler(aSvc, v, zerolog.Nop())

		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/auth/register", bytes.NewBufferString("invalid"))
		w := httptest.NewRecorder()

		h.Register(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Validation Error", func(t *testing.T) {
		aSvc := new(mocks.AuthServiceInterface)
		v := validator.New()
		h := NewAuthHandler(aSvc, v, zerolog.Nop())

		reqBody := dto.RegisterRequest{Email: "invalid"}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/auth/register", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		h.Register(w, req)
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})
}

func TestAuthHandler_Login(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		aSvc := new(mocks.AuthServiceInterface)
		v := validator.New()
		h := NewAuthHandler(aSvc, v, zerolog.Nop())

		reqBody := dto.LoginRequest{Email: "u@e.com", Password: "p"}
		body, _ := json.Marshal(reqBody) // nolint:gosec
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/auth/login", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		aSvc.On("Login", mock.Anything, &reqBody).Return(&dto.LoginResponse{AccessToken: "at"}, nil)

		h.Login(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Validation Error", func(t *testing.T) {
		aSvc := new(mocks.AuthServiceInterface)
		v := validator.New()
		h := NewAuthHandler(aSvc, v, zerolog.Nop())

		reqBody := dto.LoginRequest{Email: "invalid"}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/auth/login", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		h.Login(w, req)
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("Invalid Credentials", func(t *testing.T) {
		aSvc := new(mocks.AuthServiceInterface)
		v := validator.New()
		h := NewAuthHandler(aSvc, v, zerolog.Nop())

		reqBody := dto.LoginRequest{Email: "u@e.com", Password: "p"}
		body, _ := json.Marshal(reqBody) // nolint:gosec
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/auth/login", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		aSvc.On("Login", mock.Anything, &reqBody).Return(nil, service.ErrInvalidCredentials)

		h.Login(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestAuthHandler_Refresh(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		aSvc := new(mocks.AuthServiceInterface)
		v := validator.New()
		h := NewAuthHandler(aSvc, v, zerolog.Nop())

		reqBody := dto.RefreshRequest{RefreshToken: "rt"}
		body, _ := json.Marshal(reqBody) // nolint:gosec
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/auth/refresh", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		aSvc.On("Refresh", mock.Anything, &reqBody).Return(&dto.RefreshResponse{AccessToken: "new-at"}, nil)

		h.Refresh(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Validation Error", func(t *testing.T) {
		aSvc := new(mocks.AuthServiceInterface)
		v := validator.New()
		h := NewAuthHandler(aSvc, v, zerolog.Nop())

		reqBody := dto.RefreshRequest{RefreshToken: ""}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/auth/refresh", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		h.Refresh(w, req)
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})
}

func TestAuthHandler_Logout(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		aSvc := new(mocks.AuthServiceInterface)
		v := validator.New()
		log := zerolog.Nop()
		h := NewAuthHandler(aSvc, v, log)

		// Setup JWT and Middleware
		jwtMgr := jwt.New("secret", time.Hour, time.Hour)
		token, _ := jwtMgr.GenerateAccessToken("u123", "test@example.com")
		authMiddleware := middleware.Authenticate(jwtMgr)

		// Mock Service
		aSvc.On("Logout", mock.Anything, "u123").Return(nil)

		// Request
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/auth/logout", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		// Wrap handler with middleware to inject user_id
		handler := authMiddleware(http.HandlerFunc(h.Logout))
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		aSvc.AssertExpectations(t)
	})
}

func TestAuthHandler_Me(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		aSvc := new(mocks.AuthServiceInterface)
		v := validator.New()
		log := zerolog.Nop()
		h := NewAuthHandler(aSvc, v, log)

		jwtMgr := jwt.New("secret", time.Hour, time.Hour)
		token, _ := jwtMgr.GenerateAccessToken("u123", "test@example.com")
		authMiddleware := middleware.Authenticate(jwtMgr)

		aSvc.On("Me", mock.Anything, "u123").Return(&dto.MeResponse{ID: "u123", Email: "test@example.com"}, nil)

		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/auth/me", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		handler := authMiddleware(http.HandlerFunc(h.Me))
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		aSvc.AssertExpectations(t)
	})
}
