package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"dermify-api/internal/domain"
	"dermify-api/internal/service"
)

// PostgresAddendumRepository implements service.AddendumRepository using PostgreSQL.
type PostgresAddendumRepository struct {
	db *sql.DB
}

// NewPostgresAddendumRepository creates a new PostgresAddendumRepository.
func NewPostgresAddendumRepository(db *sql.DB) *PostgresAddendumRepository {
	return &PostgresAddendumRepository{db: db}
}

// Create inserts a new addendum record and sets the generated ID on the provided struct.
func (r *PostgresAddendumRepository) Create(ctx context.Context, addendum *domain.Addendum) error {
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO session_addendums (session_id, author_id, reason, content, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`,
		addendum.SessionID, addendum.AuthorID, addendum.Reason, addendum.Content, addendum.CreatedAt,
	).Scan(&addendum.ID)
	if err != nil {
		return fmt.Errorf("inserting addendum: %w", err)
	}

	return nil
}

// GetByID retrieves an addendum by its ID.
func (r *PostgresAddendumRepository) GetByID(ctx context.Context, id int64) (*domain.Addendum, error) {
	var a domain.Addendum

	err := r.db.QueryRowContext(ctx,
		`SELECT id, session_id, author_id, reason, content, created_at
		FROM session_addendums WHERE id = $1`, id,
	).Scan(
		&a.ID, &a.SessionID, &a.AuthorID, &a.Reason, &a.Content, &a.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrAddendumNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("querying addendum: %w", err)
	}

	return &a, nil
}

// ListBySession returns all addendums for a session ordered by created_at DESC.
func (r *PostgresAddendumRepository) ListBySession(ctx context.Context, sessionID int64) ([]domain.Addendum, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, session_id, author_id, reason, content, created_at
		FROM session_addendums WHERE session_id = $1
		ORDER BY created_at DESC`, sessionID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying addendums by session: %w", err)
	}
	defer rows.Close()

	addendums := []domain.Addendum{}

	for rows.Next() {
		var a domain.Addendum
		if err := rows.Scan(
			&a.ID, &a.SessionID, &a.AuthorID, &a.Reason, &a.Content, &a.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning addendum: %w", err)
		}

		addendums = append(addendums, a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating addendums: %w", err)
	}

	return addendums, nil
}
