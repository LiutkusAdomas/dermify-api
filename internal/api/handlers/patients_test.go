//go:build ignore
// +build ignore

// Remove //go:build ignore after Plan 01-04 creates patient handlers.

package handlers_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Suppress unused import lint.
var _ = assert.Equal

// TestDoctorAccess_PatientsEndpoint tests that a doctor can access patient endpoints (RBAC-02).
func TestDoctorAccess_PatientsEndpoint(t *testing.T) {
	t.Skip("Implemented in Plan 01-04")
}

// TestAdminAccess_PatientsEndpoint tests that an admin can access patient management endpoints (RBAC-03).
func TestAdminAccess_PatientsEndpoint(t *testing.T) {
	t.Skip("Implemented in Plan 01-04")
}

// TestUnauthorizedAccess_PatientsEndpoint tests that a user without a role receives 403.
func TestUnauthorizedAccess_PatientsEndpoint(t *testing.T) {
	t.Skip("Implemented in Plan 01-04")
}
