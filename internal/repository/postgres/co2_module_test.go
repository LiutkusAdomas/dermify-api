package postgres

import (
	"dermify-api/internal/service"
)

// Compile-time interface checks.
var _ service.CO2ModuleRepository = (*PostgresCO2ModuleRepository)(nil)
