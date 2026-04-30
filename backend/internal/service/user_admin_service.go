package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/dto"
	"github.com/moedahpos/backend/internal/repository"
)

var (
	ErrAdminUserNotFound  = errors.New("user not found")
	ErrAdminEmailConflict = errors.New("email already in use")
)

// UserAdminService handles system-wide user management operations.
type UserAdminService struct {
	userRepo repository.UserRepository
	roleRepo repository.RoleRepository
	bcrypt   int
	log      zerolog.Logger
}

func NewUserAdminService(
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	bcryptCost int,
	log zerolog.Logger,
) *UserAdminService {
	return &UserAdminService{userRepo: userRepo, roleRepo: roleRepo, bcrypt: bcryptCost, log: log}
}

// ListUsers returns a paginated, optionally filtered user list with store counts.
func (s *UserAdminService) ListUsers(ctx context.Context, search string, includeInactive bool, page, perPage int) ([]dto.UserResponse, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	users, total, err := s.userRepo.ListAll(ctx, search, includeInactive, page, perPage)
	if err != nil {
		return nil, 0, fmt.Errorf("ListUsers: %w", err)
	}

	resp := make([]dto.UserResponse, 0, len(users))
	for _, u := range users {
		stores, _ := s.userRepo.FindStoresByUserID(ctx, u.ID)
		r := toUserResponse(u, stores)
		resp = append(resp, r)
	}
	return resp, total, nil
}

// GetUser returns a single user with full store membership list.
func (s *UserAdminService) GetUser(ctx context.Context, id string) (*dto.UserResponse, error) {
	u, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("GetUser: %w", err)
	}
	if u == nil {
		return nil, ErrAdminUserNotFound
	}
	stores, _ := s.userRepo.FindStoresByUserID(ctx, u.ID)
	r := toUserResponse(u, stores)
	return &r, nil
}

// CreateUser creates a new user with hashed password and optional store assignments.
func (s *UserAdminService) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error) {
	exists, _ := s.userRepo.ExistsByEmail(ctx, req.Email)
	if exists {
		return nil, ErrAdminEmailConflict
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), s.bcrypt)
	if err != nil {
		return nil, fmt.Errorf("CreateUser hash: %w", err)
	}

	u, err := s.userRepo.Create(ctx, &domain.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: string(hash),
	})
	if err != nil {
		return nil, fmt.Errorf("CreateUser: %w", err)
	}

	if len(req.Stores) > 0 {
		assignments := make([]domain.StoreAssignment, 0, len(req.Stores))
		for _, s := range req.Stores {
			assignments = append(assignments, domain.StoreAssignment{StoreID: s.StoreID, RoleID: s.RoleID})
		}
		_ = s.userRepo.SetStores(ctx, u.ID, assignments)
	}

	stores, _ := s.userRepo.FindStoresByUserID(ctx, u.ID)
	r := toUserResponse(u, stores)
	return &r, nil
}

// UpdateUser changes name and/or email.
func (s *UserAdminService) UpdateUser(ctx context.Context, id string, req *dto.UpdateUserRequest) (*dto.UserResponse, error) {
	// Check email conflict (excluding self)
	existing, _ := s.userRepo.FindByEmail(ctx, req.Email)
	if existing != nil && existing.ID != id {
		return nil, ErrAdminEmailConflict
	}

	u, err := s.userRepo.Update(ctx, id, req.Name, req.Email)
	if err != nil {
		if errors.Is(err, ErrAdminUserNotFound) {
			return nil, ErrAdminUserNotFound
		}
		return nil, fmt.Errorf("UpdateUser: %w", err)
	}
	stores, _ := s.userRepo.FindStoresByUserID(ctx, u.ID)
	r := toUserResponse(u, stores)
	return &r, nil
}

// DeactivateUser soft-deletes (archives) a user — they cannot log in.
func (s *UserAdminService) DeactivateUser(ctx context.Context, id string) error {
	if err := s.userRepo.SoftDelete(ctx, id); err != nil {
		if errors.Is(err, ErrAdminUserNotFound) {
			return ErrAdminUserNotFound
		}
		return fmt.Errorf("DeactivateUser: %w", err)
	}
	return nil
}

// ResetPassword updates a user's password hash.
func (s *UserAdminService) ResetPassword(ctx context.Context, id string, req *dto.ResetPasswordRequest) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), s.bcrypt)
	if err != nil {
		return fmt.Errorf("ResetPassword hash: %w", err)
	}
	if err := s.userRepo.ResetPassword(ctx, id, string(hash)); err != nil {
		if errors.Is(err, ErrAdminUserNotFound) {
			return ErrAdminUserNotFound
		}
		return fmt.Errorf("ResetPassword: %w", err)
	}
	return nil
}

