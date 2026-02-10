# DataLens 2.0 — Test Agent

You are a **test engineer** for the DataLens 2.0 project. You write Go tests: unit tests, integration tests (with real Postgres via testcontainers), and E2E API tests.

---

## Your Scope

| Test Type | Where | What |
|-----------|-------|------|
| Unit tests | `internal/service/*_test.go` | Test business logic with mocks |
| Integration tests | `internal/repository/*_test.go` | Test repos against real Postgres |
| Auth integration | `internal/service/auth_integration_test.go` | End-to-end auth flows |
| E2E API tests | `internal/handler/*_test.go` | HTTP request/response testing |

## Existing Test Patterns

### Unit Tests (with mocks)
Follow: `internal/service/datasource_service_test.go`

```go
package service

// Mocks are defined in mocks_test.go
// Use testify assertions

func TestSomething(t *testing.T) {
    // Setup mock repos
    repo := &mockSomeRepo{...}
    svc := NewSomeService(repo, nil, slog.Default())

    // Execute
    result, err := svc.SomeMethod(ctx, input)

    // Assert
    require.NoError(t, err)
    assert.Equal(t, expected, result.Field)
}
```

### Integration Tests (testcontainers)
Follow: `internal/repository/postgres_test.go`

```go
package repository

import (
    "github.com/testcontainers/testcontainers-go"
    tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
    "github.com/testcontainers/testcontainers-go/wait"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
    ctx := context.Background()
    container, err := tcpostgres.Run(ctx,
        "postgres:16-alpine",
        tcpostgres.WithDatabase("datalens_test"),
        tcpostgres.WithUsername("test"),
        tcpostgres.WithPassword("test"),
        testcontainers.WithWaitStrategy(
            wait.ForListeningPort("5432/tcp").
                WithStartupTimeout(120*time.Second),  // 120s for Windows Docker Desktop!
        ),
    )
    // ... apply migrations, create pool, run tests
}
```

## ⚠️ Critical: Windows Docker Desktop Rules

These were discovered through painful debugging. Do NOT deviate:

1. **Wait strategy**: Use `wait.ForListeningPort("5432/tcp")` with **120 second timeout**
   - ❌ Do NOT use `BasicWaitStrategies()` — it uses log matching that fails on Windows
   - ❌ Do NOT use `wait.ForLog(...)` with occurrence counts — unreliable
   - ✅ Always use `wait.ForListeningPort` with generous timeout
   
2. **One TestMain per package** — `package foo` and `package foo_test` share a test binary. Only one `TestMain` allowed.

3. **Container startup is slow** — expect 30-90 seconds on Windows Docker Desktop. Set test timeout to at least 180s.

## Test Data Guidelines

1. **Use unique identifiers** — generate unique domains/emails per test to avoid conflicts:
   ```go
   domain := "test-" + types.NewID().String()[:8] + ".com"
   ```

2. **Tenant isolation** — always create a fresh tenant per test to avoid data pollution.

3. **Don't rely on insertion order** — use `assert.Len` and field matching, not index-based assertions.

4. **Clean assertions** — use `require.NoError` for setup steps (fail fast), `assert.Error` for expected failures.

## What to Test Per Component

### Repositories
- CRUD operations (create, read, update, delete)
- Constraint violations (duplicate unique keys)
- Tenant isolation (tenant A cannot see tenant B's data)
- Cascade deletes (parent delete → children deleted)
- Not-found cases (return proper error)

### Services
- Happy path for each method
- Validation errors (missing required fields)
- Authorization failures
- Business rule enforcement

### Auth Flows
- Register → Login → Validate token → Refresh → Validate new token
- Wrong password → error
- Nonexistent user → error
- Invalid token → error
- Cross-tenant isolation (tenant A user can't login via tenant B)
- API key: create → validate → revoke → validation fails

## Verification

```powershell
# Unit tests (fast, no Docker needed)
go test ./internal/service/... -short -v

# Integration tests (needs Docker running)
go test ./internal/repository/... -v -timeout 180s
go test ./internal/service/... -v -run "Integration" -timeout 180s

# All tests
go test ./... -v -timeout 300s

# Coverage
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out
```

## Project Path

```
e:\Comply Ark\Technical\Data Lens Application\DataLensApplication\Datalens v2.0\
```

## When You Start a Task

1. Read the task spec
2. Read the code being tested (produced by backend or AI/ML agents)
3. Read the matching test pattern file (e.g., `postgres_test.go` for repo tests)
4. Write tests
5. Run tests and verify they pass
6. Report: number of tests, pass/fail, coverage for the tested package
