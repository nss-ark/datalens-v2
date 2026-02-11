# 10. API Reference

## Overview

DataLens provides RESTful APIs for both the Control Centre and Agent components.

---

## Base URLs

| Environment | CONTROL CENTRE URL | Agent URL |
|-------------|----------|-----------|
| Production | `https://api.datalens.complyark.com/v1` | `http://localhost:8080/api` |
| Staging | `https://api-staging.datalens.complyark.com/v1` | `http://localhost:8080/api` |

---

## Authentication

### Control Centre API

```http
POST /auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}

Response:
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "refresh_token_here",
  "expires_in": 3600,
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "role": "ADMIN"
  }
}
```

### Agent API

```http
# All agent endpoints require X-API-Key header
X-API-Key: your-agent-api-key
```

---

## Control Centre API Endpoints

### Authentication

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/auth/login` | User login |
| POST | `/auth/logout` | User logout |
| POST | `/auth/refresh` | Refresh token |
| GET | `/auth/me` | Get current user |
| POST | `/auth/forgot-password` | Request password reset |
| POST | `/auth/reset-password` | Reset password |

### Agents

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/agents` | List all agents |
| POST | `/agents` | Register new agent |
| GET | `/agents/:id` | Get agent details |
| PUT | `/agents/:id` | Update agent |
| DELETE | `/agents/:id` | Delete agent |
| GET | `/agents/:id/status` | Get agent health status |
| POST | `/agents/:id/regenerate-key` | Regenerate API key |

**Example: List Agents**
```http
GET /agents
Authorization: Bearer {token}

Response:
{
  "data": [
    {
      "id": "uuid",
      "name": "HR Agent",
      "status": "ACTIVE",
      "last_heartbeat": "2026-02-10T10:30:00Z",
      "version": "2.1.0",
      "data_sources_count": 3
    }
  ],
  "total": 5,
  "page": 1,
  "page_size": 20
}
```

### Data Sources

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/datasources` | List data sources |
| POST | `/datasources` | Add data source |
| GET | `/datasources/:id` | Get data source |
| PUT | `/datasources/:id` | Update data source |
| DELETE | `/datasources/:id` | Delete data source |
| POST | `/datasources/:id/test` | Test connection |
| POST | `/datasources/:id/scan` | Trigger scan |

### PII Inventory

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/pii/inventory` | Get verified PII inventory |
| GET | `/pii/review-queue` | Get pending verifications |
| GET | `/pii/review-queue/:id` | Get single PII candidate |
| POST | `/pii/review-queue/:id/verify` | Verify as PII |
| POST | `/pii/review-queue/:id/reject` | Reject as false positive |

**Example: Verify PII**
```http
POST /pii/review-queue/uuid/verify
Authorization: Bearer {token}
Content-Type: application/json

{
  "pii_category": "EMAIL_ADDRESS",
  "sensitivity": "MEDIUM",
  "purpose_ids": ["uuid1", "uuid2"],
  "lawful_basis": "CONSENT"
}
```

### DSR Requests

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/dsr` | List DSR requests |
| POST | `/dsr` | Create DSR request |
| GET | `/dsr/:id` | Get DSR details |
| PUT | `/dsr/:id` | Update DSR |
| POST | `/dsr/:id/verify` | Verify identity |
| POST | `/dsr/:id/execute` | Execute DSR |
| POST | `/dsr/:id/complete` | Mark complete |

**Example: Create DSR**
```http
POST /dsr
Authorization: Bearer {token}
Content-Type: application/json

{
  "request_type": "ERASURE",
  "data_subject": {
    "email": "john@example.com",
    "name": "John Doe"
  },
  "notes": "Customer requested account deletion"
}
```

### Consent

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/consent/records` | List consent records |
| POST | `/consent/records` | Record consent |
| GET | `/consent/check` | Check consent status |
| POST | `/consent/withdraw` | Withdraw consent |
| GET | `/consent/analytics` | Consent analytics |

### Notice Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/consent/notices` | List consent notices |
| POST | `/consent/notices` | Create notice (English source) |
| GET | `/consent/notices/:id` | Get notice with translations |
| PUT | `/consent/notices/:id` | Update notice (creates new version) |
| POST | `/consent/notices/:id/publish` | Publish notice version |
| POST | `/consent/notices/:id/archive` | Archive notice |
| POST | `/consent/notices/:id/translate` | Trigger HuggingFace translation for all 22 languages |
| GET | `/consent/notices/:id/translations` | List translations for a notice version |
| PUT | `/consent/notices/:id/translations/:lang` | Override/edit a specific translation |

