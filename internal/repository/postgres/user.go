package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"dermify-api/internal/domain"
	"dermify-api/internal/service"
)

// PostgresUserRepository implements service.UserRepository using PostgreSQL.
type PostgresUserRepository struct {
	db *sql.DB
}

// NewPostgresUserRepository creates a new PostgresUserRepository.
func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

// Create inserts a new user record and sets the ID on the provided struct.
func (r *PostgresUserRepository) Create(ctx context.Context, user *domain.User) error {
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO users (username, email, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		user.Username, user.Email, user.PasswordHash, user.CreatedAt, user.UpdatedAt,
	).Scan(&user.ID)
	if err != nil {
		if isUniqueViolation(err) {
			return service.ErrUserAlreadyExists
		}

		return fmt.Errorf("inserting user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by ID.
func (r *PostgresUserRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	var u domain.User
	var bio sql.NullString

	err := r.db.QueryRowContext(ctx,
		`SELECT id, username, email, password_hash, bio, COALESCE(role, ''), language, timezone, must_change_password, created_at, updated_at
		FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &bio, &u.Role, &u.Language, &u.Timezone, &u.MustChangePassword, &u.CreatedAt, &u.UpdatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrUserNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("querying user: %w", err)
	}

	if bio.Valid {
		u.Bio = &bio.String
	}

	return &u, nil
}

// GetByEmail retrieves a user by email (case-insensitive).
func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User
	var bio sql.NullString

	err := r.db.QueryRowContext(ctx,
		`SELECT id, username, email, password_hash, bio, COALESCE(role, ''), language, timezone, must_change_password, created_at, updated_at
		FROM users WHERE LOWER(email) = LOWER($1)`, email,
	).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &bio, &u.Role, &u.Language, &u.Timezone, &u.MustChangePassword, &u.CreatedAt, &u.UpdatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrUserNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("querying user by email: %w", err)
	}

	if bio.Valid {
		u.Bio = &bio.String
	}

	return &u, nil
}

// Update modifies a user record.
func (r *PostgresUserRepository) Update(ctx context.Context, user *domain.User) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE users SET username = $1, email = $2, bio = $3, updated_at = $4
		WHERE id = $5`,
		user.Username, user.Email, user.Bio, user.UpdatedAt, user.ID,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return service.ErrUserAlreadyExists
		}

		return fmt.Errorf("updating user: %w", err)
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

// Delete removes a user by ID.
func (r *PostgresUserRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting user: %w", err)
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

// List returns all users ordered by ID.
func (r *PostgresUserRepository) List(ctx context.Context) ([]*domain.User, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, username, email, password_hash, bio, COALESCE(role, ''), language, timezone, must_change_password, created_at, updated_at
		FROM users ORDER BY id ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("querying users: %w", err)
	}
	defer rows.Close()

	var users []*domain.User

	for rows.Next() {
		var u domain.User
		var bio sql.NullString

		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &bio, &u.Role, &u.Language, &u.Timezone, &u.MustChangePassword, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning user row: %w", err)
		}

		if bio.Valid {
			u.Bio = &bio.String
		}

		users = append(users, &u)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating user rows: %w", err)
	}

	return users, nil
}

// UpdatePreferences updates the user's language and timezone preferences.
func (r *PostgresUserRepository) UpdatePreferences(ctx context.Context, userID int64, language, timezone string) error {
	if language != "" && timezone != "" {
		_, err := r.db.ExecContext(ctx,
			`UPDATE users SET language = $1, timezone = $2, updated_at = NOW() WHERE id = $3`,
			language, timezone, userID,
		)
		return err
	}

	if language != "" {
		_, err := r.db.ExecContext(ctx,
			`UPDATE users SET language = $1, updated_at = NOW() WHERE id = $2`,
			language, userID,
		)
		return err
	}

	if timezone != "" {
		_, err := r.db.ExecContext(ctx,
			`UPDATE users SET timezone = $1, updated_at = NOW() WHERE id = $2`,
			timezone, userID,
		)
		return err
	}

	return nil
}

// UpdatePassword updates password hash and optionally clears must-change-password.
func (r *PostgresUserRepository) UpdatePassword(ctx context.Context, userID int64, passwordHash string, clearMustChange bool) error {
	if clearMustChange {
		result, err := r.db.ExecContext(ctx,
			`UPDATE users SET password_hash = $1, must_change_password = false, updated_at = NOW() WHERE id = $2`,
			passwordHash, userID,
		)
		if err != nil {
			return err
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return service.ErrUserNotFound
		}
		return nil
	}

	result, err := r.db.ExecContext(ctx,
		`UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2`,
		passwordHash, userID,
	)
	if err != nil {
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return service.ErrUserNotFound
	}
	return nil
}

// SetMustChangePassword toggles must-change-password flag.
func (r *PostgresUserRepository) SetMustChangePassword(ctx context.Context, userID int64, mustChange bool) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE users SET must_change_password = $1, updated_at = NOW() WHERE id = $2`,
		mustChange, userID,
	)
	if err != nil {
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return service.ErrUserNotFound
	}
	return nil
}

// isUniqueViolation checks if the error is a PostgreSQL unique constraint violation.
func isUniqueViolation(err error) bool {
	return strings.Contains(err.Error(), "duplicate key value violates unique constraint")
}
