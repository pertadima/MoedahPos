package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/moedahpos/backend/internal/dto"
	"github.com/moedahpos/backend/internal/middleware"
	"github.com/moedahpos/backend/internal/service"
	"github.com/moedahpos/backend/internal/validator"
	"github.com/moedahpos/backend/pkg/response"
)

// StoreHandler handles HTTP requests for store and member endpoints.
type StoreHandler struct {
	storeSvc  service.StoreServiceInterface
	validator *validator.Validator
	log       zerolog.Logger
}

func NewStoreHandler(storeSvc service.StoreServiceInterface, v *validator.Validator, log zerolog.Logger) *StoreHandler {
	return &StoreHandler{storeSvc: storeSvc, validator: v, log: log}
}

// GET /stores
func (h *StoreHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := dto.StoreListFilter{}
	filter.Page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	filter.PerPage, _ = strconv.Atoi(r.URL.Query().Get("per_page"))
	filter.Search = r.URL.Query().Get("search")
	if v := r.URL.Query().Get("is_active"); v != "" {
		b, err := strconv.ParseBool(v)
		if err == nil {
			filter.IsActive = &b
		}
	}

	stores, meta, err := h.storeSvc.ListStores(r.Context(), filter)
	if err != nil {
		h.log.Error().Err(err).Msg("list stores failed")
		response.InternalError(w)
		return
	}
	response.Success(w, dto.ListResponse{Data: stores, Meta: meta})
}

// POST /stores
func (h *StoreHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateStoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	if errs := h.validator.ValidateStruct(req); errs != nil {
		response.ValidationError(w, errs)
		return
	}
	result, err := h.storeSvc.CreateStore(r.Context(), &req)
	if err != nil {
		h.log.Error().Err(err).Msg("create store failed")
		response.InternalError(w)
		return
	}
	response.Created(w, result)
}

// GET /stores/:storeId
func (h *StoreHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "storeId")
	result, err := h.storeSvc.GetStore(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrStoreNotFound) {
			response.NotFound(w, "Store")
			return
		}
		response.InternalError(w)
		return
	}
	response.Success(w, result)
}

// PUT /stores/:storeId
func (h *StoreHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "storeId")
	var req dto.UpdateStoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	if errs := h.validator.ValidateStruct(req); errs != nil {
		response.ValidationError(w, errs)
		return
	}
	result, err := h.storeSvc.UpdateStore(r.Context(), id, &req)
	if err != nil {
		if errors.Is(err, service.ErrStoreNotFound) {
			response.NotFound(w, "Store")
			return
		}
		response.InternalError(w)
		return
	}
	response.Success(w, result)
}

// DELETE /stores/:storeId  (soft delete — sets deleted_at)
func (h *StoreHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "storeId")
	if err := h.storeSvc.DeleteStore(r.Context(), id); err != nil {
		if errors.Is(err, service.ErrStoreNotFound) {
			response.NotFound(w, "Store")
			return
		}
		response.InternalError(w)
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Store deleted successfully",
	})
}

// ─── Members ──────────────────────────────────────────────────────────────────

// GET /stores/:storeId/members
func (h *StoreHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	members, err := h.storeSvc.ListMembers(r.Context(), storeID)
	if err != nil {
		if errors.Is(err, service.ErrStoreNotFound) {
			response.NotFound(w, "Store")
			return
		}
		response.InternalError(w)
		return
	}
	response.Success(w, members)
}

// POST /stores/:storeId/members
func (h *StoreHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	var req dto.AddMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	if errs := h.validator.ValidateStruct(req); errs != nil {
		response.ValidationError(w, errs)
		return
	}
	if err := h.storeSvc.AddMember(r.Context(), storeID, &req); err != nil {
		if errors.Is(err, service.ErrMemberAlreadyExists) {
			response.Error(w, http.StatusConflict, "User is already an active member of this store")
			return
		}
		if errors.Is(err, service.ErrStoreNotFound) {
			response.NotFound(w, "Store")
			return
		}
		response.InternalError(w)
		return
	}
	response.JSON(w, http.StatusCreated, map[string]interface{}{
		"success": true, "message": "Member added successfully",
	})
}

// PUT /stores/:storeId/members/:userId
func (h *StoreHandler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	userID := chi.URLParam(r, "userId")
	var req dto.UpdateMemberRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	if errs := h.validator.ValidateStruct(req); errs != nil {
		response.ValidationError(w, errs)
		return
	}
	if err := h.storeSvc.UpdateMemberRole(r.Context(), storeID, userID, &req); err != nil {
		if errors.Is(err, service.ErrMemberNotFound) {
			response.NotFound(w, "Member")
			return
		}
		response.InternalError(w)
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true, "message": "Member role updated",
	})
}

// DELETE /stores/:storeId/members/:userId  (soft — sets is_active=false)
func (h *StoreHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	userID := chi.URLParam(r, "userId")
	if err := h.storeSvc.RemoveMember(r.Context(), storeID, userID); err != nil {
		if errors.Is(err, service.ErrMemberNotFound) {
			response.NotFound(w, "Member")
			return
		}
		response.InternalError(w)
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true, "message": "Member removed from store",
	})
}

// userIDFromCtx is a shorthand helper used in handlers.
func userIDFromCtx(r *http.Request) string {
	return middleware.UserIDFromContext(r.Context())
}
