package service

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/complyark/datalens/internal/domain/consent"
	"github.com/complyark/datalens/internal/infrastructure/cache"
	"github.com/complyark/datalens/pkg/eventbus"
	"github.com/complyark/datalens/pkg/types"
)

// =============================================================================
// Request / Response Types
// =============================================================================

// CreateWidgetRequest holds input for creating a consent widget.
type CreateWidgetRequest struct {
	Name           string               `json:"name"`
	Type           string               `json:"type"`
	Domain         string               `json:"domain"`
	Config         consent.WidgetConfig `json:"config"`
	AllowedOrigins []string             `json:"allowed_origins"`
}

// UpdateWidgetRequest holds input for updating a consent widget.
type UpdateWidgetRequest struct {
	Name           *string               `json:"name,omitempty"`
	Type           *string               `json:"type,omitempty"`
	Domain         *string               `json:"domain,omitempty"`
	Config         *consent.WidgetConfig `json:"config,omitempty"`
	AllowedOrigins *[]string             `json:"allowed_origins,omitempty"`
}

// RecordConsentRequest holds input for recording consent decisions.
type RecordConsentRequest struct {
	WidgetID      types.ID                  `json:"widget_id"`
	SubjectID     *types.ID                 `json:"subject_id,omitempty"`
	Decisions     []consent.ConsentDecision `json:"decisions"`
	IPAddress     string                    `json:"ip_address"`
	UserAgent     string                    `json:"user_agent"`
	PageURL       string                    `json:"page_url"`
	NoticeVersion string                    `json:"notice_version"`
}

// WithdrawConsentRequest holds input for withdrawing consent.
type WithdrawConsentRequest struct {
	SubjectID     types.ID `json:"subject_id"`
	PurposeID     types.ID `json:"purpose_id"`
	PurposeName   string   `json:"purpose_name"`
	Source        string   `json:"source"`
	IPAddress     string   `json:"ip_address"`
	UserAgent     string   `json:"user_agent"`
	NoticeVersion string   `json:"notice_version"`
}

// =============================================================================
// ConsentService
// =============================================================================

// ConsentService implements consent widget management and consent lifecycle.
type ConsentService struct {
	widgetRepo  consent.ConsentWidgetRepository
	sessionRepo consent.ConsentSessionRepository
	historyRepo consent.ConsentHistoryRepository

	eventBus   eventbus.EventBus
	cache      cache.ConsentCache
	signingKey string
	logger     *slog.Logger
	cacheTTL   time.Duration
}

// NewConsentService creates a new ConsentService.
func NewConsentService(
	widgetRepo consent.ConsentWidgetRepository,
	sessionRepo consent.ConsentSessionRepository,
	historyRepo consent.ConsentHistoryRepository,
	eventBus eventbus.EventBus,
	cache cache.ConsentCache,
	signingKey string,
	logger *slog.Logger,
	cacheTTL time.Duration,
) *ConsentService {
	return &ConsentService{
		widgetRepo:  widgetRepo,
		sessionRepo: sessionRepo,
		historyRepo: historyRepo,

		eventBus:   eventBus,
		cache:      cache,
		signingKey: signingKey,
		logger:     logger.With("service", "consent"),
		cacheTTL:   cacheTTL,
	}
}

// =============================================================================
// Widget CRUD
// =============================================================================

// CreateWidget creates a new consent widget with a generated API key.
func (s *ConsentService) CreateWidget(ctx context.Context, req CreateWidgetRequest) (*consent.ConsentWidget, error) {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, types.NewForbiddenError("tenant context required")
	}

	if req.Name == "" {
		return nil, types.NewValidationError("name is required", nil)
	}

	// Generate API key (32 random bytes → 64 hex chars)
	apiKey, err := generateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("generate api key: %w", err)
	}

	now := time.Now().UTC()
	widget := &consent.ConsentWidget{
		TenantEntity: types.TenantEntity{
			BaseEntity: types.BaseEntity{
				ID:        types.NewID(),
				CreatedAt: now,
				UpdatedAt: now,
			},
			TenantID: tenantID,
		},
		Name:           req.Name,
		Type:           consent.WidgetType(req.Type),
		Domain:         req.Domain,
		Status:         consent.WidgetStatusDraft,
		Config:         req.Config,
		APIKey:         apiKey,
		AllowedOrigins: req.AllowedOrigins,
		Version:        1,
	}

	// Generate embed code snippet
	widget.EmbedCode = fmt.Sprintf(
		`<script src="https://cdn.datalens.io/widget.js" data-widget-id="%s" data-api-key="%s"></script>`,
		widget.ID.String(), apiKey,
	)

	if err := s.widgetRepo.Create(ctx, widget); err != nil {
		return nil, fmt.Errorf("create widget: %w", err)
	}

	// Publish event
	s.publishEvent(ctx, eventbus.EventConsentWidgetCreated, tenantID, widget)

	s.logger.Info("consent widget created",
		slog.String("tenant_id", tenantID.String()),
		slog.String("widget_id", widget.ID.String()),
		slog.String("name", widget.Name),
	)

	return widget, nil
}

