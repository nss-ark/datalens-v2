package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/complyark/datalens/internal/domain/audit"
	"github.com/complyark/datalens/internal/domain/breach"
	"github.com/complyark/datalens/internal/domain/compliance"
	"github.com/complyark/datalens/internal/domain/consent"
	"github.com/complyark/datalens/internal/domain/governance"
	"github.com/complyark/datalens/pkg/types"
)

// =============================================================================
// ReportService — Compliance Snapshot + Data Export
// =============================================================================

// ReportService aggregates data from existing services for compliance reporting.
type ReportService struct {
	dsrSvc       *DSRService
	breachSvc    *BreachService
	consentSvc   *ConsentService
	deptSvc      *DepartmentService
	tpSvc        *ThirdPartyService
	purposeSvc   *PurposeService
	auditSvc     *AuditService
	retentionSvc *RetentionService
	logger       *slog.Logger
}

// NewReportService creates a new ReportService.
func NewReportService(
	dsrSvc *DSRService,
	breachSvc *BreachService,
	consentSvc *ConsentService,
	deptSvc *DepartmentService,
	tpSvc *ThirdPartyService,
	purposeSvc *PurposeService,
	auditSvc *AuditService,
	retentionSvc *RetentionService,
	logger *slog.Logger,
) *ReportService {
	return &ReportService{
		dsrSvc:       dsrSvc,
		breachSvc:    breachSvc,
		consentSvc:   consentSvc,
		deptSvc:      deptSvc,
		tpSvc:        tpSvc,
		purposeSvc:   purposeSvc,
		auditSvc:     auditSvc,
		retentionSvc: retentionSvc,
		logger:       logger.With("service", "report"),
	}
}

// =============================================================================
// DTOs
// =============================================================================

// ComplianceSnapshot is the top-level report response.
type ComplianceSnapshot struct {
	GeneratedAt     time.Time        `json:"generated_at"`
	PeriodFrom      time.Time        `json:"period_from"`
	PeriodTo        time.Time        `json:"period_to"`
	OverallScore    float64          `json:"overall_score"`
	Pillars         []PillarScore    `json:"pillars"`
	Recommendations []Recommendation `json:"recommendations"`
	Summary         SnapshotSummary  `json:"summary"`
}

// PillarScore represents one of the 4 DPDPA compliance pillars.
type PillarScore struct {
	Name   string  `json:"name"`
	Score  float64 `json:"score"`
	Weight float64 `json:"weight"`
	Detail string  `json:"detail"`
}

// Recommendation is an actionable compliance suggestion.
type Recommendation struct {
	Priority string `json:"priority"` // HIGH, MEDIUM, LOW
	Category string `json:"category"`
	Message  string `json:"message"`
}

// SnapshotSummary holds raw counts for the report.
type SnapshotSummary struct {
	TotalDSRs            int `json:"total_dsrs"`
	CompletedDSRs        int `json:"completed_dsrs"`
	OverdueDSRs          int `json:"overdue_dsrs"`
	TotalBreaches        int `json:"total_breaches"`
	OpenBreaches         int `json:"open_breaches"`
	TotalConsentSessions int `json:"total_consent_sessions"`
	TotalDepartments     int `json:"total_departments"`
	DepartmentsWithOwner int `json:"departments_with_owner"`
	TotalThirdParties    int `json:"total_third_parties"`
	ThirdPartiesWithDPA  int `json:"third_parties_with_dpa"`
	TotalPurposes        int `json:"total_purposes"`
	ActivePurposes       int `json:"active_purposes"`
	RetentionPolicies    int `json:"retention_policies"`
}

// =============================================================================
// GenerateComplianceSnapshot
// =============================================================================

