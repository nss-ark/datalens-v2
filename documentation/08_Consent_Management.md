# 08. Consent Management

## Overview

Consent Management handles the capture, tracking, and management of consent from data subjects (data principals). It provides an embeddable Consent Management System (CMS) that companies deploy on their digital touchpoints, a self-service Data Principal Portal, and an immutable consent log stored in the Control Centre.

---

## Consent Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         CONSENT MANAGEMENT FLOW                              │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│   CAPTURE (Embed)          IMMUTABLE LOG            PORTAL (Self-Service)  │
│        │                       │                         │                  │
│  ┌─────▼──────────┐  ┌────────▼─────────────┐  ┌───────▼──────────┐      │
│  │ JS Snippet     │──►│  Control Centre      │◄─│ Data Principal   │      │
│  │ or Iframe       │  │  Consent Records     │  │ Portal           │      │
│  │                 │  │                       │  │                  │      │
│  │ • Banner        │  │  ┌─────────────────┐  │  │ • View consents  │      │
│  │ • Pref Center   │  │  │ subject_id      │  │  │ • Consent history│      │
│  │ • Inline Form   │  │  │ purpose_id      │  │  │ • Submit DPR     │      │
│  │ • Full Portal   │  │  │ consent_status  │  │  │ • Track requests │      │
│  │                 │  │  │ signature ✓     │  │  │ • Appeal         │      │
│  └─────────────────┘  │  └─────────────────┘  │  └──────────────────┘      │
│                        └───────────────────────┘                             │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Embeddable Consent Widget (CMS)

### Integration Methods

Companies integrate the consent collection system via two lightweight mechanisms:

#### 1. JavaScript Snippet (Banner / Preference Center / Inline Form)

```html
<!-- Paste in <head> or before </body> -->
<script src="https://cdn.datalens.io/consent.min.js"
        data-widget-id="wdg_abc123"
        data-api-key="pk_live_xxx"></script>
```

**Size**: ~15 KB gzipped. No dependencies. Framework-agnostic.

#### 2. Iframe (Full Data Principal Portal)

```html
<iframe src="https://portal.datalens.io/t/{tenant-slug}/portal"
        width="100%" height="600" frameborder="0"
        allow="clipboard-write"></iframe>
```

**Priority**: Iframe is first priority as it's more lightweight to implement.

### Widget Types

| Type | Use Case | Layout |
|------|----------|--------|
| **BANNER** | Cookie consent on websites | Bottom bar, top bar, or modal |
| **PREFERENCE_CENTER** | Granular purpose toggles | Modal or sidebar |
| **INLINE_FORM** | Embedded in registration/checkout | Inline within page |
| **PORTAL** | Full Data Principal Portal | Full page (iframe) |

### Widget Configuration

Companies configure widgets through the Control Centre:

```yaml
widget:
  name: "Website Consent Banner"
  type: BANNER
  domain: "*.acme.com"
  
  config:
    # Visual
    theme:
      primary_color: "#1a73e8"
      background_color: "#ffffff"
      text_color: "#333333"
      font_family: "Inter, sans-serif"
      logo_url: "https://acme.com/logo.svg"
      border_radius: "8px"
    layout: BOTTOM_BAR  # BOTTOM_BAR, TOP_BAR, MODAL, SIDEBAR
    
    # Behavior
    purpose_ids: ["marketing", "analytics", "partners"]
    default_state: "OPT_OUT"       # DPDPA requires explicit opt-in
    granular_toggle: true           # Per-purpose toggle switches
    block_until_consent: false      # Don't block page access
    
    # Content
    languages: ["en", "hi"]
    default_language: "en"
    translations:
      en:
        title: "We value your privacy"
        description: "We use cookies to improve your experience."
        accept_all: "Accept All"
        reject_all: "Reject All"
        customize: "Customize"
      hi:
        title: "हम आपकी गोपनीयता का सम्मान करते हैं"
        description: "हम आपके अनुभव को बेहतर बनाने के लिए कुकीज़ का उपयोग करते हैं।"
    
    # Compliance
    regulation_ref: "DPDPA"
    require_explicit: true
    consent_expiry_days: 365
```

---

## Consent Components

### 1. Purpose Definitions

Before collecting consent, organizations must define processing purposes:

| Purpose | Lawful Basis | Description |
|---------|--------------|-------------|
| Marketing | Consent | Sending promotional emails |
| Service Delivery | Contractual | Providing purchased services |
| HR Operations | Legitimate Interest | Managing employees |
| Analytics | Consent | Usage analytics |
| Legal Compliance | Legal Obligation | Tax records, audits |

