package breach

import (
	"time"

	"github.com/complyark/datalens/pkg/types"
)

type IncidentStatus string

const (
	StatusOpen          IncidentStatus = "OPEN"
	StatusInvestigating IncidentStatus = "INVESTIGATING"
	StatusContained     IncidentStatus = "CONTAINED"
	StatusResolved      IncidentStatus = "RESOLVED"
	StatusReported      IncidentStatus = "REPORTED"
	StatusClosed        IncidentStatus = "CLOSED"
)

type IncidentSeverity string

const (
	SeverityLow      IncidentSeverity = "LOW"
	SeverityMedium   IncidentSeverity = "MEDIUM"
	SeverityHigh     IncidentSeverity = "HIGH"
	SeverityCritical IncidentSeverity = "CRITICAL"
)

type BreachIncident struct {
	types.BaseEntity
	TenantID types.ID `json:"tenant_id"`

	Title       string           `json:"title"`
	Description string           `json:"description"`
	Type        string           `json:"type"` // CERT-In Category (e.g., "Data Breach", "Malware")
	Severity    IncidentSeverity `json:"severity"`
	Status      IncidentStatus   `json:"status"`

	// Timestamps
	DetectedAt         time.Time  `json:"detected_at"`
	OccurredAt         time.Time  `json:"occurred_at"` // Optional/Estimated
	ReportedToCertInAt *time.Time `json:"reported_to_cert_in_at,omitempty"`
	ReportedToDPBAt    *time.Time `json:"reported_to_dpb_at,omitempty"`
	ClosedAt           *time.Time `json:"closed_at,omitempty"`

	// Impact
	AffectedSystems          []string `json:"affected_systems"` // List of System Names/IPs
	AffectedDataSubjectCount int      `json:"affected_data_subject_count"`
	PiiCategories            []string `json:"pii_categories"`

	// Response
	IsReportableToCertIn bool `json:"is_reportable_cert_in"` // Calculated or Manual
	IsReportableToDPB    bool `json:"is_reportable_dpb"`     // Calculated or Manual

	// PoC for this incident
	PoCName  string `json:"poc_name"`
	PoCRole  string `json:"poc_role"`
	PoCEmail string `json:"poc_email"`
}
