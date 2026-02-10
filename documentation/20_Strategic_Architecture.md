# 20. Strategic Architecture Design

## Vision Statement

> **DataLens 2.0**: The world's most automated, reliable, and evidence-ready privacy compliance platform—regulation-agnostic at its core, AI-leveraged where it matters.

---

## Design Principles

### 1. Regulation-Agnostic Core

```
❌ WRONG: Build for DPDPA, then adapt for GDPR
✅ RIGHT: Build universal compliance engine, configure for any regulation
```

The core engine knows nothing about DPDPA, GDPR, or CCPA. It only understands:
- **Data** (personal, sensitive, categories)
- **Subjects** (individuals with rights)
- **Purposes** (why data is processed)
- **Actions** (discover, consent, delete, correct, export)
- **Evidence** (immutable proof of compliance actions)

### 2. Plugin Architecture

Every feature should be pluggable:
- **Data Sources**: Add new connectors without touching core
- **Regulations**: Add compliance frameworks as configuration
- **AI Models**: Swap LLMs without code changes
- **Notifications**: Email, SMS, webhook—all pluggable
- **Storage**: PostgreSQL, MongoDB, S3—abstracted

### 3. Event-Driven Everything

```
┌─────────┐     ┌─────────────┐     ┌──────────────────────┐
│ Action  │ ──► │   Event     │ ──► │  Subscribers         │
│ Occurs  │     │   Published │     │  • Audit Log         │
└─────────┘     └─────────────┘     │  • Notification      │
                                     │  • Analytics         │
                                     │  • Compliance Check  │
                                     │  • Webhook           │
                                     └──────────────────────┘
```

Every action emits events. Components react—they never directly call each other.

### 4. Evidence-First Design

Every operation must create immutable evidence:
- **What** happened
- **When** it happened (cryptographic timestamp)
- **Who** triggered it
- **Proof** it was executed
- **Hash** for tamper detection

---

## System Architecture

### Layer Model

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              PRESENTATION LAYER                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │    Web UI   │  │  Mobile UI  │  │   REST API  │  │    GraphQL API      │ │
│  │   (React)   │  │  (React N.) │  │   (Public)  │  │    (Internal)       │ │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘  └──────────┬──────────┘ │
└─────────┼────────────────┼────────────────┼─────────────────────┼───────────┘
          └────────────────┴────────────────┴─────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                              APPLICATION LAYER                               │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                         API GATEWAY                                  │    │