// UpdateWidget updates an existing consent widget.
func (s *ConsentService) UpdateWidget(ctx context.Context, id types.ID, req UpdateWidgetRequest) (*consent.ConsentWidget, error) {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, types.NewForbiddenError("tenant context required")
	}

	widget, err := s.widgetRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if widget.TenantID != tenantID {
		return nil, types.NewNotFoundError("consent widget", id)
	}

	configChanged := false
	if req.Name != nil {
		widget.Name = *req.Name
	}
	if req.Type != nil {
		widget.Type = consent.WidgetType(*req.Type)
	}
	if req.Domain != nil {
		widget.Domain = *req.Domain
	}
	if req.Config != nil {
		widget.Config = *req.Config
		configChanged = true
	}
	if req.AllowedOrigins != nil {
		widget.AllowedOrigins = *req.AllowedOrigins
	}

	// Bump version on config change
	if configChanged {
		widget.Version++
	}

	if err := s.widgetRepo.Update(ctx, widget); err != nil {
		return nil, fmt.Errorf("update widget: %w", err)
	}

	s.logger.Info("consent widget updated",
		slog.String("tenant_id", tenantID.String()),
		slog.String("widget_id", widget.ID.String()),
		slog.Int("version", widget.Version),
	)

	return widget, nil
}

// GetWidget retrieves a consent widget by ID.
func (s *ConsentService) GetWidget(ctx context.Context, id types.ID) (*consent.ConsentWidget, error) {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, types.NewForbiddenError("tenant context required")
	}

	widget, err := s.widgetRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if widget.TenantID != tenantID {
		return nil, types.NewNotFoundError("consent widget", id)
	}

	return widget, nil
}

// ListWidgets lists all consent widgets for the current tenant.
func (s *ConsentService) ListWidgets(ctx context.Context) ([]consent.ConsentWidget, error) {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, types.NewForbiddenError("tenant context required")
	}

	return s.widgetRepo.GetByTenant(ctx, tenantID)
}

// DeleteWidget deletes a consent widget.
func (s *ConsentService) DeleteWidget(ctx context.Context, id types.ID) error {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return types.NewForbiddenError("tenant context required")
	}

	// Verify ownership before delete
	widget, err := s.widgetRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if widget.TenantID != tenantID {
		return types.NewNotFoundError("consent widget", id)
	}

	if err := s.widgetRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete widget: %w", err)
	}

	s.logger.Info("consent widget deleted",
		slog.String("tenant_id", tenantID.String()),
		slog.String("widget_id", id.String()),
	)

	return nil
}

// ActivateWidget sets a widget's status to ACTIVE.
func (s *ConsentService) ActivateWidget(ctx context.Context, id types.ID) (*consent.ConsentWidget, error) {
	return s.setWidgetStatus(ctx, id, consent.WidgetStatusActive)
}

// PauseWidget sets a widget's status to PAUSED.
func (s *ConsentService) PauseWidget(ctx context.Context, id types.ID) (*consent.ConsentWidget, error) {
	return s.setWidgetStatus(ctx, id, consent.WidgetStatusPaused)
}

func (s *ConsentService) setWidgetStatus(ctx context.Context, id types.ID, status consent.WidgetStatus) (*consent.ConsentWidget, error) {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, types.NewForbiddenError("tenant context required")
	}

	widget, err := s.widgetRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if widget.TenantID != tenantID {
		return nil, types.NewNotFoundError("consent widget", id)
	}

	widget.Status = status
	if err := s.widgetRepo.Update(ctx, widget); err != nil {
		return nil, fmt.Errorf("update widget status: %w", err)
	}

	s.logger.Info("consent widget status changed",
		slog.String("widget_id", id.String()),
		slog.String("status", string(status)),
	)

	return widget, nil
}

// =============================================================================
// Public API — Widget Config
// =============================================================================

