package testutil

import (
	"context"
	"time"
)

// MockAuthRepository is a test double for service.AuthRepository.
type MockAuthRepository struct {
	StoreRefreshTokenFn         func(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error
	ValidateRefreshTokenFn      func(ctx context.Context, tokenHash string) (int64, error)
	RevokeRefreshTokenFn        func(ctx context.Context, tokenHash string) error
	RevokeAllUserRefreshTokensFn func(ctx context.Context, userID int64) error
}

// StoreRefreshToken delegates to StoreRefreshTokenFn if set, otherwise returns nil.
func (m *MockAuthRepository) StoreRefreshToken(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error {
	if m.StoreRefreshTokenFn != nil {
		return m.StoreRefreshTokenFn(ctx, userID, tokenHash, expiresAt)
	}
	return nil
}

// ValidateRefreshToken delegates to ValidateRefreshTokenFn if set, otherwise returns 0 and nil.
func (m *MockAuthRepository) ValidateRefreshToken(ctx context.Context, tokenHash string) (int64, error) {
	if m.ValidateRefreshTokenFn != nil {
		return m.ValidateRefreshTokenFn(ctx, tokenHash)
	}
	return 0, nil
}

// RevokeRefreshToken delegates to RevokeRefreshTokenFn if set, otherwise returns nil.
func (m *MockAuthRepository) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	if m.RevokeRefreshTokenFn != nil {
		return m.RevokeRefreshTokenFn(ctx, tokenHash)
	}
	return nil
}

// RevokeAllUserRefreshTokens delegates to RevokeAllUserRefreshTokensFn if set, otherwise returns nil.
func (m *MockAuthRepository) RevokeAllUserRefreshTokens(ctx context.Context, userID int64) error {
	if m.RevokeAllUserRefreshTokensFn != nil {
		return m.RevokeAllUserRefreshTokensFn(ctx, userID)
	}
	return nil
}