// GenerateComplianceSnapshot aggregates data across services and computes pillar scores.
func (s *ReportService) GenerateComplianceSnapshot(ctx context.Context, from, to time.Time) (*ComplianceSnapshot, error) {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, types.NewForbiddenError("tenant context required")
	}

	// Use a large page to get counts — this is a reporting endpoint, not a list.
	bigPage := types.Pagination{Page: 1, PageSize: 10000}

	// 1. DSR data
	dsrResult, err := s.dsrSvc.GetDSRs(ctx, bigPage, nil, nil)
	if err != nil {
		s.logger.Error("report: failed to get DSRs", "error", err)
		dsrResult = &types.PaginatedResult[compliance.DSR]{}
	}

	overdueDSRs, err := s.dsrSvc.GetOverdue(ctx)
	if err != nil {
		s.logger.Error("report: failed to get overdue DSRs", "error", err)
	}

	totalDSRs := dsrResult.Total
	completedDSRs := 0
	for _, d := range dsrResult.Items {
		if d.Status == compliance.DSRStatusCompleted || d.Status == compliance.DSRStatusVerified {
			completedDSRs++
		}
	}

	// 2. Breach data
	breachResult, err := s.breachSvc.ListIncidents(ctx, breach.Filter{}, bigPage)
	if err != nil {
		s.logger.Error("report: failed to get breaches", "error", err)
		breachResult = &types.PaginatedResult[breach.BreachIncident]{}
	}

	totalBreaches := breachResult.Total
	openBreaches := 0
	reportedBreaches := 0
	for _, b := range breachResult.Items {
		if b.Status == breach.StatusOpen || b.Status == breach.StatusInvestigating {
			openBreaches++
		}
		if b.ReportedToCertInAt != nil || b.Status == breach.StatusReported || b.Status == breach.StatusClosed {
			reportedBreaches++
		}
	}

	// 3. Consent data
	consentResult, err := s.consentSvc.ListSessionsByTenant(ctx, consent.ConsentSessionFilters{}, bigPage)
	if err != nil {
		s.logger.Error("report: failed to get consent sessions", "error", err)
		consentResult = &types.PaginatedResult[consent.ConsentSession]{}
	}
	totalConsent := consentResult.Total

	// 4. Department data
	depts, err := s.deptSvc.List(ctx)
	if err != nil {
		s.logger.Error("report: failed to get departments", "error", err)
	}
	totalDepts := len(depts)
	deptsWithOwner := 0
	for _, d := range depts {
		if d.OwnerEmail != "" {
			deptsWithOwner++
		}
	}

	// 5. Third-party data
	thirdParties, err := s.tpSvc.List(ctx)
	if err != nil {
		s.logger.Error("report: failed to get third parties", "error", err)
	}
	totalTP := len(thirdParties)
	tpWithDPA := 0
	for _, tp := range thirdParties {
		if tp.DPAStatus == governance.DPAStatusSigned {
			tpWithDPA++
		}
	}

	// 6. Purpose data
	purposes, err := s.purposeSvc.ListByTenant(ctx, tenantID)
	if err != nil {
		s.logger.Error("report: failed to get purposes", "error", err)
	}
	totalPurposes := len(purposes)
	activePurposes := 0
	for _, p := range purposes {
		if p.IsActive {
			activePurposes++
		}
	}

	// 7. Retention data
	retPolicies, err := s.retentionSvc.GetByTenant(ctx)
	if err != nil {
		s.logger.Error("report: failed to get retention policies", "error", err)
	}
	totalRetention := len(retPolicies)

	// =================================================================
	// Compute Pillar Scores
	// =================================================================

	// Consent Management (25%): % of consent sessions recorded
	consentScore := 100.0
	if totalConsent == 0 && totalPurposes > 0 {
		consentScore = 50.0 // Partial: purposes defined but no sessions yet
	} else if totalPurposes == 0 {
		consentScore = 0.0 // No purposes defined at all
	}

	// DSR Compliance (30%): % completed on time
	dsrScore := 100.0
	if totalDSRs > 0 {
		overdueCount := len(overdueDSRs)
		onTimeRate := float64(totalDSRs-overdueCount) / float64(totalDSRs)
		dsrScore = onTimeRate * 100
	}

	// Breach Management (20%): 0 breaches = 100; else reported/total
	breachScore := 100.0
	if totalBreaches > 0 {
		breachScore = float64(reportedBreaches) / float64(totalBreaches) * 100
	}

	// Data Governance (25%): average of dept ownership, DPA coverage, purposes mapped, retention defined
	govMetrics := []float64{0, 0, 0, 0}
	if totalDepts > 0 {
		govMetrics[0] = float64(deptsWithOwner) / float64(totalDepts) * 100
	} else {
		govMetrics[0] = 0
	}
	if totalTP > 0 {
		govMetrics[1] = float64(tpWithDPA) / float64(totalTP) * 100
	} else {
		govMetrics[1] = 100 // No third parties = no risk
	}
	if totalPurposes > 0 {
		govMetrics[2] = float64(activePurposes) / float64(totalPurposes) * 100
	}
	if totalPurposes > 0 && totalRetention > 0 {
		govMetrics[3] = 100 // Retention policies exist
	} else if totalPurposes > 0 {
		govMetrics[3] = 0
	}
	govScore := (govMetrics[0] + govMetrics[1] + govMetrics[2] + govMetrics[3]) / 4

	// Overall weighted score
	overall := consentScore*0.25 + dsrScore*0.30 + breachScore*0.20 + govScore*0.25

	pillars := []PillarScore{
		{Name: "Consent Management", Score: consentScore, Weight: 0.25,
			Detail: fmt.Sprintf("%d sessions, %d purposes defined", totalConsent, totalPurposes)},
		{Name: "DSR Compliance", Score: dsrScore, Weight: 0.30,
			Detail: fmt.Sprintf("%d total, %d completed, %d overdue", totalDSRs, completedDSRs, len(overdueDSRs))},
		{Name: "Breach Management", Score: breachScore, Weight: 0.20,
			Detail: fmt.Sprintf("%d total, %d open, %d reported", totalBreaches, openBreaches, reportedBreaches)},
		{Name: "Data Governance", Score: govScore, Weight: 0.25,
			Detail: fmt.Sprintf("Depts: %d/%d owned, TPs: %d/%d w/DPA, Retention: %d policies",
				deptsWithOwner, totalDepts, tpWithDPA, totalTP, totalRetention)},
	}

	// =================================================================
	// Recommendations
	// =================================================================
	var recs []Recommendation

	if len(overdueDSRs) > 0 {
		recs = append(recs, Recommendation{
			Priority: "HIGH",
			Category: "DSR Compliance",
			Message:  fmt.Sprintf("%d DSR(s) are overdue — immediate action required to avoid DPDPA penalties.", len(overdueDSRs)),
		})
	}
	if openBreaches > 0 {
		recs = append(recs, Recommendation{
			Priority: "HIGH",
			Category: "Breach Management",
			Message:  fmt.Sprintf("%d breach(es) still open — ensure CERT-In reporting within 6 hours.", openBreaches),
		})
	}
	if totalDepts > 0 && deptsWithOwner < totalDepts {
		recs = append(recs, Recommendation{
			Priority: "MEDIUM",
			Category: "Data Governance",
			Message:  fmt.Sprintf("%d department(s) without an assigned owner.", totalDepts-deptsWithOwner),
		})
	}
	if totalTP > 0 && tpWithDPA < totalTP {
		recs = append(recs, Recommendation{
			Priority: "MEDIUM",
			Category: "Data Governance",
			Message:  fmt.Sprintf("%d third-party processor(s) without a signed DPA.", totalTP-tpWithDPA),
		})
	}
	if totalPurposes > 0 && totalRetention == 0 {
		recs = append(recs, Recommendation{
			Priority: "MEDIUM",
			Category: "Data Governance",
			Message:  "No retention policies defined — define policies to comply with DPDPA data minimisation requirements.",
		})
	}
	if totalPurposes == 0 {
		recs = append(recs, Recommendation{
			Priority: "HIGH",
			Category: "Consent Management",
			Message:  "No data processing purposes defined. Define purposes before collecting consent.",
		})
	}
	if totalConsent == 0 && totalPurposes > 0 {
		recs = append(recs, Recommendation{
			Priority: "LOW",
			Category: "Consent Management",
			Message:  "Purposes defined but no consent sessions recorded. Deploy consent widgets to start collecting consent.",
		})
	}

	snapshot := &ComplianceSnapshot{
		GeneratedAt:     time.Now().UTC(),
		PeriodFrom:      from,
		PeriodTo:        to,
		OverallScore:    overall,
		Pillars:         pillars,
		Recommendations: recs,
		Summary: SnapshotSummary{
			TotalDSRs:            totalDSRs,
			CompletedDSRs:        completedDSRs,
			OverdueDSRs:          len(overdueDSRs),
			TotalBreaches:        totalBreaches,
			OpenBreaches:         openBreaches,
			TotalConsentSessions: totalConsent,
			TotalDepartments:     totalDepts,
			DepartmentsWithOwner: deptsWithOwner,
			TotalThirdParties:    totalTP,
			ThirdPartiesWithDPA:  tpWithDPA,
			TotalPurposes:        totalPurposes,
			ActivePurposes:       activePurposes,
			RetentionPolicies:    totalRetention,
		},
	}

	s.logger.Info("compliance snapshot generated",
		slog.String("tenant_id", tenantID.String()),
		slog.Float64("overall_score", overall),
	)

	return snapshot, nil
}

