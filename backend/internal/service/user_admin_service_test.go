package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/dto"
	"github.com/moedahpos/backend/mocks"
)

func TestUserAdminService_ListUsers(t *testing.T) {
	userRepo := new(mocks.UserRepository)
	roleRepo := new(mocks.RoleRepository)
	log := zerolog.Nop()
	svc := NewUserAdminService(userRepo, roleRepo, 10, log)

	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		users := []*domain.User{
			{ID: "u1", Name: "User 1", Email: "u1@example.com", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		}
		userRepo.On("ListAll", ctx, "search", false, 1, 20).Return(users, 1, nil).Once()
		userRepo.On("FindStoresByUserID", ctx, "u1").Return([]domain.UserStore{}, nil).Once()

		resp, total, err := svc.ListUsers(ctx, "search", false, 1, 20)
		assert.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, resp, 1)
		assert.Equal(t, "u1", resp[0].ID)
		userRepo.AssertExpectations(t)
	})

	t.Run("Repository Error", func(t *testing.T) {
		userRepo.On("ListAll", ctx, "", false, 1, 20).Return(nil, 0, errors.New("db error")).Once()
		resp, total, err := svc.ListUsers(ctx, "", false, 1, 20)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, 0, total)
	})
}

func TestUserAdminService_GetUser(t *testing.T) {
	userRepo := new(mocks.UserRepository)
	roleRepo := new(mocks.RoleRepository)
	log := zerolog.Nop()
	svc := NewUserAdminService(userRepo, roleRepo, 10, log)

	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		user := &domain.User{ID: "u1", Name: "User 1", Email: "u1@example.com", CreatedAt: time.Now(), UpdatedAt: time.Now()}
		userRepo.On("FindByID", ctx, "u1").Return(user, nil).Once()
		userRepo.On("FindStoresByUserID", ctx, "u1").Return([]domain.UserStore{{StoreName: "Store 1"}}, nil).Once()

		resp, err := svc.GetUser(ctx, "u1")
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "User 1", resp.Name)
		assert.Equal(t, 1, resp.StoreCount)
		userRepo.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		userRepo.On("FindByID", ctx, "u2").Return(nil, nil).Once()
		resp, err := svc.GetUser(ctx, "u2")
		assert.ErrorIs(t, err, ErrAdminUserNotFound)
		assert.Nil(t, resp)
	})
}

func TestUserAdminService_CreateUser(t *testing.T) {
	userRepo := new(mocks.UserRepository)
	roleRepo := new(mocks.RoleRepository)
	log := zerolog.Nop()
	svc := NewUserAdminService(userRepo, roleRepo, 10, log)

	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		req := &dto.CreateUserRequest{
			Name:     "New User",
			Email:    "new@example.com",
			Password: "password123",
			Stores: []dto.StoreAssignDTO{
				{StoreID: "s1", RoleID: "r1"},
			},
		}

		userRepo.On("ExistsByEmail", ctx, req.Email).Return(false, nil).Once()
		userRepo.On("Create", ctx, mock.MatchedBy(func(u *domain.User) bool {
			return u.Email == req.Email && u.Name == req.Name
		})).Return(&domain.User{ID: "u1", Name: req.Name, Email: req.Email, CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil).Once()
		userRepo.On("SetStores", ctx, "u1", mock.Anything).Return(nil).Once()
		userRepo.On("FindStoresByUserID", ctx, "u1").Return([]domain.UserStore{{StoreID: "s1", RoleID: "r1"}}, nil).Once()

		resp, err := svc.CreateUser(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "New User", resp.Name)
		userRepo.AssertExpectations(t)
	})

	t.Run("Email Conflict", func(t *testing.T) {
		req := &dto.CreateUserRequest{Email: "exists@example.com", Password: "pw"}
		userRepo.On("ExistsByEmail", ctx, req.Email).Return(true, nil).Once()

		resp, err := svc.CreateUser(ctx, req)
		assert.ErrorIs(t, err, ErrAdminEmailConflict)
		assert.Nil(t, resp)
	})
}

