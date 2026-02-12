// Package service contains unit tests for domain services.
// This file provides in-memory mock implementations of repository
// interfaces and the EventBus, so tests run without infrastructure.
package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/complyark/datalens/internal/domain/audit"
	"github.com/complyark/datalens/internal/domain/consent"
	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/complyark/datalens/internal/domain/governance"
	"github.com/complyark/datalens/internal/domain/identity"
	"github.com/complyark/datalens/pkg/eventbus"
	"github.com/complyark/datalens/pkg/types"
	"github.com/stretchr/testify/mock"
)

// =============================================================================
// Mock Event Bus
// =============================================================================

type mockEventBus struct {
	mu       sync.Mutex
	Events   []eventbus.Event
	handlers []eventbus.EventHandler
}

func newMockEventBus() *mockEventBus { return &mockEventBus{} }
func (m *mockEventBus) Close() error { return nil }
func (m *mockEventBus) Publish(_ context.Context, e eventbus.Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Events = append(m.Events, e)
	return nil
}
func (m *mockEventBus) Subscribe(_ context.Context, _ string, h eventbus.EventHandler) (eventbus.Subscription, error) {
	m.handlers = append(m.handlers, h)
	return &mockSub{}, nil
}

type mockSub struct{}

func (s *mockSub) Unsubscribe() error { return nil }

// =============================================================================
// Mock User Repository
// =============================================================================

type mockUserRepo struct {
	mu    sync.Mutex
	users map[types.ID]*identity.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[types.ID]*identity.User)}
}

func (r *mockUserRepo) Create(_ context.Context, u *identity.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if u.ID == (types.ID{}) {
		u.ID = types.NewID()
	}
	r.users[u.ID] = u
	return nil
}
func (r *mockUserRepo) GetByID(_ context.Context, id types.ID) (*identity.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.users[id]
	if !ok {
		return nil, fmt.Errorf("user not found")
	}
	return u, nil
}
func (r *mockUserRepo) GetByEmail(_ context.Context, tenantID types.ID, email string) (*identity.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, u := range r.users {
		if u.TenantID == tenantID && u.Email == email {
			return u, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}
func (r *mockUserRepo) GetByTenant(_ context.Context, tenantID types.ID) ([]identity.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []identity.User
	for _, u := range r.users {
		if u.TenantID == tenantID {
			result = append(result, *u)
		}
	}
	return result, nil
}
func (r *mockUserRepo) Update(_ context.Context, u *identity.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[u.ID] = u
	return nil
}
func (r *mockUserRepo) Delete(_ context.Context, id types.ID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.users, id)
	return nil
}

// =============================================================================
// Mock Tenant Repository
// =============================================================================

type mockTenantRepo struct {
	mu      sync.Mutex
	tenants map[types.ID]*identity.Tenant
}

func newMockTenantRepo() *mockTenantRepo {
	return &mockTenantRepo{tenants: make(map[types.ID]*identity.Tenant)}
}

func (r *mockTenantRepo) Create(_ context.Context, t *identity.Tenant) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if t.ID == (types.ID{}) {
		t.ID = types.NewID()
	}
	r.tenants[t.ID] = t
	return nil
}
func (r *mockTenantRepo) GetByID(_ context.Context, id types.ID) (*identity.Tenant, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.tenants[id]
	if !ok {
		return nil, fmt.Errorf("tenant not found")
	}
	return t, nil
}
func (r *mockTenantRepo) GetByDomain(_ context.Context, domain string) (*identity.Tenant, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, t := range r.tenants {
		if t.Domain == domain {
			return t, nil
		}
	}
	return nil, fmt.Errorf("tenant not found")
}
func (r *mockTenantRepo) Update(_ context.Context, t *identity.Tenant) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tenants[t.ID] = t
	return nil
}

// ... (existing code)
func (r *mockTenantRepo) Delete(_ context.Context, id types.ID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.tenants, id)
	return nil
}

func (r *mockTenantRepo) GetAll(_ context.Context) ([]identity.Tenant, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []identity.Tenant
	for _, t := range r.tenants {
		result = append(result, *t)
	}
	return result, nil
}

// =============================================================================
// Mock Role Repository
// =============================================================================

