package service

import (
	"context"
	"errors"

	"dermify-api/internal/domain"
)

// ErrInvalidRole is returned when an unrecognized role is provided.
var ErrInvalidRole = errors.New("invalid role") //nolint:gochecknoglobals // sentinel error

// RoleRepository defines the data access contract for role operations.
type RoleRepository interface {
	// UpdateUserRole sets the role for a given user.
	UpdateUserRole(ctx context.Context, userID int64, role string) error
	// GetUserRole returns the role for a given user, or empty string if unset.
	GetUserRole(ctx context.Context, userID int64) (string, error)
	// CountUsers returns the total number of registered users.
	CountUsers(ctx context.Context) (int64, error)
}

// RoleService handles role-related business logic.
type RoleService struct {
	repo RoleRepository
}

// NewRoleService creates a new RoleService with the given repository.
func NewRoleService(repo RoleRepository) *RoleService {
	return &RoleService{repo: repo}
}

// AssignRole validates the role and delegates to the repository.
func (s *RoleService) AssignRole(ctx context.Context, userID int64, role string) error {
	if !domain.ValidRole(role) {
		return ErrInvalidRole
	}
	return s.repo.UpdateUserRole(ctx, userID, role)
}

// GetUserRole retrieves the current role for a user.
func (s *RoleService) GetUserRole(ctx context.Context, userID int64) (string, error) {
	return s.repo.GetUserRole(ctx, userID)
}

// IsFirstUser checks if the system has at most one registered user.
func (s *RoleService) IsFirstUser(ctx context.Context) (bool, error) {
	count, err := s.repo.CountUsers(ctx)
	if err != nil {
		return false, err
	}
	return count <= 1, nil
}
