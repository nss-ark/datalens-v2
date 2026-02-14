# DataLens 2.0 — Context Sync

> **⚠️ ALL AGENTS: Read this file at the start of EVERY session.**
> Last updated: 2026-02-14 (Review Batches R1–R3 Complete)

---

## Architecture Overview (Post R1-R3)

### Frontend: Monorepo (R1)

The frontend is a **4-package npm workspace** under `frontend/packages/`:

| Package | Port | Proxy Target | Domain (dev) |
|---------|------|--------------|--------------|
| `@datalens/shared` | — | — | — |
| `@datalens/control-centre` | 3000 | :8080 | `cc.localhost:8000` |
| `@datalens/admin` | 3001 | :8081 | `admin.localhost:8000` |
| `@datalens/portal` | 3002 | :8082 | `portal.localhost:8000` |

- **Shared imports:** `import { Button, api, useAuthStore } from '@datalens/shared'`
- **App-local imports:** `import { Dashboard } from '@/pages/Dashboard'`
- **`frontend/src/` no longer exists.** All code lives in `packages/`.
- **`frontend/widget/`** (vanilla JS consent SDK) is independent — NOT part of the monorepo.

### Backend: Mode-Based Process Splitting (R2)

Single binary, multiple deployment modes via `--mode` flag:

```
go run cmd/api/main.go --mode=all --port=8080   # Development (default)
go run cmd/api/main.go --mode=cc --port=8080     # CC only
go run cmd/api/main.go --mode=admin --port=8081  # Admin only
go run cmd/api/main.go --mode=portal --port=8082 # Portal only
```

- Route mounting extracted to `cmd/api/routes.go` (4 functions: `mountSharedRoutes`, `mountCCRoutes`, `mountAdminRoutes`, `mountPortalRoutes`)
- Services/handlers conditionally initialized per mode via `shouldInit()` helper
- `/health` response includes `"mode"` field

### Infrastructure: Nginx Reverse Proxy (R3)

- **Dev:** `nginx/dev.conf` runs as Docker service on `:8000`, routes `*.localhost:8000` sub-domains
- **Prod:** `docker-compose.prod.yml` has 3 API instances (`api-cc`, `api-admin`, `api-portal`) + nginx gateway
- **CORS:** Now env-driven via `CORS_ALLOWED_ORIGINS` in `internal/config/config.go` (no more `*`)
- **Start command:** `.\scripts\start-all.ps1` launches backend + 3 frontend apps

---

## Key Files Changed

| File | Change |
|------|--------|
| `frontend/package.json` | Workspace root (`"workspaces": ["packages/*"]`) |
| `frontend/packages/shared/src/index.ts` | Barrel exports for shared lib |
| `frontend/packages/*/vite.config.ts` | Per-app Vite config with correct ports |
| `cmd/api/main.go` | `--mode`/`--port` flags, conditional init |
| `cmd/api/routes.go` | **NEW** — 4 route-mounting functions |
| `nginx/dev.conf` | **NEW** — Dev reverse proxy |
| `internal/config/config.go` | Added `CORSConfig`, `getEnvSlice` |
| `docker-compose.dev.yml` | Added nginx service |
| `docker-compose.prod.yml` | 3 API instances + nginx gateway |
| `scripts/start-all.ps1` | Launches 4 processes |

---

## Rules for All Agents

1. **Frontend Agent:** Always work within `frontend/packages/<app>/`. Import shared code from `@datalens/shared`, not relative `../../shared/` paths.
2. **Backend Agent:** When adding new routes, add them to the correct function in `cmd/api/routes.go`. When adding new services, wrap init in `shouldInit()` in `main.go`.
3. **DevOps Agent:** Backend Dockerfile produces ONE binary. Use `command:` in Docker Compose to set mode. Frontend Dockerfile builds all 3 apps.
4. **Test Agent:** Backend tests should pass with `--mode=all` (default). Frontend builds must pass per-workspace (`npm run build -w @datalens/control-centre`).
5. **All Agents:** CORS origins are in `.env` — don't hardcode them in Go code.
