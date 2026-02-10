package repository_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/complyark/datalens/internal/domain/governance"
	"github.com/complyark/datalens/internal/domain/identity"
	"github.com/complyark/datalens/internal/repository"
	"github.com/complyark/datalens/pkg/types"
)

// =============================================================================
// Test Infrastructure — shared Postgres container across all tests
// =============================================================================

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Start Postgres container
	container, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("datalens_test"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("5432/tcp").
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		panic("failed to start postgres container: " + err.Error())
	}
	defer container.Terminate(ctx)

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic("failed to get connection string: " + err.Error())
	}

	testPool, err = pgxpool.New(ctx, connStr)
	if err != nil {
		panic("failed to create pool: " + err.Error())
	}
	defer testPool.Close()

	// Apply migrations
	if err := applyMigrations(ctx, testPool); err != nil {
		panic("failed to apply migrations: " + err.Error())
	}

	os.Exit(m.Run())
}

func applyMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	// Find migrations directory relative to this test file
	_, filename, _, _ := runtime.Caller(0)
	migrationsDir := filepath.Join(filepath.Dir(filename), "..", "..", "migrations")

	files := []string{"001_initial_schema.sql", "002_api_keys.sql", "003_detection_feedback.sql"}
	for _, f := range files {
		sql, err := os.ReadFile(filepath.Join(migrationsDir, f))
		if err != nil {
			return err
		}
		if _, err := pool.Exec(ctx, string(sql)); err != nil {
			return err
		}
	}
	return nil
}

// =============================================================================
// TenantRepo Tests
// =============================================================================

func TestTenantRepo_CRUD(t *testing.T) {
	repo := repository.NewTenantRepo(testPool)
	ctx := context.Background()

	tenant := &identity.Tenant{
		Name:     "IntegrationCo",
		Domain:   "integration.example.com",
		Industry: "technology",
		Country:  "IN",
		Plan:     identity.PlanStarter,
		Status:   identity.TenantActive,
		Settings: identity.TenantSettings{
			DefaultRegulation:  "DPDPA",
			EnabledRegulations: []string{"DPDPA"},
			RetentionDays:      365,
		},
	}

	// Create
	err := repo.Create(ctx, tenant)
	require.NoError(t, err)
	assert.NotEqual(t, types.ID{}, tenant.ID)
	assert.False(t, tenant.CreatedAt.IsZero())

	// GetByID
	got, err := repo.GetByID(ctx, tenant.ID)
	require.NoError(t, err)
	assert.Equal(t, tenant.Name, got.Name)
	assert.Equal(t, tenant.Domain, got.Domain)
	assert.Equal(t, tenant.Industry, got.Industry)

	// GetByDomain
	got, err = repo.GetByDomain(ctx, "integration.example.com")
	require.NoError(t, err)
	assert.Equal(t, tenant.ID, got.ID)

	// Update
	got.Name = "IntegrationCo Updated"
	err = repo.Update(ctx, got)
	require.NoError(t, err)

	got2, _ := repo.GetByID(ctx, tenant.ID)
	assert.Equal(t, "IntegrationCo Updated", got2.Name)

	// Delete (soft)
	err = repo.Delete(ctx, tenant.ID)
	require.NoError(t, err)

	_, err = repo.GetByID(ctx, tenant.ID)
	assert.Error(t, err, "should not find deleted tenant")
}

func TestTenantRepo_GetByID_NotFound(t *testing.T) {
	repo := repository.NewTenantRepo(testPool)
	_, err := repo.GetByID(context.Background(), types.NewID())
	assert.Error(t, err)
}

// =============================================================================
// UserRepo Tests
// =============================================================================