func TestUserAdminService_UpdateUser(t *testing.T) {
	userRepo := new(mocks.UserRepository)
	roleRepo := new(mocks.RoleRepository)
	log := zerolog.Nop()
	svc := NewUserAdminService(userRepo, roleRepo, 10, log)

	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		req := &dto.UpdateUserRequest{Name: "Updated Name", Email: "updated@example.com"}
		userRepo.On("FindByEmail", ctx, req.Email).Return(nil, nil).Once()
		userRepo.On("Update", ctx, "u1", req.Name, req.Email).Return(&domain.User{ID: "u1", Name: req.Name, Email: req.Email, CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil).Once()
		userRepo.On("FindStoresByUserID", ctx, "u1").Return([]domain.UserStore{}, nil).Once()

		resp, err := svc.UpdateUser(ctx, "u1", req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "Updated Name", resp.Name)
	})

	t.Run("Email Conflict", func(t *testing.T) {
		req := &dto.UpdateUserRequest{Email: "other@example.com"}
		userRepo.On("FindByEmail", ctx, req.Email).Return(&domain.User{ID: "u2", Email: req.Email}, nil).Once()

		resp, err := svc.UpdateUser(ctx, "u1", req)
		assert.ErrorIs(t, err, ErrAdminEmailConflict)
		assert.Nil(t, resp)
	})
}

func TestUserAdminService_DeactivateUser(t *testing.T) {
	userRepo := new(mocks.UserRepository)
	roleRepo := new(mocks.RoleRepository)
	log := zerolog.Nop()
	svc := NewUserAdminService(userRepo, roleRepo, 10, log)

	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		userRepo.On("SoftDelete", ctx, "u1").Return(nil).Once()
		err := svc.DeactivateUser(ctx, "u1")
		assert.NoError(t, err)
	})

	t.Run("Not Found", func(t *testing.T) {
		userRepo.On("SoftDelete", ctx, "u2").Return(ErrAdminUserNotFound).Once()
		err := svc.DeactivateUser(ctx, "u2")
		assert.ErrorIs(t, err, ErrAdminUserNotFound)
	})
}

func TestUserAdminService_ResetPassword(t *testing.T) {
	userRepo := new(mocks.UserRepository)
	roleRepo := new(mocks.RoleRepository)
	log := zerolog.Nop()
	svc := NewUserAdminService(userRepo, roleRepo, 10, log)

	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		req := &dto.ResetPasswordRequest{Password: "newpassword"}
		userRepo.On("ResetPassword", ctx, "u1", mock.Anything).Return(nil).Once()
		err := svc.ResetPassword(ctx, "u1", req)
		assert.NoError(t, err)
	})
}

func TestUserAdminService_SetUserStores(t *testing.T) {
	userRepo := new(mocks.UserRepository)
	roleRepo := new(mocks.RoleRepository)
	log := zerolog.Nop()
	svc := NewUserAdminService(userRepo, roleRepo, 10, log)

	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		req := &dto.SetUserStoresRequest{
			Stores: []dto.StoreAssignDTO{{StoreID: "s1", RoleID: "r1"}},
		}
		userRepo.On("FindByID", ctx, "u1").Return(&domain.User{ID: "u1", CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil).Once()
		userRepo.On("SetStores", ctx, "u1", mock.Anything).Return(nil).Once()
		userRepo.On("FindStoresByUserID", ctx, "u1").Return([]domain.UserStore{{StoreID: "s1"}}, nil).Once()

		resp, err := svc.SetUserStores(ctx, "u1", req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})
}

func TestUserAdminService_ListRoles(t *testing.T) {
	userRepo := new(mocks.UserRepository)
	roleRepo := new(mocks.RoleRepository)
	log := zerolog.Nop()
	svc := NewUserAdminService(userRepo, roleRepo, 10, log)

	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		roles := []*domain.Role{{ID: "r1", Name: "Admin"}}
		roleRepo.On("ListRoles", ctx).Return(roles, nil).Once()

		resp, err := svc.ListRoles(ctx)
		assert.NoError(t, err)
		assert.Len(t, resp, 1)
		assert.Equal(t, "Admin", resp[0].Name)
	})
}