type mockRoleRepo struct {
	mu    sync.Mutex
	roles map[types.ID]*identity.Role
}

func newMockRoleRepo() *mockRoleRepo {
	return &mockRoleRepo{roles: make(map[types.ID]*identity.Role)}
}

func (r *mockRoleRepo) Create(_ context.Context, role *identity.Role) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if role.ID == (types.ID{}) {
		role.ID = types.NewID()
	}
	r.roles[role.ID] = role
	return nil
}
func (r *mockRoleRepo) GetByID(_ context.Context, id types.ID) (*identity.Role, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	role, ok := r.roles[id]
	if !ok {
		return nil, fmt.Errorf("role not found")
	}
	return role, nil
}
func (r *mockRoleRepo) GetSystemRoles(_ context.Context) ([]identity.Role, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []identity.Role
	for _, role := range r.roles {
		if role.IsSystem {
			result = append(result, *role)
		}
	}
	return result, nil
}
func (r *mockRoleRepo) GetByTenant(_ context.Context, tenantID types.ID) ([]identity.Role, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []identity.Role
	for _, role := range r.roles {
		if role.TenantID != nil && *role.TenantID == tenantID {
			result = append(result, *role)
		}
	}
	return result, nil
}
func (r *mockRoleRepo) Update(_ context.Context, role *identity.Role) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.roles[role.ID] = role
	return nil
}

// =============================================================================
// Mock DataSource Repository
// =============================================================================

type mockDataSourceRepo struct {
	mu      sync.Mutex
	sources map[types.ID]*discovery.DataSource
}

func newMockDataSourceRepo() *mockDataSourceRepo {
	return &mockDataSourceRepo{sources: make(map[types.ID]*discovery.DataSource)}
}

func (r *mockDataSourceRepo) Create(_ context.Context, ds *discovery.DataSource) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if ds.ID == (types.ID{}) {
		ds.ID = types.NewID()
	}
	r.sources[ds.ID] = ds
	return nil
}
func (r *mockDataSourceRepo) GetByID(_ context.Context, id types.ID) (*discovery.DataSource, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	ds, ok := r.sources[id]
	if !ok {
		return nil, fmt.Errorf("data source not found")
	}
	return ds, nil
}
func (r *mockDataSourceRepo) GetByTenant(_ context.Context, tenantID types.ID) ([]discovery.DataSource, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []discovery.DataSource
	for _, ds := range r.sources {
		if ds.TenantID == tenantID {
			result = append(result, *ds)
		}
	}
	return result, nil
}
func (r *mockDataSourceRepo) Update(_ context.Context, ds *discovery.DataSource) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sources[ds.ID] = ds
	return nil
}
func (r *mockDataSourceRepo) Delete(_ context.Context, id types.ID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sources, id)
	return nil
}

// =============================================================================
// Mock Purpose Repository
// =============================================================================

type mockPurposeRepo struct {
	mu       sync.Mutex
	purposes map[types.ID]*governance.Purpose
}

func newMockPurposeRepo() *mockPurposeRepo {
	return &mockPurposeRepo{purposes: make(map[types.ID]*governance.Purpose)}
}

func (r *mockPurposeRepo) Create(_ context.Context, p *governance.Purpose) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if p.ID == (types.ID{}) {
		p.ID = types.NewID()
	}
	r.purposes[p.ID] = p
	return nil
}
func (r *mockPurposeRepo) GetByID(_ context.Context, id types.ID) (*governance.Purpose, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.purposes[id]
	if !ok {
		return nil, fmt.Errorf("purpose not found")
	}
	return p, nil
}
func (r *mockPurposeRepo) GetByTenant(_ context.Context, tenantID types.ID) ([]governance.Purpose, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []governance.Purpose
	for _, p := range r.purposes {
		if p.TenantID == tenantID {
			result = append(result, *p)
		}
	}
	return result, nil
}
func (r *mockPurposeRepo) GetByCode(_ context.Context, tenantID types.ID, code string) (*governance.Purpose, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, p := range r.purposes {
		if p.TenantID == tenantID && p.Code == code {
			return p, nil
		}
	}
	return nil, fmt.Errorf("purpose not found")
}
func (r *mockPurposeRepo) Update(_ context.Context, p *governance.Purpose) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.purposes[p.ID] = p
	return nil
}
func (r *mockPurposeRepo) Delete(_ context.Context, id types.ID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.purposes, id)
	return nil
}

