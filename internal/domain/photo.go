package domain

import "time"

// Photo type constants identify the category of a clinical photograph.
const (
	PhotoTypeBefore = "before"
	PhotoTypeLabel  = "label"
)

// MaxPhotoSize is the maximum allowed upload size (10 MB).
const MaxPhotoSize = 10 << 20

// AllowedPhotoContentTypes lists the MIME types accepted for clinical photos.
var AllowedPhotoContentTypes = []string{"image/jpeg", "image/png"} //nolint:gochecknoglobals // constant slice

// Photo represents a clinical photograph associated with a treatment session.
type Photo struct {
	ID           int64     `json:"id"`
	SessionID    int64     `json:"session_id"`
	ModuleID     *int64    `json:"module_id,omitempty"`
	PhotoType    string    `json:"photo_type"`
	FilePath     string    `json:"file_path"`
	OriginalName string    `json:"original_name"`
	ContentType  string    `json:"content_type"`
	SizeBytes    int64     `json:"size_bytes"`
	Version      int       `json:"version"`
	CreatedAt    time.Time `json:"created_at"`
	CreatedBy    int64     `json:"created_by"`
	UpdatedAt    time.Time `json:"updated_at"`
	UpdatedBy    int64     `json:"updated_by"`
}