// GetWidgetConfig retrieves widget configuration by API key (public endpoint).
func (s *ConsentService) GetWidgetConfig(ctx context.Context, apiKey string) (*consent.WidgetConfig, error) {
	widget, err := s.widgetRepo.GetByAPIKey(ctx, apiKey)
	if err != nil {
		return nil, err
	}

	if widget.Status != consent.WidgetStatusActive {
		return nil, types.NewNotFoundError("consent widget", nil)
	}

	return &widget.Config, nil
}

// =============================================================================
// Consent Session Capture
// =============================================================================

// RecordConsent records a consent session and creates history entries.
func (s *ConsentService) RecordConsent(ctx context.Context, req RecordConsentRequest) (*consent.ConsentSession, error) {
	// Resolve tenantID from context (set by widget middleware)
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, types.NewForbiddenError("tenant context required")
	}

	if len(req.Decisions) == 0 {
		return nil, types.NewValidationError("at least one consent decision is required", nil)
	}

	// Get the widget to read version
	widget, err := s.widgetRepo.GetByID(ctx, req.WidgetID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()

	// Build signature from canonical decision data
	signature := s.signDecisions(req.Decisions, now)

	session := &consent.ConsentSession{
		BaseEntity: types.BaseEntity{
			ID:        types.NewID(),
			CreatedAt: now,
		},
		TenantID:      tenantID,
		WidgetID:      req.WidgetID,
		SubjectID:     req.SubjectID,
		Decisions:     req.Decisions,
		IPAddress:     req.IPAddress,
		UserAgent:     req.UserAgent,
		PageURL:       req.PageURL,
		WidgetVersion: widget.Version,
		NoticeVersion: req.NoticeVersion,
		Signature:     signature,
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("create consent session: %w", err)
	}

	// Create history entries for each decision
	for _, decision := range req.Decisions {
		newStatus := "WITHDRAWN"
		if decision.Granted {
			newStatus = "GRANTED"
		}

		entry := &consent.ConsentHistoryEntry{
			BaseEntity: types.BaseEntity{
				ID:        types.NewID(),
				CreatedAt: now,
			},
			TenantID:      tenantID,
			SubjectID:     s.resolveSubjectID(req.SubjectID),
			WidgetID:      &req.WidgetID,
			PurposeID:     decision.PurposeID,
			PurposeName:   "", // Denormalized name can be filled in future lookups
			NewStatus:     newStatus,
			Source:        "BANNER",
			IPAddress:     req.IPAddress,
			UserAgent:     req.UserAgent,
			NoticeVersion: req.NoticeVersion,
			Signature:     s.signRecord(fmt.Sprintf("%s:%s:%s:%s", decision.PurposeID, newStatus, tenantID, now.Format(time.RFC3339))),
		}

		if err := s.historyRepo.Create(ctx, entry); err != nil {
			s.logger.Error("failed to create consent history entry",
				slog.String("purpose_id", decision.PurposeID.String()),
				slog.String("error", err.Error()),
			)
			// Continue — session is the source of truth; history is best-effort
		}

		// Emit events
		if decision.Granted {
			s.publishEvent(ctx, eventbus.EventConsentGranted, tenantID, map[string]any{
				"session_id": session.ID.String(),
				"purpose_id": decision.PurposeID.String(),
				"subject_id": session.SubjectID.String(),
			})
		}
	}

	s.logger.Info("consent session recorded",
		slog.String("tenant_id", tenantID.String()),
		slog.String("session_id", session.ID.String()),
		slog.Int("decisions", len(req.Decisions)),
	)

	return session, nil
}

// =============================================================================
// Consent Check
// =============================================================================

// CheckConsent checks whether consent is currently granted for a subject+purpose.
func (s *ConsentService) CheckConsent(ctx context.Context, tenantID, subjectID, purposeID types.ID) (bool, error) {
	// 1. Check Cache
	if s.cache != nil {
		cached, err := s.cache.GetConsentStatus(ctx, tenantID, subjectID, purposeID)
		if err != nil {
			s.logger.Warn("failed to get consent status from cache", "error", err)
		} else if cached != nil {
			return *cached, nil
		}
	}

	// 2. Cache Miss — Check DB
	entry, err := s.historyRepo.GetLatestState(ctx, tenantID, subjectID, purposeID)
	if err != nil {
		return false, fmt.Errorf("check consent: %w", err)
	}

	granted := false
	if entry != nil && entry.NewStatus == "GRANTED" {
		granted = true
	}

	// 3. Populate Cache
	if s.cache != nil {
		// Asynchronously populate cache to not block response?
		// No, for consistency we should just do it, it's fast.
		// Use a safe TTL (e.g. 5 mins defined in config)
		if err := s.cache.SetConsentStatus(ctx, tenantID, subjectID, purposeID, granted, s.cacheTTL); err != nil {
			s.logger.Warn("failed to set consent status in cache", "error", err)
		}
	}

	return granted, nil
}