│  │  • Authentication  • Rate Limiting  • Request Routing  • Logging    │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                      │                                       │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                    ORCHESTRATION ENGINE                              │    │
│  │  • Workflow Management  • Task Scheduling  • Event Routing          │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                              DOMAIN LAYER (CORE)                             │
│                                                                              │
│  ┌───────────────┐  ┌───────────────┐  ┌───────────────┐  ┌─────────────┐  │
│  │ PII Detection │  │ Data Mapping  │  │ DSR Engine    │  │ Audit &     │  │
│  │ Engine        │  │ Service       │  │               │  │ Evidence    │  │
│  │               │  │               │  │               │  │             │  │
│  │ • AI Analysis │  │ • Discovery   │  │ • Workflow    │  │ • Immutable │  │
│  │ • Validation  │  │ • Mapping     │  │ • Execution   │  │ • Signed    │  │
│  │ • Confidence  │  │ • Lineage     │  │ • Verification│  │ • Queryable │  │
│  └───────────────┘  └───────────────┘  └───────────────┘  └─────────────┘  │
│                                                                              │
│  ┌───────────────┐  ┌───────────────┐  ┌───────────────┐  ┌─────────────┐  │
│  │ Consent       │  │ Governance    │  │ Notification  │  │ Analytics   │  │
│  │ Manager       │  │ Policy Engine │  │ Service       │  │ Engine      │  │
│  │               │  │               │  │               │  │             │  │
│  │ • Capture     │  │ • Rules       │  │ • Multi-chan  │  │ • Insights  │  │
│  │ • Tracking    │  │ • Enforcement │  │ • Templates   │  │ • Reports   │  │
│  │ • Expiry      │  │ • Alerts      │  │ • Scheduling  │  │ • Dashboard │  │
│  └───────────────┘  └───────────────┘  └───────────────┘  └─────────────┘  │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                        COMPLIANCE ADAPTER LAYER                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │   DPDPA     │  │    GDPR     │  │    CCPA     │  │   Custom/Future     │ │
│  │   Adapter   │  │   Adapter   │  │   Adapter   │  │   Adapters          │ │
│  │             │  │             │  │             │  │                     │ │
│  │ • Rights    │  │ • Rights    │  │ • Rights    │  │ • Configurable      │ │
│  │ • Timelines │  │ • Timelines │  │ • Timelines │  │ • Template-based    │ │
│  │ • Notices   │  │ • Notices   │  │ • Notices   │  │                     │ │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         INTEGRATION LAYER                                    │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │ AI Gateway  │  │ Data Source │  │ Message     │  │ External Services   │ │
│  │             │  │ Connectors  │  │ Queue       │  │                     │ │
│  │ • OpenAI    │  │ • SQL       │  │ • NATS      │  │ • DigiLocker        │ │
│  │ • Claude    │  │ • NoSQL     │  │ • Kafka     │  │ • Payment (KYC)     │ │
│  │ • Local LLM │  │ • S3/Cloud  │  │ • Redis     │  │ • Email/SMS         │ │
│  │ • Fallback  │  │ • Control Centre APIs │  │             │  │ • Webhooks          │ │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         INFRASTRUCTURE LAYER                                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │ PostgreSQL  │  │ Redis       │  │ Object      │  │ Observability       │ │
│  │ (Primary)   │  │ (Cache)     │  │ Storage     │  │                     │ │
│  │             │  │             │  │ (Evidence)  │  │ • Prometheus        │ │
│  │ • Entities  │  │ • Sessions  │  │ • S3/Minio  │  │ • Jaeger            │ │
│  │ • Events    │  │ • Rate Lim  │  │ • Backups   │  │ • Loki              │ │
│  │ • Audit     │  │ • Pub/Sub   │  │             │  │ • Grafana           │ │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Core Engine Specifications

### 1. PII Detection Engine

**Purpose**: Detect, classify, and score personal data across any data source.

```go
// Interface - regulation-agnostic
type PIIDetector interface {
    // Detect PII in raw content
    Detect(ctx context.Context, input DetectionInput) ([]Detection, error)
    
    // Classify detected PII into categories
    Classify(ctx context.Context, detections []Detection) ([]ClassifiedPII, error)
    
    // Learn from human corrections
    Learn(ctx context.Context, feedback DetectionFeedback) error
}

// Implementation composes multiple strategies
type ComposablePIIDetector struct {
    strategies []DetectionStrategy  // AI, Regex, Heuristic
    validator  PIIValidator
    cache      DetectionCache
    learner    FeedbackLearner
}

// Strategy pattern for different detection methods
type DetectionStrategy interface {
    Name() string
    Priority() int
    Detect(ctx context.Context, input DetectionInput) ([]Detection, error)
}
```

**Strategies** (composable, orderable):
1. **AIStrategy**: LLM-based contextual analysis
2. **PatternStrategy**: Regex patterns for known formats
3. **HeuristicStrategy**: Column name, data structure hints
4. **IndustryStrategy**: Sector-specific patterns

### 2. Data Mapping Service

**Purpose**: Universal data discovery and lineage tracking.