func TestUserRepo_CRUD(t *testing.T) {
	repo := repository.NewUserRepo(testPool)
	tenantRepo := repository.NewTenantRepo(testPool)
	ctx := context.Background()

	// Create a tenant first
	tenant := &identity.Tenant{
		Name:   "UserTestCo",
		Domain: "usertest-" + types.NewID().String()[:8] + ".com",
		Plan:   identity.PlanFree,
		Status: identity.TenantActive,
		Settings: identity.TenantSettings{
			DefaultRegulation: "DPDPA",
		},
	}
	require.NoError(t, tenantRepo.Create(ctx, tenant))

	user := &identity.User{
		Email:      "alice@example.com",
		Name:       "Alice",
		Password:   "$2a$10$hashedpassword",
		Status:     identity.UserActive,
		MFAEnabled: false,
	}
	user.TenantID = tenant.ID

	// Create
	err := repo.Create(ctx, user)
	require.NoError(t, err)
	assert.NotEqual(t, types.ID{}, user.ID)

	// GetByID
	got, err := repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, "Alice", got.Name)
	assert.Equal(t, "alice@example.com", got.Email)

	// GetByEmail
	got, err = repo.GetByEmail(ctx, tenant.ID, "alice@example.com")
	require.NoError(t, err)
	assert.Equal(t, user.ID, got.ID)

	// GetByTenant
	users, err := repo.GetByTenant(ctx, tenant.ID)
	require.NoError(t, err)
	assert.Len(t, users, 1)

	// Update
	now := time.Now().UTC()
	got.Name = "Alice Updated"
	got.LastLoginAt = &now
	err = repo.Update(ctx, got)
	require.NoError(t, err)

	got2, _ := repo.GetByID(ctx, user.ID)
	assert.Equal(t, "Alice Updated", got2.Name)
	assert.NotNil(t, got2.LastLoginAt)

	// Delete
	err = repo.Delete(ctx, user.ID)
	require.NoError(t, err)
	_, err = repo.GetByID(ctx, user.ID)
	assert.Error(t, err)
}

// =============================================================================
// RoleRepo Tests
// =============================================================================

func TestRoleRepo_CRUD(t *testing.T) {
	repo := repository.NewRoleRepo(testPool)
	tenantRepo := repository.NewTenantRepo(testPool)
	ctx := context.Background()

	tenant := &identity.Tenant{
		Name:   "RoleTestCo",
		Domain: "roletest-" + types.NewID().String()[:8] + ".com",
		Plan:   identity.PlanFree,
		Status: identity.TenantActive,
		Settings: identity.TenantSettings{
			DefaultRegulation: "DPDPA",
		},
	}
	require.NoError(t, tenantRepo.Create(ctx, tenant))

	role := &identity.Role{
		TenantID:    &tenant.ID,
		Name:        "Data Analyst",
		Description: "Can view and analyze data",
		Permissions: []identity.Permission{
			{Resource: "DATA_SOURCE", Actions: []string{"READ"}},
			{Resource: "PII", Actions: []string{"READ", "VERIFY"}},
		},
		IsSystem: false,
	}

	// Create
	err := repo.Create(ctx, role)
	require.NoError(t, err)
	assert.NotEqual(t, types.ID{}, role.ID)

	// GetByID
	got, err := repo.GetByID(ctx, role.ID)
	require.NoError(t, err)
	assert.Equal(t, "Data Analyst", got.Name)
	assert.Len(t, got.Permissions, 2)
	assert.Equal(t, "DATA_SOURCE", got.Permissions[0].Resource)

	// GetByTenant
	roles, err := repo.GetByTenant(ctx, tenant.ID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(roles), 1) // at least our custom role

	// Update
	got.Description = "Updated description"
	err = repo.Update(ctx, got)
	require.NoError(t, err)

	got2, _ := repo.GetByID(ctx, role.ID)
	assert.Equal(t, "Updated description", got2.Description)
}

// =============================================================================
// DataSourceRepo Tests
// =============================================================================

