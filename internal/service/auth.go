package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"dermify-api/internal/domain"
)

// Sentinel errors for authentication operations.
var (
	ErrInvalidCredentials  = errors.New("invalid credentials")             //nolint:gochecknoglobals // sentinel error
	ErrRefreshTokenInvalid = errors.New("invalid or expired refresh token") //nolint:gochecknoglobals // sentinel error
)

// AuthRepository defines the data access contract for authentication tokens.
type AuthRepository interface {
	// StoreRefreshToken inserts a hashed refresh token.
	StoreRefreshToken(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error
	// ValidateRefreshToken checks if a token hash is valid and returns the user ID.
	ValidateRefreshToken(ctx context.Context, tokenHash string) (int64, error)
	// RevokeRefreshToken marks a refresh token as revoked.
	RevokeRefreshToken(ctx context.Context, tokenHash string) error
	// RevokeAllUserRefreshTokens revokes all refresh tokens for a user.
	RevokeAllUserRefreshTokens(ctx context.Context, userID int64) error
}

// AuthService handles authentication business logic.
type AuthService struct {
	authRepo AuthRepository
	userRepo UserRepository
	roleSvc  *RoleService
}

// NewAuthService creates a new AuthService with the given repositories.
func NewAuthService(authRepo AuthRepository, userRepo UserRepository, roleSvc *RoleService) *AuthService {
	return &AuthService{
		authRepo: authRepo,
		userRepo: userRepo,
		roleSvc:  roleSvc,
	}
}

// Register creates a new user and performs first-user bootstrap if applicable.
func (s *AuthService) Register(ctx context.Context, username, email, passwordHash string) (*domain.User, error) {
	user := &domain.User{
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
	}

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// First-user bootstrap: auto-promote to Admin if this is the first user.
	isFirst, err := s.roleSvc.IsFirstUser(ctx)
	if err != nil {
		slog.Error("failed to check first user status", "error", err)
	} else if isFirst {
		if assignErr := s.roleSvc.AssignRole(ctx, user.ID, domain.RoleAdmin); assignErr != nil {
			slog.Error("failed to auto-promote first user to admin", "error", assignErr)
		} else {
			slog.Info("first user auto-promoted to admin", "user_id", user.ID)
		}
	}

	return user, nil
}

// Authenticate looks up a user by email for credential verification.
// Password comparison should be done by the caller using the returned user's PasswordHash.
func (s *AuthService) Authenticate(ctx context.Context, email string) (*domain.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

// GetUserByID retrieves a user by ID.
func (s *AuthService) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

// StoreRefreshToken stores a hashed refresh token.
func (s *AuthService) StoreRefreshToken(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error {
	return s.authRepo.StoreRefreshToken(ctx, userID, tokenHash, expiresAt)
}

// ValidateRefreshToken checks if a refresh token hash is valid and returns the user ID.
func (s *AuthService) ValidateRefreshToken(ctx context.Context, tokenHash string) (int64, error) {
	userID, err := s.authRepo.ValidateRefreshToken(ctx, tokenHash)
	if err != nil {
		return 0, ErrRefreshTokenInvalid
	}

	return userID, nil
}

// RevokeRefreshToken marks a refresh token as revoked.
func (s *AuthService) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	return s.authRepo.RevokeRefreshToken(ctx, tokenHash)
}

// RevokeAllUserRefreshTokens revokes all refresh tokens for a user.
func (s *AuthService) RevokeAllUserRefreshTokens(ctx context.Context, userID int64) error {
	return s.authRepo.RevokeAllUserRefreshTokens(ctx, userID)
}
