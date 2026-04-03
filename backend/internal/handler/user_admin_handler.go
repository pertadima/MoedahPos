package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/moedahpos/backend/internal/dto"
	"github.com/moedahpos/backend/internal/service"
	"github.com/moedahpos/backend/internal/validator"
	"github.com/moedahpos/backend/pkg/response"
)

// UserAdminHandler handles HTTP requests for system-wide user management.
type UserAdminHandler struct {
	svc       *service.UserAdminService
	validator *validator.Validator
	log       zerolog.Logger
}

// NewUserAdminHandler constructs a UserAdminHandler.
func NewUserAdminHandler(svc *service.UserAdminService, v *validator.Validator, log zerolog.Logger) *UserAdminHandler {
	return &UserAdminHandler{svc: svc, validator: v, log: log}
}

// List godoc — GET /admin/users
func (h *UserAdminHandler) List(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	includeInactive := r.URL.Query().Get("include_inactive") == "true"
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}

	users, total, err := h.svc.ListUsers(r.Context(), search, includeInactive, page, perPage)
	if err != nil {
		h.log.Error().Err(err).Msg("UserAdminHandler.List")
		response.Error(w, http.StatusInternalServerError, "Failed to fetch users")
		return
	}

	meta := dto.NewMeta(dto.PaginationQuery{Page: page, PerPage: perPage}, total)
	response.JSON(w, http.StatusOK, map[string]any{"data": users, "meta": meta})
}

// Get godoc — GET /admin/users/:userId
func (h *UserAdminHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "userId")
	u, err := h.svc.GetUser(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrAdminUserNotFound) {
			response.Error(w, http.StatusNotFound, "User not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Failed to fetch user")
		return
	}
	response.JSON(w, http.StatusOK, u)
}

// Create godoc — POST /admin/users
func (h *UserAdminHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	if errs := h.validator.ValidateStruct(req); len(errs) > 0 {
		response.ValidationError(w, errs)
		return
	}

	u, err := h.svc.CreateUser(r.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrAdminEmailConflict) {
			response.Error(w, http.StatusConflict, "Email already in use")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Failed to create user")
		return
	}
	response.JSON(w, http.StatusCreated, u)
}

// Update godoc — PUT /admin/users/:userId
func (h *UserAdminHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "userId")
	var req dto.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	if errs := h.validator.ValidateStruct(req); len(errs) > 0 {
		response.ValidationError(w, errs)
		return
	}

	u, err := h.svc.UpdateUser(r.Context(), id, &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAdminUserNotFound):
			response.Error(w, http.StatusNotFound, "User not found")
		case errors.Is(err, service.ErrAdminEmailConflict):
			response.Error(w, http.StatusConflict, "Email already in use")
		default:
			response.Error(w, http.StatusInternalServerError, "Failed to update user")
		}
		return
	}
	response.JSON(w, http.StatusOK, u)
}

// Deactivate godoc — POST /admin/users/:userId/deactivate (soft-delete)
func (h *UserAdminHandler) Deactivate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "userId")
	if err := h.svc.DeactivateUser(r.Context(), id); err != nil {
		if errors.Is(err, service.ErrAdminUserNotFound) {
			response.Error(w, http.StatusNotFound, "User not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Failed to deactivate user")
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"success": true, "message": "User deactivated"})
}

// ResetPassword godoc — POST /admin/users/:userId/reset-password
func (h *UserAdminHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "userId")
	var req dto.ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	if errs := h.validator.ValidateStruct(req); len(errs) > 0 {
		response.ValidationError(w, errs)
		return
	}
	if err := h.svc.ResetPassword(r.Context(), id, &req); err != nil {
		if errors.Is(err, service.ErrAdminUserNotFound) {
			response.Error(w, http.StatusNotFound, "User not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Failed to reset password")
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"success": true, "message": "Password updated"})
}

// SetStores godoc — PUT /admin/users/:userId/stores
func (h *UserAdminHandler) SetStores(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "userId")
	var req dto.SetUserStoresRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	u, err := h.svc.SetUserStores(r.Context(), id, &req)
	if err != nil {
		if errors.Is(err, service.ErrAdminUserNotFound) {
			response.Error(w, http.StatusNotFound, "User not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Failed to update store assignments")
		return
	}
	response.JSON(w, http.StatusOK, u)
}

// ListRoles godoc — GET /admin/roles
func (h *UserAdminHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := h.svc.ListRoles(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to fetch roles")
		return
	}
	resp := make([]dto.RoleResponse, 0, len(roles))
	for _, ro := range roles {
		resp = append(resp, dto.RoleResponse{
			ID:          ro.ID,
			Name:        ro.Name,
			Description: ro.Description,
			Permissions: ro.Permissions,
		})
	}
	response.JSON(w, http.StatusOK, map[string]any{"data": resp})
}
