// Package service contains unit tests for domain services.
// This file provides in-memory mock implementations of repository
// interfaces and the EventBus, so tests run without infrastructure.
package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/complyark/datalens/internal/domain/governance"
	"github.com/complyark/datalens/internal/domain/identity"
	"github.com/complyark/datalens/pkg/eventbus"
	"github.com/complyark/datalens/pkg/types"
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
func (r *mockTenantRepo) Delete(_ context.Context, id types.ID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.tenants, id)
	return nil
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