// SetUserStores atomically replaces all store memberships for a user.
func (s *UserAdminService) SetUserStores(ctx context.Context, userID string, req *dto.SetUserStoresRequest) (*dto.UserResponse, error) {
	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil || u == nil {
		return nil, ErrAdminUserNotFound
	}

	assignments := make([]domain.StoreAssignment, 0, len(req.Stores))
	for _, sa := range req.Stores {
		assignments = append(assignments, domain.StoreAssignment{StoreID: sa.StoreID, RoleID: sa.RoleID})
	}
	if err := s.userRepo.SetStores(ctx, userID, assignments); err != nil {
		return nil, fmt.Errorf("SetUserStores: %w", err)
	}

	stores, _ := s.userRepo.FindStoresByUserID(ctx, userID)
	r := toUserResponse(u, stores)
	return &r, nil
}

// ListRoles returns all roles with their permissions.
func (s *UserAdminService) ListRoles(ctx context.Context) ([]*domain.Role, error) {
	return s.roleRepo.ListRoles(ctx)
}

func (s *UserAdminService) CreateRole(ctx context.Context, req *dto.CreateRoleRequest) (*dto.RoleResponse, error) {
	role := &domain.Role{
		Name:        req.Name,
		Description: req.Description,
	}

	created, err := s.roleRepo.CreateRole(ctx, role, req.PermissionIDs)
	if err != nil {
		return nil, fmt.Errorf("CreateRole: %w", err)
	}

	// Fetch fresh permissions to return full name not just IDs
	roles, err := s.roleRepo.ListRoles(ctx)
	if err != nil {
		return nil, err
	}
	for _, r := range roles {
		if r.ID == created.ID {
			return &dto.RoleResponse{
				ID:          r.ID,
				Name:        r.Name,
				Description: r.Description,
				Permissions: r.Permissions,
			}, nil
		}
	}
	return nil, fmt.Errorf("role created but not found")
}

func (s *UserAdminService) UpdateRole(ctx context.Context, id string, req *dto.UpdateRoleRequest) (*dto.RoleResponse, error) {
	role := &domain.Role{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
	}

	updated, err := s.roleRepo.UpdateRole(ctx, role, req.PermissionIDs)
	if err != nil {
		return nil, fmt.Errorf("UpdateRole: %w", err)
	}

	roles, err := s.roleRepo.ListRoles(ctx)
	if err != nil {
		return nil, err
	}
	for _, r := range roles {
		if r.ID == updated.ID {
			return &dto.RoleResponse{
				ID:          r.ID,
				Name:        r.Name,
				Description: r.Description,
				Permissions: r.Permissions,
			}, nil
		}
	}
	return nil, fmt.Errorf("role updated but not found")
}

func (s *UserAdminService) DeleteRole(ctx context.Context, id string) error {
	return s.roleRepo.DeleteRole(ctx, id)
}

func (s *UserAdminService) ListPermissions(ctx context.Context) ([]dto.PermissionResponse, error) {
	perms, err := s.roleRepo.ListPermissions(ctx)
	if err != nil {
		return nil, fmt.Errorf("ListPermissions: %w", err)
	}
	resp := make([]dto.PermissionResponse, 0, len(perms))
	for _, p := range perms {
		resp = append(resp, dto.PermissionResponse{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
		})
	}
	return resp, nil
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func toUserResponse(u *domain.User, stores []domain.UserStore) dto.UserResponse {
	sr := make([]dto.UserStoreResponse, 0, len(stores))
	for _, s := range stores {
		sr = append(sr, dto.UserStoreResponse{
			StoreID:   s.StoreID,
			StoreName: s.StoreName,
			StoreType: s.StoreType,
			RoleID:    s.RoleID,
			RoleName:  s.RoleName,
			IsActive:  s.IsActive,
		})
	}

	var deletedAt *string
	if u.DeletedAt != nil {
		s := u.DeletedAt.Format("2006-01-02T15:04:05Z07:00")
		deletedAt = &s
	}

	return dto.UserResponse{
		ID:         u.ID,
		Name:       u.Name,
		Email:      u.Email,
		IsActive:   u.IsActive,
		CreatedAt:  u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  u.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		DeletedAt:  deletedAt,
		StoreCount: len(sr),
		Stores:     sr,
	}
}
