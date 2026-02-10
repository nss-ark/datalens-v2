package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/complyark/datalens/internal/domain/identity"
	"github.com/complyark/datalens/pkg/types"
)

// =============================================================================
// TenantRepo
// =============================================================================

// TenantRepo implements identity.TenantRepository.
type TenantRepo struct {
	pool *pgxpool.Pool
}

// NewTenantRepo creates a new TenantRepo.
func NewTenantRepo(pool *pgxpool.Pool) *TenantRepo {
	return &TenantRepo{pool: pool}
}

func (r *TenantRepo) Create(ctx context.Context, t *identity.Tenant) error {
	t.ID = types.NewID()
	query := `
		INSERT INTO tenants (id, name, domain, industry, country, plan, status, settings)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at, updated_at`

	return r.pool.QueryRow(ctx, query,
		t.ID, t.Name, t.Domain, t.Industry, t.Country, t.Plan, t.Status, t.Settings,
	).Scan(&t.CreatedAt, &t.UpdatedAt)
}

func (r *TenantRepo) GetByID(ctx context.Context, id types.ID) (*identity.Tenant, error) {
	query := `
		SELECT id, name, domain, industry, country, plan, status, settings, created_at, updated_at
		FROM tenants
		WHERE id = $1 AND deleted_at IS NULL`

	t := &identity.Tenant{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&t.ID, &t.Name, &t.Domain, &t.Industry, &t.Country, &t.Plan, &t.Status, &t.Settings,
		&t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, types.NewNotFoundError("Tenant", id)
		}
		return nil, fmt.Errorf("get tenant: %w", err)
	}
	return t, nil
}

func (r *TenantRepo) GetByDomain(ctx context.Context, domain string) (*identity.Tenant, error) {
	query := `
		SELECT id, name, domain, industry, country, plan, status, settings, created_at, updated_at
		FROM tenants
		WHERE domain = $1 AND deleted_at IS NULL`

	t := &identity.Tenant{}
	err := r.pool.QueryRow(ctx, query, domain).Scan(
		&t.ID, &t.Name, &t.Domain, &t.Industry, &t.Country, &t.Plan, &t.Status, &t.Settings,
		&t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, types.NewNotFoundError("Tenant", domain)
		}
		return nil, fmt.Errorf("get tenant by domain: %w", err)
	}
	return t, nil
}

func (r *TenantRepo) Update(ctx context.Context, t *identity.Tenant) error {
	query := `
		UPDATE tenants
		SET name = $2, domain = $3, industry = $4, country = $5, plan = $6,
		    status = $7, settings = $8, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING updated_at`

	err := r.pool.QueryRow(ctx, query,
		t.ID, t.Name, t.Domain, t.Industry, t.Country, t.Plan, t.Status, t.Settings,
	).Scan(&t.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return types.NewNotFoundError("Tenant", t.ID)
		}
		return fmt.Errorf("update tenant: %w", err)
	}
	return nil
}

func (r *TenantRepo) Delete(ctx context.Context, id types.ID) error {
	query := `UPDATE tenants SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	ct, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete tenant: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return types.NewNotFoundError("Tenant", id)
	}
	return nil
}

var _ identity.TenantRepository = (*TenantRepo)(nil)

// =============================================================================
// UserRepo
// =============================================================================

// UserRepo implements identity.UserRepository.
type UserRepo struct {
	pool *pgxpool.Pool
}

// NewUserRepo creates a new UserRepo.
func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func (r *UserRepo) Create(ctx context.Context, u *identity.User) error {
	u.ID = types.NewID()
	query := `
		INSERT INTO users (id, tenant_id, email, name, password, status, mfa_enabled)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at`

	return r.pool.QueryRow(ctx, query,
		u.ID, u.TenantID, u.Email, u.Name, u.Password, u.Status, u.MFAEnabled,
	).Scan(&u.CreatedAt, &u.UpdatedAt)
}

func (r *UserRepo) GetByID(ctx context.Context, id types.ID) (*identity.User, error) {
	query := `
		SELECT u.id, u.tenant_id, u.email, u.name, u.password, u.status,
		       u.mfa_enabled, u.last_login_at, u.created_at, u.updated_at,
		       COALESCE(array_agg(ur.role_id) FILTER (WHERE ur.role_id IS NOT NULL), '{}')
		FROM users u
		LEFT JOIN user_roles ur ON ur.user_id = u.id
		WHERE u.id = $1 AND u.deleted_at IS NULL
		GROUP BY u.id`

	u := &identity.User{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.TenantID, &u.Email, &u.Name, &u.Password, &u.Status,
		&u.MFAEnabled, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt, &u.RoleIDs,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, types.NewNotFoundError("User", id)
		}
		return nil, fmt.Errorf("get user: %w", err)
	}
	return u, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, tenantID types.ID, email string) (*identity.User, error) {
	query := `
		SELECT u.id, u.tenant_id, u.email, u.name, u.password, u.status,
		       u.mfa_enabled, u.last_login_at, u.created_at, u.updated_at,
		       COALESCE(array_agg(ur.role_id) FILTER (WHERE ur.role_id IS NOT NULL), '{}')
		FROM users u
		LEFT JOIN user_roles ur ON ur.user_id = u.id
		WHERE u.tenant_id = $1 AND u.email = $2 AND u.deleted_at IS NULL
		GROUP BY u.id`

	u := &identity.User{}
	err := r.pool.QueryRow(ctx, query, tenantID, email).Scan(
		&u.ID, &u.TenantID, &u.Email, &u.Name, &u.Password, &u.Status,
		&u.MFAEnabled, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt, &u.RoleIDs,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, types.NewNotFoundError("User", email)
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return u, nil
}

func (r *UserRepo) GetByTenant(ctx context.Context, tenantID types.ID) ([]identity.User, error) {
	query := `
		SELECT id, tenant_id, email, name, status, mfa_enabled, last_login_at, created_at, updated_at
		FROM users
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY name ASC`

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var results []identity.User
	for rows.Next() {
		var u identity.User
		if err := rows.Scan(
			&u.ID, &u.TenantID, &u.Email, &u.Name, &u.Status,
			&u.MFAEnabled, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		results = append(results, u)
	}
	return results, rows.Err()
}

func (r *UserRepo) Update(ctx context.Context, u *identity.User) error {
	query := `
		UPDATE users
		SET name = $2, status = $3, mfa_enabled = $4, last_login_at = $5, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING updated_at`

	err := r.pool.QueryRow(ctx, query,
		u.ID, u.Name, u.Status, u.MFAEnabled, u.LastLoginAt,
	).Scan(&u.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return types.NewNotFoundError("User", u.ID)
		}
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}

func (r *UserRepo) Delete(ctx context.Context, id types.ID) error {
	query := `UPDATE users SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	ct, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return types.NewNotFoundError("User", id)
	}
	return nil
}