// =============================================================================
// Mock PII Classification Repository
// =============================================================================

type mockPIIClassificationRepo struct {
	mu              sync.Mutex
	classifications map[types.ID]*discovery.PIIClassification
}

func newMockPIIClassificationRepo() *mockPIIClassificationRepo {
	return &mockPIIClassificationRepo{
		classifications: make(map[types.ID]*discovery.PIIClassification),
	}
}

func (r *mockPIIClassificationRepo) Create(_ context.Context, c *discovery.PIIClassification) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if c.ID == (types.ID{}) {
		c.ID = types.NewID()
	}
	r.classifications[c.ID] = c
	return nil
}
func (r *mockPIIClassificationRepo) BulkCreate(_ context.Context, classifications []discovery.PIIClassification) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, c := range classifications {
		if c.ID == (types.ID{}) {
			c.ID = types.NewID()
		}
		// Create a copy to avoid pointer issues if needed, but here value receiver on range so it's a copy
		// We need to store pointers though
		val := c
		r.classifications[c.ID] = &val
	}
	return nil
}
func (r *mockPIIClassificationRepo) GetByID(_ context.Context, id types.ID) (*discovery.PIIClassification, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.classifications[id]
	if !ok {
		return nil, fmt.Errorf("classification not found")
	}
	return c, nil
}
func (r *mockPIIClassificationRepo) GetByDataSource(_ context.Context, dsID types.ID, p types.Pagination) (*types.PaginatedResult[discovery.PIIClassification], error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var items []discovery.PIIClassification
	for _, c := range r.classifications {
		if c.DataSourceID == dsID {
			items = append(items, *c)
		}
	}
	return &types.PaginatedResult[discovery.PIIClassification]{Items: items, Total: len(items)}, nil
}
func (r *mockPIIClassificationRepo) Update(_ context.Context, c *discovery.PIIClassification) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.classifications[c.ID] = c
	return nil
}
func (r *mockPIIClassificationRepo) GetPending(_ context.Context, tenantID types.ID, p types.Pagination) (*types.PaginatedResult[discovery.PIIClassification], error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var items []discovery.PIIClassification
	return &types.PaginatedResult[discovery.PIIClassification]{Items: items, Total: 0}, nil
}

func (r *mockPIIClassificationRepo) GetClassifications(_ context.Context, tenantID types.ID, filter discovery.ClassificationFilter) (*types.PaginatedResult[discovery.PIIClassification], error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var items []discovery.PIIClassification
	for _, c := range r.classifications {
		if filter.DataSourceID != nil && c.DataSourceID != *filter.DataSourceID {
			continue
		}
		if filter.Status != nil && c.Status != *filter.Status {
			continue
		}
		items = append(items, *c)
	}
	return &types.PaginatedResult[discovery.PIIClassification]{Items: items, Total: len(items), Page: 1, PageSize: 20, TotalPages: 1}, nil
}

func (r *mockPIIClassificationRepo) GetCounts(_ context.Context, tenantID types.ID) (*discovery.PIICounts, error) {
	return &discovery.PIICounts{Total: 0, ByCategory: make(map[string]int)}, nil
}

// =============================================================================
// Mock Detection Feedback Repository
// =============================================================================

type mockDetectionFeedbackRepo struct {
	mu       sync.Mutex
	feedback map[types.ID]*discovery.DetectionFeedback
}

func newMockDetectionFeedbackRepo() *mockDetectionFeedbackRepo {
	return &mockDetectionFeedbackRepo{
		feedback: make(map[types.ID]*discovery.DetectionFeedback),
	}
}