func TestDataSourceRepo_CRUD(t *testing.T) {
	dsRepo := repository.NewDataSourceRepo(testPool)
	tenantRepo := repository.NewTenantRepo(testPool)
	ctx := context.Background()

	tenant := &identity.Tenant{
		Name:   "DSTestCo",
		Domain: "dstest-" + types.NewID().String()[:8] + ".com",
		Plan:   identity.PlanFree,
		Status: identity.TenantActive,
		Settings: identity.TenantSettings{
			DefaultRegulation: "DPDPA",
		},
	}
	require.NoError(t, tenantRepo.Create(ctx, tenant))

	ds := &discovery.DataSource{
		Name:        "Test DB",
		Type:        types.DataSourcePostgreSQL,
		Description: "Test database",
		Host:        "localhost",
		Port:        5432,
		Database:    "testdb",
		Status:      discovery.ConnectionStatusConnected,
	}
	ds.TenantID = tenant.ID

	// Create
	err := dsRepo.Create(ctx, ds)
	require.NoError(t, err)
	assert.NotEqual(t, types.ID{}, ds.ID)

	// GetByID
	got, err := dsRepo.GetByID(ctx, ds.ID)
	require.NoError(t, err)
	assert.Equal(t, "Test DB", got.Name)
	assert.Equal(t, types.DataSourcePostgreSQL, got.Type)

	// ListByTenant
	sources, err := dsRepo.GetByTenant(ctx, tenant.ID)
	require.NoError(t, err)
	assert.Len(t, sources, 1)

	// Update
	got.Description = "Updated description"
	err = dsRepo.Update(ctx, got)
	require.NoError(t, err)

	got2, _ := dsRepo.GetByID(ctx, ds.ID)
	assert.Equal(t, "Updated description", got2.Description)

	// Delete
	err = dsRepo.Delete(ctx, ds.ID)
	require.NoError(t, err)
	_, err = dsRepo.GetByID(ctx, ds.ID)
	assert.Error(t, err)
}

// =============================================================================
// PurposeRepo Tests
// =============================================================================

func TestPurposeRepo_CRUD(t *testing.T) {
	purposeRepo := repository.NewPurposeRepo(testPool)
	tenantRepo := repository.NewTenantRepo(testPool)
	ctx := context.Background()

	tenant := &identity.Tenant{
		Name:   "PurposeTestCo",
		Domain: "purposetest-" + types.NewID().String()[:8] + ".com",
		Plan:   identity.PlanFree,
		Status: identity.TenantActive,
		Settings: identity.TenantSettings{
			DefaultRegulation: "DPDPA",
		},
	}
	require.NoError(t, tenantRepo.Create(ctx, tenant))

	p := &governance.Purpose{
		Code:            "MARKETING",
		Name:            "Marketing Communications",
		Description:     "For sending marketing emails",
		LegalBasis:      "CONSENT",
		RetentionDays:   90,
		IsActive:        true,
		RequiresConsent: true,
	}
	p.TenantID = tenant.ID

	// Create
	err := purposeRepo.Create(ctx, p)
	require.NoError(t, err)
	assert.NotEqual(t, types.ID{}, p.ID)

	// GetByID
	got, err := purposeRepo.GetByID(ctx, p.ID)
	require.NoError(t, err)
	assert.Equal(t, "MARKETING", got.Code)
	assert.Equal(t, 90, got.RetentionDays)

	// ListByTenant
	purposes, err := purposeRepo.GetByTenant(ctx, tenant.ID)
	require.NoError(t, err)
	assert.Len(t, purposes, 1)

	// Update
	got.Name = "Marketing Updated"
	err = purposeRepo.Update(ctx, got)
	require.NoError(t, err)

	got2, _ := purposeRepo.GetByID(ctx, p.ID)
	assert.Equal(t, "Marketing Updated", got2.Name)

	// Delete
	err = purposeRepo.Delete(ctx, p.ID)
	require.NoError(t, err)
	_, err = purposeRepo.GetByID(ctx, p.ID)
	assert.Error(t, err)
}

// =============================================================================
// DataInventoryRepo Tests
// =============================================================================

