# Agent Communication Log

## SA-1: SuperAdmin Portal Updates (2026-02-19)

### Backend API Changes
- **New SuperAdmin Login**: `POST /api/v2/superadmin/login`
  - Accepts: `{ email, password }`
  - Returns: `{ access_token, refresh_token, expires_in }`
  - Logic: Global login (no tenant_id needed), validates credentials.
  - **Note**: Strict `PLATFORM_ADMIN` role check is deferred to route middleware on protected routes.
- **Me Endpoint**: `GET /api/v2/admin/me`
  - Returns: Authenticated user details (same structure as Shared User entity).
- **Tenant Management**:
  - `GET /api/v2/admin/tenants/{id}`: Fetch tenant details.
  - `PATCH /api/v2/admin/tenants/{id}`: Update tenant details (partial).
- **Wiring**:
  - `AdminHandler` now uses `AuthService` for login delegation.
  - Public routes mounted at `/api/v2/superadmin`.

### Frontend Changes (Admin Package)
- **Service Layer**: Fully rewritten `adminService.ts` to use real API calls via `@datalens/shared/api`. Mocks removed.
- **Authentication**:
  - `Login.tsx` uses `adminService.login` -> `authStore.login`.
  - Auth flow fetches user details from `/api/v2/admin/me` immediately after token acquisition.
- **Routing**:
  - `AdminRoute.tsx` enforces `isAuthenticated` check.
  - **TODO**: Add strict `PLATFORM_ADMIN` role check once role ID constants are available in shared package.
- **UI Renaming**: "Admin Portal" -> "SuperAdmin Portal" standardization.
