package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/moedahpos/backend/internal/dto"
	"github.com/moedahpos/backend/internal/middleware"
	"github.com/moedahpos/backend/internal/service"
	"github.com/moedahpos/backend/internal/validator"
	"github.com/moedahpos/backend/pkg/response"
)

// AuthHandler handles HTTP requests for the /auth routes.
type AuthHandler struct {
	authSvc   *service.AuthService
	validator *validator.Validator
	log       zerolog.Logger
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authSvc *service.AuthService, v *validator.Validator, log zerolog.Logger) *AuthHandler {
	return &AuthHandler{authSvc: authSvc, validator: v, log: log}
}

// Register handles POST /auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	if errs := h.validator.ValidateStruct(req); errs != nil {
		response.ValidationError(w, errs)
		return
	}

	result, err := h.authSvc.Register(r.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrEmailTaken) {
			response.ValidationError(w, []dto.FieldError{
				{Field: "email", Message: "email already taken"},
			})
			return
		}
		h.log.Error().Err(err).Msg("register failed")
		response.InternalError(w)
		return
	}

	response.Created(w, result)
}

// Login handles POST /auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	if errs := h.validator.ValidateStruct(req); errs != nil {
		response.ValidationError(w, errs)
		return
	}

	result, err := h.authSvc.Login(r.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) || errors.Is(err, service.ErrUserInactive) {
			response.Unauthorized(w, "Invalid email or password")
			return
		}
		h.log.Error().Err(err).Msg("login failed")
		response.InternalError(w)
		return
	}

	response.Success(w, result)
}

// Refresh handles POST /auth/refresh
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	if errs := h.validator.ValidateStruct(req); errs != nil {
		response.ValidationError(w, errs)
		return
	}

	result, err := h.authSvc.Refresh(r.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidToken) {
			response.Unauthorized(w, "Invalid or expired refresh token")
			return
		}
		h.log.Error().Err(err).Msg("token refresh failed")
		response.InternalError(w)
		return
	}

	response.Success(w, result)
}

// Logout handles POST /auth/logout (requires JWT)
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		response.Unauthorized(w, "")
		return
	}

	if err := h.authSvc.Logout(r.Context(), userID); err != nil {
		h.log.Error().Err(err).Msg("logout failed")
		response.InternalError(w)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Logged out successfully",
	})
}

// Me handles GET /auth/me (requires JWT)
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		response.Unauthorized(w, "")
		return
	}

	result, err := h.authSvc.Me(r.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			response.NotFound(w, "User")
			return
		}
		h.log.Error().Err(err).Msg("me failed")
		response.InternalError(w)
		return
	}

	response.Success(w, result)
}
