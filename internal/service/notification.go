package service

import (
	"context"
	"log/slog"
	"time"

	"dermify-api/internal/domain"
)

// NotificationRepository defines the data access contract for notifications.
type NotificationRepository interface {
	Create(ctx context.Context, n *domain.Notification) error
	UpdateStatus(ctx context.Context, id int64, status string, sentAt *time.Time, errMsg string) error
}

// NotificationService handles notification business logic.
// The default implementation logs and stores notifications; real delivery is pluggable.
type NotificationService struct {
	repo NotificationRepository
}

// NewNotificationService creates a new NotificationService.
func NewNotificationService(repo NotificationRepository) *NotificationService {
	return &NotificationService{repo: repo}
}

// Send records a notification and attempts delivery (currently logs only).
func (s *NotificationService) Send(ctx context.Context, n *domain.Notification) error {
	n.Status = domain.NotificationStatusPending
	n.CreatedAt = time.Now()

	if err := s.repo.Create(ctx, n); err != nil {
		return err
	}

	// Placeholder: in production, dispatch to an external messaging gateway here.
	slog.Info("notification dispatched",
		"id", n.ID,
		"type", n.Type,
		"channel", n.Channel,
		"recipient", n.Recipient,
	)

	now := time.Now()
	n.Status = domain.NotificationStatusSent
	n.SentAt = &now
	_ = s.repo.UpdateStatus(ctx, n.ID, domain.NotificationStatusSent, &now, "")

	return nil
}

// SendAppointmentNotification is a convenience method for appointment-related notifications.
func (s *NotificationService) SendAppointmentNotification(ctx context.Context, orgID, patientID, appointmentID int64, notifType, recipient, subject, body string) {
	n := &domain.Notification{
		OrgID:       orgID,
		PatientID:   patientID,
		Channel:     domain.NotificationChannelEmail,
		Type:        notifType,
		Recipient:   recipient,
		Subject:     subject,
		Body:        body,
		ReferenceID: appointmentID,
	}
	if err := s.Send(ctx, n); err != nil {
		slog.Error("failed to send notification", "error", err, "type", notifType, "appointmentId", appointmentID)
	}
}
