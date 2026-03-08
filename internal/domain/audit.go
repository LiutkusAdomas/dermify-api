package domain

import (
	"encoding/json"
	"time"
)

// AuditEntry represents a single entry in the append-only audit trail.
// Audit entries are created by database triggers and cannot be modified or deleted.
type AuditEntry struct {
	ID          int64           `json:"id"`
	Action      string          `json:"action"`
	PerformedAt time.Time       `json:"performed_at"`
	UserID      *int64          `json:"user_id,omitempty"`
	EntityType  string          `json:"entity_type"`
	EntityID    int64           `json:"entity_id"`
	OldValues   json.RawMessage `json:"old_values,omitempty"`
	NewValues   json.RawMessage `json:"new_values,omitempty"`
}
