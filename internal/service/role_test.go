package service_test

import (
	"context"
	"errors"
	"testing"

	"dermify-api/internal/service"
	"dermify-api/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAssignRole_ValidRole tests assigning a valid role to a user.
func TestAssignRole_ValidRole(t *testing.T) {
	var calledWith struct {
		userID int64
		role   string
	}

	mock := &testutil.MockRoleRepository{
		UpdateUserRoleFn: func(_ context.Context, userID int64, role string) error {
			calledWith.userID = userID
			calledWith.role = role
			return nil
		},
	}

	svc := service.NewRoleService(mock)
	err := svc.AssignRole(context.Background(), 42, "doctor")

	require.NoError(t, err)
	assert.Equal(t, int64(42), calledWith.userID)
	assert.Equal(t, "doctor", calledWith.role)
}

// TestAssignRole_InvalidRole tests that assigning an invalid role returns an error.
func TestAssignRole_InvalidRole(t *testing.T) {
	repoCalled := false

	mock := &testutil.MockRoleRepository{
		UpdateUserRoleFn: func(_ context.Context, _ int64, _ string) error {
			repoCalled = true
			return nil
		},
	}

	svc := service.NewRoleService(mock)
	err := svc.AssignRole(context.Background(), 42, "superuser")

	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrInvalidRole))
	assert.False(t, repoCalled, "repository should not be called for invalid roles")
}

// TestIsFirstUser tests detection of the first registered user for admin bootstrap.
func TestIsFirstUser(t *testing.T) {
	tests := []struct {
		name     string
		count    int64
		expected bool
	}{
		{name: "single user is first", count: 1, expected: true},
		{name: "no users is first", count: 0, expected: true},
		{name: "multiple users is not first", count: 5, expected: false},
		{name: "two users is not first", count: 2, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &testutil.MockRoleRepository{
				CountUsersFn: func(_ context.Context) (int64, error) {
					return tt.count, nil
				},
			}

			svc := service.NewRoleService(mock)
			result, err := svc.IsFirstUser(context.Background())

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGetUserRole tests retrieving a user's current role.
func TestGetUserRole(t *testing.T) {
	mock := &testutil.MockRoleRepository{
		GetUserRoleFn: func(_ context.Context, userID int64) (string, error) {
			if userID == 1 {
				return "doctor", nil
			}
			return "", nil
		},
	}

	svc := service.NewRoleService(mock)

	role, err := svc.GetUserRole(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "doctor", role)

	role, err = svc.GetUserRole(context.Background(), 99)
	require.NoError(t, err)
	assert.Equal(t, "", role)
}