func TestDataInventoryRepo_CRUD(t *testing.T) {
	invRepo := repository.NewDataInventoryRepo(testPool)
	dsRepo := repository.NewDataSourceRepo(testPool)
	tenantRepo := repository.NewTenantRepo(testPool)
	ctx := context.Background()

	// Setup tenant + data source
	tenant := &identity.Tenant{
		Name:   "InvTestCo",
		Domain: "invtest-" + types.NewID().String()[:8] + ".com",
		Plan:   identity.PlanFree,
		Status: identity.TenantActive,
		Settings: identity.TenantSettings{
			DefaultRegulation: "DPDPA",
		},
	}
	require.NoError(t, tenantRepo.Create(ctx, tenant))

	ds := &discovery.DataSource{
		Name:     "Inventory Source",
		Type:     types.DataSourcePostgreSQL,
		Host:     "localhost",
		Port:     5432,
		Database: "testdb",
		Status:   discovery.ConnectionStatusConnected,
	}
	ds.TenantID = tenant.ID
	require.NoError(t, dsRepo.Create(ctx, ds))

	inv := &discovery.DataInventory{
		DataSourceID:   ds.ID,
		TotalEntities:  10,
		TotalFields:    50,
		PIIFieldsCount: 5,
		SchemaVersion:  "v1",
	}

	// Create
	err := invRepo.Create(ctx, inv)
	require.NoError(t, err)
	assert.NotEqual(t, types.ID{}, inv.ID)

	// GetByID
	got, err := invRepo.GetByID(ctx, inv.ID)
	require.NoError(t, err)
	assert.Equal(t, 10, got.TotalEntities)
	assert.Equal(t, 5, got.PIIFieldsCount)

	// GetByDataSource
	got, err = invRepo.GetByDataSource(ctx, ds.ID)
	require.NoError(t, err)
	assert.Equal(t, inv.ID, got.ID)

	// Update
	got.TotalEntities = 20
	err = invRepo.Update(ctx, got)
	require.NoError(t, err)

	got2, _ := invRepo.GetByID(ctx, inv.ID)
	assert.Equal(t, 20, got2.TotalEntities)
}

// =============================================================================
// DataEntityRepo Tests
// =============================================================================

func TestDataEntityRepo_CRUD(t *testing.T) {
	entityRepo := repository.NewDataEntityRepo(testPool)
	invRepo := repository.NewDataInventoryRepo(testPool)
	dsRepo := repository.NewDataSourceRepo(testPool)
	tenantRepo := repository.NewTenantRepo(testPool)
	ctx := context.Background()

	// Setup chain: tenant → data source → inventory
	tenant := &identity.Tenant{
		Name:   "EntityTestCo",
		Domain: "entitytest-" + types.NewID().String()[:8] + ".com",
		Plan:   identity.PlanFree,
		Status: identity.TenantActive,
		Settings: identity.TenantSettings{
			DefaultRegulation: "DPDPA",
		},
	}
	require.NoError(t, tenantRepo.Create(ctx, tenant))

	ds := &discovery.DataSource{
		Name:     "Entity Source",
		Type:     types.DataSourcePostgreSQL,
		Host:     "localhost",
		Port:     5432,
		Database: "testdb",
		Status:   discovery.ConnectionStatusConnected,
	}
	ds.TenantID = tenant.ID
	require.NoError(t, dsRepo.Create(ctx, ds))

	inv := &discovery.DataInventory{
		DataSourceID:  ds.ID,
		SchemaVersion: "v1",
	}
	require.NoError(t, invRepo.Create(ctx, inv))

	entity := &discovery.DataEntity{
		InventoryID:   inv.ID,
		Name:          "users",
		Schema:        "public",
		Type:          discovery.EntityTypeTable,
		PIIConfidence: 0.95,
	}

	// Create
	err := entityRepo.Create(ctx, entity)
	require.NoError(t, err)
	assert.NotEqual(t, types.ID{}, entity.ID)

	// GetByID
	got, err := entityRepo.GetByID(ctx, entity.ID)
	require.NoError(t, err)
	assert.Equal(t, "users", got.Name)
	assert.Equal(t, "public", got.Schema)
	assert.InDelta(t, 0.95, got.PIIConfidence, 0.001)

	// GetByInventory
	entities, err := entityRepo.GetByInventory(ctx, inv.ID)
	require.NoError(t, err)
	assert.Len(t, entities, 1)

	// Update
	got.Name = "customers"
	err = entityRepo.Update(ctx, got)
	require.NoError(t, err)

	got2, _ := entityRepo.GetByID(ctx, entity.ID)
	assert.Equal(t, "customers", got2.Name)

	// Delete
	err = entityRepo.Delete(ctx, entity.ID)
	require.NoError(t, err)
	_, err = entityRepo.GetByID(ctx, entity.ID)
	assert.Error(t, err)
}