```go
type DataMapper interface {
    // Discover data structure from any source
    Discover(ctx context.Context, source DataSource) (*DataInventory, error)
    
    // Map data fields to universal categories
    MapFields(ctx context.Context, fields []Field) ([]MappedField, error)
    
    // Track data lineage (where data flows)
    TrackLineage(ctx context.Context, dataID string) (*DataLineage, error)
    
    // Sync changes (incremental)
    Sync(ctx context.Context, source DataSource, since time.Time) (*SyncResult, error)
}

// Universal field mapping (regulation-agnostic)
type MappedField struct {
    SourceField    FieldReference
    Category       DataCategory      // IDENTITY, CONTACT, FINANCIAL, HEALTH, etc.
    Sensitivity    SensitivityLevel  // LOW, MEDIUM, HIGH, CRITICAL
    PIIType        string            // NAME, EMAIL, PHONE, AADHAAR, SSN, etc.
    Purposes       []Purpose         // Why this data is collected
    LegalBasis     LegalBasis        // CONSENT, CONTRACT, LEGAL_OBLIGATION, etc.
    RetentionDays  int               // Auto-computed or manual
    DataSubjectRef string            // Link to subject identifier field
}
```

### 3. DSR Engine

**Purpose**: Execute data subject requests across any regulation.

```go
type DSREngine interface {
    // Create a new request (type-agnostic)
    Create(ctx context.Context, request DSRRequest) (*DSR, error)
    
    // Execute across all relevant data sources
    Execute(ctx context.Context, dsrID string) (*ExecutionResult, error)
    
    // Verify execution completed successfully
    Verify(ctx context.Context, dsrID string) (*VerificationResult, error)
    
    // Generate evidence package
    GenerateEvidence(ctx context.Context, dsrID string) (*EvidencePackage, error)
}

// Universal DSR request (works for any regulation)
type DSRRequest struct {
    ID              string
    Type            DSRType          // ACCESS, ERASURE, CORRECTION, PORTABILITY, OBJECTION
    SubjectID       string           // Data subject identifier
    Scope           DSRScope         // Which data sources, categories
    RequestedBy     Requester        // Who made the request
    RegulationRef   string           // "DPDPA", "GDPR", etc. (for SLA/rules)
    Deadline        time.Time        // Auto-computed from regulation
    Priority        Priority
    Metadata        map[string]any   // Regulation-specific fields
}

// Type is universal, regulations just enable/disable certain types
type DSRType string

const (
    DSRAccess      DSRType = "ACCESS"       // DPDPA: Yes, GDPR: Yes, CCPA: Yes
    DSRErasure     DSRType = "ERASURE"      // DPDPA: Yes, GDPR: Yes, CCPA: Yes
    DSRCorrection  DSRType = "CORRECTION"   // DPDPA: Yes, GDPR: Yes, CCPA: No
    DSRPortability DSRType = "PORTABILITY"  // DPDPA: Yes, GDPR: Yes, CCPA: No
    DSRObjection   DSRType = "OBJECTION"    // DPDPA: No,  GDPR: Yes, CCPA: No
    DSRRestriction DSRType = "RESTRICTION"  // DPDPA: No,  GDPR: Yes, CCPA: No
)
```

### 4. Audit & Evidence Engine

**Purpose**: Create tamper-proof, legally admissible evidence.

```go
type EvidenceEngine interface {
    // Record any action with full context
    Record(ctx context.Context, event AuditEvent) (*EvidenceRecord, error)
    
    // Generate evidence package for legal/audit purposes
    GeneratePackage(ctx context.Context, query EvidenceQuery) (*EvidencePackage, error)
    
    // Verify evidence integrity (hash chain)
    Verify(ctx context.Context, recordID string) (*VerificationResult, error)
    
    // Export for external audit
    Export(ctx context.Context, query EvidenceQuery, format ExportFormat) (io.Reader, error)
}

// Every action creates immutable evidence
type EvidenceRecord struct {
    ID              string            // UUID
    Timestamp       time.Time         // Cryptographic timestamp
    EventType       string            // CONSENT_GRANTED, DSR_EXECUTED, PII_DISCOVERED, etc.
    Actor           Actor             // Who performed action
    Subject         *DataSubject      // Affected data subject (if any)
    Action          ActionDetails     // What was done
    Before          *Snapshot         // State before (for changes)
    After           *Snapshot         // State after (for changes)
    ProofHash       string            // SHA-256 of (previous_hash + this_record)
    Signature       string            // Digital signature
    RegulationRefs  []string          // Which regulations this satisfies
    Metadata        map[string]any
}
```

