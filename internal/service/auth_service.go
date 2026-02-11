package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/complyark/datalens/internal/domain/identity"
	"github.com/complyark/datalens/pkg/types"
)

// AuthService handles user authentication and JWT token management.
type AuthService struct {
	userRepo   identity.UserRepository
	roleRepo   identity.RoleRepository
	secretKey  []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
	logger     *slog.Logger
}

// NewAuthService creates a new AuthService.
func NewAuthService(
	userRepo identity.UserRepository,
	roleRepo identity.RoleRepository,
	secretKey string,
	accessTTL, refreshTTL time.Duration,
	logger *slog.Logger,
) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		roleRepo:   roleRepo,
		secretKey:  []byte(secretKey),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
		logger:     logger.With("service", "auth"),
	}
}

// Claims represents the JWT token claims.
type Claims struct {
	jwt.RegisteredClaims
	UserID   types.ID `json:"user_id"`
	TenantID types.ID `json:"tenant_id"`
	Email    string   `json:"email"`
	Name     string   `json:"name"`
}

// TokenPair holds access and refresh tokens.
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// RegisterInput holds fields for user registration.
type RegisterInput struct {
	TenantID types.ID
	Email    string
	Name     string
	Password string
}

// Register creates a new user with hashed password.
func (s *AuthService) Register(ctx context.Context, in RegisterInput) (*identity.User, error) {
	if in.Email == "" {
		return nil, types.NewValidationError("email is required", nil)
	}
	if in.Name == "" {
		return nil, types.NewValidationError("name is required", nil)
	}
	if len(in.Password) < 8 {
		return nil, types.NewValidationError("password must be at least 8 characters", nil)
	}

	// Check for duplicate email within tenant
	existing, err := s.userRepo.GetByEmail(ctx, in.TenantID, in.Email)
	if err == nil && existing != nil {
		return nil, types.NewConflictError("User", "email", in.Email)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &identity.User{
		Email:    in.Email,
		Name:     in.Name,
		Password: string(hash),
		Status:   identity.UserActive,
	}
	user.TenantID = in.TenantID

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	s.logger.InfoContext(ctx, "user registered",
		slog.String("tenant_id", in.TenantID.String()),
		slog.String("id", user.ID.String()),
		slog.String("email", user.Email),
	)
	return user, nil
}

// LoginInput holds credentials for authentication.
type LoginInput struct {
	TenantID types.ID
	Email    string
	Password string
}

// Login authenticates a user and returns a token pair.
func (s *AuthService) Login(ctx context.Context, in LoginInput) (*TokenPair, error) {
	user, err := s.userRepo.GetByEmail(ctx, in.TenantID, in.Email)
	if err != nil {
		return nil, types.NewUnauthorizedError("invalid email or password")
	}

	if user.Status != identity.UserActive {
		return nil, types.NewForbiddenError("account is not active")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(in.Password)); err != nil {
		return nil, types.NewUnauthorizedError("invalid email or password")
	}

	pair, err := s.generateTokenPair(user)
	if err != nil {
		return nil, err
	}

	// Update last login
	now := time.Now().UTC()
	user.LastLoginAt = &now
	_ = s.userRepo.Update(ctx, user)

	s.logger.InfoContext(ctx, "user logged in",
		slog.String("tenant_id", in.TenantID.String()),

		slog.String("id", user.ID.String()),
		slog.String("email", user.Email),
	)
	return pair, nil
}

// RefreshToken validates a refresh token and returns a new token pair.
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	claims, err := s.ValidateToken(refreshToken)
	if err != nil {
		return nil, types.NewUnauthorizedError("invalid or expired refresh token")
	}

	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, types.NewUnauthorizedError("user not found")
	}

	if user.Status != identity.UserActive {
		return nil, types.NewForbiddenError("account is not active")
	}

	return s.generateTokenPair(user)
}

// ValidateToken parses and validates a JWT token, returning the claims.
func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secretKey, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

// GetCurrentUser retrieves the user from a validated token's claims.
func (s *AuthService) GetCurrentUser(ctx context.Context, userID types.ID) (*identity.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

// GetUserRoles loads all roles assigned to a user by their RoleIDs.
func (s *AuthService) GetUserRoles(ctx context.Context, userID types.ID) ([]identity.Role, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user for roles: %w", err)
	}

	roles := make([]identity.Role, 0, len(user.RoleIDs))
	for _, rid := range user.RoleIDs {
		role, err := s.roleRepo.GetByID(ctx, rid)
		if err != nil {
			s.logger.WarnContext(ctx, "role not found for user",
				slog.String("tenant_id", user.TenantID.String()),
				slog.String("role_id", rid.String()),
				slog.String("user_id", userID.String()),
			)
			continue
		}
		roles = append(roles, *role)
	}
	return roles, nil
}

// HasPermission checks if any of the given roles grant access to the resource+action.
func HasPermission(roles []identity.Role, resource, action string) bool {
	for _, role := range roles {
		// ADMIN role has full access to everything
		if role.Name == identity.RoleAdmin {
			return true
		}
		for _, perm := range role.Permissions {
			if perm.Resource == resource || perm.Resource == "*" {
				for _, a := range perm.Actions {
					if a == action || a == "*" {
						return true
					}
				}
			}
		}
	}
	return false
}

func (s *AuthService) generateTokenPair(user *identity.User) (*TokenPair, error) {
	now := time.Now().UTC()

	// Access token
	accessClaims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTTL)),
			Issuer:    "datalens",
		},
		UserID:   user.ID,
		TenantID: user.TenantID,
		Email:    user.Email,
		Name:     user.Name,
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(s.secretKey)
	if err != nil {
		return nil, fmt.Errorf("sign access token: %w", err)
	}

	// Refresh token (longer-lived, minimal claims)
	refreshClaims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.refreshTTL)),
			Issuer:    "datalens",
		},
		UserID:   user.ID,
		TenantID: user.TenantID,
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(s.secretKey)
	if err != nil {
		return nil, fmt.Errorf("sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    now.Add(s.accessTTL),
	}, nil
}