**Example: Trigger Translation**
```http
POST /consent/notices/uuid/translate
Authorization: Bearer {token}

Response:
{
  "data": {
    "notice_id": "uuid",
    "version": 2,
    "translations_requested": 22,
    "status": "IN_PROGRESS"
  },
  "message": "Translation triggered for 22 languages"
}
```

### Consent Notifications

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/consent/notifications` | List sent notifications |
| GET | `/consent/notifications/templates` | List notification templates |
| POST | `/consent/notifications/templates` | Create notification template |
| PUT | `/consent/notifications/templates/:id` | Update template |
| POST | `/consent/notifications/send` | Manually send notification |

### DigiLocker Integration

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/integrations/digilocker/authorize` | Initiate DigiLocker OAuth 2.0 + PKCE flow |
| POST | `/integrations/digilocker/callback` | Handle OAuth callback, exchange code for token |
| GET | `/integrations/digilocker/user` | Get DigiLocker user details (identity verification) |
| POST | `/integrations/digilocker/push` | Push consent artifact URI to user's DigiLocker |
| GET | `/integrations/digilocker/documents` | List user's issued documents (for KYC/age verification) |

### Grievances

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/grievances` | List grievances |
| POST | `/grievances` | Create grievance |
| GET | `/grievances/:id` | Get grievance |
| PUT | `/grievances/:id` | Update grievance |
| POST | `/grievances/:id/resolve` | Resolve grievance |

### Reports

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/reports` | List reports |
| POST | `/reports/generate` | Generate report |
| GET | `/reports/:id` | Get report |
| GET | `/reports/:id/download` | Download report |

### Users

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/users` | List users |
| POST | `/users` | Create user |
| GET | `/users/:id` | Get user |
| PUT | `/users/:id` | Update user |
| DELETE | `/users/:id` | Delete user |

---

## Agent API Endpoints

### Status & Health

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/status` | Agent status |
| GET | `/health` | Health check |
| GET | `/config` | Get configuration |

### Data Sources

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/datasources` | List data sources |
| POST | `/datasources` | Add data source |
| GET | `/datasources/:id` | Get data source |
| PUT | `/datasources/:id` | Update data source |
| DELETE | `/datasources/:id` | Delete data source |
| POST | `/datasources/:id/test` | Test connection |
| POST | `/datasources/:id/scan` | Start scan |
| GET | `/datasources/:id/tables` | List tables |
| GET | `/datasources/:id/tables/:table/columns` | List columns |

**Example: Add Data Source**
```http
POST /datasources
X-API-Key: {agent-api-key}
Content-Type: application/json

{
  "name": "HR Database",
  "type": "postgresql",
  "connection_details": {
    "host": "hr-db.internal",
    "port": 5432,
    "database": "hr_production",
    "username": "datalens_reader",
    "password": "secure-password",
    "ssl_mode": "require"
  }
}
```

### PII Detection

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/pii/candidates` | List PII candidates |
| GET | `/pii/candidates/:id` | Get candidate |
| POST | `/pii/detect` | Manual detection |

### Data Subjects

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/subjects` | List data subjects |
| GET | `/subjects/:id` | Get subject |
| GET | `/subjects/:id/pii-locations` | Get PII locations |

### Scan Runs

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/scans` | List scan runs |
| GET | `/scans/:id` | Get scan details |
| GET | `/scans/:id/logs` | Get scan logs |

---

## Response Formats

### Success Response

```json
{
  "data": { ... },
  "message": "Success"
}
```

### Error Response

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request",
    "details": [
      {
        "field": "email",
        "message": "Email is required"
      }
    ]
  }
}
```

### Pagination

```json
{
  "data": [ ... ],
  "meta": {
    "total": 100,
    "page": 1,
    "page_size": 20,
    "total_pages": 5
  }
}
```

---

## Error Codes

| Code | Status | Description |
|------|--------|-------------|
| `UNAUTHORIZED` | 401 | Missing/invalid token |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `NOT_FOUND` | 404 | Resource not found |
| `VALIDATION_ERROR` | 400 | Invalid input |
| `CONFLICT` | 409 | Resource already exists |
| `INTERNAL_ERROR` | 500 | Server error |

---

## Rate Limiting

| Endpoint Type | Limit |
|---------------|-------|
| Authentication | 10/minute |
| Read operations | 100/minute |
| Write operations | 30/minute |
| Scan operations | 5/minute |

Headers returned:
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1707552000
```
