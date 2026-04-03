package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/dto"
	"github.com/moedahpos/backend/internal/validator"
	"github.com/moedahpos/backend/pkg/response"
)

// ─── Interfaces ───────────────────────────────────────────────────────────────

type tableService interface {
	List(ctx context.Context, storeID string) ([]*dto.TableResponse, error)
	Create(ctx context.Context, storeID string, req *dto.CreateTableRequest) (*dto.TableResponse, error)
	Update(ctx context.Context, id string, req *dto.UpdateTableRequest) (*dto.TableResponse, error)
	UpdateStatus(ctx context.Context, id string, status domain.TableStatus) error
	Delete(ctx context.Context, id string) error
}

type menuItemService interface {
	List(ctx context.Context, storeID string) ([]*dto.MenuItemResponse, error)
	Create(ctx context.Context, storeID string, req *dto.CreateMenuItemRequest) (*dto.MenuItemResponse, error)
	Update(ctx context.Context, id string, req *dto.UpdateMenuItemRequest) (*dto.MenuItemResponse, error)
	Delete(ctx context.Context, id string) error
}

// ─── Handler ──────────────────────────────────────────────────────────────────

type RestaurantHandler struct {
	tableSvc tableService
	menuSvc  menuItemService
	validate *validator.Validator
	log      zerolog.Logger
}

func NewRestaurantHandler(
	tableSvc tableService,
	menuSvc menuItemService,
	v *validator.Validator,
	log zerolog.Logger,
) *RestaurantHandler {
	return &RestaurantHandler{tableSvc: tableSvc, menuSvc: menuSvc, validate: v, log: log}
}

// ── Tables ────────────────────────────────────────────────────────────────────

func (h *RestaurantHandler) ListTables(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	tables, err := h.tableSvc.List(r.Context(), storeID)
	if err != nil {
		h.log.Error().Err(err).Msg("list tables failed")
		response.InternalError(w)
		return
	}
	response.Success(w, tables)
}

func (h *RestaurantHandler) CreateTable(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	var req dto.CreateTableRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	if errs := h.validate.ValidateStruct(req); errs != nil {
		response.ValidationError(w, errs)
		return
	}
	table, err := h.tableSvc.Create(r.Context(), storeID, &req)
	if err != nil {
		h.log.Error().Err(err).Msg("create table failed")
		response.InternalError(w)
		return
	}
	response.Created(w, table)
}

func (h *RestaurantHandler) UpdateTable(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "tableId")
	var req dto.UpdateTableRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	if errs := h.validate.ValidateStruct(req); errs != nil {
		response.ValidationError(w, errs)
		return
	}
	table, err := h.tableSvc.Update(r.Context(), id, &req)
	if err != nil {
		h.log.Error().Err(err).Msg("update table failed")
		response.InternalError(w)
		return
	}
	response.Success(w, table)
}

func (h *RestaurantHandler) UpdateTableStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "tableId")
	var req dto.UpdateTableStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	if errs := h.validate.ValidateStruct(req); errs != nil {
		response.ValidationError(w, errs)
		return
	}
	if err := h.tableSvc.UpdateStatus(r.Context(), id, domain.TableStatus(req.Status)); err != nil {
		h.log.Error().Err(err).Msg("update table status failed")
		response.InternalError(w)
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{"success": true, "status": req.Status})
}

func (h *RestaurantHandler) DeleteTable(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "tableId")
	if err := h.tableSvc.Delete(r.Context(), id); err != nil {
		h.log.Error().Err(err).Msg("delete table failed")
		response.InternalError(w)
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{"success": true, "message": "Table deleted"})
}

// ── Menu Items ────────────────────────────────────────────────────────────────

func (h *RestaurantHandler) ListMenuItems(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	items, err := h.menuSvc.List(r.Context(), storeID)
	if err != nil {
		h.log.Error().Err(err).Msg("list menu items failed")
		response.InternalError(w)
		return
	}
	response.Success(w, items)
}

func (h *RestaurantHandler) CreateMenuItem(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	var req dto.CreateMenuItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	if errs := h.validate.ValidateStruct(req); errs != nil {
		response.ValidationError(w, errs)
		return
	}
	item, err := h.menuSvc.Create(r.Context(), storeID, &req)
	if err != nil {
		h.log.Error().Err(err).Msg("create menu item failed")
		response.InternalError(w)
		return
	}
	response.Created(w, item)
}

func (h *RestaurantHandler) UpdateMenuItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "menuItemId")
	var req dto.UpdateMenuItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	if errs := h.validate.ValidateStruct(req); errs != nil {
		response.ValidationError(w, errs)
		return
	}
	item, err := h.menuSvc.Update(r.Context(), id, &req)
	if err != nil {
		h.log.Error().Err(err).Msg("update menu item failed")
		response.InternalError(w)
		return
	}
	response.Success(w, item)
}

func (h *RestaurantHandler) DeleteMenuItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "menuItemId")
	if err := h.menuSvc.Delete(r.Context(), id); err != nil {
		h.log.Error().Err(err).Msg("delete menu item failed")
		response.InternalError(w)
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{"success": true, "message": "Menu item deleted"})
}
