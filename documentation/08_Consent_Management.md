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

## Consent Lifecycle (MeITY BRD § 4.1)

The consent lifecycle defines five distinct phases through which every consent record progresses. Each phase transition is immutably logged and triggers downstream notifications.

```
  COLLECTION ──► VALIDATION ──► ACTIVE ──► UPDATE/RENEWAL ──► WITHDRAWAL
      │              │            │              │                  │
      ▼              ▼            ▼              ▼                  ▼
  User provides  Signature &  Processing   Purpose change      Processing
  affirmative    metadata     permitted    or expiry nearing   must stop
  action         verified                  → re-consent
```

### Phase Descriptions

| Phase | Trigger | Actions | BRD Ref |
|-------|---------|---------|---------|
| **Collection** | User interacts with consent widget | Present notice in preferred language, capture explicit opt-in per purpose, record metadata (IP, UA, timestamp, language) | § 4.1.1 |
| **Validation** | Consent submitted | Verify affirmative action, generate digital signature (SHA-256), store immutable record, emit `consent.granted` event | § 4.1.2 |
| **Update** | Purpose definition changes, notice version changes | Identify affected consents, re-present updated notice, collect fresh consent, link to previous record | § 4.1.3 |
| **Renewal** | Consent approaching expiry (`consent_expiry_days`) | Send proactive reminder (30/15/7 days before), present renewal notice, record renewed consent as new entry | § 4.1.4 |
| **Withdrawal** | Data principal withdraws via portal or API | Mark consent `WITHDRAWN`, emit `consent.withdrawn` event, notify Data Fiduciary & processors, cascade to downstream systems | § 4.1.5 |

### Consent Metadata (per record)

Every consent interaction captures the following metadata as required by the BRD:

| Field | Description | Example |
|-------|-------------|---------|
| `subject_id` | Data Principal identifier | `uuid` |
| `purpose_id` | Processing purpose granted/denied | `marketing` |
| `consent_status` | Current state | `GRANTED`, `DENIED`, `WITHDRAWN` |
| `language_preference` | Language in which notice was presented | `hi` (Hindi) |
| `timestamp` | ISO 8601 timestamp of action | `2026-02-10T10:30:00Z` |
| `widget_id` | Consent collection widget | `wdg_abc123` |
| `notice_version` | Version of the notice shown | `1.2` |
| `collection_channel` | How consent was collected | `BANNER`, `PORTAL`, `API` |
| `ip_address` | Client IP | `192.168.1.1` |
| `user_agent` | Browser/device fingerprint | `Mozilla/5.0...` |
| `signature` | SHA-256 digital signature | `sha256:abc123...` |

---

## Multi-Language Support (Eighth Schedule)

Consent notices must be available in all 22 languages listed in the Eighth Schedule of the Constitution of India. Translation is powered by a **HuggingFace API** integration — the English version of each notice is authored in the Control Centre, and translations are triggered from within the application.

### Supported Languages

| # | Language | ISO 639 | # | Language | ISO 639 |
|---|----------|---------|---|----------|---------|
| 1 | Assamese | `as` | 12 | Manipuri | `mni` |
| 2 | Bengali | `bn` | 13 | Marathi | `mr` |
| 3 | Bodo | `brx` | 14 | Nepali | `ne` |
| 4 | Dogri | `doi` | 15 | Odia | `or` |
| 5 | English | `en` | 16 | Punjabi | `pa` |
| 6 | Gujarati | `gu` | 17 | Sanskrit | `sa` |
| 7 | Hindi | `hi` | 18 | Santali | `sat` |
| 8 | Kannada | `kn` | 19 | Sindhi | `sd` |
| 9 | Kashmiri | `ks` | 20 | Tamil | `ta` |
| 10 | Konkani | `kok` | 21 | Telugu | `te` |
| 11 | Maithili | `mai` | 22 | Urdu | `ur` |

### Translation Flow

```
  Author notice    Trigger          HuggingFace       Store
  in English  ──►  translation  ──►  API call     ──►  per-language
  (Control Centre)  (in-app)         (external)        translations
                                                        │
                                                        ▼
                                                   Tied to notice
                                                   version + audit logged
```

