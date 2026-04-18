package service

import (
	"context"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/dto"
	"github.com/moedahpos/backend/mocks"
)

func TestStoreService_ListStores(t *testing.T) {
	storeRepo := new(mocks.StoreRepository)
	userRepo := new(mocks.UserRepository)
	log := zerolog.Nop()
	svc := NewStoreService(storeRepo, userRepo, log)

	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		filter := dto.StoreListFilter{}
		stores := []*domain.Store{
			{ID: "s1", Name: "Store 1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		}
		storeRepo.On("FindAll", ctx, mock.Anything).Return(stores, 1, nil).Once()

		resp, meta, err := svc.ListStores(ctx, filter)
		assert.NoError(t, err)
		assert.Len(t, resp, 1)
		assert.Equal(t, 1, meta.Total)
		storeRepo.AssertExpectations(t)
	})
}

func TestStoreService_GetStore(t *testing.T) {
	storeRepo := new(mocks.StoreRepository)
	userRepo := new(mocks.UserRepository)
	log := zerolog.Nop()
	svc := NewStoreService(storeRepo, userRepo, log)

	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		store := &domain.Store{ID: "s1", Name: "Store 1", CreatedAt: time.Now(), UpdatedAt: time.Now()}
		storeRepo.On("FindByID", ctx, "s1").Return(store, nil).Once()

		resp, err := svc.GetStore(ctx, "s1")
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "Store 1", resp.Name)
	})

	t.Run("Not Found", func(t *testing.T) {
		storeRepo.On("FindByID", ctx, "s2").Return(nil, nil).Once()
		resp, err := svc.GetStore(ctx, "s2")
		assert.ErrorIs(t, err, ErrStoreNotFound)
		assert.Nil(t, resp)
	})
}

func TestStoreService_CreateStore(t *testing.T) {
	storeRepo := new(mocks.StoreRepository)
	userRepo := new(mocks.UserRepository)
	log := zerolog.Nop()
	svc := NewStoreService(storeRepo, userRepo, log)

	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		req := &dto.CreateStoreRequest{Name: "New Store"}
		store := &domain.Store{ID: "s1", Name: "New Store", CreatedAt: time.Now(), UpdatedAt: time.Now()}
		storeRepo.On("Create", ctx, mock.Anything).Return(store, nil).Once()

		resp, err := svc.CreateStore(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "New Store", resp.Name)
	})
}

func TestStoreService_UpdateStore(t *testing.T) {
	storeRepo := new(mocks.StoreRepository)
	userRepo := new(mocks.UserRepository)
	log := zerolog.Nop()
	svc := NewStoreService(storeRepo, userRepo, log)

	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		req := &dto.UpdateStoreRequest{Name: "Updated Store"}
		store := &domain.Store{ID: "s1", Name: "Old Store", CreatedAt: time.Now(), UpdatedAt: time.Now()}
		storeRepo.On("FindByID", ctx, "s1").Return(store, nil).Once()
		storeRepo.On("Update", ctx, mock.Anything).Return(store, nil).Once()

		resp, err := svc.UpdateStore(ctx, "s1", req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "Updated Store", resp.Name)
	})
}

func TestStoreService_DeleteStore(t *testing.T) {
	storeRepo := new(mocks.StoreRepository)
	userRepo := new(mocks.UserRepository)
	log := zerolog.Nop()
	svc := NewStoreService(storeRepo, userRepo, log)

	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		store := &domain.Store{ID: "s1"}
		storeRepo.On("FindByID", ctx, "s1").Return(store, nil).Once()
		storeRepo.On("SoftDelete", ctx, "s1").Return(nil).Once()

		err := svc.DeleteStore(ctx, "s1")
		assert.NoError(t, err)
	})
}

func TestStoreService_Members(t *testing.T) {
	storeRepo := new(mocks.StoreRepository)
	userRepo := new(mocks.UserRepository)
	log := zerolog.Nop()
	svc := NewStoreService(storeRepo, userRepo, log)

	ctx := context.Background()

	t.Run("ListMembers_Success", func(t *testing.T) {
		storeRepo.On("FindByID", ctx, "s1").Return(&domain.Store{ID: "s1"}, nil).Once()
		members := []*domain.StoreMember{
			{UserID: "u1", UserName: "User 1", JoinedAt: time.Now()},
		}
		storeRepo.On("ListMembers", ctx, "s1").Return(members, nil).Once()

		resp, err := svc.ListMembers(ctx, "s1")
		assert.NoError(t, err)
		assert.Len(t, resp, 1)
		assert.Equal(t, "User 1", resp[0].UserName)
	})

	t.Run("AddMember_Success", func(t *testing.T) {
		req := &dto.AddMemberRequest{UserID: "u1", RoleID: "r1"}
		storeRepo.On("FindByID", ctx, "s1").Return(&domain.Store{ID: "s1"}, nil).Once()
		storeRepo.On("FindMember", ctx, "u1", "s1").Return(nil, nil).Once()
		storeRepo.On("AddMember", ctx, mock.Anything).Return(nil).Once()

		err := svc.AddMember(ctx, "s1", req)
		assert.NoError(t, err)
	})

	t.Run("AddMember_AlreadyExists", func(t *testing.T) {
		req := &dto.AddMemberRequest{UserID: "u1", RoleID: "r1"}
		storeRepo.On("FindByID", ctx, "s1").Return(&domain.Store{ID: "s1"}, nil).Once()
		storeRepo.On("FindMember", ctx, "u1", "s1").Return(&domain.UserStore{IsActive: true}, nil).Once()

		err := svc.AddMember(ctx, "s1", req)
		assert.ErrorIs(t, err, ErrMemberAlreadyExists)
	})

	t.Run("UpdateMemberRole_Success", func(t *testing.T) {
		req := &dto.UpdateMemberRoleRequest{RoleID: "r2"}
		storeRepo.On("FindMember", ctx, "u1", "s1").Return(&domain.UserStore{UserID: "u1"}, nil).Once()
		storeRepo.On("UpdateMemberRole", ctx, "u1", "s1", "r2").Return(nil).Once()

		err := svc.UpdateMemberRole(ctx, "s1", "u1", req)
		assert.NoError(t, err)
	})

	t.Run("RemoveMember_Success", func(t *testing.T) {
		storeRepo.On("FindMember", ctx, "u1", "s1").Return(&domain.UserStore{UserID: "u1"}, nil).Once()
		storeRepo.On("DeactivateMember", ctx, "u1", "s1").Return(nil).Once()

		err := svc.RemoveMember(ctx, "s1", "u1")
		assert.NoError(t, err)
	})

	t.Run("RemoveMember_NotFound", func(t *testing.T) {
		storeRepo.On("FindMember", ctx, "u2", "s1").Return(nil, nil).Once()
		err := svc.RemoveMember(ctx, "s1", "u2")
		assert.ErrorIs(t, err, ErrMemberNotFound)
	})
}
