package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"dermify-api/internal/domain"
	"dermify-api/internal/service"
)

// PostgresAuditRepository implements service.AuditRepository using PostgreSQL.
type PostgresAuditRepository struct {
	db *sql.DB
}

// NewPostgresAuditRepository creates a new PostgresAuditRepository.
func NewPostgresAuditRepository(db *sql.DB) *PostgresAuditRepository {
	return &PostgresAuditRepository{db: db}
}

// ListByEntity returns all audit entries for a specific entity type and ID,
// ordered by performed_at DESC.
func (r *PostgresAuditRepository) ListByEntity(ctx context.Context, entityType string, entityID int64) ([]domain.AuditEntry, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, action, performed_at, user_id, entity_type, entity_id, old_values, new_values
		FROM audit_trail
		WHERE entity_type = $1 AND entity_id = $2
		ORDER BY performed_at DESC`, entityType, entityID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying audit entries by entity: %w", err)
	}
	defer rows.Close()

	return scanAuditEntries(rows)
}

// List returns paginated audit entries matching the given filter.
// It supports filtering by entity_type, entity_id, and user_id.
func (r *PostgresAuditRepository) List(ctx context.Context, filter service.AuditFilter) (*service.AuditListResult, error) {
	var conditions []string
	args := []interface{}{}
	argIndex := 1

	if filter.EntityType != "" {
		conditions = append(conditions, fmt.Sprintf("entity_type = $%d", argIndex))
		args = append(args, filter.EntityType)
		argIndex++
	}

	if filter.EntityID > 0 {
		conditions = append(conditions, fmt.Sprintf("entity_id = $%d", argIndex))
		args = append(args, filter.EntityID)
		argIndex++
	}

	if filter.UserID > 0 {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIndex))
		args = append(args, filter.UserID)
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total matching records.
	var total int

	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM audit_trail"+whereClause, args...,
	).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("counting audit entries: %w", err)
	}

	// Apply ordering and pagination.
	offset := (filter.Page - 1) * filter.PerPage
	limitClause := fmt.Sprintf(" ORDER BY performed_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, filter.PerPage, offset)

	rows, err := r.db.QueryContext(ctx,
		`SELECT id, action, performed_at, user_id, entity_type, entity_id, old_values, new_values
		FROM audit_trail`+whereClause+limitClause, args...,
	)
	if err != nil {
		return nil, fmt.Errorf("querying audit entries: %w", err)
	}
	defer rows.Close()

	entries, err := scanAuditEntries(rows)
	if err != nil {
		return nil, err
	}

	return &service.AuditListResult{
		Entries: entries,
		Total:   total,
	}, nil
}

// scanAuditEntries scans rows into an AuditEntry slice, handling nullable
// user_id and JSONB old_values/new_values columns.
func scanAuditEntries(rows *sql.Rows) ([]domain.AuditEntry, error) {
	entries := []domain.AuditEntry{}

	for rows.Next() {
		var e domain.AuditEntry

		// Use nullable byte slices for JSONB that may be NULL.
		var oldValues *[]byte
		var newValues *[]byte

		if err := rows.Scan(
			&e.ID, &e.Action, &e.PerformedAt, &e.UserID,
			&e.EntityType, &e.EntityID, &oldValues, &newValues,
		); err != nil {
			return nil, fmt.Errorf("scanning audit entry: %w", err)
		}

		if oldValues != nil {
			e.OldValues = *oldValues
		}

		if newValues != nil {
			e.NewValues = *newValues
		}

		entries = append(entries, e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating audit entries: %w", err)
	}

	return entries, nil
}
