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
	repomocks "github.com/moedahpos/backend/internal/repository/mocks"
	"github.com/moedahpos/backend/internal/service/mocks"
	"github.com/moedahpos/backend/pkg/hash"
	"github.com/moedahpos/backend/pkg/jwt"
)

func TestAuthService_Register(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	log := zerolog.Nop()
	bcryptCost := 10

	tests := []struct {
		name    string
		req     *dto.RegisterRequest
		setup   func(u *repomocks.UserRepository)
		wantErr error
	}{
		{
			name: "Success",
			req: &dto.RegisterRequest{
				Name:     "Test User",
				Email:    "test@example.com",
				Password: "password123",
			},
			setup: func(u *repomocks.UserRepository) {
				u.On("ExistsByEmail", ctx, "test@example.com").Return(false, nil)
				u.On("Create", ctx, mock.MatchedBy(func(user *domain.User) bool {
					return user.Name == "Test User" && user.Email == "test@example.com"
				})).Return(&domain.User{
					ID:        "user-123",
					Name:      "Test User",
					Email:     "test@example.com",
					CreatedAt: time.Now(),
				}, nil)
			},
			wantErr: nil,
		},
		{
			name: "Email Taken",
			req: &dto.RegisterRequest{
				Email: "taken@example.com",
			},
			setup: func(u *repomocks.UserRepository) {
				u.On("ExistsByEmail", ctx, "taken@example.com").Return(true, nil)
			},
			wantErr: ErrEmailTaken,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uRepo := new(repomocks.UserRepository)
			tRepo := new(repomocks.RefreshTokenRepository)
			aSvc := new(mocks.ActivityLogServiceInterface)

			if tt.setup != nil {
				tt.setup(uRepo)
			}

			s := NewAuthService(uRepo, tRepo, aSvc, nil, bcryptCost, log)
			resp, err := s.Register(ctx, tt.req)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.req.Name, resp.Name)
			}

			uRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_Register_Errors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	log := zerolog.Nop()
	bcryptCost := 10

	tests := []struct {
		name    string
		req     *dto.RegisterRequest
		setup   func(u *repomocks.UserRepository)
		wantErr error
	}{
		{
			name: "Repository Error on Exists",
			req: &dto.RegisterRequest{
				Email: "err@example.com",
			},
			setup: func(u *repomocks.UserRepository) {
				u.On("ExistsByEmail", ctx, "err@example.com").Return(false, errors.New("db error"))
			},
			wantErr: errors.New("checking email: db error"),
		},
		{
			name: "Repository Error on Create",
			req: &dto.RegisterRequest{
				Name:     "Fail",
				Email:    "fail@example.com",
				Password: "pass",
			},
			setup: func(u *repomocks.UserRepository) {
				u.On("ExistsByEmail", ctx, "fail@example.com").Return(false, nil)
				u.On("Create", ctx, mock.Anything).Return(nil, errors.New("create error"))
			},
			wantErr: errors.New("creating user: create error"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uRepo := new(repomocks.UserRepository)
			tRepo := new(repomocks.RefreshTokenRepository)
			aSvc := new(mocks.ActivityLogServiceInterface)

			if tt.setup != nil {
				tt.setup(uRepo)
			}

			s := NewAuthService(uRepo, tRepo, aSvc, nil, bcryptCost, log)
			resp, err := s.Register(ctx, tt.req)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.req.Name, resp.Name)
			}

			uRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	ctx := context.Background()
	log := zerolog.Nop()
	jwtMgr := jwt.New("secret", time.Hour, time.Hour*24)
	bcryptCost := 10

	// Pre-hash password for success case
	pass := "password123"
	// We'll use a real hashing for the setup to ensure logic works,
	// but we can also mock the hashing if it's external.
	// In this codebase, hash.Password is used.

	tests := []struct {
		name    string
		req     *dto.LoginRequest
		setup   func(u *repomocks.UserRepository, t *repomocks.RefreshTokenRepository, a *mocks.ActivityLogServiceInterface)
		wantErr error
	}{
		{
			name: "Success",
			req: &dto.LoginRequest{
				Email:    "login@example.com",
				Password: pass,
			},
			setup: func(u *repomocks.UserRepository, rtRepo *repomocks.RefreshTokenRepository, a *mocks.ActivityLogServiceInterface) {
				hashed, _ := hash.Password(pass, bcryptCost)
				u.On("FindByEmail", ctx, "login@example.com").Return(&domain.User{
					ID:           "u1",
					Email:        "login@example.com",
					PasswordHash: hashed,
					IsActive:     true,
				}, nil)
				rtRepo.On("Create", ctx, mock.Anything).Return(nil)
				a.On("LogActivity", ctx, "u1", "", domain.ActionAuthLogin, domain.ModuleAuth, "", mock.Anything).Return()
			},
			wantErr: nil,
		},
		{
			name: "User Not Found",
			req:  &dto.LoginRequest{Email: "none@example.com"},
			setup: func(u *repomocks.UserRepository, _ *repomocks.RefreshTokenRepository, _ *mocks.ActivityLogServiceInterface) {
				u.On("FindByEmail", ctx, "none@example.com").Return(nil, nil)
			},
			wantErr: ErrInvalidCredentials,
		},
		{
			name: "User Inactive",
			req:  &dto.LoginRequest{Email: "inactive@example.com"},
			setup: func(u *repomocks.UserRepository, _ *repomocks.RefreshTokenRepository, _ *mocks.ActivityLogServiceInterface) {
				u.On("FindByEmail", ctx, "inactive@example.com").Return(&domain.User{IsActive: false}, nil)
			},
			wantErr: ErrUserInactive,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uRepo := new(repomocks.UserRepository)
			tRepo := new(repomocks.RefreshTokenRepository)
			aSvc := new(mocks.ActivityLogServiceInterface)

			if tt.setup != nil {
				tt.setup(uRepo, tRepo, aSvc)
			}

			s := NewAuthService(uRepo, tRepo, aSvc, jwtMgr, bcryptCost, log)
			resp, err := s.Login(ctx, tt.req)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotEmpty(t, resp.AccessToken)
			}
		})
	}
}

