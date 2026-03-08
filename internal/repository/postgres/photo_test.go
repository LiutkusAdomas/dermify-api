package postgres

import (
	"dermify-api/internal/service"
)

// Compile-time interface checks.
var _ service.PhotoRepository = (*PostgresPhotoRepository)(nil)