// =============================================================================
// DataFieldRepo Tests
// =============================================================================

func TestDataFieldRepo_CRUD(t *testing.T) {
	fieldRepo := repository.NewDataFieldRepo(testPool)
	entityRepo := repository.NewDataEntityRepo(testPool)
	invRepo := repository.NewDataInventoryRepo(testPool)
	dsRepo := repository.NewDataSourceRepo(testPool)
	tenantRepo := repository.NewTenantRepo(testPool)
	ctx := context.Background()

	// Setup full chain
	tenant := &identity.Tenant{
		Name:   "FieldTestCo",
		Domain: "fieldtest-" + types.NewID().String()[:8] + ".com",
		Plan:   identity.PlanFree,
		Status: identity.TenantActive,
		Settings: identity.TenantSettings{
			DefaultRegulation: "DPDPA",
		},
	}
	require.NoError(t, tenantRepo.Create(ctx, tenant))

	ds := &discovery.DataSource{
		Name:     "Field Source",
		Type:     types.DataSourcePostgreSQL,
		Host:     "localhost",
		Port:     5432,
		Database: "testdb",
		Status:   discovery.ConnectionStatusConnected,
	}
	ds.TenantID = tenant.ID
	require.NoError(t, dsRepo.Create(ctx, ds))

	inv := &discovery.DataInventory{
		DataSourceID:  ds.ID,
		SchemaVersion: "v1",
	}
	require.NoError(t, invRepo.Create(ctx, inv))

	entity := &discovery.DataEntity{
		InventoryID: inv.ID,
		Name:        "users",
		Schema:      "public",
		Type:        discovery.EntityTypeTable,
	}
	require.NoError(t, entityRepo.Create(ctx, entity))

	field := &discovery.DataField{
		EntityID:     entity.ID,
		Name:         "email",
		DataType:     "VARCHAR(255)",
		Nullable:     false,
		IsPrimaryKey: false,
		IsForeignKey: false,
	}

	// Create
	err := fieldRepo.Create(ctx, field)
	require.NoError(t, err)
	assert.NotEqual(t, types.ID{}, field.ID)

	// GetByID
	got, err := fieldRepo.GetByID(ctx, field.ID)
	require.NoError(t, err)
	assert.Equal(t, "email", got.Name)
	assert.Equal(t, "VARCHAR(255)", got.DataType)
	assert.False(t, got.Nullable)

	// GetByEntity
	fields, err := fieldRepo.GetByEntity(ctx, entity.ID)
	require.NoError(t, err)
	assert.Len(t, fields, 1)

	// Update
	got.Nullable = true
	err = fieldRepo.Update(ctx, got)
	require.NoError(t, err)

	got2, _ := fieldRepo.GetByID(ctx, field.ID)
	assert.True(t, got2.Nullable)

	// Delete
	err = fieldRepo.Delete(ctx, field.ID)
	require.NoError(t, err)
	_, err = fieldRepo.GetByID(ctx, field.ID)
	assert.Error(t, err)
}

// =============================================================================
// Cross-cutting: Tenant Isolation
// =============================================================================