func TestAuthService_Login_Errors(t *testing.T) {
	ctx := context.Background()
	log := zerolog.Nop()
	jwtMgr := jwt.New("secret", time.Hour, time.Hour*24)
	bcryptCost := 10

	tests := []struct {
		name    string
		req     *dto.LoginRequest
		setup   func(u *repomocks.UserRepository, t *repomocks.RefreshTokenRepository, a *mocks.ActivityLogServiceInterface)
		wantErr error
	}{
		{
			name: "Invalid Password",
			req:  &dto.LoginRequest{Email: "u@example.com", Password: "wrong"},
			setup: func(u *repomocks.UserRepository, _ *repomocks.RefreshTokenRepository, _ *mocks.ActivityLogServiceInterface) {
				hashed, _ := hash.Password("right", bcryptCost)
				u.On("FindByEmail", ctx, "u@example.com").Return(&domain.User{
					PasswordHash: hashed,
					IsActive:     true,
				}, nil)
			},
			wantErr: ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uRepo := new(repomocks.UserRepository)
			tRepo := new(repomocks.RefreshTokenRepository)
			aSvc := new(mocks.ActivityLogServiceInterface)

			if tt.setup != nil {
				tt.setup(uRepo, tRepo, aSvc)
			}

			s := NewAuthService(uRepo, tRepo, aSvc, jwtMgr, bcryptCost, log)
			resp, err := s.Login(ctx, tt.req)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotEmpty(t, resp.AccessToken)
			}
		})
	}
}