- **Trigger**: Admin clicks "Translate" for a notice version in the Control Centre
- **API**: Application calls HuggingFace translation model endpoint with English text + target language code
- **Storage**: Each translation is stored in `consent_notice_translations` linked to the notice version
- **Versioning**: Translations are immutable per notice version — new notice version = new translations
- **Fallback**: If translation fails, the English version is served with a language warning indicator

---

## Consent Notifications (MeITY BRD § 4.4)

Notifications are sent to three stakeholder groups at key lifecycle events:

### Notification Matrix

| Event | Data Principal | Data Fiduciary | Data Processor |
|-------|:--------------:|:--------------:|:--------------:|
| Consent granted | ✅ Confirmation | ✅ Alert | — |
| Consent withdrawn | ✅ Acknowledgment | ✅ Alert | ✅ Revocation notice |
| Consent expiring (30d) | ✅ Reminder | ✅ Summary | — |
| Consent renewed | ✅ Confirmation | ✅ Alert | — |
| Notice updated | ✅ Re-consent request | ✅ Action required | — |
| DPR submitted | ✅ Acknowledgment | ✅ Assignment | — |
| DPR completed | ✅ Result + download | ✅ Closure | — |

### Notification Channels

| Channel | Use Case | Priority |
|---------|----------|----------|
| **Email** | Primary notification channel | High |
| **SMS** | OTP, urgent alerts | High |
| **In-App** | Dashboard notifications | Medium |
| **Webhook** | External system integration (DF/Processor) | Medium |

### Notification Schema

```json
{
  "id": "uuid",
  "tenant_id": "uuid",
  "recipient_type": "DATA_PRINCIPAL",
  "recipient_id": "uuid",
  "event_type": "CONSENT_WITHDRAWN",
  "channel": "EMAIL",
  "template_id": "uuid",
  "payload": {
    "subject_name": "John Doe",
    "purpose": "Marketing Communications",
    "action_url": "https://portal.datalens.io/..."
  },
  "status": "SENT",
  "sent_at": "2026-02-10T10:30:00Z",
  "created_at": "2026-02-10T10:30:00Z"
}
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

---

## Consent Enforcement Middleware (Planned)

> **Status**: Planned for future release. The API endpoint foundation is already in place.

Organizations will be able to embed a **single-line consent check** in their application code to verify real-time consent validity before any processing activity. This acts as a middleware/SDK that calls the consent module's check endpoint.

### Architecture

```
  Organization's App Code              DataLens
  ─────────────────────              ────────
  
  if datalens.HasConsent(userId, "marketing") {     ──► GET /api/public/consent/check
      processMarketingData()                              ?subject_id=X&purpose=marketing
  }                                                   ◄── { "has_consent": true, ... }
```

### Existing Foundation (API Endpoint)

The low-latency consent check endpoint is already defined:

```http
GET /api/public/consent/check?subject_id=uuid&purpose=marketing
X-API-Key: pk_live_xxx

Response (< 50ms target):
{
  "has_consent": true,
  "consent_status": "GRANTED",
  "granted_at": "2026-01-15T10:30:00Z",
  "expires_at": "2027-01-15T10:30:00Z"
}
```

### Planned SDK/Middleware Features

| Feature | Description |
|---------|-------------|
| **One-line embed** | `datalens.HasConsent(subjectId, purposeId)` — single function call |
| **Language SDKs** | Go, Python, Node.js, Java wrappers around the REST API |
| **Local caching** | Short-lived cache (configurable TTL) to minimize API calls |
| **Fail-safe mode** | Configurable: deny processing on API timeout (conservative) or allow (permissive) |
| **Batch check** | Check multiple purposes in a single call |
| **Event hook** | Emit events when consent is checked (for audit completeness) |
| **Middleware pattern** | Express/Gin/FastAPI middleware that auto-blocks request if consent missing |

### Design Considerations

- **Latency**: Target < 50ms p99 via Redis-backed consent cache
- **Cache invalidation**: Consent withdrawal events instantly invalidate cache via pub/sub
- **Idempotency**: Check endpoint is read-only, safe to retry
- **Rate limiting**: Higher limits for consent check vs. write operations (1000/min vs. 30/min)
