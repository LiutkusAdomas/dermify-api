package service_test

import (
	"context"
	"errors"
	"testing"

	"dermify-api/internal/domain"
	"dermify-api/internal/service"
	"dermify-api/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validUser() *domain.User {
	return &domain.User{
		Username:     "johndoe",
		Email:        "john@example.com",
		PasswordHash: "$2a$10$fakehash",
	}
}

func TestCreateUser_ValidData(t *testing.T) {
	repoCalled := false

	mock := &testutil.MockUserRepository{
		CreateFn: func(_ context.Context, user *domain.User) error {
			repoCalled = true
			user.ID = 1
			return nil
		},
	}

	svc := service.NewUserService(mock)
	user := validUser()
	err := svc.Create(context.Background(), user)

	require.NoError(t, err)
	assert.True(t, repoCalled, "repository Create should be called")
	assert.Equal(t, int64(1), user.ID)
	assert.False(t, user.CreatedAt.IsZero(), "created_at should be set")
	assert.False(t, user.UpdatedAt.IsZero(), "updated_at should be set")
}

func TestCreateUser_MissingFields(t *testing.T) {
	mock := &testutil.MockUserRepository{}
	svc := service.NewUserService(mock)

	tests := []struct {
		name string
		user *domain.User
	}{
		{
			name: "missing username",
			user: &domain.User{
				Email:        "john@example.com",
				PasswordHash: "$2a$10$fakehash",
			},
		},
		{
			name: "missing email",
			user: &domain.User{
				Username:     "johndoe",
				PasswordHash: "$2a$10$fakehash",
			},
		},
		{
			name: "missing password hash",
			user: &domain.User{
				Username: "johndoe",
				Email:    "john@example.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.Create(context.Background(), tt.user)

			require.Error(t, err)
			assert.True(t, errors.Is(err, service.ErrInvalidUserData))
		})
	}
}

func TestCreateUser_AlreadyExists(t *testing.T) {
	mock := &testutil.MockUserRepository{
		CreateFn: func(_ context.Context, _ *domain.User) error {
			return service.ErrUserAlreadyExists
		},
	}

	svc := service.NewUserService(mock)
	user := validUser()
	err := svc.Create(context.Background(), user)

	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrUserAlreadyExists))
}

func TestGetUserByID_Found(t *testing.T) {
	expected := &domain.User{ID: 42, Username: "johndoe", Email: "john@example.com"}

	mock := &testutil.MockUserRepository{
		GetByIDFn: func(_ context.Context, id int64) (*domain.User, error) {
			assert.Equal(t, int64(42), id)
			return expected, nil
		},
	}

	svc := service.NewUserService(mock)
	user, err := svc.GetByID(context.Background(), 42)

	require.NoError(t, err)
	assert.Equal(t, expected, user)
}

func TestGetUserByID_NotFound(t *testing.T) {
	mock := &testutil.MockUserRepository{
		GetByIDFn: func(_ context.Context, _ int64) (*domain.User, error) {
			return nil, service.ErrUserNotFound
		},
	}

	svc := service.NewUserService(mock)
	_, err := svc.GetByID(context.Background(), 999)

	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrUserNotFound))
}

func TestUpdateUser_ValidData(t *testing.T) {
	repoCalled := false

	mock := &testutil.MockUserRepository{
		UpdateFn: func(_ context.Context, _ *domain.User) error {
			repoCalled = true
			return nil
		},
	}

	svc := service.NewUserService(mock)
	user := &domain.User{ID: 1, Username: "updated", Email: "updated@example.com"}
	err := svc.Update(context.Background(), user)

	require.NoError(t, err)
	assert.True(t, repoCalled, "repository Update should be called")
	assert.False(t, user.UpdatedAt.IsZero(), "updated_at should be set")
}

func TestUpdateUser_MissingFields(t *testing.T) {
	mock := &testutil.MockUserRepository{}
	svc := service.NewUserService(mock)

	tests := []struct {
		name string
		user *domain.User
	}{
		{
			name: "missing username",
			user: &domain.User{ID: 1, Email: "john@example.com"},
		},
		{
			name: "missing email",
			user: &domain.User{ID: 1, Username: "johndoe"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.Update(context.Background(), tt.user)

			require.Error(t, err)
			assert.True(t, errors.Is(err, service.ErrInvalidUserData))
		})
	}
}

func TestDeleteUser_Success(t *testing.T) {
	repoCalled := false

	mock := &testutil.MockUserRepository{
		DeleteFn: func(_ context.Context, id int64) error {
			repoCalled = true
			assert.Equal(t, int64(1), id)
			return nil
		},
	}

	svc := service.NewUserService(mock)
	err := svc.Delete(context.Background(), 1)

	require.NoError(t, err)
	assert.True(t, repoCalled, "repository Delete should be called")
}

func TestDeleteUser_NotFound(t *testing.T) {
	mock := &testutil.MockUserRepository{
		DeleteFn: func(_ context.Context, _ int64) error {
			return service.ErrUserNotFound
		},
	}

	svc := service.NewUserService(mock)
	err := svc.Delete(context.Background(), 999)

	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrUserNotFound))
}

func TestListUsers_Empty(t *testing.T) {
	mock := &testutil.MockUserRepository{
		ListFn: func(_ context.Context) ([]*domain.User, error) {
			return []*domain.User{}, nil
		},
	}

	svc := service.NewUserService(mock)
	users, err := svc.List(context.Background())

	require.NoError(t, err)
	assert.Empty(t, users)
}

func TestListUsers_WithResults(t *testing.T) {
	expected := []*domain.User{
		{ID: 1, Username: "user1", Email: "user1@example.com"},
		{ID: 2, Username: "user2", Email: "user2@example.com"},
	}

	mock := &testutil.MockUserRepository{
		ListFn: func(_ context.Context) ([]*domain.User, error) {
			return expected, nil
		},
	}

	svc := service.NewUserService(mock)
	users, err := svc.List(context.Background())

	require.NoError(t, err)
	assert.Len(t, users, 2)
	assert.Equal(t, expected, users)
}
