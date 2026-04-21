package service

import (
	"context"
	"errors"
	"time"

	"dermify-api/internal/domain"
)

// Sentinel errors for user operations.
var (
	ErrUserNotFound      = errors.New("user not found")      //nolint:gochecknoglobals // sentinel error
	ErrUserAlreadyExists = errors.New("user already exists") //nolint:gochecknoglobals // sentinel error
	ErrInvalidUserData   = errors.New("invalid user data")   //nolint:gochecknoglobals // sentinel error
)

// UserRepository defines the data access contract for users.
type UserRepository interface {
	// Create inserts a new user and sets the ID on the provided struct.
	Create(ctx context.Context, user *domain.User) error
	// GetByID retrieves a user by ID.
	GetByID(ctx context.Context, id int64) (*domain.User, error)
	// GetByEmail retrieves a user by email (case-insensitive).
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	// Update modifies a user record.
	Update(ctx context.Context, user *domain.User) error
	// Delete removes a user by ID.
	Delete(ctx context.Context, id int64) error
	// List returns all users.
	List(ctx context.Context) ([]*domain.User, error)
	// UpdatePreferences updates the user's language and timezone.
	UpdatePreferences(ctx context.Context, userID int64, language, timezone string) error
	// UpdatePassword updates user password hash and optionally clears must-change flag.
	UpdatePassword(ctx context.Context, userID int64, passwordHash string, clearMustChange bool) error
	// SetMustChangePassword toggles must-change-password flag for user.
	SetMustChangePassword(ctx context.Context, userID int64, mustChange bool) error
}

// UserService handles user business logic.
type UserService struct {
	repo UserRepository
}

// NewUserService creates a new UserService with the given repository.
func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

// Create validates and creates a new user record.
func (s *UserService) Create(ctx context.Context, user *domain.User) error {
	if err := validateUser(user); err != nil {
		return err
	}

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	return s.repo.Create(ctx, user)
}

// GetByID retrieves a user by ID.
func (s *UserService) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	return s.repo.GetByID(ctx, id)
}

// GetByEmail retrieves a user by email.
func (s *UserService) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return s.repo.GetByEmail(ctx, email)
}

// Update validates and updates a user record.
func (s *UserService) Update(ctx context.Context, user *domain.User) error {
	if user.Username == "" || user.Email == "" {
		return ErrInvalidUserData
	}

	user.UpdatedAt = time.Now()

	return s.repo.Update(ctx, user)
}

// Delete removes a user by ID.
func (s *UserService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

// List returns all users.
func (s *UserService) List(ctx context.Context) ([]*domain.User, error) {
	return s.repo.List(ctx)
}

// UpdatePreferences updates the user's language and timezone.
func (s *UserService) UpdatePreferences(ctx context.Context, userID int64, language, timezone string) error {
	return s.repo.UpdatePreferences(ctx, userID, language, timezone)
}

// validateUser checks required fields on a user record.
func validateUser(user *domain.User) error {
	if user.Username == "" {
		return ErrInvalidUserData
	}

	if user.Email == "" {
		return ErrInvalidUserData
	}

	if user.PasswordHash == "" {
		return ErrInvalidUserData
	}

	return nil
}
