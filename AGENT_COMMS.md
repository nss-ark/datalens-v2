# Agent Communication Board

**System Instructions**:
- Post your status here when starting or finishing a task.
- If you are blocked, post a message here with `[BLOCKED]` prefix.
- If you need to hand off work to another agent, post a message with `[HANDOFF]` prefix.
- The Orchestrator reads this file at the start of every session.

## Current Sprint Goals (Batch 12: Identity Assurance)
| Goal | Owner | Status | Details |
|------|-------|--------|---------|
| **Identity Core** | Backend | [ ] | IAL Model (Basic/Substantial) & Verification Service. |
| **DigiLocker** | Backend | [ ] | OAuth2 with HMAC, Profile Mapping. |
| **Policy UI** | Frontend | [x] | Admin UI to configure Verification rules. |
| **Portal Identity** | Frontend | [x] | User Profile with Verification Badges. |

## Active Messages
*(Newest on top)*

### [2026-02-12] [FROM: Frontend] → [TO: ALL]
**Subject**: Identity & Verification UI Implementation
**Type**: HANDOFF

**Changes**:
- **Admin**: Identity Settings Page (`/compliance/settings/identity`)
- **Portal**: Identity Card with Verification Level & DigiLocker Flow
- **Services**: `identityService` (Admin), updated `portalService` (User)

**Features Enabled**:
- Admins can toggle DigiLocker/Govt ID requirements.
- Users can view verification status and initiate DigiLocker link.

**Verification**: `npm run build` ✅

**Action Required**:
- **Backend**: Ensure `/public/portal/identity/status` and `/link` endpoints are live.

### [2026-02-12] [FROM: Backend] → [TO: ALL]
**Subject**: Identity Architecture & DigiLocker Integration
**Type**: HANDOFF

**Changes**:
- **Domain**: `IdentityProfile`, `AssuranceLevel`, `IdentityProvider` interface (`internal/domain/identity/identity_profile.go`)
- **Infrastructure**: `DigiLockerProvider` with OAuth2/HMAC (`internal/infrastructure/identity/provider/digilocker.go`)
- **Service**: `IdentityService` implementing IAL upgrade logic (`internal/service/identity_service.go`)
- **API**: `IdentityHandler` with `/status` and `/link` endpoints (`internal/handler/identity_handler.go`)
- **Tests**: Unit tests for `IdentityService` passed.

**API Contracts** (for Frontend agent):
- `GET /api/v2/identity/status` — Response: `{success: true, data: {assurance_level: "NONE|BASIC|SUBSTANTIAL", ...}}`
- `POST /api/v2/identity/link` — Request: `{"provider_name": "DigiLocker", "auth_code": "..."}` — Response: `{success: true, data: {assurance_level: "SUBSTANTIAL"}}`

**Action Required**:
- **Test**: Verify DigiLocker OAuth flow end-to-end with real credentials if available.
- **Frontend**: Implement the "Verify Identity" button in the Data Principal Portal and handle the OAuth redirect/popup.

### [2026-02-12] [FROM: Orchestrator] → [TO: ALL]
**Subject**: Batch 8 Complete — Setup for Batch 9
**Type**: ANNOUNCEMENT

**Status**:
- Batch 8 (Audit, Connectors, Lineage) is **COMPLETE**.
- I am now planning Batch 9.

**Upcoming Focus**:
- **Breach Management**: DPDPA Section 28 Compliance.
- **Microsoft 365**: High-value data source integration.

## Resolved / Archived
*(Move resolved threads here)*

### [2026-02-12] [FROM: Backend] → [TO: ALL]
**Subject**: Enterprise Audit Logging Implementation
**Type**: HANDOFF
... (Archived)

### [2026-02-12] [FROM: Frontend] → [TO: ALL]
**Subject**: Data Lineage Visualization Implementation
**Type**: HANDOFF
... (Archived)

### [2026-02-12] [FROM: Orchestrator] → [TO: ALL]
**Subject**: Batch 8 Started — Enterprise Features
**Type**: ANNOUNCEMENT
... (Archived)

## Resolved / Archived
*(Move resolved threads here)*

### [2026-02-12] [FROM: Backend] → [TO: DevOps, Test]
**Subject**: Database Seeder Tool Ready
**Type**: STATUS_UPDATE
... (Archived)

### [2026-02-12] [FROM: Backend] → [TO: ALL]
**Subject**: E2E Integration Tests Complete
**Type**: STATUS_UPDATE
... (Archived)

### [2026-02-12] [FROM: Frontend] → [TO: ALL]
**Subject**: Frontend Polish & Hardening Complete
**Type**: HANDOFF
... (Archived)

### [2026-02-11] [FROM: Frontend] → [TO: ALL]
**Subject**: Data Principal Portal UI Implementation
**Type**: HANDOFF
... (Archived)
### [2026-02-12] [FROM: Test] → [TO: ALL]
**Subject**: Identity Matrix Tests Complete
**Type**: HANDOFF

**Changes**:
- **Service**: `PolicyEnforcer` (`internal/service/policy_enforcer_test.go`)
- **Tests**: Policy Scenarios (Strict/Flexible) coverage.
- **Tests**: DigiLocker mock parsing verification.

**Results**:
- PolicyEnforcer tests: **PASS**
- DigiLocker Provider tests: **PASS**
- `go build ./...`: **PASS** (Fixed `identity_handler.go` compilation)

**Issues Found**:
- `identity_handler.go` had incorrect `SubjectIDFromContext` usage and ID comparison, causing build failure. **FIXED**.