func TestTenantIsolation_DataSources(t *testing.T) {
	dsRepo := repository.NewDataSourceRepo(testPool)
	tenantRepo := repository.NewTenantRepo(testPool)
	ctx := context.Background()

	// Create two tenants
	t1 := &identity.Tenant{Name: "TenantA", Domain: "a-" + types.NewID().String()[:8] + ".com", Plan: identity.PlanFree, Status: identity.TenantActive, Settings: identity.TenantSettings{DefaultRegulation: "DPDPA"}}
	t2 := &identity.Tenant{Name: "TenantB", Domain: "b-" + types.NewID().String()[:8] + ".com", Plan: identity.PlanFree, Status: identity.TenantActive, Settings: identity.TenantSettings{DefaultRegulation: "DPDPA"}}
	require.NoError(t, tenantRepo.Create(ctx, t1))
	require.NoError(t, tenantRepo.Create(ctx, t2))

	// Create DS in each
	ds1 := &discovery.DataSource{Name: "DS-A", Type: types.DataSourcePostgreSQL, Host: "a", Port: 5432, Database: "a", Status: discovery.ConnectionStatusConnected}
	ds1.TenantID = t1.ID
	ds2 := &discovery.DataSource{Name: "DS-B", Type: types.DataSourceMySQL, Host: "b", Port: 3306, Database: "b", Status: discovery.ConnectionStatusConnected}
	ds2.TenantID = t2.ID
	require.NoError(t, dsRepo.Create(ctx, ds1))
	require.NoError(t, dsRepo.Create(ctx, ds2))

	// TenantA should only see their DS
	list, err := dsRepo.GetByTenant(ctx, t1.ID)
	require.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, "DS-A", list[0].Name)

	// TenantB should only see their DS
	list, err = dsRepo.GetByTenant(ctx, t2.ID)
	require.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, "DS-B", list[0].Name)
}

// =============================================================================
// PIIClassificationRepo Tests
// =============================================================================

