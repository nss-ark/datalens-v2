# 12. Security & Compliance

## Overview

DataLens is designed with security-first principles to protect personal data and ensure DPDPA compliance.

---

## Security Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         SECURITY LAYERS                                      │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │  NETWORK SECURITY                                                    │    │
│  │  • TLS 1.3 for all connections                                      │    │
│  │  • mTLS optional for Agent-CONTROL CENTRE                                     │    │
│  │  • Firewall rules                                                   │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │  APPLICATION SECURITY                                                │    │
│  │  • JWT authentication                                               │    │
│  │  • Role-based access control (RBAC)                                 │    │
│  │  • Input validation                                                 │    │
│  │  • Rate limiting                                                    │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │  DATA SECURITY                                                       │    │
│  │  • Encryption at rest (AES-256)                                     │    │
│  │  • Encryption in transit (TLS)                                      │    │
│  │  • Data masking in samples                                          │    │
│  │  • Zero-PII architecture                                            │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │  AUDIT & MONITORING                                                  │    │
│  │  • Complete audit trail                                             │    │
│  │  • Log aggregation                                                  │    │
│  │  • Anomaly detection                                                │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Authentication

### Control Centre

| Method | Use Case |
|--------|----------|
| Username/Password | User login to dashboard |
| JWT Tokens | API authentication |
| Refresh Tokens | Session extension |
| MFA (optional) | Enhanced security |

### Agent Authentication

| Method | Use Case |
|--------|----------|
| API Key | Agent-to-CONTROL CENTRE communication |
| mTLS Certificates | Enhanced agent security |
| Client ID | Tenant identification |

### JWT Token Structure

```json
{
  "sub": "user-uuid",
  "client_id": "client-uuid",
  "role": "ADMIN",
  "permissions": ["pii:read", "pii:verify", "dsr:manage"],
  "iat": 1707552000,
  "exp": 1707555600
}
```

---

## Role-Based Access Control (RBAC)

### User Roles

| Role | Description | Typical User |
|------|-------------|--------------|
| **SUPER_ADMIN** | Platform-level access | ComplyArk staff |
| **ADMIN** | Full client access | IT Admin |
| **DPO** | Compliance management | Data Protection Officer |
| **ANALYST** | Read/write specific modules | Compliance team |
| **VIEWER** | Read-only access | Auditors |

### Permission Matrix

| Permission | Super Admin | Admin | DPO | Analyst | Viewer |
|------------|:-----------:|:-----:|:---:|:-------:|:------:|
| View Dashboard | ✅ | ✅ | ✅ | ✅ | ✅ |
| Manage Agents | ✅ | ✅ | ❌ | ❌ | ❌ |
| Verify PII | ✅ | ✅ | ✅ | ✅ | ❌ |
| Manage DSR | ✅ | ✅ | ✅ | ✅ | ❌ |
| Manage Consent | ✅ | ✅ | ✅ | ✅ | ❌ |
| View Audit Logs | ✅ | ✅ | ✅ | ✅ | ✅ |
| Manage Users | ✅ | ✅ | ❌ | ❌ | ❌ |
| Generate Reports | ✅ | ✅ | ✅ | ✅ | ✅ |
| Configure Settings | ✅ | ✅ | ✅ | ❌ | ❌ |

---

## Encryption

### At Rest

| Data | Encryption | Key Management |
|------|------------|----------------|
| Database | PostgreSQL TDE | Managed keys |
| Connection details | AES-256-GCM | Environment variable |
| Backups | AES-256 | Cloud KMS |
| File uploads | AES-256 | Per-tenant keys |

### In Transit

| Connection | Protocol |
|------------|----------|
| Browser ↔ CONTROL CENTRE | HTTPS (TLS 1.3) |
| Agent ↔ CONTROL CENTRE | HTTPS/gRPC (TLS 1.3) |
| Agent ↔ Agent | gRPC (TLS 1.3) |
| Agent ↔ Data Sources | Database-specific TLS |

### Credential Encryption

```go
// Agent encrypts connection details before storage
type EncryptionService struct {
    key []byte // 32 bytes for AES-256
}

func (e *EncryptionService) Encrypt(plaintext []byte) ([]byte, error) {
    // AES-256-GCM encryption
}

func (e *EncryptionService) Decrypt(ciphertext []byte) ([]byte, error) {
    // AES-256-GCM decryption
}
```

---

## Zero-PII Architecture

### What Never Leaves Client Infrastructure

| Data | Stored Where |
|------|--------------|
| Actual personal data | Client databases |
| Full email addresses | Client systems |
| Phone numbers | Client systems |
| Aadhaar/PAN numbers | Client systems |
| File contents | Client file systems |

### What Gets Sent to CONTROL CENTRE

| Data | Purpose |
|------|---------|
| Object identifiers | e.g., "hr.employees.email" |
| PII categories | e.g., "EMAIL_ADDRESS" |
| Confidence scores | e.g., 0.95 |
| Record counts | e.g., 1500 records |
| Masked samples | e.g., "j***@e***.com" |

---

## Audit Trail

### What Gets Logged

| Category | Events |
|----------|--------|
| Authentication | Login, logout, failed attempts |
| User Management | Create, update, delete users |
| PII | View, verify, reject |
| DSR | Create, verify, execute |
| Consent | Grant, withdraw |
| Configuration | Settings changes |
| Agent | Connect, disconnect, scan |

