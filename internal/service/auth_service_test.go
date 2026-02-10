package service

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/complyark/datalens/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestAuthService() (*AuthService, *mockUserRepo) {
	repo := newMockUserRepo()
	roleRepo := newMockRoleRepo()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	svc := NewAuthService(repo, roleRepo, "test-secret-key-32chars!!", 15*time.Minute, 7*24*time.Hour, logger)
	return svc, repo
}

var testTenantID = types.NewID()

func TestAuthService_Register_Success(t *testing.T) {
	svc, _ := newTestAuthService()
	ctx := context.Background()

	user, err := svc.Register(ctx, RegisterInput{
		TenantID: testTenantID,
		Email:    "alice@acme.local",
		Name:     "Alice Admin",
		Password: "securepassword123",
	})

	require.NoError(t, err)
	assert.NotEqual(t, types.ID{}, user.ID)
	assert.Equal(t, "alice@acme.local", user.Email)
	assert.Equal(t, "Alice Admin", user.Name)
	assert.NotEqual(t, "securepassword123", user.Password, "password should be hashed")
}

func TestAuthService_Register_EmptyEmail(t *testing.T) {
	svc, _ := newTestAuthService()
	ctx := context.Background()

	_, err := svc.Register(ctx, RegisterInput{
		TenantID: testTenantID,
		Email:    "",
		Name:     "Alice",
		Password: "securepassword123",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "email")
}

func TestAuthService_Register_ShortPassword(t *testing.T) {
	svc, _ := newTestAuthService()
	ctx := context.Background()

	_, err := svc.Register(ctx, RegisterInput{
		TenantID: testTenantID,
		Email:    "alice@acme.local",
		Name:     "Alice",
		Password: "short",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "password")
}

func TestAuthService_Register_DuplicateEmail(t *testing.T) {
	svc, _ := newTestAuthService()
	ctx := context.Background()

	_, err := svc.Register(ctx, RegisterInput{
		TenantID: testTenantID,
		Email:    "alice@acme.local",
		Name:     "Alice",
		Password: "securepassword123",
	})
	require.NoError(t, err)

	_, err = svc.Register(ctx, RegisterInput{
		TenantID: testTenantID,
		Email:    "alice@acme.local",
		Name:     "Alice Duplicate",
		Password: "anotherpassword123",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "conflict")
}

func TestAuthService_Login_Success(t *testing.T) {
	svc, _ := newTestAuthService()
	ctx := context.Background()

	_, err := svc.Register(ctx, RegisterInput{
		TenantID: testTenantID,
		Email:    "alice@acme.local",
		Name:     "Alice",
		Password: "securepassword123",
	})
	require.NoError(t, err)

	pair, err := svc.Login(ctx, LoginInput{
		TenantID: testTenantID,
		Email:    "alice@acme.local",
		Password: "securepassword123",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
	assert.False(t, pair.ExpiresAt.IsZero())
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	svc, _ := newTestAuthService()
	ctx := context.Background()

	_, _ = svc.Register(ctx, RegisterInput{
		TenantID: testTenantID,
		Email:    "alice@acme.local",
		Name:     "Alice",
		Password: "securepassword123",
	})

	_, err := svc.Login(ctx, LoginInput{
		TenantID: testTenantID,
		Email:    "alice@acme.local",
		Password: "wrongpassword",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestAuthService_Login_NonexistentUser(t *testing.T) {
	svc, _ := newTestAuthService()
	ctx := context.Background()

	_, err := svc.Login(ctx, LoginInput{
		TenantID: testTenantID,
		Email:    "nobody@acme.local",
		Password: "securepassword123",
	})

	require.Error(t, err)
}

func TestAuthService_ValidateToken(t *testing.T) {
	svc, _ := newTestAuthService()
	ctx := context.Background()

	_, _ = svc.Register(ctx, RegisterInput{
		TenantID: testTenantID,
		Email:    "alice@acme.local",
		Name:     "Alice",
		Password: "securepassword123",
	})

	pair, _ := svc.Login(ctx, LoginInput{
		TenantID: testTenantID,
		Email:    "alice@acme.local",
		Password: "securepassword123",
	})

	claims, err := svc.ValidateToken(pair.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, testTenantID, claims.TenantID)
	assert.Equal(t, "alice@acme.local", claims.Email)
}

func TestAuthService_RefreshToken(t *testing.T) {
	svc, _ := newTestAuthService()
	ctx := context.Background()

	_, _ = svc.Register(ctx, RegisterInput{
		TenantID: testTenantID,
		Email:    "alice@acme.local",
		Name:     "Alice",
		Password: "securepassword123",
	})

	pair, _ := svc.Login(ctx, LoginInput{
		TenantID: testTenantID,
		Email:    "alice@acme.local",
		Password: "securepassword123",
	})

	newPair, err := svc.RefreshToken(ctx, pair.RefreshToken)
	require.NoError(t, err)
	assert.NotEmpty(t, newPair.AccessToken)
	assert.NotEmpty(t, newPair.RefreshToken)

	// Validate the new access token is a valid JWT
	claims, err := svc.ValidateToken(newPair.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, "alice@acme.local", claims.Email)
}

func TestAuthService_RefreshToken_InvalidToken(t *testing.T) {
	svc, _ := newTestAuthService()
	ctx := context.Background()

	_, err := svc.RefreshToken(ctx, "invalid-token-string")
	require.Error(t, err)
}

func TestAuthService_GetCurrentUser(t *testing.T) {
	svc, _ := newTestAuthService()
	ctx := context.Background()

	user, _ := svc.Register(ctx, RegisterInput{
		TenantID: testTenantID,
		Email:    "alice@acme.local",
		Name:     "Alice",
		Password: "securepassword123",
	})

	fetched, err := svc.GetCurrentUser(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.Email, fetched.Email)
}