// =============================================================================
// ExportEntity
// =============================================================================

// ExportEntity exports a flat list of entity records in CSV or JSON format.
// Returns: data bytes, filename, content-type, error.
func (s *ReportService) ExportEntity(ctx context.Context, entity, format string) ([]byte, string, string, error) {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, "", "", types.NewForbiddenError("tenant context required")
	}

	bigPage := types.Pagination{Page: 1, PageSize: 10000}
	timestamp := time.Now().Format("20060102")

	switch entity {
	case "dsr":
		result, err := s.dsrSvc.GetDSRs(ctx, bigPage, nil, nil)
		if err != nil {
			return nil, "", "", fmt.Errorf("export dsr: %w", err)
		}
		return marshalExport(result.Items, entity, format, timestamp,
			[]string{"ID", "Type", "Status", "Subject Name", "Subject Email", "Priority", "SLA Deadline", "Created At"},
			func(d compliance.DSR) []string {
				return []string{d.ID.String(), string(d.RequestType), string(d.Status), d.SubjectName, d.SubjectEmail, d.Priority, d.SLADeadline.Format(time.RFC3339), d.CreatedAt.Format(time.RFC3339)}
			},
		)

	case "breaches":
		result, err := s.breachSvc.ListIncidents(ctx, breach.Filter{}, bigPage)
		if err != nil {
			return nil, "", "", fmt.Errorf("export breaches: %w", err)
		}
		return marshalExport(result.Items, entity, format, timestamp,
			[]string{"ID", "Title", "Severity", "Status", "Detected At", "Type", "Affected Count"},
			func(b breach.BreachIncident) []string {
				return []string{b.ID.String(), b.Title, string(b.Severity), string(b.Status), b.DetectedAt.Format(time.RFC3339), b.Type, fmt.Sprintf("%d", b.AffectedDataSubjectCount)}
			},
		)

	case "consent-records":
		result, err := s.consentSvc.ListSessionsByTenant(ctx, consent.ConsentSessionFilters{}, bigPage)
		if err != nil {
			return nil, "", "", fmt.Errorf("export consent-records: %w", err)
		}
		return marshalExport(result.Items, entity, format, timestamp,
			[]string{"ID", "Widget ID", "Subject ID", "IP Address", "Created At"},
			func(c consent.ConsentSession) []string {
				return []string{c.ID.String(), c.WidgetID.String(), c.SubjectID.String(), c.IPAddress, c.CreatedAt.Format(time.RFC3339)}
			},
		)

	case "audit-logs":
		result, err := s.auditSvc.ListByTenant(ctx, tenantID, audit.AuditFilters{}, bigPage)
		if err != nil {
			return nil, "", "", fmt.Errorf("export audit-logs: %w", err)
		}
		return marshalExport(result.Items, entity, format, timestamp,
			[]string{"ID", "Action", "Resource Type", "Resource ID", "User ID", "IP Address", "Created At"},
			func(a audit.AuditLog) []string {
				return []string{a.ID.String(), a.Action, a.ResourceType, a.ResourceID.String(), a.UserID.String(), a.IPAddress, a.CreatedAt.Format(time.RFC3339)}
			},
		)

	case "departments":
		items, err := s.deptSvc.List(ctx)
		if err != nil {
			return nil, "", "", fmt.Errorf("export departments: %w", err)
		}
		return marshalExport(items, entity, format, timestamp,
			[]string{"ID", "Name", "Owner Name", "Owner Email", "Is Active", "Created At"},
			func(d governance.Department) []string {
				return []string{d.ID.String(), d.Name, d.OwnerName, d.OwnerEmail, fmt.Sprintf("%t", d.IsActive), d.CreatedAt.Format(time.RFC3339)}
			},
		)

	case "third-parties":
		items, err := s.tpSvc.List(ctx)
		if err != nil {
			return nil, "", "", fmt.Errorf("export third-parties: %w", err)
		}
		return marshalExport(items, entity, format, timestamp,
			[]string{"ID", "Name", "Type", "Country", "DPA Status", "Contact Name", "Contact Email"},
			func(tp governance.ThirdParty) []string {
				return []string{tp.ID.String(), tp.Name, string(tp.Type), tp.Country, tp.DPAStatus, tp.ContactName, tp.ContactEmail}
			},
		)

	case "purposes":
		items, err := s.purposeSvc.ListByTenant(ctx, tenantID)
		if err != nil {
			return nil, "", "", fmt.Errorf("export purposes: %w", err)
		}
		return marshalExport(items, entity, format, timestamp,
			[]string{"ID", "Code", "Name", "Legal Basis", "Retention Days", "Is Active", "Requires Consent"},
			func(p governance.Purpose) []string {
				return []string{p.ID.String(), p.Code, p.Name, string(p.LegalBasis), fmt.Sprintf("%d", p.RetentionDays), fmt.Sprintf("%t", p.IsActive), fmt.Sprintf("%t", p.RequiresConsent)}
			},
		)

	default:
		return nil, "", "", types.NewValidationError("unsupported entity: "+entity, nil)
	}
}

// marshalExport converts a slice of items to CSV or JSON bytes.
func marshalExport[T any](
	items []T,
	entity, format, timestamp string,
	headers []string,
	toRow func(T) []string,
) ([]byte, string, string, error) {
	switch format {
	case "csv":
		var buf bytes.Buffer
		w := csv.NewWriter(&buf)
		_ = w.Write(headers)
		for _, item := range items {
			_ = w.Write(toRow(item))
		}
		w.Flush()
		if err := w.Error(); err != nil {
			return nil, "", "", fmt.Errorf("csv write: %w", err)
		}
		filename := fmt.Sprintf("%s_%s.csv", entity, timestamp)
		return buf.Bytes(), filename, "text/csv", nil

	case "json", "":
		data, err := json.MarshalIndent(items, "", "  ")
		if err != nil {
			return nil, "", "", fmt.Errorf("json marshal: %w", err)
		}
		filename := fmt.Sprintf("%s_%s.json", entity, timestamp)
		return data, filename, "application/json", nil

	default:
		return nil, "", "", types.NewValidationError("unsupported format: "+format+", use csv or json", nil)
	}
}