func TestAuthService_Refresh(t *testing.T) {
	ctx := context.Background()
	log := zerolog.Nop()
	jwtMgr := jwt.New("secret", time.Hour, time.Hour*24)

	t.Run("Success", func(t *testing.T) {
		uRepo := new(repomocks.UserRepository)
		tRepo := new(repomocks.RefreshTokenRepository)
		aSvc := new(mocks.ActivityLogServiceInterface)

		token := "refresh-token"
		tokenHash := hash.SHA256(token)

		tRepo.On("FindByHash", ctx, tokenHash).Return(&domain.RefreshToken{
			ID:     "t1",
			UserID: "u1",
		}, nil)
		uRepo.On("FindByID", ctx, "u1").Return(&domain.User{
			ID:       "u1",
			Email:    "test@example.com",
			IsActive: true,
		}, nil)
		tRepo.On("RevokeByID", ctx, "t1").Return(nil)

		s := NewAuthService(uRepo, tRepo, aSvc, jwtMgr, 10, log)
		resp, err := s.Refresh(ctx, &dto.RefreshRequest{RefreshToken: token})

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.AccessToken)
	})

	t.Run("Invalid Token", func(t *testing.T) {
		uRepo := new(repomocks.UserRepository)
		tRepo := new(repomocks.RefreshTokenRepository)
		aSvc := new(mocks.ActivityLogServiceInterface)

		tRepo.On("FindByHash", ctx, mock.Anything).Return(nil, nil)

		s := NewAuthService(uRepo, tRepo, aSvc, jwtMgr, 10, log)
		resp, err := s.Refresh(ctx, &dto.RefreshRequest{RefreshToken: "bad"})

		assert.ErrorIs(t, err, ErrInvalidToken)
		assert.Nil(t, resp)
	})
}

func TestAuthService_Logout(t *testing.T) {
	ctx := context.Background()
	log := zerolog.Nop()

	t.Run("Success", func(t *testing.T) {
		uRepo := new(repomocks.UserRepository)
		tRepo := new(repomocks.RefreshTokenRepository)
		aSvc := new(mocks.ActivityLogServiceInterface)

		tRepo.On("RevokeAllByUserID", ctx, "u1").Return(nil)
		aSvc.On("LogActivity", ctx, "u1", "", domain.ActionAuthLogout, domain.ModuleAuth, "", nil).Return()

		s := NewAuthService(uRepo, tRepo, aSvc, nil, 10, log)
		err := s.Logout(ctx, "u1")

		assert.NoError(t, err)
	})
}

func TestAuthService_Me(t *testing.T) {
	ctx := context.Background()
	log := zerolog.Nop()

	t.Run("Success", func(t *testing.T) {
		uRepo := new(repomocks.UserRepository)
		tRepo := new(repomocks.RefreshTokenRepository)
		aSvc := new(mocks.ActivityLogServiceInterface)

		uRepo.On("FindByID", ctx, "u1").Return(&domain.User{
			ID:    "u1",
			Email: "me@example.com",
			Name:  "Me",
		}, nil)
		uRepo.On("FindStoresByUserID", ctx, "u1").Return([]domain.UserStore{
			{StoreID: "s1", StoreName: "Store 1", RoleName: "Admin"},
		}, nil)

		s := NewAuthService(uRepo, tRepo, aSvc, nil, 10, log)
		resp, err := s.Me(ctx, "u1")

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "me@example.com", resp.Email)
		assert.Len(t, resp.Stores, 1)
	})

	t.Run("User Not Found", func(t *testing.T) {
		uRepo := new(repomocks.UserRepository)
		tRepo := new(repomocks.RefreshTokenRepository)
		aSvc := new(mocks.ActivityLogServiceInterface)

		uRepo.On("FindByID", ctx, "u1").Return(nil, nil)

		s := NewAuthService(uRepo, tRepo, aSvc, nil, 10, log)
		resp, err := s.Me(ctx, "u1")

		assert.ErrorIs(t, err, ErrUserNotFound)
		assert.Nil(t, resp)
	})
}
