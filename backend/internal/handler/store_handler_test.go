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

	"github.com/moedahpos/backend/internal/dto"
	"github.com/moedahpos/backend/internal/service/mocks"
	"github.com/moedahpos/backend/internal/validator"
)

func setupStoreHandler() (*mocks.StoreServiceInterface, *StoreHandler) {
	svc := new(mocks.StoreServiceInterface)
	v := validator.New()
	log := zerolog.Nop()
	h := NewStoreHandler(svc, v, log)
	return svc, h
}

func TestStoreHandler_List(t *testing.T) {
	svc, h := setupStoreHandler()

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/stores", nil)
	w := httptest.NewRecorder()

	svc.On("ListStores", mock.Anything, mock.Anything).Return([]*dto.StoreResponse{{ID: "s1"}}, dto.PaginationMeta{Total: 1}, nil).Once()

	h.List(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestStoreHandler_Get(t *testing.T) {
	svc, h := setupStoreHandler()

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/stores/s1", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("storeId", "s1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	svc.On("GetStore", mock.Anything, "s1").Return(&dto.StoreResponse{ID: "s1"}, nil).Once()

	h.Get(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestStoreHandler_Create(t *testing.T) {
	svc, h := setupStoreHandler()

	t.Run("success", func(t *testing.T) {
		reqBody := dto.CreateStoreRequest{
			Name:                   "New Store",
			Currency:               "IDR",
			StoreType:              "retail",
			LoyaltyPointsPerRupiah: 1,
			LoyaltyRupiahPerPoint:  1,
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/stores", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		svc.On("CreateStore", mock.Anything, mock.Anything).Return(&dto.StoreResponse{ID: "s1"}, nil).Once()

		h.Create(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("Validation Error", func(t *testing.T) {
		reqBody := dto.CreateStoreRequest{Name: ""}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/stores", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		h.Create(w, req)
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})
}

func TestStoreHandler_Update(t *testing.T) {
	svc, h := setupStoreHandler()

	t.Run("success", func(t *testing.T) {
		reqBody := dto.UpdateStoreRequest{
			Name:      "Updated Name",
			Currency:  "IDR",
			StoreType: "retail",
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/stores/s1", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		svc.On("UpdateStore", mock.Anything, "s1", mock.Anything).Return(&dto.StoreResponse{ID: "s1"}, nil).Once()

		h.Update(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("Validation Error", func(t *testing.T) {
		reqBody := dto.UpdateStoreRequest{Name: ""}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/stores/s1", bytes.NewBuffer(body))
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		h.Update(w, req)
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})
}

func TestStoreHandler_Delete(t *testing.T) {
	svc, h := setupStoreHandler()

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodDelete, "/stores/s1", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("storeId", "s1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	svc.On("DeleteStore", mock.Anything, "s1").Return(nil).Once()

	h.Delete(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestStoreHandler_ListMembers(t *testing.T) {
	svc, h := setupStoreHandler()

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/stores/s1/members", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("storeId", "s1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	svc.On("ListMembers", mock.Anything, "s1").Return([]*dto.MemberResponse{{UserID: "u1"}}, nil).Once()

	h.ListMembers(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestStoreHandler_AddMember(t *testing.T) {
	svc, h := setupStoreHandler()

	t.Run("success", func(t *testing.T) {
		reqBody := dto.AddMemberRequest{UserID: "00000000-0000-0000-0000-000000000001", RoleID: "00000000-0000-0000-0000-000000000002"}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/stores/s1/members", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		svc.On("AddMember", mock.Anything, "s1", mock.Anything).Return(nil).Once()

		h.AddMember(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("Validation Error", func(t *testing.T) {
		reqBody := dto.AddMemberRequest{UserID: "invalid"}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/stores/s1/members", bytes.NewBuffer(body))
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		h.AddMember(w, req)
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})
}

func TestStoreHandler_UpdateMemberRole(t *testing.T) {
	svc, h := setupStoreHandler()

	reqBody := dto.UpdateMemberRoleRequest{RoleID: "00000000-0000-0000-0000-000000000002"}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPut, "/stores/s1/members/u1", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("storeId", "s1")
	rctx.URLParams.Add("userId", "u1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	svc.On("UpdateMemberRole", mock.Anything, "s1", "u1", mock.Anything).Return(nil).Once()

	h.UpdateMemberRole(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestStoreHandler_RemoveMember(t *testing.T) {
	svc, h := setupStoreHandler()

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodDelete, "/stores/s1/members/u1", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("storeId", "s1")
	rctx.URLParams.Add("userId", "u1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	svc.On("RemoveMember", mock.Anything, "s1", "u1").Return(nil).Once()

	h.RemoveMember(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}