### 2. Consent Notice

The notice presented to data subjects:

```yaml
consent_notice:
  title: "Data Processing Consent"
  description: "We collect and process your personal data for the following purposes."
  purposes:
    - id: "marketing"
      name: "Marketing Communications"
      description: "Send promotional offers and newsletters"
      required: false
      default: false
    - id: "analytics"
      name: "Analytics"
      description: "Understand how you use our services"
      required: false
      default: false   # Must be opt-in under DPDPA
    - id: "service"
      name: "Service Delivery"
      description: "Provide the services you requested"
      required: true
      default: true
```

### 3. Consent Record (Immutable)

Every consent interaction is stored as a signed, immutable record:

```json
{
  "id": "uuid",
  "tenant_id": "uuid",
  "subject_id": "uuid",
  "widget_id": "wdg_abc123",
  "decisions": [
    { "purpose_id": "marketing", "granted": false },
    { "purpose_id": "analytics", "granted": true }
  ],
  "ip_address": "192.168.1.1",
  "user_agent": "Mozilla/5.0...",
  "page_url": "https://acme.com/signup",
  "notice_version": "1.2",
  "widget_version": 3,
  "signature": "sha256:abc123...",
  "created_at": "2026-02-10T10:30:00Z"
}
```

---

## Consent History Timeline

Every consent state change is recorded as a `ConsentHistoryEntry`, forming an immutable, chronological timeline viewable by both the company (in the Control Centre) and the data principal (in the Portal).

```
┌─────────────────────────────────────────────────────────────────────┐
│  Consent History for: user@example.com                               │
├─────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  📅 2026-02-10  Marketing Communications    GRANTED → WITHDRAWN     │
│     Source: PORTAL   •   Notice v1.2   •   Signature: ✓             │
│                                                                       │
│  📅 2026-01-15  Analytics                   — → GRANTED              │
│     Source: BANNER   •   Notice v1.2   •   Signature: ✓             │
│                                                                       │
│  📅 2026-01-15  Marketing Communications    — → GRANTED              │
│     Source: BANNER   •   Notice v1.2   •   Signature: ✓             │
│                                                                       │
│  📅 2026-01-15  Service Delivery            — → GRANTED              │
│     Source: BANNER   •   Notice v1.0   •   Signature: ✓             │
│                                                                       │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Consent Database Schema

```sql
-- Consent widgets (embeddable configurations)
consent_widgets (
  id                UUID PRIMARY KEY,
  tenant_id         UUID REFERENCES tenants(id),
  name              VARCHAR(255),
  type              VARCHAR(50),
  domain            VARCHAR(255),
  status            VARCHAR(20),
  config            JSONB,
  embed_code        TEXT,
  api_key           VARCHAR(255) UNIQUE,
  allowed_origins   TEXT[],
  version           INT DEFAULT 1,
  created_at        TIMESTAMP,
  updated_at        TIMESTAMP
)

-- Consent sessions (immutable log)
consent_sessions (
  id                UUID PRIMARY KEY,
  tenant_id         UUID REFERENCES tenants(id),
  widget_id         UUID REFERENCES consent_widgets(id),
  subject_id        UUID,
  decisions         JSONB,
  ip_address        INET,
  user_agent        TEXT,
  page_url          TEXT,
  widget_version    INT,
  notice_version    VARCHAR(20),
  signature         TEXT NOT NULL,
  created_at        TIMESTAMP
)

-- Consent history (immutable timeline)
consent_history (
  id                UUID PRIMARY KEY,
  tenant_id         UUID REFERENCES tenants(id),
  subject_id        UUID NOT NULL,
  widget_id         UUID REFERENCES consent_widgets(id),
  purpose_id        UUID NOT NULL,
  purpose_name      VARCHAR(255),
  previous_status   VARCHAR(20),
  new_status        VARCHAR(20) NOT NULL,
  source            VARCHAR(50),
  ip_address        INET,
  user_agent        TEXT,
  notice_version    VARCHAR(20),
  signature         TEXT NOT NULL,
  created_at        TIMESTAMP
)

-- Data principal profiles (portal identity)
data_principal_profiles (
  id                UUID PRIMARY KEY,
  tenant_id         UUID REFERENCES tenants(id),
  email             VARCHAR(255) NOT NULL,
  phone             VARCHAR(50),
  verification_status VARCHAR(20),
  verified_at       TIMESTAMP,
  verification_method VARCHAR(50),
  subject_id        UUID,
  last_access_at    TIMESTAMP,
  preferred_lang    VARCHAR(10) DEFAULT 'en',
  created_at        TIMESTAMP,
  updated_at        TIMESTAMP,
  UNIQUE(tenant_id, email)
)

