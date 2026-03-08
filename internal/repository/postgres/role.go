package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"dermify-api/internal/service"
)

// PostgresRoleRepository implements service.RoleRepository using PostgreSQL.
type PostgresRoleRepository struct {
	db *sql.DB
}

// NewPostgresRoleRepository creates a new PostgresRoleRepository.
func NewPostgresRoleRepository(db *sql.DB) *PostgresRoleRepository {
	return &PostgresRoleRepository{db: db}
}

// UpdateUserRole sets the role for the given user. Returns an error if the user does not exist.
func (r *PostgresRoleRepository) UpdateUserRole(ctx context.Context, userID int64, role string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE users SET role = $1 WHERE id = $2",
		role, userID,
	)
	if err != nil {
		return fmt.Errorf("updating user role: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return service.ErrUserNotFound
	}

	return nil
}

// GetUserRole returns the role for the given user. Returns empty string if the role is NULL.
func (r *PostgresRoleRepository) GetUserRole(ctx context.Context, userID int64) (string, error) {
	var role string
	err := r.db.QueryRowContext(ctx,
		"SELECT COALESCE(role, '') FROM users WHERE id = $1",
		userID,
	).Scan(&role)

	if errors.Is(err, sql.ErrNoRows) {
		return "", service.ErrUserNotFound
	}

	if err != nil {
		return "", fmt.Errorf("querying user role: %w", err)
	}

	return role, nil
}

// CountUsers returns the total number of registered users.
func (r *PostgresRoleRepository) CountUsers(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("counting users: %w", err)
	}
	return count, nil
}
