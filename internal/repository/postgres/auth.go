package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// PostgresAuthRepository implements service.AuthRepository using PostgreSQL.
type PostgresAuthRepository struct {
	db *sql.DB
}

// NewPostgresAuthRepository creates a new PostgresAuthRepository.
func NewPostgresAuthRepository(db *sql.DB) *PostgresAuthRepository {
	return &PostgresAuthRepository{db: db}
}

// StoreRefreshToken inserts a hashed refresh token into the database.
func (r *PostgresAuthRepository) StoreRefreshToken(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`,
		userID, tokenHash, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("storing refresh token: %w", err)
	}

	return nil
}

// ValidateRefreshToken checks if a refresh token hash exists, is not expired, and is not revoked.
// Returns the user_id if valid.
func (r *PostgresAuthRepository) ValidateRefreshToken(ctx context.Context, tokenHash string) (int64, error) {
	var userID int64

	err := r.db.QueryRowContext(ctx,
		`SELECT user_id FROM refresh_tokens
		 WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > NOW()`,
		tokenHash,
	).Scan(&userID)
	if err != nil {
		return 0, fmt.Errorf("validating refresh token: %w", err)
	}

	return userID, nil
}

// RevokeRefreshToken marks a refresh token as revoked.
func (r *PostgresAuthRepository) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE refresh_tokens SET revoked_at = NOW() WHERE token_hash = $1`,
		tokenHash,
	)
	if err != nil {
		return fmt.Errorf("revoking refresh token: %w", err)
	}

	return nil
}

// RevokeAllUserRefreshTokens revokes all refresh tokens for a given user.
func (r *PostgresAuthRepository) RevokeAllUserRefreshTokens(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE refresh_tokens SET revoked_at = NOW() WHERE user_id = $1 AND revoked_at IS NULL`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("revoking all user refresh tokens: %w", err)
	}

	return nil
}