func (r *mockDetectionFeedbackRepo) Create(_ context.Context, fb *discovery.DetectionFeedback) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if fb.ID == (types.ID{}) {
		fb.ID = types.NewID()
	}
	r.feedback[fb.ID] = fb
	return nil
}
func (r *mockDetectionFeedbackRepo) GetByID(_ context.Context, id types.ID) (*discovery.DetectionFeedback, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	fb, ok := r.feedback[id]
	if !ok {
		return nil, fmt.Errorf("feedback not found")
	}
	return fb, nil
}
func (r *mockDetectionFeedbackRepo) GetByClassification(_ context.Context, cID types.ID) ([]discovery.DetectionFeedback, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []discovery.DetectionFeedback
	for _, fb := range r.feedback {
		if fb.ClassificationID == cID {
			result = append(result, *fb)
		}
	}
	return result, nil
}
func (r *mockDetectionFeedbackRepo) GetByTenant(_ context.Context, tID types.ID, p types.Pagination) (*types.PaginatedResult[discovery.DetectionFeedback], error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var items []discovery.DetectionFeedback
	for _, fb := range r.feedback {
		if fb.TenantID == tID {
			items = append(items, *fb)
		}
	}
	return &types.PaginatedResult[discovery.DetectionFeedback]{Items: items, Total: len(items)}, nil

}
func (r *mockDetectionFeedbackRepo) GetCorrectionPatterns(_ context.Context, tID types.ID, colPattern string) ([]discovery.DetectionFeedback, error) {
	return nil, nil
}
func (r *mockDetectionFeedbackRepo) GetAccuracyStats(_ context.Context, tID types.ID, method types.DetectionMethod) (*discovery.AccuracyStats, error) {
	return &discovery.AccuracyStats{}, nil
}

// =============================================================================
// Mock Data Inventory Repository
// =============================================================================

type mockDataInventoryRepo struct {
	mu          sync.Mutex
	inventories map[types.ID]*discovery.DataInventory
}

func newMockDataInventoryRepo() *mockDataInventoryRepo {
	return &mockDataInventoryRepo{
		inventories: make(map[types.ID]*discovery.DataInventory),
	}
}

func (r *mockDataInventoryRepo) Create(_ context.Context, inv *discovery.DataInventory) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if inv.ID == (types.ID{}) {
		inv.ID = types.NewID()
	}
	r.inventories[inv.ID] = inv
	return nil
}
func (r *mockDataInventoryRepo) GetByID(_ context.Context, id types.ID) (*discovery.DataInventory, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	inv, ok := r.inventories[id]
	if !ok {
		return nil, fmt.Errorf("inventory not found")
	}
	return inv, nil
}
func (r *mockDataInventoryRepo) GetByDataSource(_ context.Context, dsID types.ID) (*discovery.DataInventory, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, inv := range r.inventories {
		if inv.DataSourceID == dsID {
			return inv, nil
		}
	}
	return nil, types.NewNotFoundError("DataInventory", dsID)
}
func (r *mockDataInventoryRepo) Update(_ context.Context, inv *discovery.DataInventory) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.inventories[inv.ID] = inv
	return nil
}
func (r *mockDataInventoryRepo) Delete(_ context.Context, id types.ID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.inventories, id)
	return nil
}

// =============================================================================
// Mock Data Entity Repository
// =============================================================================

type mockDataEntityRepo struct {
	mu       sync.Mutex
	entities map[types.ID]*discovery.DataEntity
}

func newMockDataEntityRepo() *mockDataEntityRepo {
	return &mockDataEntityRepo{
		entities: make(map[types.ID]*discovery.DataEntity),
	}
}

func (r *mockDataEntityRepo) Create(_ context.Context, e *discovery.DataEntity) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if e.ID == (types.ID{}) {
		e.ID = types.NewID()
	}
	r.entities[e.ID] = e
	return nil
}
func (r *mockDataEntityRepo) GetByID(_ context.Context, id types.ID) (*discovery.DataEntity, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	e, ok := r.entities[id]
	if !ok {
		return nil, fmt.Errorf("entity not found")
	}
	return e, nil
}
func (r *mockDataEntityRepo) GetByInventory(_ context.Context, invID types.ID) ([]discovery.DataEntity, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []discovery.DataEntity
	for _, e := range r.entities {
		if e.InventoryID == invID {
			result = append(result, *e)
		}
	}
	return result, nil
}
func (r *mockDataEntityRepo) Update(_ context.Context, e *discovery.DataEntity) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entities[e.ID] = e
	return nil
}
func (r *mockDataEntityRepo) Delete(_ context.Context, id types.ID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.entities, id)
	return nil
}

