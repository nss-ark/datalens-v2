# 26. Breach Management Design (Batch 9)

## Overview

The Breach Management module addresses **DPDPA Section 28** and **CERT-In** reporting obligations. It allows users to log incidents, track reporting deadlines (6h / 72h), and generate notification data.

> **Note**: This module does **NOT** automatically detect breaches. It is a workflow tool for human operators.

---

## 1. Domain Model

**Package**: `internal/domain/breach`

### Entity: `BreachIncident`

```go
type IncidentStatus string

const (
    StatusOpen         IncidentStatus = "OPEN"
    StatusInvestigating IncidentStatus = "INVESTIGATING"
    StatusContained    IncidentStatus = "CONTAINED"
    StatusResolved     IncidentStatus = "RESOLVED"
    StatusReported     IncidentStatus = "REPORTED"
    StatusClosed       IncidentStatus = "CLOSED"
)

type IncidentSeverity string

const (
    SeverityLow      IncidentSeverity = "LOW"
    SeverityMedium   IncidentSeverity = "MEDIUM"
    SeverityHigh     IncidentSeverity = "HIGH"
    SeverityCritical IncidentSeverity = "CRITICAL"
)

type BreachIncident struct {
    ID              types.ID           `json:"id"`
    TenantID        types.ID           `json:"tenant_id"`
    Title           string             `json:"title"`
    Description     string             `json:"description"`
    Type            string             `json:"type"` // CERT-In Category (e.g., "Data Breach", "Malware")
    Severity        IncidentSeverity   `json:"severity"`
    Status          IncidentStatus     `json:"status"`
    
    // Timestamps
    DetectedAt      time.Time          `json:"detected_at"`
    OccurredAt      time.Time          `json:"occurred_at"` // Optional/Estimated
    ReportedToCertInAt *time.Time      `json:"reported_to_cert_in_at,omitempty"`
    ReportedToDPBAt    *time.Time      `json:"reported_to_dpb_at,omitempty"`
    ClosedAt          *time.Time       `json:"closed_at,omitempty"`
    
    // Impact
    AffectedSystems []string           `json:"affected_systems"` // List of System Names/IPs
    AffectedDataSubjectCount int       `json:"affected_data_subject_count"`
    PiiCategories   []string           `json:"pii_categories"`

    // Response
    IsReportableToCertIn bool          `json:"is_reportable_cert_in"` // Calculated or Manual
    IsReportableToDPB    bool          `json:"is_reportable_dpb"`     // Calculated or Manual
    
    // PoC for this incident
    PoCName         string             `json:"poc_name"`
    PoCRole         string             `json:"poc_role"`
    PoCEmail        string             `json:"poc_email"`
    
    CreatedAt       time.Time          `json:"created_at"`
    UpdatedAt       time.Time          `json:"updated_at"`
}
```

### SLA Calculation (Virtual Fields)

- `TimeRemainingCertIn`: (DetectedAt + 6h) - Now
- `TimeRemainingDPB`: (DetectedAt + 72h) - Now

---

## 2. API Endpoints

**Base Path**: `/api/v2/breach`

| Method | Path | Description |
|--------|------|-------------|
| GET | `/incidents` | List incidents (filter by status, severity) |
| POST | `/incidents` | Log new incident |
| GET | `/incidents/{id}` | Get incident details |
| PUT | `/incidents/{id}` | Update incident (status, details) |
| GET | `/incidents/{id}/report/cert-in` | Generate CERT-In JSON payload |
| GET | `/incidents/{id}/report/dpb` | Generate DPB JSON payload |

---

## 3. Microsoft 365 Integration (Batch 9)

### Entity: `IntegrationConfig` (or reuse DataSource)

We will reuse `DataSource` but with `Type: "MICROSOFT_365"`.

**Credential Structure (JSON in `DataSource.Credentials`)**:
```json
{
  "client_id": "...",
  "client_secret": "...",
  "tenant_id": "...", // Azure AD Tenant
  "scopes": ["User.Read.All", "Files.Read.All", "Mail.Read"],
  "refresh_token": "..." // Stored securely
}
```

### Auth Flow
1. **Initiate**: `GET /api/v2/integration/m365/auth-url` -> Returns Microsoft OAuth2 URL.
2. **Callback**: `POST /api/v2/integration/m365/callback` -> Exchanger code for tokens, creates/updates DataSource.

---

## 4. UI Requirements (Frontend)

### Dashboard (`/breach/dashboard`)
- **Stats**: Open Incidents, SLA Breaches (Next 6h, Next 72h).
- **List**: Sortable table of incidents with Status Badge and Severity Color.

### Incident Form (`/breach/new` & `/breach/{id}`)
- **Incident Types**: Dropdown with CERT-In 21 categories.
- **Timestamps**: Date-Time pickers.
- **SLA Countdown**: Prominent timer showing time left to report.
- **Report Generation**: Buttons to "View CERT-In Report" (Displays data in the CERT-In format).

---

## 5. Compliance Rules

1. **6-Hour Rule**: If `Type` is in [List of Reportable Incidents], show WARNING if (Now - DetectedAt) > 4 hours.
2. **Logs**: All actions on Breach entities must be Audit Logged (using the Audit system built in Batch 8).