### 5. Consent Manager

**Purpose**: Universal consent capture, tracking, and enforcement.

```go
type ConsentManager interface {
    // Capture new consent
    Capture(ctx context.Context, consent Consent) (*ConsentRecord, error)
    
    // Check if action is allowed
    IsAllowed(ctx context.Context, query ConsentQuery) (*ConsentDecision, error)
    
    // Withdraw consent
    Withdraw(ctx context.Context, consentID string, reason string) error
    
    // Get consent status for subject
    GetStatus(ctx context.Context, subjectID string) (*ConsentStatus, error)
    
    // Generate consent receipt
    GenerateReceipt(ctx context.Context, consentID string) (*ConsentReceipt, error)
}

// Universal consent model
type Consent struct {
    SubjectID       string
    Purposes        []Purpose         // What data is used for
    DataCategories  []DataCategory    // What types of data
    GrantedAt       time.Time
    ExpiresAt       *time.Time        // Optional expiry
    Mechanism       ConsentMechanism  // EXPLICIT, OPT_IN, OPT_OUT
    Proof           ConsentProof      // IP, timestamp, UI screenshot hash
    LegalBasis      LegalBasis
    RegulationRef   string            // Which regulation governs
}
```

### 6. Governance Policy Engine

**Purpose**: Define and enforce data governance rules.

```go
type PolicyEngine interface {
    // Define a policy
    Define(ctx context.Context, policy Policy) error
    
    // Evaluate action against policies
    Evaluate(ctx context.Context, action ProposedAction) (*PolicyDecision, error)
    
    // Get violations
    GetViolations(ctx context.Context, query ViolationQuery) ([]Violation, error)
    
    // Auto-remediate (if configured)
    Remediate(ctx context.Context, violationID string) (*RemediationResult, error)
}

// Policy is regulation-agnostic
type Policy struct {
    ID          string
    Name        string
    Description string
    Rules       []PolicyRule
    Severity    Severity        // INFO, WARNING, CRITICAL
    Actions     []RemediationAction  // ALERT, BLOCK, AUTO_FIX
    Enabled     bool
}

// Example rules (configured, not coded)
// - "PII without purpose must be flagged within 7 days"
// - "Consent expiring in 30 days triggers renewal notification"
// - "DSR must complete within regulation SLA"
// - "Cross-border transfer requires legal basis documentation"
```

---

## Compliance Adapter Pattern

### How Regulations Plug In

```go
// Each regulation is just configuration
type ComplianceAdapter interface {
    // Name of regulation
    Name() string  // "DPDPA", "GDPR", "CCPA"
    
    // Which DSR types are supported
    SupportedDSRTypes() []DSRType
    
    // SLA for each DSR type (days)
    DSRDeadlines() map[DSRType]int
    
    // Required consent mechanisms
    ConsentRequirements() ConsentRequirements
    
    // Breach notification rules
    BreachRules() BreachNotificationRules
    
    // Data subject rights configuration
    RightsConfiguration() RightsConfig
    
    // Notice/disclosure requirements
    NoticeRequirements() NoticeConfig
}

// DPDPA Adapter (example)
type DPDPAAdapter struct{}

func (a *DPDPAAdapter) Name() string { return "DPDPA" }

func (a *DPDPAAdapter) SupportedDSRTypes() []DSRType {
    return []DSRType{DSRAccess, DSRErasure, DSRCorrection, DSRPortability}
}

func (a *DPDPAAdapter) DSRDeadlines() map[DSRType]int {
    return map[DSRType]int{
        DSRAccess:      30,  // Section 11
        DSRErasure:     30,  // Section 12
        DSRCorrection:  30,  // Section 11
        DSRPortability: 30,  // Section 11
    }
}

func (a *DPDPAAdapter) BreachRules() BreachNotificationRules {
    return BreachNotificationRules{
        NotifyBoardDeadlineHours: 72,
        NotifySubjectsRequired:   true,
        IncidentTypes:           21,  // CERT-In categories
    }
}
```