### Log Entry Structure

```json
{
  "id": "uuid",
  "timestamp": "2026-02-10T10:30:00Z",
  "client_id": "uuid",
  "user_id": "uuid",
  "action": "PII_VERIFIED",
  "resource_type": "PII_DISCOVERY",
  "resource_id": "uuid",
  "details": {
    "old_status": "PENDING",
    "new_status": "VERIFIED",
    "pii_category": "EMAIL_ADDRESS"
  },
  "ip_address": "192.168.1.100",
  "user_agent": "Mozilla/5.0..."
}
```

### Log Retention

| Log Type | Retention |
|----------|-----------|
| Security events | 2 years |
| DSR activity | 5 years |
| PII verification | 5 years |
| System logs | 90 days |

---

## DPDPA Compliance Features

### Section 5: Consent

| Requirement | Implementation |
|-------------|----------------|
| Free consent | No dark patterns, clear UI |
| Informed | Detailed notices in 22 Eighth Schedule languages |
| Specific | Per-purpose consent |
| Unambiguous | Explicit actions required |
| Withdrawable | Easy withdrawal mechanism |

### Section 8: Children's Data

| Requirement | Implementation |
|-------------|----------------|
| Identify minors | Age verification (DigiLocker integration) |
| Guardian consent | Parent/guardian OTP approval flow |
| Special handling | Enhanced restrictions |

### Section 11-13: Data Subject Rights

| Right | Implementation |
|-------|----------------|
| Access | DSR ACCESS request type |
| Correction | DSR RECTIFICATION type |
| Erasure | DSR ERASURE type |
| Portability | Data export feature |

### Section 14: Grievance Redressal

| Requirement | Implementation |
|-------------|----------------|
| Grievance mechanism | Dedicated module |
| Acknowledgment | Automatic confirmation |
| Resolution tracking | SLA monitoring |
| Escalation | Built-in workflow |

### Section 8(i): Nomination

| Requirement | Implementation |
|-------------|----------------|
| Nominate person | Nomination management |
| Death handling | Nominee can exercise rights |

---

## MeITY BRD Compliance Matrix

Mapping each BRD section to DataLens features:

| BRD Section | Requirement | DataLens Feature | Status |
|-------------|-------------|------------------|--------|
| § 4.1.1 | Consent Collection | Embeddable widget with per-purpose toggles | ✅ Documented |
| § 4.1.2 | Consent Validation | Digital signature (SHA-256), immutable records | ✅ Documented |
| § 4.1.3 | Consent Update | Notice versioning, re-consent on purpose change | ✅ Documented |
| § 4.1.4 | Consent Renewal | Proactive expiry reminders, renewal workflow | ✅ Documented |
| § 4.1.5 | Consent Withdrawal | Portal + API withdrawal, cascading notifications | ✅ Documented |
| § 4.3 | User Dashboard | Data Principal Portal (consent history, DPR, appeal) | ✅ Documented |
| § 4.4 | Consent Notifications | DP/DF/Processor alerts via Email, SMS, Webhook | ✅ Documented |
| § 4.5 | Grievance Redressal | Dedicated grievance module with SLA tracking | ✅ Documented |
| § 4.6 | System Administration | RBAC (5 roles), MFA, tenant management | ✅ Documented |
| § 4.7 | Logging & Audit | Immutable audit logs with hash chain, tamper-proof | ✅ Documented |

---

## WCAG 2.1 Compliance (BRD Requirement)

Consent UIs must meet WCAG 2.1 Level AA accessibility standards:

| Principle | Requirements |
|-----------|-------------|
| **Perceivable** | Text alternatives, captions, color contrast ≥ 4.5:1, resizable text |
| **Operable** | Keyboard navigable, no time limits on consent decisions, clear focus indicators |
| **Understandable** | Readable language, predictable navigation, input assistance |
| **Robust** | Compatible with assistive technologies, valid semantic HTML |

---

## Immutable Audit Logging (BRD § 4.7)

All consent-related actions are logged as tamper-proof, hash-chained records:

| Requirement | Implementation |
|-------------|----------------|
| Comprehensive | Every consent action logged (grant, deny, withdraw, renew, update) |
| Tamper-proof | Hash chain: each record includes SHA-256 of previous record |
| Immutable | Append-only storage, no updates or deletes permitted |
| Queryable | Full-text search by subject, purpose, date range, action type |
| Exportable | CSV/JSON export for regulatory audits |
| Retention | Configurable per regulation (DPDPA: minimum 5 years) |

---

## Security Controls Checklist

### Infrastructure

- [ ] TLS 1.3 everywhere
- [ ] Firewall configured
- [ ] DDoS protection
- [ ] Regular patching
- [ ] Intrusion detection

### Application

- [ ] Input validation
- [ ] Output encoding
- [ ] CSRF protection
- [ ] Rate limiting
- [ ] Secure headers

### Data

- [ ] Encryption at rest
- [ ] Encryption in transit
- [ ] Key rotation
- [ ] Backup encryption
- [ ] Secure deletion

### Access

- [ ] Strong passwords
- [ ] MFA enabled
- [ ] Role-based access
- [ ] Session timeout
- [ ] Principle of least privilege

### Monitoring

- [ ] Audit logging
- [ ] Log aggregation
- [ ] Alerting
- [ ] Incident response
- [ ] Regular audits