-- DPR requests (Data Principal Rights)
dpr_requests (
  id                UUID PRIMARY KEY,
  tenant_id         UUID REFERENCES tenants(id),
  profile_id        UUID REFERENCES data_principal_profiles(id),
  dsr_id            UUID,
  type              VARCHAR(50) NOT NULL,
  description       TEXT,
  status            VARCHAR(30),
  submitted_at      TIMESTAMP NOT NULL,
  deadline          TIMESTAMP,
  verified_at       TIMESTAMP,
  verification_ref  VARCHAR(255),
  is_minor          BOOLEAN DEFAULT false,
  guardian_name     VARCHAR(255),
  guardian_email    VARCHAR(255),
  guardian_relation VARCHAR(50),
  guardian_verified BOOLEAN DEFAULT false,
  completed_at      TIMESTAMP,
  response_summary  TEXT,
  download_url      TEXT,
  appeal_of         UUID REFERENCES dpr_requests(id),
  appeal_reason     TEXT,
  is_escalated      BOOLEAN DEFAULT false,
  escalated_to      VARCHAR(255),
  created_at        TIMESTAMP,
  updated_at        TIMESTAMP
)
```

---

## Consent Status Flow

```
                   GRANTED
                      │
         ┌────────────┼────────────┐
         │            │            │
         ▼            ▼            ▼
      ACTIVE      EXPIRED     WITHDRAWN
         │            │            │
         │            └────────────┘
         │                   │
         ▼                   ▼
   [Processing OK]    [No Processing]
```

| Status | Description | Can Process? |
|--------|-------------|--------------|
| `GRANTED` | Consent given, not expired | ✅ Yes |
| `DENIED` | Consent explicitly refused | ❌ No |
| `WITHDRAWN` | Consent revoked by subject | ❌ No |
| `EXPIRED` | Consent past expiration date | ❌ No |
| `PENDING` | Awaiting subject response | ❌ No |

---

## Data Principal Portal

The Data Principal Portal is a self-service interface for data principals (data subjects) to manage their consent and exercise their rights. It is served as a standalone page and embeddable via iframe.

### Portal Pages

```
┌──────────────────────────────────────────────────────────────────────┐
│  Data Principal Portal                                                │
├──────────────────────────────────────────────────────────────────────┤
│                                                                        │
│  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐    │
│  │ 🔐 Verify        │  │ 📋 Consent       │  │ 📜 History       │    │
│  │   Identity       │  │   Dashboard      │  │   Timeline       │    │
│  └──────────────────┘  └──────────────────┘  └──────────────────┘    │
│                                                                        │
│  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐    │
│  │ 📝 Submit        │  │ 🔍 Track         │  │ ⚖️ Appeal        │    │
│  │   DPR Request    │  │   My Requests    │  │   Decision       │    │
│  └──────────────────┘  └──────────────────┘  └──────────────────┘    │
│                                                                        │
└──────────────────────────────────────────────────────────────────────┘
```

### Portal Features

| Feature | Description |
|---------|-------------|
| **Identity Verification** | Email or Phone OTP to prove identity |
| **Consent Dashboard** | View current consent status per purpose, toggle on/off |
| **Consent History** | Immutable timeline of all consent changes with signatures |
| **DPR Submission** | Submit ACCESS, ERASURE, CORRECTION, NOMINATION requests |
| **DPR Tracking** | Track request status, view responses, download ACCESS data |
| **Guardian Consent** | Minor verification flow (DPDPA Section 9) |
| **Appeal** | Appeal rejected requests (DPDPA Section 18) |

---

## DPR (Data Principal Rights) Flow

```
┌──────────────┐    ┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│   Submit     │───►│   Verify     │───►│   Process    │───►│   Complete   │
│   Request    │    │   Identity   │    │   (Agent)    │    │   & Download │
└──────┬───────┘    └──────┬───────┘    └──────────────┘    └──────┬───────┘
       │                   │                                       │
       ▼                   ▼                                       ▼
  [Is Minor?]         [Rejected]                              [Appeal?]
       │                                                          │
       ▼                                                          ▼
  Guardian              ─────                              ┌──────────────┐
  Verification                                             │  Escalate to │
                                                           │  DPA Board   │
                                                           └──────────────┘
```

### DPR Status Flow

```
SUBMITTED → PENDING_VERIFICATION → VERIFIED → IN_PROGRESS → COMPLETED
                   │                  │                          │
                   ▼                  ▼                          ▼
            GUARDIAN_PENDING     REJECTED                    APPEALED
                                                                │
                                                                ▼
                                                            ESCALATED
