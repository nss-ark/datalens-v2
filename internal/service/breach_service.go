package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/complyark/datalens/internal/domain/breach"
	"github.com/complyark/datalens/internal/domain/consent"
	"github.com/complyark/datalens/pkg/eventbus"
	"github.com/complyark/datalens/pkg/types"
)

type CreateIncidentRequest struct {
	Title                    string                  `json:"title"`
	Description              string                  `json:"description"`
	Type                     string                  `json:"type"`
	Severity                 breach.IncidentSeverity `json:"severity"`
	DetectedAt               time.Time               `json:"detected_at"`
	OccurredAt               time.Time               `json:"occurred_at"`
	AffectedSystems          []string                `json:"affected_systems"`
	AffectedDataSubjectCount int                     `json:"affected_data_subject_count"`
	PiiCategories            []string                `json:"pii_categories"`
	PoCName                  string                  `json:"poc_name"`
	PoCRole                  string                  `json:"poc_role"`
	PoCEmail                 string                  `json:"poc_email"`
}

type UpdateIncidentRequest struct {
	Title                    *string                  `json:"title"`
	Description              *string                  `json:"description"`
	Type                     *string                  `json:"type"`
	Severity                 *breach.IncidentSeverity `json:"severity"`
	Status                   *breach.IncidentStatus   `json:"status"`
	DetectedAt               *time.Time               `json:"detected_at"`
	OccurredAt               *time.Time               `json:"occurred_at"`
	ReportedToCertInAt       *time.Time               `json:"reported_to_cert_in_at"`
	ReportedToDPBAt          *time.Time               `json:"reported_to_dpb_at"`
	ClosedAt                 *time.Time               `json:"closed_at"`
	AffectedSystems          []string                 `json:"affected_systems"`
	AffectedDataSubjectCount *int                     `json:"affected_data_subject_count"`
	PiiCategories            []string                 `json:"pii_categories"`
	PoCName                  *string                  `json:"poc_name"`
	PoCRole                  *string                  `json:"poc_role"`
	PoCEmail                 *string                  `json:"poc_email"`
}

type BreachService struct {
	repo                breach.Repository
	profileRepo         consent.DataPrincipalProfileRepository
	notificationService *NotificationService
	auditService        *AuditService
	eventBus            eventbus.EventBus
	logger              *slog.Logger
}

func NewBreachService(
	repo breach.Repository,
	profileRepo consent.DataPrincipalProfileRepository,
	notificationService *NotificationService,
	auditService *AuditService,
	eventBus eventbus.EventBus,
	logger *slog.Logger,
) *BreachService {
	return &BreachService{
		repo:                repo,
		profileRepo:         profileRepo,
		notificationService: notificationService,
		auditService:        auditService,
		eventBus:            eventBus,
		logger:              logger.With("service", "breach"),
	}
}

// Helper to convert struct to map for audit logging
func toMap(v interface{}) map[string]any {
	var m map[string]any
	b, _ := json.Marshal(v)
	json.Unmarshal(b, &m)
	return m
}

func (s *BreachService) CreateIncident(ctx context.Context, req CreateIncidentRequest) (*breach.BreachIncident, error) {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, types.NewForbiddenError("tenant context required")
	}

	userID, ok := types.UserIDFromContext(ctx)
	if !ok {
		return nil, types.NewForbiddenError("user context required")
	}

	if req.Title == "" {
		return nil, types.NewValidationError("title is required", nil)
	}
	if req.DetectedAt.IsZero() {
		return nil, types.NewValidationError("detected_at is required", nil)
	}

	incident := &breach.BreachIncident{
		BaseEntity: types.BaseEntity{
			ID:        types.NewID(),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		},
		TenantID:                 tenantID,
		Title:                    req.Title,
		Description:              req.Description,
		Type:                     req.Type,
		Severity:                 req.Severity,
		Status:                   breach.StatusOpen,
		DetectedAt:               req.DetectedAt,
		OccurredAt:               req.OccurredAt,
		AffectedSystems:          req.AffectedSystems,
		AffectedDataSubjectCount: req.AffectedDataSubjectCount,
		PiiCategories:            req.PiiCategories,
		PoCName:                  req.PoCName,
		PoCRole:                  req.PoCRole,
		PoCEmail:                 req.PoCEmail,
	}

	if req.Severity == breach.SeverityHigh || req.Severity == breach.SeverityCritical {
		incident.IsReportableToCertIn = true
		incident.IsReportableToDPB = true
	}

	if err := s.repo.Create(ctx, incident); err != nil {
		return nil, err
	}

	_ = s.eventBus.Publish(ctx, eventbus.NewEvent(eventbus.EventBreachIncidentCreated, "breach", incident.TenantID, incident))
	s.auditService.Log(ctx, userID, "CREATE_INCIDENT", "BREACH_INCIDENT", incident.ID, nil, toMap(incident), tenantID)

	return incident, nil
}

func (s *BreachService) GetIncident(ctx context.Context, id types.ID) (*breach.BreachIncident, map[string]interface{}, error) {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, nil, types.NewForbiddenError("tenant context required")
	}

	incident, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	if incident.TenantID != tenantID {
		return nil, nil, types.NewNotFoundError("breach incident", map[string]any{"id": id})
	}

	sla := s.calculateSLA(incident)
	return incident, sla, nil
}