### Adding a New Regulation

To add GDPR support:
1. Create `GDPRAdapter` implementing `ComplianceAdapter`
2. Configure in system settings
3. Done. No code changes to core.

```yaml
# config/regulations/gdpr.yaml
name: GDPR
jurisdiction: EU
dsr_types:
  - ACCESS
  - ERASURE
  - CORRECTION
  - PORTABILITY
  - OBJECTION
  - RESTRICTION
deadlines:
  default: 30
  erasure: 30
  access: 30
consent:
  explicit_required: true
  granular_purposes: true
  withdrawal_easy: true
breach:
  notify_authority_hours: 72
  notify_subjects: conditional
```

---

## Data Source Connector Pattern

### Universal Connector Interface

```go
type DataSourceConnector interface {
    // Connect to the data source
    Connect(ctx context.Context, config ConnectionConfig) error
    
    // Discover schema/structure
    DiscoverSchema(ctx context.Context) (*Schema, error)
    
    // Sample data for PII detection
    Sample(ctx context.Context, entity string, limit int) ([]Record, error)
    
    // Execute DSR actions
    ExecuteDSR(ctx context.Context, dsr DSRAction) (*DSRResult, error)
    
    // Close connection
    Close() error
    
    // Capabilities (what can this connector do?)
    Capabilities() ConnectorCapabilities
}

type ConnectorCapabilities struct {
    CanDiscover     bool  // Can discover schema automatically
    CanSample       bool  // Can sample data
    CanDelete       bool  // Can perform deletions
    CanUpdate       bool  // Can perform corrections
    CanExport       bool  // Can export data
    SupportsStreaming bool // Can stream large datasets
    SupportsIncremental bool // Can sync incrementally
}
```

### Example: Adding a New Connector

```go
// Snowflake connector (example)
type SnowflakeConnector struct {
    db     *sql.DB
    config SnowflakeConfig
}

func (c *SnowflakeConnector) Capabilities() ConnectorCapabilities {
    return ConnectorCapabilities{
        CanDiscover:        true,
        CanSample:          true,
        CanDelete:          true,  // Snowflake supports DELETE
        CanUpdate:          true,
        CanExport:          true,
        SupportsStreaming:  true,
        SupportsIncremental: true,
    }
}

func (c *SnowflakeConnector) DiscoverSchema(ctx context.Context) (*Schema, error) {
    // Snowflake-specific schema discovery
    query := `SELECT table_schema, table_name, column_name, data_type 
              FROM information_schema.columns`
    // ... implementation
}
```

---

## Event System

### Event Categories

```go
// All events in the system
const (
    // PII Events
    EventPIIDiscovered     = "pii.discovered"
    EventPIIVerified       = "pii.verified"
    EventPIIRejected       = "pii.rejected"
    EventPIIClassified     = "pii.classified"
    
    // DSR Events
    EventDSRCreated        = "dsr.created"
    EventDSRExecuting      = "dsr.executing"
    EventDSRCompleted      = "dsr.completed"
    EventDSRFailed         = "dsr.failed"
    EventDSRVerified       = "dsr.verified"
    
    // Consent Events
    EventConsentGranted    = "consent.granted"
    EventConsentWithdrawn  = "consent.withdrawn"
    EventConsentExpiring   = "consent.expiring"
    
    // Breach Events
    EventBreachDetected    = "breach.detected"
    EventBreachNotified    = "breach.notified"
    EventBreachResolved    = "breach.resolved"
    
    // Scan Events
    EventScanStarted       = "scan.started"
    EventScanCompleted     = "scan.completed"
    EventScanFailed        = "scan.failed"
    
    // Policy Events
    EventPolicyViolation   = "policy.violation"
    EventPolicyRemediated  = "policy.remediated"
)
```

