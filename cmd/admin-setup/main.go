package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"github.com/complyark/datalens/internal/config"
	"github.com/complyark/datalens/internal/domain/identity"
	"github.com/complyark/datalens/internal/repository"
	"github.com/complyark/datalens/internal/service"
	"github.com/complyark/datalens/pkg/database"
	"github.com/complyark/datalens/pkg/types"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	_ = godotenv.Load()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	db, err := database.New(cfg.DB)
	if err != nil {
		return fmt.Errorf("connect db: %w", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Initialize Repos
	roleRepo := repository.NewRoleRepo(db)
	tenantRepo := repository.NewTenantRepo(db)
	userRepo := repository.NewUserRepo(db)

	// dummy auth svc just for hashing password if needed, but we can reuse create logic
	// Actually we need AuthService to hash password properly if we employ manual creation
	// Or we use TenantService.Onboard logic which uses AuthService.

	// But simpler: Ensure role exists first.
	platformRoleID, err := ensurePlatformRole(ctx, roleRepo, logger)
	if err != nil {
		return err
	}

	// Ensure system tenant exists (optional, but good for platform admin to "live" somewhere)
	// We'll use 'datalens-system' domain
	tenantID, err := ensureSystemTenant(ctx, tenantRepo, logger)
	if err != nil {
		return err
	}

	// Ensure platform admin user exists
	if err := ensurePlatformUser(ctx, userRepo, db, tenantID, platformRoleID, cfg.App.SecretKey, logger); err != nil {
		return err
	}

	return nil
}

func ensurePlatformRole(ctx context.Context, repo identity.RoleRepository, logger *slog.Logger) (types.ID, error) {
	roles, err := repo.GetSystemRoles(ctx)
	if err != nil {
		return types.ID{}, fmt.Errorf("get system roles: %w", err)
	}

	for _, r := range roles {
		if r.Name == identity.RolePlatformAdmin {
			logger.Info("Platform Admin role exists", "id", r.ID)
			return r.ID, nil
		}
	}

	// Create if not exists
	role := &identity.Role{
		Name:        identity.RolePlatformAdmin,
		Description: "Super Administrator with full platform access",
		Permissions: []identity.Permission{
			{Resource: "*", Actions: []string{"*"}}, // Full access
		},
		IsSystem: true,
	}

	if err := repo.Create(ctx, role); err != nil {
		return types.ID{}, fmt.Errorf("create platform role: %w", err)
	}
	logger.Info("Created Platform Admin role", "id", role.ID)
	return role.ID, nil
}

func ensureSystemTenant(ctx context.Context, repo identity.TenantRepository, logger *slog.Logger) (types.ID, error) {
	const systemDomain = "system.datalens.io"

	t, err := repo.GetByDomain(ctx, systemDomain)
	if err == nil && t != nil {
		logger.Info("System tenant exists", "id", t.ID)
		return t.ID, nil
	}

	// Create
	sysTenant := &identity.Tenant{
		Name:     "DataLens System",
		Domain:   systemDomain,
		Industry: "TECHNOLOGY",
		Country:  "IN",
		Plan:     identity.PlanEnterprise,
		Status:   identity.TenantActive,
		Settings: identity.TenantSettings{
			DefaultRegulation:  "DPDPA",
			EnabledRegulations: []string{"DPDPA"},
			EnableAI:           true,
		},
	}

	if err := repo.Create(ctx, sysTenant); err != nil {
		return types.ID{}, fmt.Errorf("create system tenant: %w", err)
	}
	logger.Info("Created System tenant", "id", sysTenant.ID)
	return sysTenant.ID, nil
}

func ensurePlatformUser(ctx context.Context, repo identity.UserRepository, db *pgxpool.Pool, tenantID types.ID, roleID types.ID, jwtSecret string, logger *slog.Logger) error {
	const email = "platform@datalens.io"

	u, err := repo.GetByEmailGlobal(ctx, email)
	if err == nil && u != nil {
		// Ensure role is assigned
		hasRole := false
		for _, rid := range u.RoleIDs {
			if rid == roleID {
				hasRole = true
				break
			}
		}
		if !hasRole {
			// Add role
			// This requires manual DB update since UserRepo Update doesn't handle RoleIDs usually (handled by user_roles table)
			// We'll Insert into user_roles manually
			if _, err := db.Exec(ctx, "INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", u.ID, roleID); err != nil {
				return fmt.Errorf("assign role: %w", err)
			}
			logger.Info("Assigned Platform Admin role to existing user")
		} else {
			logger.Info("Platform Admin user already configured")
		}
		return nil
	}

	// Create User
	// We need AuthService to hash password or reproduce hashing logic.
	// Hashing logic is usually bcrypt.
	// Importing golang.org/x/crypto/bcrypt is best.
	// But to avoid dependency issues if not present?
	// AuthService uses bcrypt. Let's restart AuthService here locally.

	// Create a minimal AuthService just for hashing
	// authSvc := service.NewAuthService(repo, nil, jwtSecret, 0, 0, logger, nil)
	// But that requires RoleRepo and AuditService. Excessive.
	// We'll implement basic bcrypt here if possible, or just copy from AuthService.
	// "golang.org/x/crypto/bcrypt" is standard.

	// Actually, easier to invoke AuthService Register if available.
	// We need 'roleRepo' for NewAuthService. We have it.
	// AuditService we can pass nil or mock? It might panic if it calls audit.
	// Let's assume database package has bcrypt? No.
	// We'll invoke AuthService.Register logic manually without the service instance to avoid dependency hell if possible.
	// But `service` package imports `golang.org/x/crypto/bcrypt`.
	// So we can use `service.HashPassword` if exported?
	// It's likely private `hashPassword`.

	// Let's instantiate AuthService properly.
	auditRepo := repository.NewPostgresAuditRepository(db)
	auditSvc := service.NewAuditService(auditRepo, logger)
	roleRepo := repository.NewRoleRepo(db)

	authSvc := service.NewAuthService(repo, roleRepo, jwtSecret, 0, 0, logger, auditSvc)

	// Register
	regInput := service.RegisterInput{
		TenantID: tenantID,
		Email:    email,
		Name:     "Platform Administrator",
		Password: "StrongPass123!", // Hardcoded initial password
	}

	user, err := authSvc.Register(ctx, regInput)
	if err != nil {
		return fmt.Errorf("register user: %w", err)
	}

	// Assign Platform Admin Role
	if _, err := db.Exec(ctx, "INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)", user.ID, roleID); err != nil {
		return fmt.Errorf("assign role: %w", err)
	}

	logger.Info("Created Platform Admin user", "email", email, "password", "StrongPass123!")
	return nil
}
