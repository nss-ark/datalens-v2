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

func (r *TenantRepo) GetAll(ctx context.Context) ([]identity.Tenant, error) {
	query := `
		SELECT id, name, domain, industry, country, plan, status, settings, created_at, updated_at
		FROM tenants
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list all tenants: %w", err)
	}
	defer rows.Close()

	var tenants []identity.Tenant
	for rows.Next() {
		var t identity.Tenant
		if err := rows.Scan(
			&t.ID, &t.Name, &t.Domain, &t.Industry, &t.Country, &t.Plan, &t.Status, &t.Settings,
			&t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan tenant: %w", err)
		}
		tenants = append(tenants, t)
	}
	return tenants, rows.Err()
}

func (r *TenantRepo) Search(ctx context.Context, filter identity.TenantFilter) ([]identity.Tenant, int, error) {
	where := "WHERE deleted_at IS NULL"
	args := []any{}
	argIdx := 1

	if filter.Status != nil {
		where += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, *filter.Status)
		argIdx++
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM tenants %s", where)
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count tenants: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT id, name, domain, industry, country, plan, status, settings, created_at, updated_at
		FROM tenants
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("search tenants: %w", err)
	}
	defer rows.Close()

	var tenants []identity.Tenant
	for rows.Next() {
		var t identity.Tenant
		if err := rows.Scan(
			&t.ID, &t.Name, &t.Domain, &t.Industry, &t.Country, &t.Plan, &t.Status, &t.Settings,
			&t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan tenant: %w", err)
		}
		tenants = append(tenants, t)
	}
	return tenants, total, rows.Err()
}

func (r *TenantRepo) GetStats(ctx context.Context) (*identity.TenantStats, error) {
	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = 'ACTIVE') as active
		FROM tenants
		WHERE deleted_at IS NULL`

	stats := &identity.TenantStats{}
	err := r.pool.QueryRow(ctx, query).Scan(&stats.TotalTenants, &stats.ActiveTenants)
	if err != nil {
		return nil, fmt.Errorf("get tenant stats: %w", err)
	}
	return stats, nil
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

func (r *UserRepo) GetByEmailGlobal(ctx context.Context, email string) (*identity.User, error) {
	// Find the user by email across all tenants (limit to 1 for now)
	query := `
		SELECT u.id, u.tenant_id, u.email, u.name, u.password, u.status,
		       u.mfa_enabled, u.last_login_at, u.created_at, u.updated_at,
		       COALESCE(array_agg(ur.role_id) FILTER (WHERE ur.role_id IS NOT NULL), '{}')
		FROM users u
		LEFT JOIN user_roles ur ON ur.user_id = u.id
		WHERE u.email = $1 AND u.deleted_at IS NULL
		GROUP BY u.id
		LIMIT 1`

	u := &identity.User{}
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&u.ID, &u.TenantID, &u.Email, &u.Name, &u.Password, &u.Status,
		&u.MFAEnabled, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt, &u.RoleIDs,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, types.NewNotFoundError("User", email)
		}
		return nil, fmt.Errorf("get user by email global: %w", err)
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

func (r *UserRepo) CountGlobal(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`
	var count int64
	err := r.pool.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count global users: %w", err)
	}
	return count, nil
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

func (r *UserRepo) SearchGlobal(ctx context.Context, filter identity.UserFilter) ([]identity.User, int, error) {
	where := "WHERE u.deleted_at IS NULL"
	args := []any{}
	argIdx := 1

	if filter.TenantID != nil {
		where += fmt.Sprintf(" AND u.tenant_id = $%d", argIdx)
		args = append(args, *filter.TenantID)
		argIdx++
	}

	if filter.Status != nil {
		where += fmt.Sprintf(" AND u.status = $%d", argIdx)
		args = append(args, *filter.Status)
		argIdx++
	}

	if filter.Search != "" {
		where += fmt.Sprintf(" AND (u.name ILIKE $%d OR u.email ILIKE $%d)", argIdx, argIdx)
		args = append(args, "%"+filter.Search+"%")
		argIdx++
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM users u %s", where)
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count global users: %w", err)
	}

	// Use the same projection as other Get methods
	query := fmt.Sprintf(`
		SELECT u.id, u.tenant_id, u.email, u.name, u.password, u.status,
		       u.mfa_enabled, u.last_login_at, u.created_at, u.updated_at,
		       COALESCE(array_agg(ur.role_id) FILTER (WHERE ur.role_id IS NOT NULL), '{}')
		FROM users u
		LEFT JOIN user_roles ur ON ur.user_id = u.id
		%s
		GROUP BY u.id
		ORDER BY u.created_at DESC
		LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("search global users: %w", err)
	}
	defer rows.Close()

	var results []identity.User
	for rows.Next() {
		var u identity.User
		if err := rows.Scan(
			&u.ID, &u.TenantID, &u.Email, &u.Name, &u.Password, &u.Status,
			&u.MFAEnabled, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt, &u.RoleIDs,
		); err != nil {
			return nil, 0, fmt.Errorf("scan user: %w", err)
		}
		results = append(results, u)
	}
	return results, total, rows.Err()
}

func (r *UserRepo) AssignRoles(ctx context.Context, userID types.ID, roleIDs []types.ID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Delete existing roles
	if _, err := tx.Exec(ctx, "DELETE FROM user_roles WHERE user_id = $1", userID); err != nil {
		return fmt.Errorf("delete existing roles: %w", err)
	}

	// 2. Insert new roles
	if len(roleIDs) > 0 {
		query := "INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)"
		for _, roleID := range roleIDs {
			if _, err := tx.Exec(ctx, query, userID, roleID); err != nil {
				return fmt.Errorf("insert role %s: %w", roleID, err)
			}
		}
	}

	return tx.Commit(ctx)
}

func (r *UserRepo) UpdateStatus(ctx context.Context, id types.ID, status identity.UserStatus) error {
	query := `UPDATE users SET status = $2, updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	ct, err := r.pool.Exec(ctx, query, id, status)
	if err != nil {
		return fmt.Errorf("update user status: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return types.NewNotFoundError("User", id)
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