### Event Bus

```go
type EventBus interface {
    // Publish an event
    Publish(ctx context.Context, event Event) error
    
    // Subscribe to events
    Subscribe(ctx context.Context, pattern string, handler EventHandler) (Subscription, error)
    
    // Replay events (for audit/debugging)
    Replay(ctx context.Context, query EventQuery) (<-chan Event, error)
}

// Subscribers are decoupled
// Example: Consent withdrawn -> Multiple reactions
// 1. Audit logger records it
// 2. Notification service alerts user
// 3. Analytics updates metrics
// 4. DSR engine checks for pending requests
// 5. Webhook delivers to external systems
```

---

## Embeddable Consent SDK Architecture

### Overview

Companies embed a lightweight JavaScript snippet into their digital touchpoints (websites, web apps, kiosks) to collect consent. The SDK communicates directly with the Control Centre's public API. All consent records are stored as an immutable log on the Control Centre.

### Delivery Model

```
Company Website / App
    │
    ├── <script src="https://cdn.datalens.io/consent.min.js"
    │          data-widget-id="wdg_xxx" data-api-key="pk_xxx"></script>
    │
    └── OR: <iframe src="https://portal.datalens.io/t/{tenant}/portal"></iframe>
```

**Two deployment options:**

| Method | Use Case | Weight |
|--------|----------|--------|
| **JS Snippet** | Consent banners, preference centers, inline forms | ~15 KB gzipped |
| **Iframe** | Full Data Principal Portal (consent history, DPR submission) | Self-contained page |

### Public API Surface

```
/api/public/consent/                   (Public — API-key authenticated)
├── POST   /sessions                   # Record consent decisions
├── GET    /check?subject=X&purpose=Y  # Check consent status
├── POST   /withdraw                   # Withdraw consent
├── GET    /widget/{id}/config         # Fetch widget configuration
│
/api/public/portal/                    (Public — OTP authenticated)
├── POST   /verify                     # Email/Phone OTP verification
├── GET    /profile                    # Data principal profile
├── GET    /consent-history            # Consent timeline
├── POST   /dpr                        # Submit DPR request
├── GET    /dpr/{id}                   # Track DPR status
├── POST   /dpr/{id}/appeal            # Appeal a rejected DPR
└── GET    /dpr/{id}/download          # Download ACCESS data
```

### Consent Session Flow

```
┌──────────────┐     ┌──────────────────┐     ┌──────────────────┐
│  JS Snippet  │────►│ Public API       │────►│ Control Centre   │
│  (Browser)   │     │ /consent/session │     │ Immutable Log    │
└──────────────┘     └──────────────────┘     └──────────────────┘
       │                      │                        │
       │   1. Load config     │                        │
       │◄─────────────────────│                        │
       │                      │                        │
       │   2. Show banner     │                        │
       │   (user interacts)   │                        │
       │                      │                        │
       │   3. POST decisions  │   4. Store + sign      │
       │─────────────────────►│───────────────────────►│
       │                      │                        │
       │   5. Ack + set       │                        │
       │◄─────────────────────│                        │
       │      cookie/flag     │                        │
       ▼                      ▼                        ▼
```

### Data Principal Portal (Iframe)

The portal is a standalone web page served by the Control Centre, embeddable via iframe. It provides:

1. **Identity Verification** — Email or Phone OTP
2. **Consent Dashboard** — Current consent status per purpose
3. **Consent History Timeline** — Immutable, chronological record of all consent changes
4. **DPR Submission** — Submit ACCESS, ERASURE, CORRECTION, NOMINATION requests
5. **DPR Tracking** — Track request status and download results
6. **Guardian Consent** — Minor verification flow (DPDPA Section 9)
7. **Appeal** — Appeal rejected requests (DPDPA Section 18)