var _ identity.UserRepository = (*UserRepo)(nil)

// =============================================================================
// RoleRepo
// =============================================================================

// RoleRepo implements identity.RoleRepository.
type RoleRepo struct {
	pool *pgxpool.Pool
}

// NewRoleRepo creates a new RoleRepo.
func NewRoleRepo(pool *pgxpool.Pool) *RoleRepo {
	return &RoleRepo{pool: pool}
}

func (r *RoleRepo) Create(ctx context.Context, role *identity.Role) error {
	role.ID = types.NewID()
	query := `
		INSERT INTO roles (id, tenant_id, name, description, permissions, is_system)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at`

	return r.pool.QueryRow(ctx, query,
		role.ID, role.TenantID, role.Name, role.Description, role.Permissions, role.IsSystem,
	).Scan(&role.CreatedAt, &role.UpdatedAt)
}

func (r *RoleRepo) GetByID(ctx context.Context, id types.ID) (*identity.Role, error) {
	query := `
		SELECT id, tenant_id, name, description, permissions, is_system, created_at, updated_at
		FROM roles
		WHERE id = $1`

	role := &identity.Role{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&role.ID, &role.TenantID, &role.Name, &role.Description, &role.Permissions,
		&role.IsSystem, &role.CreatedAt, &role.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, types.NewNotFoundError("Role", id)
		}
		return nil, fmt.Errorf("get role: %w", err)
	}
	return role, nil
}

func (r *RoleRepo) GetSystemRoles(ctx context.Context) ([]identity.Role, error) {
	query := `
		SELECT id, tenant_id, name, description, permissions, is_system, created_at, updated_at
		FROM roles
		WHERE is_system = TRUE
		ORDER BY name ASC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list system roles: %w", err)
	}
	defer rows.Close()

	var results []identity.Role
	for rows.Next() {
		var role identity.Role
		if err := rows.Scan(
			&role.ID, &role.TenantID, &role.Name, &role.Description, &role.Permissions,
			&role.IsSystem, &role.CreatedAt, &role.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan role: %w", err)
		}
		results = append(results, role)
	}
	return results, rows.Err()
}

func (r *RoleRepo) GetByTenant(ctx context.Context, tenantID types.ID) ([]identity.Role, error) {
	query := `
		SELECT id, tenant_id, name, description, permissions, is_system, created_at, updated_at
		FROM roles
		WHERE tenant_id = $1 OR is_system = TRUE
		ORDER BY is_system DESC, name ASC`

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list tenant roles: %w", err)
	}
	defer rows.Close()

	var results []identity.Role
	for rows.Next() {
		var role identity.Role
		if err := rows.Scan(
			&role.ID, &role.TenantID, &role.Name, &role.Description, &role.Permissions,
			&role.IsSystem, &role.CreatedAt, &role.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan role: %w", err)
		}
		results = append(results, role)
	}
	return results, rows.Err()
}

func (r *RoleRepo) Update(ctx context.Context, role *identity.Role) error {
	query := `
		UPDATE roles
		SET name = $2, description = $3, permissions = $4, updated_at = NOW()
		WHERE id = $1 AND is_system = FALSE
		RETURNING updated_at`

	err := r.pool.QueryRow(ctx, query, role.ID, role.Name, role.Description, role.Permissions).
		Scan(&role.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return types.NewNotFoundError("Role", role.ID)
		}
		return fmt.Errorf("update role: %w", err)
	}
	return nil
}

var _ identity.RoleRepository = (*RoleRepo)(nil)
