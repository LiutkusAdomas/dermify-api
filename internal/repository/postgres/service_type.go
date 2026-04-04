package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"dermify-api/internal/domain"
	"dermify-api/internal/service"
)

// PostgresServiceTypeRepository implements service.ServiceTypeRepository.
type PostgresServiceTypeRepository struct {
	db *sql.DB
}

// NewPostgresServiceTypeRepository creates a new PostgresServiceTypeRepository.
func NewPostgresServiceTypeRepository(db *sql.DB) *PostgresServiceTypeRepository {
	return &PostgresServiceTypeRepository{db: db}
}

// Create inserts a new service type.
func (r *PostgresServiceTypeRepository) Create(ctx context.Context, st *domain.ServiceType) error {
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO service_types (org_id, name, default_duration_minutes, description, active)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, created_at, updated_at`,
		st.OrgID, st.Name, st.DefaultDuration, st.Description, st.Active,
	).Scan(&st.ID, &st.CreatedAt, &st.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return service.ErrServiceTypeNameExists
		}
		return fmt.Errorf("inserting service type: %w", err)
	}
	return nil
}

// GetByID retrieves a service type by ID within an organization.
func (r *PostgresServiceTypeRepository) GetByID(ctx context.Context, id, orgID int64) (*domain.ServiceType, error) {
	var st domain.ServiceType
	err := r.db.QueryRowContext(ctx,
		`SELECT id, org_id, name, default_duration_minutes, COALESCE(description, ''), active, created_at, updated_at
		 FROM service_types WHERE id = $1 AND org_id = $2`,
		id, orgID,
	).Scan(&st.ID, &st.OrgID, &st.Name, &st.DefaultDuration, &st.Description, &st.Active, &st.CreatedAt, &st.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrServiceTypeNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying service type: %w", err)
	}
	return &st, nil
}

// Update modifies a service type.
func (r *PostgresServiceTypeRepository) Update(ctx context.Context, st *domain.ServiceType) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE service_types SET name = $1, default_duration_minutes = $2, description = $3, active = $4, updated_at = $5
		 WHERE id = $6 AND org_id = $7`,
		st.Name, st.DefaultDuration, st.Description, st.Active, st.UpdatedAt, st.ID, st.OrgID,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return service.ErrServiceTypeNameExists
		}
		return fmt.Errorf("updating service type: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return service.ErrServiceTypeNotFound
	}
	return nil
}

// Delete removes a service type.
func (r *PostgresServiceTypeRepository) Delete(ctx context.Context, id, orgID int64) error {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM service_types WHERE id = $1 AND org_id = $2`, id, orgID,
	)
	if err != nil {
		return fmt.Errorf("deleting service type: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return service.ErrServiceTypeNotFound
	}
	return nil
}

// ListByOrg returns service types for an organization, optionally filtered to active only.
func (r *PostgresServiceTypeRepository) ListByOrg(ctx context.Context, orgID int64, activeOnly bool) ([]*domain.ServiceType, error) {
	query := `SELECT id, org_id, name, default_duration_minutes, COALESCE(description, ''), active, created_at, updated_at
		 FROM service_types WHERE org_id = $1`
	if activeOnly {
		query += ` AND active = true`
	}
	query += ` ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("querying service types: %w", err)
	}
	defer rows.Close()

	var types []*domain.ServiceType
	for rows.Next() {
		var st domain.ServiceType
		if err := rows.Scan(&st.ID, &st.OrgID, &st.Name, &st.DefaultDuration, &st.Description, &st.Active, &st.CreatedAt, &st.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning service type: %w", err)
		}
		types = append(types, &st)
	}
	return types, rows.Err()
}
