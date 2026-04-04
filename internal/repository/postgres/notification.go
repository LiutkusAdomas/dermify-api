package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"dermify-api/internal/domain"
)

// PostgresNotificationRepository implements service.NotificationRepository.
type PostgresNotificationRepository struct {
	db *sql.DB
}

// NewPostgresNotificationRepository creates a new PostgresNotificationRepository.
func NewPostgresNotificationRepository(db *sql.DB) *PostgresNotificationRepository {
	return &PostgresNotificationRepository{db: db}
}

// Create inserts a notification record.
func (r *PostgresNotificationRepository) Create(ctx context.Context, n *domain.Notification) error {
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO notifications (org_id, patient_id, channel, type, recipient, subject, body, status, reference_id)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING id, created_at`,
		n.OrgID, n.PatientID, n.Channel, n.Type, n.Recipient, n.Subject, n.Body, n.Status, n.ReferenceID,
	).Scan(&n.ID, &n.CreatedAt)
	if err != nil {
		return fmt.Errorf("inserting notification: %w", err)
	}
	return nil
}

// UpdateStatus updates the delivery status of a notification.
func (r *PostgresNotificationRepository) UpdateStatus(ctx context.Context, id int64, status string, sentAt *time.Time, errMsg string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE notifications SET status = $1, sent_at = $2, error = $3 WHERE id = $4`,
		status, sentAt, errMsg, id,
	)
	if err != nil {
		return fmt.Errorf("updating notification status: %w", err)
	}
	return nil
}
