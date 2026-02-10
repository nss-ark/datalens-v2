# 04. DataLens Control Centre

## Overview

The DataLens Control Centre is the **central compliance management platform** hosted in the cloud. It provides the user interface and business logic for managing PII inventory, consent, DSR requests, and compliance reporting.

---

## Application Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                      DATALENS Control Centre                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                         FRONTEND (React)                             │    │
│  │  ┌────────────────────────────────────────────────────────────────┐ │    │
│  │  │  Pages (56+)                                                    │ │    │
│  │  │  ├── Dashboard, Agents, PIIInventory, PIIReviewQueue           │ │    │
│  │  │  ├── DataSources, DataSubjects, DataLineage                    │ │    │
│  │  │  ├── DSRRequests, DSRDetails, ConsentRecords                   │ │    │
│  │  │  ├── Grievances, Nominations, RoPA, Reports                    │ │    │
│  │  │  ├── Departments, ThirdParties, Purposes                       │ │    │
│  │  │  ├── UserManagement, Settings, AuditLogs                       │ │    │
│  │  │  └── DataPrincipalPortal/*, SuperAdmin/*                       │ │    │
│  │  └────────────────────────────────────────────────────────────────┘ │    │
│  │  ┌────────────────────────────────────────────────────────────────┐ │    │
│  │  │  Components (16+)                                               │ │    │
│  │  │  ├── Layout, Sidebar, Navbar, DataTable                        │ │    │
│  │  │  └── Charts, Forms, Modals, Notifications                      │ │    │
│  │  └────────────────────────────────────────────────────────────────┘ │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                     │                                        │
│                                     ▼                                        │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                          BACKEND (Go)                                │    │
│  │                                                                       │    │
│  │  ┌───────────────────────────────────────────────────────────────┐   │   │
│  │  │  HANDLERS (34 files)                                           │   │   │
│  │  │  ├── auth_handler.go          - Authentication                 │   │   │
│  │  │  ├── agent_handler.go         - Agent management               │   │   │
│  │  │  ├── consent_handler.go       - Consent management             │   │   │
│  │  │  ├── dsr_handler.go           - DSR requests                   │   │   │
│  │  │  ├── grievance_handler.go     - Grievance handling             │   │   │
│  │  │  ├── nomination_handler.go    - Nominee management             │   │   │
│  │  │  ├── pii_verification_handler - PII review                     │   │   │
│  │  │  ├── ropa_handler.go          - RoPA generation                │   │   │
│  │  │  ├── data_principal_handler   - Data subject management        │   │   │
│  │  │  ├── department_handler.go    - Department management          │   │   │
│  │  │  ├── third_party_handler.go   - Third-party management         │   │   │
│  │  │  └── super_admin_handler.go   - Platform administration        │   │   │
│  │  └───────────────────────────────────────────────────────────────┘   │   │
│  │                                                                       │    │
│  │  ┌───────────────────────────────────────────────────────────────┐   │   │
│  │  │  SERVICES (35 files)                                           │   │   │
│  │  │  ├── auth_service.go          - Authentication logic           │   │   │
│  │  │  ├── consent_service.go       - Consent business logic         │   │   │
│  │  │  ├── dsr_service.go           - DSR workflow                   │   │   │
│  │  │  ├── grievance_service.go     - Grievance workflow             │   │   │
│  │  │  ├── pii_verification_service - PII verification logic         │   │   │
│  │  │  └── ...                                                       │   │   │
│  │  └───────────────────────────────────────────────────────────────┘   │   │
│  │                                                                       │    │
│  │  ┌───────────────────────────────────────────────────────────────┐   │   │
│  │  │  REPOSITORIES (22 files)                                       │   │   │
│  │  │  Database access layer                                         │   │   │
│  │  └───────────────────────────────────────────────────────────────┘   │   │
│  │                                                                       │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                     │                                        │
│                                     ▼                                        │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                        DATABASE (PostgreSQL)                         │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Directory Structure

```
DataLensApplication/
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go             # Application entry point
│   ├── internal/
│   │   ├── domain/                 # Domain models (21 files)
│   │   ├── dto/                    # Data Transfer Objects (12 files)
│   │   ├── handler/                # HTTP handlers (34 files)
│   │   ├── middleware/             # Auth, CORS, logging (6 files)
│   │   ├── repository/             # Database access (22 files)
│   │   ├── router/                 # Route definitions
│   │   ├── scheduler/              # Background jobs
│   │   └── service/                # Business logic (35 files)
│   ├── migrations/                 # Database migrations (20 files)
│   └── go.mod
│
├── frontend/
│   ├── src/
│   │   ├── pages/                  # UI pages (56+ files)
│   │   ├── components/             # Reusable components (16 files)
│   │   ├── services/               # API clients (25 files)
│   │   ├── hooks/                  # Custom React hooks
│   │   ├── utils/                  # Utility functions
│   │   └── App.tsx
│   ├── package.json
│   └── vite.config.ts
│
└── docker-compose.yml
```

---

## Core Modules

### 1. Agent Management

**Purpose**: Register, monitor, and configure DataLens Agents

| Component | File | Purpose |
|-----------|------|---------|
| Handler | `agent_handler.go` | API endpoints |
| Service | `agent_service.go` | Business logic |
| Repository | `agent_repository.go` | Database operations |
| Page | `Agents.tsx` | Agent listing |
| Page | `AgentDetails.tsx` | Agent configuration |

**Key Features**:
- Agent registration and authentication
- Health monitoring and alerts
- Data source configuration
- Scan scheduling

### 2. PII Inventory & Verification

**Purpose**: Manage discovered PII and human verification workflow

| Component | File | Purpose |
|-----------|------|---------|
| Handler | `pii_verification_handler.go` | API endpoints |
| Service | `pii_verification_service.go` | Verification logic |
| Page | `PIIInventory.tsx` | Full inventory view |
| Page | `PIIReviewQueue.tsx` | Pending verifications |

**Key Features**:
- View all discovered PII across organization
- Verify/reject agent discoveries
- Assign purposes and lawful basis
- Mark sensitivity levels
- Track verification status

### 3. Consent Management

**Purpose**: Capture, track, and manage consent from data subjects

| Component | File | Purpose |
|-----------|------|---------|
| Handler | `consent_handler.go` | API endpoints |
| Service | `consent_service.go` | Consent logic |
| Page | `ConsentRecords.tsx` | Consent history |
| Page | `ConsentAnalytics.tsx` | Consent statistics |

**Key Features**:
- Create consent notices
- Track consent records
- Manage opt-in/opt-out
- Consent analytics dashboard

### 4. DSR Management

**Purpose**: Handle Data Subject Rights requests

| Component | File | Purpose |
|-----------|------|---------|
| Handler | `dsr_handler.go` | API endpoints |
| Service | `dsr_service.go` | DSR workflow |
| Page | `DSRRequests.tsx` | Request listing |
| Page | `DSRDetails.tsx` | Request management |

**Key Features**:
- Create/track DSR requests
- SLA monitoring
- Agent task orchestration
- Response generation

### 5. Grievance Management

**Purpose**: Track and resolve complaints from data subjects

| Component | File | Purpose |
|-----------|------|---------|
| Handler | `grievance_handler.go` | API endpoints |
| Service | `grievance_service.go` | Grievance workflow |
| Page | `Grievances.tsx` | Grievance listing |

**Key Features**:
- Log grievances
- Assign grievance officers
- Track resolution
- Escalation workflow

### 6. Nomination Management

**Purpose**: Handle nominee appointments (for deceased data subjects)

| Component | File | Purpose |
|-----------|------|---------|
| Handler | `nomination_handler.go` | API endpoints |
| Service | `nomination_service.go` | Nomination logic |
| Page | `NominationClaims.tsx` | Nomination requests |

### 7. Purpose Management

**Purpose**: Define processing purposes and lawful bases

| Component | File | Purpose |
|-----------|------|---------|
| Handler | `purpose_handler.go` | API endpoints |
| Service | `purpose_service.go` | Purpose logic |
| Page | `Purposes.tsx` | Purpose definitions |

**Key Features**:
- Define processing purposes
- Set lawful basis for each purpose
- Map PII to purposes

### 8. Recipient Management

**Purpose**: Track internal departments and external third parties

| Component | File | Purpose |
|-----------|------|---------|
| Handler | `department_handler.go` | Department APIs |
| Handler | `third_party_handler.go` | Third-party APIs |
| Page | `DepartmentsManagement.tsx` | Department listing |
| Page | `ThirdParties.tsx` | Third-party listing |

**Key Features**:
- Define departments
- Register third parties
- Map PII sharing

### 9. Compliance Reporting

**Purpose**: Generate DPDPA-required documentation

| Component | File | Purpose |
|-----------|------|---------|
| Handler | `ropa_handler.go` | RoPA APIs |
| Handler | `report_handler.go` | Report APIs |
| Page | `RoPA.tsx` | RoPA generation |
| Page | `Reports.tsx` | Report dashboard |

**Key Features**:
- Records of Processing Activities (RoPA)
- Data inventory reports
- DSR performance reports
- Consent analytics

### 10. User Management

**Purpose**: Manage users and roles

| Component | File | Purpose |
|-----------|------|---------|
| Handler | `auth_handler.go` | Authentication |
| Handler | `user_handler.go` | User management |
| Page | `UserManagement.tsx` | User listing |
| Page | `Login.tsx` | Login page |

**Roles**:
- **Super Admin**: Platform-level access
- **Admin**: Full client access
- **DPO**: Compliance management
- **Analyst**: Read/write for specific modules
- **Viewer**: Read-only access

---

## Frontend Pages Summary

### Main Navigation

| Category | Pages |
|----------|-------|
| **Overview** | Dashboard |
| **Agents** | Agents, AgentDetails, DataSources, ScanHistory |
| **PII** | PIIInventory, PIIReviewQueue, DataLineage |
| **Subjects** | DataSubjects, DataSubjectDetails |
| **Consent** | ConsentRecords, ConsentAnalytics, ConsentNotices |
| **DSR** | DSRRequests, DSRDetails, DSRCreate |
| **Grievance** | Grievances, GrievanceDetails |
| **Nomination** | NominationClaims |
| **Compliance** | RoPA, Reports, AuditLogs |
| **Settings** | Purposes, Departments, ThirdParties, RetentionPolicies |
| **Admin** | UserManagement, Settings |

### Data Principal Portal (12 pages)

| Page | Purpose |
|------|---------|
| PortalDashboard | Landing page |
| SubmitDSR | Submit access/deletion request |
| TrackDSR | Track request status |
| ConsentPreferences | Manage consent |
| VerifyIdentity | Identity verification |

### Super Admin Portal (10 pages)

| Page | Purpose |
|------|---------|
| ClientManagement | Manage tenants |
| SystemHealth | Platform monitoring |
| GlobalSettings | Platform configuration |

---

## API Structure

### Authentication

```
POST /api/auth/login
POST /api/auth/logout
POST /api/auth/refresh
GET  /api/auth/me
```

### Agents

```
GET    /api/agents
POST   /api/agents
GET    /api/agents/:id
PUT    /api/agents/:id
DELETE /api/agents/:id
GET    /api/agents/:id/status
POST   /api/agents/:id/scan
```

### PII

```
GET    /api/pii/inventory
GET    /api/pii/review-queue
POST   /api/pii/verify/:id
POST   /api/pii/reject/:id
```

### DSR

```
GET    /api/dsr
POST   /api/dsr
GET    /api/dsr/:id
PUT    /api/dsr/:id
POST   /api/dsr/:id/verify
POST   /api/dsr/:id/execute
```

### Consent

```
GET    /api/consent/records
POST   /api/consent/records
GET    /api/consent/analytics
```

---

## Multi-Tenancy

The Control Centre supports multiple clients (organizations):

```
┌────────────────────────────────────────────────────────────────┐
│                     DATALENS CONTROL CENTRE                               │
├────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐            │
│  │  Client A   │  │  Client B   │  │  Client C   │   ...      │
│  │  (Tenant 1) │  │  (Tenant 2) │  │  (Tenant 3) │            │
│  │             │  │             │  │             │             │
│  │ - Agents    │  │ - Agents    │  │ - Agents    │             │
│  │ - PII       │  │ - PII       │  │ - PII       │             │
│  │ - Users     │  │ - Users     │  │ - Users     │             │
│  │ - DSRs      │  │ - DSRs      │  │ - DSRs      │             │
│  └─────────────┘  └─────────────┘  └─────────────┘            │
│                                                                 │
│         ↓ Complete data isolation between tenants ↓            │
│                                                                 │
│  ┌─────────────────────────────────────────────────┐           │
│  │              SHARED INFRASTRUCTURE               │           │
│  │  • Database (row-level isolation)               │           │
│  │  • API (tenant context from token)              │           │
│  │  • Frontend (tenant branding)                   │           │
│  └─────────────────────────────────────────────────┘           │
│                                                                 │
└────────────────────────────────────────────────────────────────┘
```

### Tenant Isolation

- Every database query includes `client_id` filter
- JWT tokens contain tenant context
- Middleware enforces tenant boundaries