---

## Deployment Topology

### Cloud-Native Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              KUBERNETES CLUSTER                              │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                         INGRESS (LoadBalancer)                       │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                      │                                       │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                         API GATEWAY (Kong/Envoy)                     │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│           │              │              │              │                     │
│           ▼              ▼              ▼              ▼                     │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐           │
│  │ PII Service │ │ DSR Service │ │Consent Svc  │ │ Audit Svc   │           │
│  │ (3 replicas)│ │ (3 replicas)│ │ (2 replicas)│ │ (2 replicas)│           │
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘           │
│           │              │              │              │                     │
│           └──────────────┴──────────────┴──────────────┘                     │
│                                      │                                       │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                         MESSAGE QUEUE (NATS)                         │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│           │              │              │              │                     │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐           │
│  │ PostgreSQL  │ │ Redis       │ │ MinIO (S3)  │ │ Prometheus  │           │
│  │ (Primary)   │ │ (Cluster)   │ │ (Evidence)  │ │ + Grafana   │           │
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘           │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                         CUSTOMER PREMISES (Agent)                            │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                         DataLens Agent                               │    │
│  │  • Connects to local data sources                                    │    │
│  │  • Executes DSR actions locally                                      │    │
│  │  • Sends metadata only (Zero-PII to cloud)                          │    │
│  │  • Can run air-gapped with local CONTROL CENTRE                               │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Security Architecture

### Zero-Trust Model

```
1. IDENTITY: Every request is authenticated (JWT + mTLS)
2. AUTHORIZATION: Every action is authorized (ABAC/RBAC)
3. ENCRYPTION: All data encrypted at rest and in transit
4. AUDIT: Every action is logged immutably
5. SECRETS: All credentials in Vault, never in code/config
```

### Data Classification

```go
type SecurityClassification string

const (
    ClassPublic       SecurityClassification = "PUBLIC"       // Non-sensitive
    ClassInternal     SecurityClassification = "INTERNAL"     // Business data
    ClassConfidential SecurityClassification = "CONFIDENTIAL" // PII
    ClassRestricted   SecurityClassification = "RESTRICTED"   // Sensitive PII
    ClassSecret       SecurityClassification = "SECRET"       // Highly sensitive
)

// Automatic classification based on PII type
func ClassifyPII(piiType string) SecurityClassification {
    switch piiType {
    case "AADHAAR", "PASSPORT", "HEALTH_DATA", "BIOMETRIC":
        return ClassRestricted
    case "NAME", "EMAIL", "PHONE", "ADDRESS":
        return ClassConfidential
    case "JOB_TITLE", "DEPARTMENT":
        return ClassInternal
    default:
        return ClassPublic
    }
}
```

---

## Technology Decisions

| Component | Technology | Rationale |
|-----------|------------|-----------|
| **API Language** | Go | Performance, concurrency, single binary |
| **AI/ML** | Python | ML ecosystem, model availability |
| **Frontend** | React + TypeScript | Ecosystem, type safety |
| **Database** | PostgreSQL | ACID, JSON support, extensions |
| **Cache** | Redis | Performance, pub/sub, rate limiting |
| **Queue** | NATS | Lightweight, Go-native, JetStream |
| **Object Store** | MinIO/S3 | Evidence storage, backups |
| **Container** | Docker + Kubernetes | Scalability, DevOps |
| **Observability** | Prometheus + Grafana + Jaeger | Industry standard |
| **Secrets** | HashiCorp Vault | Enterprise secrets management |

---

## Next Steps

1. **[21_Domain_Model.md](./21_Domain_Model.md)** - Bounded contexts, entity relationships
2. **[22_AI_Integration_Strategy.md](./22_AI_Integration_Strategy.md)** - AI patterns, when to use, fallbacks
3. **[23_AGILE_Development_Plan.md](./23_AGILE_Development_Plan.md)** - Sprint breakdown, milestones
