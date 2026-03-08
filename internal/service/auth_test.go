package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"dermify-api/internal/domain"
	"dermify-api/internal/service"
	"dermify-api/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newAuthService(authRepo *testutil.MockAuthRepository, userRepo *testutil.MockUserRepository) *service.AuthService {
	roleRepo := &testutil.MockRoleRepository{
		CountUsersFn: func(_ context.Context) (int64, error) {
			return 10, nil // not first user by default
		},
	}
	roleSvc := service.NewRoleService(roleRepo)

	return service.NewAuthService(authRepo, userRepo, roleSvc)
}

func TestRegister_Success(t *testing.T) {
	userRepoCalled := false

	userRepo := &testutil.MockUserRepository{
		CreateFn: func(_ context.Context, user *domain.User) error {
			userRepoCalled = true
			user.ID = 1
			return nil
		},
	}

	svc := newAuthService(&testutil.MockAuthRepository{}, userRepo)
	user, err := svc.Register(context.Background(), "johndoe", "john@example.com", "$2a$10$fakehash")

	require.NoError(t, err)
	assert.True(t, userRepoCalled, "user repo Create should be called")
	assert.Equal(t, int64(1), user.ID)
	assert.Equal(t, "johndoe", user.Username)
	assert.Equal(t, "john@example.com", user.Email)
	assert.False(t, user.CreatedAt.IsZero(), "created_at should be set")
}

func TestRegister_DuplicateUser(t *testing.T) {
	userRepo := &testutil.MockUserRepository{
		CreateFn: func(_ context.Context, _ *domain.User) error {
			return service.ErrUserAlreadyExists
		},
	}

	svc := newAuthService(&testutil.MockAuthRepository{}, userRepo)
	_, err := svc.Register(context.Background(), "johndoe", "john@example.com", "$2a$10$fakehash")

	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrUserAlreadyExists))
}

func TestRegister_FirstUserBootstrap(t *testing.T) {
	roleAssigned := false

	roleRepo := &testutil.MockRoleRepository{
		CountUsersFn: func(_ context.Context) (int64, error) {
			return 1, nil // first user
		},
		UpdateUserRoleFn: func(_ context.Context, _ int64, role string) error {
			roleAssigned = true
			assert.Equal(t, domain.RoleAdmin, role)
			return nil
		},
	}
	roleSvc := service.NewRoleService(roleRepo)

	userRepo := &testutil.MockUserRepository{
		CreateFn: func(_ context.Context, user *domain.User) error {
			user.ID = 1
			return nil
		},
	}

	svc := service.NewAuthService(&testutil.MockAuthRepository{}, userRepo, roleSvc)
	_, err := svc.Register(context.Background(), "admin", "admin@example.com", "$2a$10$fakehash")

	require.NoError(t, err)
	assert.True(t, roleAssigned, "first user should be auto-promoted to admin")
}

func TestAuthenticate_Success(t *testing.T) {
	expected := &domain.User{
		ID:           1,
		Username:     "johndoe",
		Email:        "john@example.com",
		PasswordHash: "$2a$10$fakehash",
		Role:         "doctor",
	}

	userRepo := &testutil.MockUserRepository{
		GetByEmailFn: func(_ context.Context, email string) (*domain.User, error) {
			assert.Equal(t, "john@example.com", email)
			return expected, nil
		},
	}

	svc := newAuthService(&testutil.MockAuthRepository{}, userRepo)
	user, err := svc.Authenticate(context.Background(), "john@example.com")

	require.NoError(t, err)
	assert.Equal(t, expected, user)
}

func TestAuthenticate_UserNotFound(t *testing.T) {
	userRepo := &testutil.MockUserRepository{
		GetByEmailFn: func(_ context.Context, _ string) (*domain.User, error) {
			return nil, service.ErrUserNotFound
		},
	}

	svc := newAuthService(&testutil.MockAuthRepository{}, userRepo)
	_, err := svc.Authenticate(context.Background(), "unknown@example.com")

	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrInvalidCredentials))
}

func TestValidateRefreshToken_Success(t *testing.T) {
	authRepo := &testutil.MockAuthRepository{
		ValidateRefreshTokenFn: func(_ context.Context, hash string) (int64, error) {
			assert.Equal(t, "somehash", hash)
			return 42, nil
		},
	}

	svc := newAuthService(authRepo, &testutil.MockUserRepository{})
	userID, err := svc.ValidateRefreshToken(context.Background(), "somehash")

	require.NoError(t, err)
	assert.Equal(t, int64(42), userID)
}

func TestValidateRefreshToken_Invalid(t *testing.T) {
	authRepo := &testutil.MockAuthRepository{
		ValidateRefreshTokenFn: func(_ context.Context, _ string) (int64, error) {
			return 0, errors.New("no rows")
		},
	}

	svc := newAuthService(authRepo, &testutil.MockUserRepository{})
	_, err := svc.ValidateRefreshToken(context.Background(), "badhash")

	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrRefreshTokenInvalid))
}

func TestStoreRefreshToken_Success(t *testing.T) {
	storeCalled := false

	authRepo := &testutil.MockAuthRepository{
		StoreRefreshTokenFn: func(_ context.Context, userID int64, hash string, expiresAt time.Time) error {
			storeCalled = true
			assert.Equal(t, int64(1), userID)
			assert.Equal(t, "tokenhash", hash)
			assert.False(t, expiresAt.IsZero())
			return nil
		},
	}

	svc := newAuthService(authRepo, &testutil.MockUserRepository{})
	err := svc.StoreRefreshToken(context.Background(), 1, "tokenhash", time.Now().Add(24*time.Hour))

	require.NoError(t, err)
	assert.True(t, storeCalled, "auth repo StoreRefreshToken should be called")
}

func TestRevokeRefreshToken_Success(t *testing.T) {
	revokeCalled := false

	authRepo := &testutil.MockAuthRepository{
		RevokeRefreshTokenFn: func(_ context.Context, hash string) error {
			revokeCalled = true
			assert.Equal(t, "tokenhash", hash)
			return nil
		},
	}

	svc := newAuthService(authRepo, &testutil.MockUserRepository{})
	err := svc.RevokeRefreshToken(context.Background(), "tokenhash")

	require.NoError(t, err)
	assert.True(t, revokeCalled, "auth repo RevokeRefreshToken should be called")
}

func TestRevokeAllUserRefreshTokens_Success(t *testing.T) {
	revokeAllCalled := false

	authRepo := &testutil.MockAuthRepository{
		RevokeAllUserRefreshTokensFn: func(_ context.Context, userID int64) error {
			revokeAllCalled = true
			assert.Equal(t, int64(1), userID)
			return nil
		},
	}

	svc := newAuthService(authRepo, &testutil.MockUserRepository{})
	err := svc.RevokeAllUserRefreshTokens(context.Background(), 1)

	require.NoError(t, err)
	assert.True(t, revokeAllCalled, "auth repo RevokeAllUserRefreshTokens should be called")
}

func TestGetUserByID_Success(t *testing.T) {
	expected := &domain.User{ID: 1, Username: "johndoe", Email: "john@example.com"}

	userRepo := &testutil.MockUserRepository{
		GetByIDFn: func(_ context.Context, id int64) (*domain.User, error) {
			assert.Equal(t, int64(1), id)
			return expected, nil
		},
	}

	svc := newAuthService(&testutil.MockAuthRepository{}, userRepo)
	user, err := svc.GetUserByID(context.Background(), 1)

	require.NoError(t, err)
	assert.Equal(t, expected, user)
}