// =============================================================================
// Mock Data Field Repository
// =============================================================================

type mockDataFieldRepo struct {
	mu     sync.Mutex
	fields map[types.ID]*discovery.DataField
}

func newMockDataFieldRepo() *mockDataFieldRepo {
	return &mockDataFieldRepo{
		fields: make(map[types.ID]*discovery.DataField),
	}
}

func (r *mockDataFieldRepo) Create(_ context.Context, f *discovery.DataField) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if f.ID == (types.ID{}) {
		f.ID = types.NewID()
	}
	r.fields[f.ID] = f
	return nil
}
func (r *mockDataFieldRepo) GetByID(_ context.Context, id types.ID) (*discovery.DataField, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	f, ok := r.fields[id]
	if !ok {
		return nil, fmt.Errorf("field not found")
	}
	return f, nil
}
func (r *mockDataFieldRepo) GetByEntity(_ context.Context, entityID types.ID) ([]discovery.DataField, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []discovery.DataField
	for _, f := range r.fields {
		if f.EntityID == entityID {
			result = append(result, *f)
		}
	}
	return result, nil
}
func (r *mockDataFieldRepo) Update(_ context.Context, f *discovery.DataField) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.fields[f.ID] = f
	return nil
}
func (r *mockDataFieldRepo) Delete(_ context.Context, id types.ID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.fields, id)
	return nil
}

// =============================================================================
// Mock Scan Run Repository
// =============================================================================

type mockScanRunRepo struct {
	mu   sync.Mutex
	runs map[types.ID]*discovery.ScanRun
}

func newMockScanRunRepo() *mockScanRunRepo {
	return &mockScanRunRepo{
		runs: make(map[types.ID]*discovery.ScanRun),
	}
}

func (r *mockScanRunRepo) Create(_ context.Context, run *discovery.ScanRun) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if run.ID == (types.ID{}) {
		run.ID = types.NewID()
	}
	r.runs[run.ID] = run
	return nil
}

func (r *mockScanRunRepo) GetByID(_ context.Context, id types.ID) (*discovery.ScanRun, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	run, ok := r.runs[id]
	if !ok {
		return nil, fmt.Errorf("scan run not found")
	}
	return run, nil
}

func (r *mockScanRunRepo) GetByDataSource(_ context.Context, dataSourceID types.ID) ([]discovery.ScanRun, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []discovery.ScanRun
	for _, run := range r.runs {
		if run.DataSourceID == dataSourceID {
			result = append(result, *run)
		}
	}
	return result, nil
}

func (r *mockScanRunRepo) GetActive(_ context.Context, tenantID types.ID) ([]discovery.ScanRun, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []discovery.ScanRun
	for _, run := range r.runs {
		if run.TenantID == tenantID && (run.Status == discovery.ScanStatusPending || run.Status == discovery.ScanStatusRunning) {
			result = append(result, *run)
		}
	}
	return result, nil
}

func (r *mockScanRunRepo) GetRecent(_ context.Context, tenantID types.ID, limit int) ([]discovery.ScanRun, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []discovery.ScanRun
	count := 0
	for _, run := range r.runs {
		if run.TenantID == tenantID {
			result = append(result, *run)
			count++
			if count >= limit {
				break
			}
		}
	}
	return result, nil
}

func (r *mockScanRunRepo) Update(_ context.Context, run *discovery.ScanRun) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.runs[run.ID] = run
	return nil
}

// =============================================================================
// Mock Audit Repository
// =============================================================================

type mockAuditRepo struct {
	mu   sync.Mutex
	logs []audit.AuditLog
}

func newMockAuditRepo() *mockAuditRepo {
	return &mockAuditRepo{logs: []audit.AuditLog{}}
}

func (r *mockAuditRepo) Create(_ context.Context, log *audit.AuditLog) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if log.ID == (types.ID{}) {
		log.ID = types.NewID()
	}
	r.logs = append(r.logs, *log)
	return nil
}

