package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/zerolog"

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/dto"
	"github.com/moedahpos/backend/internal/repository"
	"github.com/moedahpos/backend/pkg/hash"
	"github.com/moedahpos/backend/pkg/jwt"
)

// Sentinel errors for the auth service.
var (
	ErrEmailTaken         = errors.New("email already taken")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserInactive       = errors.New("user account is inactive")
)

// AuthService implements authentication business logic.
type AuthService struct {
	userRepo    repository.UserRepository
	tokenRepo   repository.RefreshTokenRepository
	activitySvc ActivityLogServiceInterface
	jwtMgr      *jwt.Manager
	bcryptCost  int
	log         zerolog.Logger
}

// NewAuthService creates a new AuthService.
func NewAuthService(
	userRepo repository.UserRepository,
	tokenRepo repository.RefreshTokenRepository,
	activitySvc ActivityLogServiceInterface,
	jwtMgr *jwt.Manager,
	bcryptCost int,
	log zerolog.Logger,
) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		tokenRepo:   tokenRepo,
		activitySvc: activitySvc,
		jwtMgr:      jwtMgr,
		bcryptCost:  bcryptCost,
		log:         log,
	}
}

// Register creates a new user account.
func (s *AuthService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.RegisterResponse, error) {
	// Check uniqueness
	taken, err := s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("checking email: %w", err)
	}
	if taken {
		return nil, ErrEmailTaken
	}

	// Hash password
	hashed, err := hash.Password(req.Password, s.bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	// Persist user
	user, err := s.userRepo.Create(ctx, &domain.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hashed,
	})
	if err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}

	s.log.Info().Str("user_id", user.ID).Str("email", user.Email).Msg("user registered")

	return &dto.RegisterResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

// Login authenticates a user and issues JWT + refresh token.
func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("finding user: %w", err)
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}
	if !user.IsActive {
		return nil, ErrUserInactive
	}

	// Verify password
	if err := hash.CheckPassword(req.Password, user.PasswordHash); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Generate tokens
	accessToken, err := s.jwtMgr.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("generating access token: %w", err)
	}

	rawRefresh, expiresAt, err := s.jwtMgr.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("generating refresh token: %w", err)
	}

	// Store hashed refresh token
	if err := s.tokenRepo.Create(ctx, &domain.RefreshToken{
		UserID:    user.ID,
		TokenHash: hash.SHA256(rawRefresh),
		ExpiresAt: expiresAt,
	}); err != nil {
		return nil, fmt.Errorf("storing refresh token: %w", err)
	}

	s.log.Info().Str("user_id", user.ID).Msg("user logged in")

	s.activitySvc.LogActivity(ctx, user.ID, "", domain.ActionAuthLogin, domain.ModuleAuth, "", map[string]interface{}{
		"email": user.Email,
	})

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		TokenType:    "Bearer",
		ExpiresIn:    s.jwtMgr.AccessTTLSeconds(),
		User: dto.UserInfo{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		},
	}, nil
}

// Refresh issues a new access token given a valid refresh token.
func (s *AuthService) Refresh(ctx context.Context, req *dto.RefreshRequest) (*dto.RefreshResponse, error) {
	tokenHash := hash.SHA256(req.RefreshToken)

	stored, err := s.tokenRepo.FindByHash(ctx, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("finding refresh token: %w", err)
	}
	if stored == nil {
		return nil, ErrInvalidToken
	}

	// Load user
	user, err := s.userRepo.FindByID(ctx, stored.UserID)
	if err != nil {
		return nil, fmt.Errorf("finding user: %w", err)
	}
	if user == nil || !user.IsActive {
		return nil, ErrInvalidToken
	}

	// Rotate: revoke old token
	if err := s.tokenRepo.RevokeByID(ctx, stored.ID); err != nil {
		return nil, fmt.Errorf("revoking old token: %w", err)
	}

	// Issue new access token
	accessToken, err := s.jwtMgr.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("generating access token: %w", err)
	}

	return &dto.RefreshResponse{
		AccessToken: accessToken,
		ExpiresIn:   s.jwtMgr.AccessTTLSeconds(),
	}, nil
}

// Logout revokes all refresh tokens for the authenticated user.
func (s *AuthService) Logout(ctx context.Context, userID string) error {
	if err := s.tokenRepo.RevokeAllByUserID(ctx, userID); err != nil {
		return fmt.Errorf("revoking tokens: %w", err)
	}
	s.log.Info().Str("user_id", userID).Msg("user logged out")

	s.activitySvc.LogActivity(ctx, userID, "", domain.ActionAuthLogout, domain.ModuleAuth, "", nil)

	return nil
}

// Me returns the authenticated user's profile including store memberships.
func (s *AuthService) Me(ctx context.Context, userID string) (*dto.MeResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("finding user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	stores, err := s.userRepo.FindStoresByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("finding user stores: %w", err)
	}

	storeInfos := make([]dto.StoreRoleInfo, 0, len(stores))
	for _, us := range stores {
		storeInfos = append(storeInfos, dto.StoreRoleInfo{
			StoreID:                us.StoreID,
			StoreName:              us.StoreName,
			StoreType:              us.StoreType,
			Role:                   us.RoleName,
			LoyaltyPointsPerRupiah: us.LoyaltyPointsPerRupiah,
		})
	}

	return &dto.MeResponse{
		ID:     user.ID,
		Name:   user.Name,
		Email:  user.Email,
		Stores: storeInfos,
	}, nil
}