func (s *BreachService) UpdateIncident(ctx context.Context, id types.ID, req UpdateIncidentRequest) (*breach.BreachIncident, error) {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, types.NewForbiddenError("tenant context required")
	}
	userID, ok := types.UserIDFromContext(ctx)
	if !ok {
		return nil, types.NewForbiddenError("user context required")
	}

	incident, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if incident.TenantID != tenantID {
		return nil, types.NewNotFoundError("breach incident", map[string]any{"id": id})
	}

	oldValues := toMap(incident)

	if req.Title != nil {
		incident.Title = *req.Title
	}
	if req.Description != nil {
		incident.Description = *req.Description
	}
	if req.Type != nil {
		incident.Type = *req.Type
	}
	if req.Severity != nil {
		incident.Severity = *req.Severity
		if incident.Severity == breach.SeverityHigh || incident.Severity == breach.SeverityCritical {
			incident.IsReportableToCertIn = true
			incident.IsReportableToDPB = true
		}
	}
	if req.Status != nil {
		incident.Status = *req.Status
	}
	if req.DetectedAt != nil {
		incident.DetectedAt = *req.DetectedAt
	}
	if req.OccurredAt != nil {
		incident.OccurredAt = *req.OccurredAt
	}
	if req.ReportedToCertInAt != nil {
		incident.ReportedToCertInAt = req.ReportedToCertInAt
	}
	if req.ReportedToDPBAt != nil {
		incident.ReportedToDPBAt = req.ReportedToDPBAt
	}
	if req.ClosedAt != nil {
		incident.ClosedAt = req.ClosedAt
	}
	if req.AffectedSystems != nil {
		incident.AffectedSystems = req.AffectedSystems
	}
	if req.AffectedDataSubjectCount != nil {
		incident.AffectedDataSubjectCount = *req.AffectedDataSubjectCount
	}
	if req.PiiCategories != nil {
		incident.PiiCategories = req.PiiCategories
	}
	if req.PoCName != nil {
		incident.PoCName = *req.PoCName
	}
	if req.PoCRole != nil {
		incident.PoCRole = *req.PoCRole
	}
	if req.PoCEmail != nil {
		incident.PoCEmail = *req.PoCEmail
	}

	incident.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, incident); err != nil {
		return nil, err
	}

	_ = s.eventBus.Publish(ctx, eventbus.NewEvent(eventbus.EventBreachIncidentUpdated, "breach", incident.TenantID, incident))
	s.auditService.Log(ctx, userID, "UPDATE_INCIDENT", "BREACH_INCIDENT", incident.ID, oldValues, toMap(incident), tenantID)

	return incident, nil
}

func (s *BreachService) ListIncidents(ctx context.Context, filter breach.Filter, pagination types.Pagination) (*types.PaginatedResult[breach.BreachIncident], error) {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, types.NewForbiddenError("tenant context required")
	}

	return s.repo.List(ctx, tenantID, filter, pagination)
}

func (s *BreachService) GenerateCertInReport(ctx context.Context, id types.ID) (map[string]interface{}, error) {
	incident, _, err := s.GetIncident(ctx, id)
	if err != nil {
		return nil, err
	}

	report := map[string]interface{}{
		"incident_id":      incident.ID,
		"title":            incident.Title,
		"severity":         incident.Severity,
		"date_occurred":    incident.OccurredAt,
		"date_detected":    incident.DetectedAt,
		"affected_systems": incident.AffectedSystems,
		"nature_of_breach": incident.Type,
		"poc_details": map[string]string{
			"name":  incident.PoCName,
			"role":  incident.PoCRole,
			"email": incident.PoCEmail,
		},
	}

	return report, nil
}

func (s *BreachService) calculateSLA(incident *breach.BreachIncident) map[string]interface{} {
	now := time.Now().UTC()
	certInDeadline := incident.DetectedAt.Add(6 * time.Hour)
	dpbDeadline := incident.DetectedAt.Add(72 * time.Hour)

	return map[string]interface{}{
		"time_remaining_cert_in": certInDeadline.Sub(now).String(),
		"time_remaining_dpb":     dpbDeadline.Sub(now).String(),
		"cert_in_deadline":       certInDeadline,
		"dpb_deadline":           dpbDeadline,
		"overdue_cert_in":        now.After(certInDeadline),
		"overdue_dpb":            now.After(dpbDeadline),
	}
}

// NotifyDataPrincipals triggers DPDPA §28 notifications for an incident
func (s *BreachService) NotifyDataPrincipals(ctx context.Context, id types.ID) error {
	incident, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Fetch all data principals for the tenant
	// TODO: Filter based on PII category match if possible in future
	pagination := types.Pagination{Page: 1, PageSize: 1000} // Batch processing
	for {
		result, err := s.profileRepo.ListByTenant(ctx, incident.TenantID, pagination)
		if err != nil {
			return fmt.Errorf("failed to list data principals: %w", err)
		}

		for _, profile := range result.Items {
			// Construct payload
			payload := map[string]any{
				"incident_id":        incident.ID,
				"title":              incident.Title,
				"severity":           incident.Severity,
				"occurred_at":        incident.OccurredAt,
				"description":        incident.Description, // Brief?
				"affected_data":      incident.PiiCategories,
				"what_we_are_doing":  "We have contained the incident and are investigating...",
				"contact_email":      incident.PoCEmail,
				"data_principal_id":  profile.ID,
				"data_principal_sub": profile.SubjectID,
			}

			// Use NotificationService to dispatch
			// We use profile.Email as recipient
			if err := s.notificationService.DispatchNotification(
				ctx,
				"breach.notification",
				incident.TenantID,
				consent.RecipientTypeDataPrincipal,
				profile.Email,
				payload,
			); err != nil {
				s.logger.Error("failed to dispatch breach notification", "incident_id", id, "email", profile.Email, "error", err)
				// Continue to next
			}
		}

		if result.Total <= (pagination.Page * pagination.PageSize) {
			break
		}
		pagination.Page++
	}

	return nil
}