func (r *mockAuditRepo) GetByTenant(_ context.Context, tenantID types.ID, limit int) ([]audit.AuditLog, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []audit.AuditLog
	count := 0
	for _, l := range r.logs {
		if l.TenantID == tenantID {
			result = append(result, l)
			count++
			if limit > 0 && count >= limit {
				break
			}
		}
	}
	return result, nil
}

// =============================================================================
// Mock Connector (Testify)
// =============================================================================

type MockConnector struct {
	mock.Mock
}

func (m *MockConnector) Connect(ctx context.Context, ds *discovery.DataSource) error {
	args := m.Called(ctx, ds)
	return args.Error(0)
}

func (m *MockConnector) DiscoverSchema(ctx context.Context, input discovery.DiscoveryInput) (*discovery.DataInventory, []discovery.DataEntity, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).(*discovery.DataInventory), args.Get(1).([]discovery.DataEntity), args.Error(2)
}

func (m *MockConnector) GetFields(ctx context.Context, entityID string) ([]discovery.DataField, error) {
	args := m.Called(ctx, entityID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]discovery.DataField), args.Error(1)
}

func (m *MockConnector) SampleData(ctx context.Context, entity, field string, limit int) ([]string, error) {
	args := m.Called(ctx, entity, field, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockConnector) Capabilities() discovery.ConnectorCapabilities {
	args := m.Called()
	return args.Get(0).(discovery.ConnectorCapabilities)
}

func (m *MockConnector) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockConnector) Delete(ctx context.Context, entity string, filter map[string]string) (int64, error) {
	args := m.Called(ctx, entity, filter)
	return int64(args.Int(0)), args.Error(1)
}

func (m *MockConnector) Export(ctx context.Context, entity string, filter map[string]string) ([]map[string]interface{}, error) {
	args := m.Called(ctx, entity, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

// =============================================================================
// Mock Notice Repository
// =============================================================================

type mockNoticeRepo struct {
	mu      sync.Mutex
	notices map[types.ID]*consent.ConsentNotice
}

func newMockNoticeRepo() *mockNoticeRepo {
	return &mockNoticeRepo{
		notices: make(map[types.ID]*consent.ConsentNotice),
	}
}

func (r *mockNoticeRepo) Create(_ context.Context, n *consent.ConsentNotice) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if n.ID == (types.ID{}) {
		n.ID = types.NewID()
	}
	r.notices[n.ID] = n
	return nil
}

func (r *mockNoticeRepo) GetByID(_ context.Context, id types.ID) (*consent.ConsentNotice, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	n, ok := r.notices[id]
	if !ok {
		return nil, fmt.Errorf("notice not found")
	}
	return n, nil
}

func (r *mockNoticeRepo) GetByTenant(_ context.Context, tenantID types.ID) ([]consent.ConsentNotice, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []consent.ConsentNotice
	for _, n := range r.notices {
		if n.TenantID == tenantID {
			result = append(result, *n)
		}
	}
	return result, nil
}

func (r *mockNoticeRepo) Update(_ context.Context, n *consent.ConsentNotice) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.notices[n.ID] = n
	return nil
}

func (r *mockNoticeRepo) Publish(_ context.Context, id types.ID) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	n, ok := r.notices[id]
	if !ok {
		return 0, fmt.Errorf("notice not found")
	}
	n.Version++
	n.Status = consent.NoticeStatusPublished
	return n.Version, nil
}

func (r *mockNoticeRepo) Archive(_ context.Context, id types.ID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	n, ok := r.notices[id]
	if !ok {
		return fmt.Errorf("notice not found")
	}
	n.Status = consent.NoticeStatusArchived
	return nil
}

func (r *mockNoticeRepo) BindToWidgets(_ context.Context, noticeID types.ID, widgetIDs []types.ID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	n, ok := r.notices[noticeID]
	if !ok {
		return fmt.Errorf("notice not found")
	}
	n.WidgetIDs = widgetIDs
	return nil
}

func (r *mockNoticeRepo) AddTranslation(_ context.Context, t *consent.ConsentNoticeTranslation) error {
	return nil // Mock implementation
}

func (r *mockNoticeRepo) GetTranslations(_ context.Context, noticeID types.ID) ([]consent.ConsentNoticeTranslation, error) {
	return nil, nil // Mock implementation
}