```

| Status | Description |
|--------|-------------|
| `SUBMITTED` | Request received from data principal |
| `PENDING_VERIFICATION` | Awaiting identity verification (OTP) |
| `GUARDIAN_PENDING` | Minor's request — awaiting guardian consent |
| `VERIFIED` | Identity confirmed, ready to process |
| `IN_PROGRESS` | Agent is executing the request |
| `COMPLETED` | All tasks finished, results available |
| `REJECTED` | Identity verification failed or request invalid |
| `APPEALED` | Data principal has appealed the decision |
| `ESCALATED` | Escalated to Data Protection Authority |

---

## Guardian Consent for Minors (DPDPA Section 9)

For data principals under 18:

```
┌────────────────────────────────────────────────────────────────────┐
│  Guardian Consent Required                                          │
│                                                                     │
│  This data principal is under 18 years of age.                     │
│  Consent/request must be verified by parent or guardian.           │
│                                                                     │
│  Guardian Name:     [________________]                              │
│  Guardian Email:    [________________]                              │
│  Relationship:      [Parent ▼]                                      │
│                                                                     │
│  [Send Verification to Guardian]                                    │
└────────────────────────────────────────────────────────────────────┘
```

**Flow**: Minor submits → Guardian receives OTP → Guardian verifies → Request proceeds.

---

## Consent API

### Public API (Widget / Embed)

```http
POST /api/public/consent/sessions
Content-Type: application/json
X-API-Key: pk_live_xxx

{
  "widget_id": "wdg_abc123",
  "decisions": [
    { "purpose_id": "marketing", "granted": false },
    { "purpose_id": "analytics", "granted": true }
  ],
  "page_url": "https://acme.com/signup"
}
```

### Check Consent

```http
GET /api/public/consent/check?subject_id=uuid&purpose=marketing
X-API-Key: pk_live_xxx

Response:
{
  "has_consent": true,
  "consent_status": "GRANTED",
  "granted_at": "2026-01-15T10:30:00Z",
  "expires_at": "2027-01-15T10:30:00Z"
}
```

### Withdraw Consent (Portal)

```http
POST /api/public/portal/consent/withdraw
Authorization: Bearer <portal-session-token>

{
  "purpose_id": "marketing"
}
```

### Submit DPR Request (Portal)

```http
POST /api/public/portal/dpr
Authorization: Bearer <portal-session-token>

{
  "type": "ACCESS",
  "description": "I want a copy of all my personal data"
}
```

---

## Consent Analytics

The Control Centre provides analytics on consent rates:

### Metrics

| Metric | Description |
|--------|-------------|
| Consent Rate | % of subjects who granted consent |
| Withdrawal Rate | % who later withdrew consent |
| Purpose Breakdown | Consent rate per purpose |
| Channel Analysis | Consent by collection method (banner, portal, API) |
| Trend Analysis | Changes over time |

### Dashboard

```
┌────────────────────────────────────────────────────────────────────┐
│  Consent Analytics                                                  │
├────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  Overall Consent Rate: 72%                                         │
│                                                                     │
│  ┌─────────────────────────────────────────────────────┐            │
│  │ Purpose           │ Granted │ Denied │ Withdrawn │  Rate │      │
│  ├───────────────────┼─────────┼────────┼───────────┼───────┤      │
│  │ Marketing         │  8,234  │ 3,421  │    156    │  71%  │      │
│  │ Analytics         │ 10,892  │ 1,203  │     42    │  90%  │      │
│  │ Partner Sharing   │  2,156  │ 9,832  │    312    │  18%  │      │
│  └─────────────────────────────────────────────────────┘            │
│                                                                     │
└────────────────────────────────────────────────────────────────────┘
```

---

## DPDPA Compliance

### Section 6 Requirements

| Requirement | DataLens Feature |
|-------------|------------------|
| Free consent | No dark patterns |
| Informed | Detailed purpose descriptions |
| Specific | Per-purpose consent |
| Unambiguous | Explicit action required |
| Withdrawable | Withdrawal mechanism in portal |

### Section 9 — Minor's Consent

- Guardian verification required for data principals under 18
- Guardian OTP flow integrated into portal
- Consent records link guardian and minor

### Section 18 — Right to Appeal

- Data principals can appeal rejected DPR requests
- Appeal creates a new DPR linked to the original
- If appeal is denied, principal can escalate to Data Protection Board

### Consent Validity

- Must be given before processing begins
- Cannot be bundled with terms
- Must be as easy to withdraw as to give
- Must be documented with timestamp and digital signature
