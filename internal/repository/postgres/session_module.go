package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"dermify-api/internal/domain"
)

// PostgresModuleRepository implements service.ModuleRepository using PostgreSQL.
type PostgresModuleRepository struct {
	db *sql.DB
}

// NewPostgresModuleRepository creates a new PostgresModuleRepository.
func NewPostgresModuleRepository(db *sql.DB) *PostgresModuleRepository {
	return &PostgresModuleRepository{db: db}
}

// Create inserts a new session module and sets the ID on the provided struct.
func (r *PostgresModuleRepository) Create(ctx context.Context, module *domain.SessionModule, userID int64) error {
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO session_modules (session_id, module_type, sort_order, version,
			created_at, created_by, updated_at, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`,
		module.SessionID, module.ModuleType, module.SortOrder, module.Version,
		module.CreatedAt, userID, module.UpdatedAt, userID,
	).Scan(&module.ID)
	if err != nil {
		return fmt.Errorf("inserting session module: %w", err)
	}

	return nil
}

// ListBySession returns all modules for a session ordered by sort_order.
func (r *PostgresModuleRepository) ListBySession(ctx context.Context, sessionID int64) ([]domain.SessionModule, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, session_id, module_type, sort_order, version,
			created_at, created_by, updated_at, updated_by
		FROM session_modules WHERE session_id = $1
		ORDER BY sort_order ASC`, sessionID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying session modules: %w", err)
	}
	defer rows.Close()

	modules := []domain.SessionModule{}

	for rows.Next() {
		var m domain.SessionModule
		err := rows.Scan(
			&m.ID, &m.SessionID, &m.ModuleType, &m.SortOrder, &m.Version,
			&m.CreatedAt, &m.CreatedBy, &m.UpdatedAt, &m.UpdatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning session module row: %w", err)
		}

		modules = append(modules, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating session module rows: %w", err)
	}

	return modules, nil
}

// Delete removes a module from a session. Returns an error if no matching row is found.
func (r *PostgresModuleRepository) Delete(ctx context.Context, id int64, sessionID int64) error {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM session_modules WHERE id = $1 AND session_id = $2`,
		id, sessionID,
	)
	if err != nil {
		return fmt.Errorf("deleting session module: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session module not found: id=%d session_id=%d", id, sessionID)
	}

	return nil
}

// NextSortOrder returns the next available sort order for a session.
func (r *PostgresModuleRepository) NextSortOrder(ctx context.Context, sessionID int64) (int, error) {
	var next int

	err := r.db.QueryRowContext(ctx,
		`SELECT COALESCE(MAX(sort_order), 0) + 1 FROM session_modules WHERE session_id = $1`,
		sessionID,
	).Scan(&next)
	if err != nil {
		return 0, fmt.Errorf("querying next sort order: %w", err)
	}

	return next, nil
}