func (r *mockNoticeRepo) GetLatestVersion(_ context.Context, seriesID types.ID) (int, error) {
	return 0, nil
}

// =============================================================================
// Mock Renewal Repository
// =============================================================================

type mockRenewalRepo struct {
	mu   sync.Mutex
	logs []consent.ConsentRenewalLog
}

func newMockRenewalRepo() *mockRenewalRepo {
	return &mockRenewalRepo{
		logs: []consent.ConsentRenewalLog{},
	}
}

func (r *mockRenewalRepo) Create(_ context.Context, l *consent.ConsentRenewalLog) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if l.ID == (types.ID{}) {
		l.ID = types.NewID()
	}
	r.logs = append(r.logs, *l)
	return nil
}

func (r *mockRenewalRepo) GetBySubject(_ context.Context, tenantID, subjectID types.ID) ([]consent.ConsentRenewalLog, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []consent.ConsentRenewalLog
	for _, l := range r.logs {
		if l.TenantID == tenantID && l.SubjectID == subjectID {
			result = append(result, l)
		}
	}
	return result, nil
}

func (r *mockRenewalRepo) Update(_ context.Context, l *consent.ConsentRenewalLog) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	found := false
	for i, existing := range r.logs {
		if existing.ID == l.ID {
			r.logs[i] = *l
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("renewal log not found")
	}
	return nil
}

// =============================================================================
// Mock Widget Repository
// =============================================================================

type mockWidgetRepo struct {
	mu      sync.Mutex
	widgets map[types.ID]*consent.ConsentWidget
}

func newMockWidgetRepo() *mockWidgetRepo {
	return &mockWidgetRepo{
		widgets: make(map[types.ID]*consent.ConsentWidget),
	}
}

func (r *mockWidgetRepo) Create(_ context.Context, w *consent.ConsentWidget) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if w.ID == (types.ID{}) {
		w.ID = types.NewID()
	}
	r.widgets[w.ID] = w
	return nil
}

func (r *mockWidgetRepo) GetByID(_ context.Context, id types.ID) (*consent.ConsentWidget, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	w, ok := r.widgets[id]
	if !ok {
		return nil, fmt.Errorf("widget not found")
	}
	return w, nil
}

func (r *mockWidgetRepo) GetByTenant(_ context.Context, tenantID types.ID) ([]consent.ConsentWidget, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []consent.ConsentWidget
	for _, w := range r.widgets {
		if w.TenantID == tenantID {
			result = append(result, *w)
		}
	}
	return result, nil
}

func (r *mockWidgetRepo) GetByAPIKey(_ context.Context, apiKey string) (*consent.ConsentWidget, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, w := range r.widgets {
		if w.APIKey == apiKey {
			return w, nil
		}
	}
	return nil, fmt.Errorf("widget not found")
}

func (r *mockWidgetRepo) Update(_ context.Context, w *consent.ConsentWidget) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.widgets[w.ID] = w
	return nil
}

func (r *mockWidgetRepo) Delete(_ context.Context, id types.ID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.widgets, id)
	return nil
}

// =============================================================================
// Mock Session Repository
// =============================================================================

type mockSessionRepo struct {
	mu       sync.Mutex
	sessions map[types.ID]*consent.ConsentSession
}

func newMockSessionRepo() *mockSessionRepo {
	return &mockSessionRepo{
		sessions: make(map[types.ID]*consent.ConsentSession),
	}
}

func (r *mockSessionRepo) Create(_ context.Context, s *consent.ConsentSession) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if s.ID == (types.ID{}) {
		s.ID = types.NewID()
	}
	r.sessions[s.ID] = s
	return nil
}

func (r *mockSessionRepo) GetBySubject(_ context.Context, tenantID, subjectID types.ID) ([]consent.ConsentSession, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []consent.ConsentSession
	for _, s := range r.sessions {
		if s.TenantID == tenantID && s.SubjectID != nil && *s.SubjectID == subjectID {
			result = append(result, *s)
		}
	}
	return result, nil
}

