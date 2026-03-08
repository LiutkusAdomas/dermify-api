package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"dermify-api/internal/domain"
	"dermify-api/internal/service"
)

// PostgresPhotoRepository implements service.PhotoRepository using PostgreSQL.
type PostgresPhotoRepository struct {
	db *sql.DB
}

// NewPostgresPhotoRepository creates a new PostgresPhotoRepository.
func NewPostgresPhotoRepository(db *sql.DB) *PostgresPhotoRepository {
	return &PostgresPhotoRepository{db: db}
}

// Create inserts a new photo record and scans the generated ID, version, and timestamps.
func (r *PostgresPhotoRepository) Create(ctx context.Context, photo *domain.Photo) error {
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO session_photos (
			session_id, module_id, photo_type, file_path, original_name,
			content_type, size_bytes, created_by, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, version, created_at, updated_at`,
		photo.SessionID, photo.ModuleID, photo.PhotoType, photo.FilePath, photo.OriginalName,
		photo.ContentType, photo.SizeBytes, photo.CreatedBy, photo.UpdatedBy,
	).Scan(&photo.ID, &photo.Version, &photo.CreatedAt, &photo.UpdatedAt)
	if err != nil {
		return fmt.Errorf("inserting session photo: %w", err)
	}

	return nil
}

// GetByID retrieves a photo by ID.
func (r *PostgresPhotoRepository) GetByID(ctx context.Context, id int64) (*domain.Photo, error) {
	var p domain.Photo

	err := r.db.QueryRowContext(ctx,
		`SELECT id, session_id, module_id, photo_type, file_path, original_name,
			content_type, size_bytes, version, created_at, created_by, updated_at, updated_by
		FROM session_photos WHERE id = $1`, id,
	).Scan(
		&p.ID, &p.SessionID, &p.ModuleID, &p.PhotoType, &p.FilePath, &p.OriginalName,
		&p.ContentType, &p.SizeBytes, &p.Version, &p.CreatedAt, &p.CreatedBy, &p.UpdatedAt, &p.UpdatedBy,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrPhotoNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("querying session photo: %w", err)
	}

	return &p, nil
}

// ListBySession returns all photos for a session ordered by creation time.
func (r *PostgresPhotoRepository) ListBySession(ctx context.Context, sessionID int64) ([]*domain.Photo, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, session_id, module_id, photo_type, file_path, original_name,
			content_type, size_bytes, version, created_at, created_by, updated_at, updated_by
		FROM session_photos WHERE session_id = $1 ORDER BY created_at ASC`, sessionID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying session photos: %w", err)
	}
	defer rows.Close()

	photos := make([]*domain.Photo, 0)

	for rows.Next() {
		var p domain.Photo
		if err := rows.Scan(
			&p.ID, &p.SessionID, &p.ModuleID, &p.PhotoType, &p.FilePath, &p.OriginalName,
			&p.ContentType, &p.SizeBytes, &p.Version, &p.CreatedAt, &p.CreatedBy, &p.UpdatedAt, &p.UpdatedBy,
		); err != nil {
			return nil, fmt.Errorf("scanning session photo: %w", err)
		}

		photos = append(photos, &p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating session photos: %w", err)
	}

	return photos, nil
}

// ListByModule returns all photos for a module ordered by creation time.
func (r *PostgresPhotoRepository) ListByModule(ctx context.Context, moduleID int64) ([]*domain.Photo, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, session_id, module_id, photo_type, file_path, original_name,
			content_type, size_bytes, version, created_at, created_by, updated_at, updated_by
		FROM session_photos WHERE module_id = $1 ORDER BY created_at ASC`, moduleID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying module photos: %w", err)
	}
	defer rows.Close()

	photos := make([]*domain.Photo, 0)

	for rows.Next() {
		var p domain.Photo
		if err := rows.Scan(
			&p.ID, &p.SessionID, &p.ModuleID, &p.PhotoType, &p.FilePath, &p.OriginalName,
			&p.ContentType, &p.SizeBytes, &p.Version, &p.CreatedAt, &p.CreatedBy, &p.UpdatedAt, &p.UpdatedBy,
		); err != nil {
			return nil, fmt.Errorf("scanning module photo: %w", err)
		}

		photos = append(photos, &p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating module photos: %w", err)
	}

	return photos, nil
}

// Delete removes a photo record by ID.
func (r *PostgresPhotoRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM session_photos WHERE id = $1`, id,
	)
	if err != nil {
		return fmt.Errorf("deleting session photo: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return service.ErrPhotoNotFound
	}

	return nil
}