// =============================================================================
// Consent Withdrawal
// =============================================================================

// WithdrawConsent records a consent withdrawal.
func (s *ConsentService) WithdrawConsent(ctx context.Context, req WithdrawConsentRequest) error {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return types.NewForbiddenError("tenant context required")
	}

	now := time.Now().UTC()

	entry := &consent.ConsentHistoryEntry{
		BaseEntity: types.BaseEntity{
			ID:        types.NewID(),
			CreatedAt: now,
		},
		TenantID:       tenantID,
		SubjectID:      req.SubjectID,
		PurposeID:      req.PurposeID,
		PurposeName:    req.PurposeName,
		PreviousStatus: types.Ptr("GRANTED"),
		NewStatus:      "WITHDRAWN",
		Source:         req.Source,
		IPAddress:      req.IPAddress,
		UserAgent:      req.UserAgent,
		NoticeVersion:  req.NoticeVersion,
		Signature:      s.signRecord(fmt.Sprintf("%s:WITHDRAWN:%s:%s", req.PurposeID, tenantID, now.Format(time.RFC3339))),
	}

	if err := s.historyRepo.Create(ctx, entry); err != nil {
		return fmt.Errorf("record consent withdrawal: %w", err)
	}

	// Emit withdrawal event
	s.publishEvent(ctx, eventbus.EventConsentWithdrawn, tenantID, map[string]any{
		"subject_id": req.SubjectID.String(),
		"purpose_id": req.PurposeID.String(),
	})

	s.logger.Info("consent withdrawn",
		slog.String("tenant_id", tenantID.String()),
		slog.String("purpose_id", req.PurposeID.String()),
	)

	return nil
}

// =============================================================================
// Session Listing (for internal handler)
// =============================================================================

// GetSessionsBySubject retrieves consent sessions for a subject.
func (s *ConsentService) GetSessionsBySubject(ctx context.Context, subjectID types.ID) ([]consent.ConsentSession, error) {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, types.NewForbiddenError("tenant context required")
	}

	return s.sessionRepo.GetBySubject(ctx, tenantID, subjectID)
}

// GetHistory retrieves paginated consent history for a subject.
func (s *ConsentService) GetHistory(ctx context.Context, subjectID types.ID, pagination types.Pagination) (*types.PaginatedResult[consent.ConsentHistoryEntry], error) {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, types.NewForbiddenError("tenant context required")
	}

	return s.historyRepo.GetBySubject(ctx, tenantID, subjectID, pagination)
}

// =============================================================================
// Helpers
// =============================================================================

// signDecisions creates an HMAC-SHA256 signature over the canonical consent decisions.
func (s *ConsentService) signDecisions(decisions []consent.ConsentDecision, ts time.Time) string {
	canonical := struct {
		Decisions []consent.ConsentDecision `json:"decisions"`
		Timestamp string                    `json:"timestamp"`
	}{
		Decisions: decisions,
		Timestamp: ts.Format(time.RFC3339Nano),
	}

	data, _ := json.Marshal(canonical)
	return s.signRecord(string(data))
}

// signRecord creates an HMAC-SHA256 signature.
func (s *ConsentService) signRecord(data string) string {
	mac := hmac.New(sha256.New, []byte(s.signingKey))
	mac.Write([]byte(data))
	return "sha256:" + hex.EncodeToString(mac.Sum(nil))
}

// publishEvent publishes a domain event (best-effort — never fails the caller).
func (s *ConsentService) publishEvent(ctx context.Context, eventType string, tenantID types.ID, data any) {
	event := eventbus.NewEvent(eventType, "consent", tenantID, data)
	if err := s.eventBus.Publish(ctx, event); err != nil {
		s.logger.Error("failed to publish event",
			slog.String("event_type", eventType),
			slog.String("error", err.Error()),
		)
	}
}

// generateAPIKey generates a cryptographically random 32-byte hex API key.
func generateAPIKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// resolveSubjectID returns the subject ID or a zero UUID for anonymous sessions.
func (s *ConsentService) resolveSubjectID(subjectID *types.ID) types.ID {
	if subjectID != nil {
		return *subjectID
	}
	return types.ID{} // Zero UUID for anonymous
}
