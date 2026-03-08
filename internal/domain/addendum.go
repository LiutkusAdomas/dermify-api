package domain

import "time"

// Addendum represents an immutable addendum to a locked session.
// Addendums are insert-only: once created, they cannot be modified or deleted.
type Addendum struct {
	ID        int64     `json:"id"`
	SessionID int64     `json:"session_id"`
	AuthorID  int64     `json:"author_id"`
	Reason    string    `json:"reason"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}