func (r *mockSessionRepo) GetConversionStats(_ context.Context, tenantID types.ID, from, to time.Time, interval string) ([]consent.ConversionStat, error) {
	return nil, nil // Mock
}

func (r *mockSessionRepo) GetPurposeStats(_ context.Context, tenantID types.ID, from, to time.Time) ([]consent.PurposeStat, error) {
	return nil, nil // Mock
}

func (r *mockSessionRepo) GetExpiringSessions(_ context.Context, withinDays int) ([]consent.ConsentSession, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []consent.ConsentSession
	// now := time.Now().UTC() // Unused
	// This is tricky without widget config access in the repo mock.
	// For testing, we might assume the caller sets "CreatedAt" such that it expires.
	// We'll just return all sessions for the mock and let the service filter?
	// No, the service calls this to GET candidates.
	// In the test, we'll probably rely on the test specifically creating sessions that should be returned.
	// But `GetExpiringSessions` usually generates a SQL query based on expiry date.
	// Here we can't easily simulate "Expiry Date" because it depends on Widget Config (which is in WidgetRepo).
	// So for the MOCK, we will return ALL sessions, or maybe filter by some convention?
	// ACTUALLY: The test `TestExpiryChecker_DetectsExpiringConsent` will set up a session.
	// We can add a helper to the mock to "inject" expected return values for GetExpiringSessions.

	// A better approach for the MOCK: just return all sessions. The Service loop (processSessionExpiry)
	// re-checks the expiry math using the WidgetRepo. So returning extra sessions is fine (inefficient but correct).
	// Wait, the Service:
	// sessions, err := s.sessionRepo.GetExpiringSessions(ctx, 31)
	// for _, session := range sessions { ... processSessionExpiry ... }
	// processSessionExpiry gets the widget and checks expiry.
	// So returning ALL sessions is safe for the mock.
	for _, s := range r.sessions {
		result = append(result, *s)
	}
	return result, nil
}

// Helper to manually seed expiring sessions if needed
func (r *mockSessionRepo) SetSessions(sessions []consent.ConsentSession) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions = make(map[types.ID]*consent.ConsentSession)
	for _, s := range sessions {
		val := s
		r.sessions[s.ID] = &val
	}
}

// =============================================================================
// Mock History Repository
// =============================================================================

type mockHistoryRepo struct {
	mu      sync.Mutex
	entries []consent.ConsentHistoryEntry
}

func newMockHistoryRepo() *mockHistoryRepo {
	return &mockHistoryRepo{
		entries: []consent.ConsentHistoryEntry{},
	}
}

func (r *mockHistoryRepo) Create(_ context.Context, entry *consent.ConsentHistoryEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if entry.ID == (types.ID{}) {
		entry.ID = types.NewID()
	}
	r.entries = append(r.entries, *entry)
	return nil
}

func (r *mockHistoryRepo) GetBySubject(_ context.Context, tenantID, subjectID types.ID, pagination types.Pagination) (*types.PaginatedResult[consent.ConsentHistoryEntry], error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var items []consent.ConsentHistoryEntry
	for _, e := range r.entries {
		if e.TenantID == tenantID && e.SubjectID == subjectID {
			items = append(items, e)
		}
	}
	return &types.PaginatedResult[consent.ConsentHistoryEntry]{Items: items, Total: len(items)}, nil
}

func (r *mockHistoryRepo) GetByPurpose(_ context.Context, tenantID, purposeID types.ID) ([]consent.ConsentHistoryEntry, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []consent.ConsentHistoryEntry
	for _, e := range r.entries {
		if e.TenantID == tenantID && e.PurposeID == purposeID {
			result = append(result, e)
		}
	}
	return result, nil
}

func (r *mockHistoryRepo) GetLatestState(_ context.Context, tenantID, subjectID, purposeID types.ID) (*consent.ConsentHistoryEntry, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var latest *consent.ConsentHistoryEntry
	for _, e := range r.entries {
		if e.TenantID == tenantID && e.SubjectID == subjectID && e.PurposeID == purposeID {
			// Find the one with latest CreatedAt
			if latest == nil || e.CreatedAt.After(latest.CreatedAt) {
				val := e
				latest = &val
			}
		}
	}
	return latest, nil
}
