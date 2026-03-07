//go:build ignore
// +build ignore

// Remove //go:build ignore after Plan 01-01 creates service interfaces.

package service_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Suppress unused import lint.
var _ = assert.Equal

// TestAssignRole_ValidRole tests assigning a valid role to a user.
func TestAssignRole_ValidRole(t *testing.T) {
	t.Skip("Implemented in Plan 01-01")
}

// TestAssignRole_InvalidRole tests that assigning an invalid role returns an error.
func TestAssignRole_InvalidRole(t *testing.T) {
	t.Skip("Implemented in Plan 01-01")
}

// TestIsFirstUser tests detection of the first registered user for admin bootstrap.
func TestIsFirstUser(t *testing.T) {
	t.Skip("Implemented in Plan 01-01")
}

// TestGetUserRole tests retrieving a user's current role.
func TestGetUserRole(t *testing.T) {
	t.Skip("Implemented in Plan 01-01")
}