func TestPIIClassificationRepo_CRUD(t *testing.T) {
	// Setup Repos
	repo := repository.NewPIIClassificationRepo(testPool)
	dsRepo := repository.NewDataSourceRepo(testPool)
	tenantRepo := repository.NewTenantRepo(testPool)
	userRepo := repository.NewUserRepo(testPool)

	fieldRepo := repository.NewDataFieldRepo(testPool)
	entityRepo := repository.NewDataEntityRepo(testPool)
	invRepo := repository.NewDataInventoryRepo(testPool)
	ctx := context.Background()

	// Setup hierarchy: Tenant -> DS -> Inv -> Entity -> Field
	tenant := &identity.Tenant{
		Name:   "PIITestCo",
		Domain: "piitest-" + types.NewID().String()[:8] + ".com",
		Plan:   identity.PlanFree,
		Status: identity.TenantActive,
		Settings: identity.TenantSettings{
			DefaultRegulation: "DPDPA",
		},
	}
	require.NoError(t, tenantRepo.Create(ctx, tenant))

	// Create User for VerifiedBy
	admin := &identity.User{
		TenantEntity: types.TenantEntity{
			TenantID: tenant.ID,
		},
		Email:      "admin@" + tenant.Domain,
		Name:       "Admin User",
		Password:   "hash",
		Status:     identity.UserActive,
		MFAEnabled: false,
	}
	require.NoError(t, userRepo.Create(ctx, admin))

	ds := &discovery.DataSource{
		Name:     "PII Source",
		Type:     types.DataSourcePostgreSQL,
		Host:     "localhost",
		Port:     5432,
		Database: "testdb",
		Status:   discovery.ConnectionStatusConnected,
	}
	ds.TenantID = tenant.ID
	require.NoError(t, dsRepo.Create(ctx, ds))

	inv := &discovery.DataInventory{
		DataSourceID:  ds.ID,
		SchemaVersion: "v1",
	}
	require.NoError(t, invRepo.Create(ctx, inv))

	entity := &discovery.DataEntity{
		InventoryID: inv.ID,
		Name:        "users",
		Schema:      "public",
		Type:        discovery.EntityTypeTable,
	}
	require.NoError(t, entityRepo.Create(ctx, entity))

	field := &discovery.DataField{
		EntityID:     entity.ID,
		Name:         "email",
		DataType:     "VARCHAR",
		Nullable:     false,
		IsPrimaryKey: false,
		IsForeignKey: false,
	}
	require.NoError(t, fieldRepo.Create(ctx, field))

	// 1. Create Classification
	c := &discovery.PIIClassification{
		FieldID:         field.ID,
		DataSourceID:    ds.ID,
		EntityName:      "users",
		FieldName:       "email",
		Category:        types.PIICategoryContact,
		Type:            types.PIITypeEmail,
		Sensitivity:     types.SensitivityMedium,
		Confidence:      0.95,
		DetectionMethod: types.DetectionMethodAI,
		Status:          types.VerificationPending,
		Reasoning:       "Looks like an email",
	}

	err := repo.Create(ctx, c)
	require.NoError(t, err)
	assert.NotEqual(t, types.ID{}, c.ID)

	// 2. GetByID
	got, err := repo.GetByID(ctx, c.ID)
	require.NoError(t, err)
	assert.Equal(t, "email", got.FieldName)
	assert.Equal(t, types.PIICategoryContact, got.Category)
	assert.Equal(t, types.VerificationPending, got.Status)

	// 3. Update
	got.Status = types.VerificationVerified
	got.VerifiedBy = &admin.ID
	now := time.Now().UTC()
	got.VerifiedAt = &now
	err = repo.Update(ctx, got)
	require.NoError(t, err)

	got2, _ := repo.GetByID(ctx, c.ID)
	assert.Equal(t, types.VerificationVerified, got2.Status)
	assert.Equal(t, &admin.ID, got2.VerifiedBy)

	// 4. GetByDataSource (Pagination)
	page, err := repo.GetByDataSource(ctx, ds.ID, types.Pagination{Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, 1, page.Total)
	assert.Equal(t, c.ID, page.Items[0].ID)

	// 5. Bulk Create
	c2 := discovery.PIIClassification{
		FieldID:         field.ID,
		DataSourceID:    ds.ID,
		EntityName:      "users",
		FieldName:       "phone",
		Category:        types.PIICategoryContact,
		Type:            types.PIITypePhone,
		Sensitivity:     types.SensitivityMedium,
		Confidence:      0.88,
		DetectionMethod: types.DetectionMethodRegex,
		Status:          types.VerificationPending,
		Reasoning:       "Regex match",
	}
	c3 := discovery.PIIClassification{
		FieldID:         field.ID,
		DataSourceID:    ds.ID,
		EntityName:      "users",
		FieldName:       "ip_address",
		Category:        types.PIICategoryBehavioral,
		Type:            types.PIITypeIPAddress,
		Sensitivity:     types.SensitivityLow,
		Confidence:      0.60,
		DetectionMethod: types.DetectionMethodHeuristic,
		Status:          types.VerificationPending,
		Reasoning:       "Heuristic match",
	}

	err = repo.BulkCreate(ctx, []discovery.PIIClassification{c2, c3})
	require.NoError(t, err)

	// 6. GetPending
	// c was verified, so only c2 and c3 are pending
	pending, err := repo.GetPending(ctx, tenant.ID, types.Pagination{Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, 2, pending.Total)
}

// =============================================================================
// DetectionFeedbackRepo Tests
// =============================================================================

func TestFeedbackRepo_CRUD(t *testing.T) {
	fbRepo := repository.NewDetectionFeedbackRepo(testPool)
	piiRepo := repository.NewPIIClassificationRepo(testPool)
	dsRepo := repository.NewDataSourceRepo(testPool)
	tenantRepo := repository.NewTenantRepo(testPool)
	userRepo := repository.NewUserRepo(testPool)
	fieldRepo := repository.NewDataFieldRepo(testPool)
	entityRepo := repository.NewDataEntityRepo(testPool)
	invRepo := repository.NewDataInventoryRepo(testPool)
	ctx := context.Background()

	// 1. Setup Data Hierarchy
	tenant := &identity.Tenant{
		Name:     "FeedbackTestCo",
		Domain:   "fbtest-" + types.NewID().String()[:8] + ".com",
		Status:   identity.TenantActive,
		Settings: identity.TenantSettings{DefaultRegulation: "DPDPA"},
	}
	require.NoError(t, tenantRepo.Create(ctx, tenant))

	admin := &identity.User{
		TenantEntity: types.TenantEntity{
			TenantID: tenant.ID,
		},
		Email:    "admin@" + tenant.Domain,
		Name:     "Admin",
		Password: "hash",
		Status:   identity.UserActive,
	}
	require.NoError(t, userRepo.Create(ctx, admin))

	ds := &discovery.DataSource{
		Name:   "Feedback DB",
		Type:   types.DataSourcePostgreSQL,
		Status: discovery.ConnectionStatusConnected,
	}
	ds.TenantID = tenant.ID
	require.NoError(t, dsRepo.Create(ctx, ds))

	inv := &discovery.DataInventory{DataSourceID: ds.ID}
	require.NoError(t, invRepo.Create(ctx, inv))

	entity := &discovery.DataEntity{InventoryID: inv.ID, Name: "users", Type: discovery.EntityTypeTable}
	require.NoError(t, entityRepo.Create(ctx, entity))

	field := &discovery.DataField{EntityID: entity.ID, Name: "email", DataType: "VARCHAR"}
	require.NoError(t, fieldRepo.Create(ctx, field))

	// 2. Create Classification (Parent of Feedback)
	classification := &discovery.PIIClassification{
		FieldID:         field.ID,
		DataSourceID:    ds.ID,
		EntityName:      "users",
		FieldName:       "email",
		Category:        types.PIICategoryContact,
		Type:            types.PIITypeEmail,
		Confidence:      0.9,
		DetectionMethod: types.DetectionMethodAI,
		Status:          types.VerificationPending,
	}
	require.NoError(t, piiRepo.Create(ctx, classification))

	// 3. Create Feedback (Verify)
	now := time.Now().UTC()
	fb1 := &discovery.DetectionFeedback{
		ClassificationID:   classification.ID,
		TenantID:           tenant.ID,
		FeedbackType:       discovery.FeedbackVerified,
		OriginalCategory:   classification.Category,
		OriginalType:       classification.Type,
		OriginalConfidence: classification.Confidence,
		OriginalMethod:     classification.DetectionMethod,
		CorrectedBy:        admin.ID,
		CorrectedAt:        now,
		Notes:              "Looks good",
		ColumnName:         "email",
		TableName:          "users",
	}
	err := fbRepo.Create(ctx, fb1)
	require.NoError(t, err)
	assert.NotEqual(t, types.ID{}, fb1.ID)

	// 4. GetByID
	got, err := fbRepo.GetByID(ctx, fb1.ID)
	require.NoError(t, err)
	assert.Equal(t, discovery.FeedbackVerified, got.FeedbackType)
	assert.Equal(t, "email", got.ColumnName)

	// 5. Create Feedback 2 (Correction on same classification just for list test)
	// In reality, one classification usually has one feedback, but let's test listing.
	fb2 := &discovery.DetectionFeedback{
		ClassificationID:  classification.ID,
		TenantID:          tenant.ID,
		FeedbackType:      discovery.FeedbackCorrected,
		CorrectedBy:       admin.ID,
		OriginalMethod:    types.DetectionMethodAI,
		CorrectedCategory: &classification.Category, // same for test
		Notes:             "Correction",
	}
	require.NoError(t, fbRepo.Create(ctx, fb2))

	// 6. GetByClassification
	list, err := fbRepo.GetByClassification(ctx, classification.ID)
	require.NoError(t, err)
	assert.Len(t, list, 2) // fb1 and fb2 (ordered by created_at desc)

	// 7. GetByTenant
	page, err := fbRepo.GetByTenant(ctx, tenant.ID, types.Pagination{Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, 2, page.Total)

	// 8. GetAccuracyStats
	stats, err := fbRepo.GetAccuracyStats(ctx, tenant.ID, types.DetectionMethodAI)
	require.NoError(t, err)
	assert.Equal(t, 2, stats.Total) // fb1 (verified) + fb2 (corrected)
	assert.Equal(t, 1, stats.Verified)
	assert.Equal(t, 1, stats.Corrected)
	assert.Equal(t, 0.5, stats.Accuracy) // 1 verified / 2 total
}
