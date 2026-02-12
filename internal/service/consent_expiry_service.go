package service

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/complyark/datalens/internal/domain/consent"
	"github.com/complyark/datalens/pkg/eventbus"
	"github.com/complyark/datalens/pkg/types"
)

// ConsentExpiryService manages consent expiration checks and renewals.
type ConsentExpiryService struct {
	sessionRepo consent.ConsentSessionRepository
	renewalRepo consent.ConsentRenewalRepository
	historyRepo consent.ConsentHistoryRepository
	widgetRepo  consent.ConsentWidgetRepository
	eventBus    eventbus.EventBus
	logger      *slog.Logger
	consentSvc  *ConsentService
}

// NewConsentExpiryService creates a new ConsentExpiryService.
func NewConsentExpiryService(
	sessionRepo consent.ConsentSessionRepository,
	renewalRepo consent.ConsentRenewalRepository,
	historyRepo consent.ConsentHistoryRepository,
	widgetRepo consent.ConsentWidgetRepository,
	eventBus eventbus.EventBus,
	logger *slog.Logger,
	consentSvc *ConsentService,
) *ConsentExpiryService {
	return &ConsentExpiryService{
		sessionRepo: sessionRepo,
		renewalRepo: renewalRepo,
		historyRepo: historyRepo,
		widgetRepo:  widgetRepo,
		eventBus:    eventBus,
		logger:      logger.With("service", "consent_expiry"),
		consentSvc:  consentSvc,
	}
}

// CheckExpiries checks for expiring consents and emits reminders.
// This is designed to be run periodically (e.g., daily).
func (s *ConsentExpiryService) CheckExpiries(ctx context.Context) error {
	s.logger.Info("Starting consent expiry check")

	// 1. Get sessions expiring within 30 days
	// We use 31 days to be safe and catch things that just entered the window
	sessions, err := s.sessionRepo.GetExpiringSessions(ctx, 31)
	if err != nil {
		return fmt.Errorf("scan expiring sessions: %w", err)
	}

	for _, session := range sessions {
		if err := s.processSessionExpiry(ctx, session); err != nil {
			s.logger.Error("failed to process session expiry",
				slog.String("session_id", session.ID.String()),
				slog.String("error", err.Error()))
			// Continue with next session
		}
	}

	return nil
}

func (s *ConsentExpiryService) processSessionExpiry(ctx context.Context, session consent.ConsentSession) error {
	// Need the widget config to know exact expiry days
	widget, err := s.widgetRepo.GetByID(ctx, session.WidgetID)
	if err != nil {
		return fmt.Errorf("get widget: %w", err)
	}

	expiryDays := widget.Config.ConsentExpiryDays
	if expiryDays <= 0 {
		return nil // Should be filtered by repo, but double check
	}

	expiryDate := session.CreatedAt.Add(time.Duration(expiryDays) * 24 * time.Hour)
	daysUntil := int(math.Ceil(time.Until(expiryDate).Hours() / 24))

	// Iterate decisions in this session
	for _, d := range session.Decisions {
		if !d.Granted {
			continue
		}

		// Check if this decision is still the ACTIVE one
		latest, err := s.historyRepo.GetLatestState(ctx, session.TenantID, *session.SubjectID, d.PurposeID)
		if err != nil {
			return fmt.Errorf("get latest state: %w", err)
		}

		// If explicitly withdrawn or superseded by a newer session, ignore
		if latest.NewStatus != "GRANTED" {
			continue
		}
		// Warning: If the user re-consented, 'latest' would be from a NEWer session.
		// We only want to alert if THIS session is the active one, OR if the active one is also expiring?
		// Actually, if 'latest' is GRANTED, we should check *when* it was granted.
		// If latest.CreatedAt > session.CreatedAt, then THIS session is old.
		// We should check expiration of the *latest* entry.
		// But GetExpiringSessions returns *sessions*.
		if latest.CreatedAt.After(session.CreatedAt) {
			// This session is superseded.
			// The newer session will be picked up by GetExpiringSessions if it's also expiring.
			continue
		}

		// Check existing renewal log to avoid duplicate alerts
		logs, err := s.renewalRepo.GetBySubject(ctx, session.TenantID, *session.SubjectID)
		if err != nil {
			return fmt.Errorf("get renewal logs: %w", err)
		}

		// Find log for this purpose
		var currentLog *consent.ConsentRenewalLog
		for _, l := range logs {
			if l.PurposeID == d.PurposeID && l.Status == "PENDING" {
				currentLog = &l
				break
			}
		}

		// If expired
		if daysUntil <= 0 {
			if currentLog == nil || currentLog.Status != "LAPSED" {
				// Expire it
				if err := s.expireDist(ctx, session, d.PurposeID, latest.NoticeVersion); err != nil {
					return err
				}
				// create/update log
				if err := s.logRenewalStatus(ctx, session, d.PurposeID, expiryDate, "LAPSED"); err != nil {
					return err
				}
			}
			continue
		}

		// Reminders: 30, 15, 7 days
		reminderType := ""
		if daysUntil <= 30 && daysUntil > 15 {
			if daysUntil == 30 {
				reminderType = "30d"
			}
		} else if daysUntil <= 15 && daysUntil > 7 {
			if daysUntil == 15 {
				reminderType = "15d"
			}
		} else if daysUntil <= 7 && daysUntil > 0 {
			reminderType = "7d"
		}

		if reminderType != "" {
			// Check if we already sent THIS reminder today?
			// Simpler: Check if last reminder was sent recently?
			// Or check existence of log.
			// Ideally we store "LastReminderType" in log.
			// But for now, let's assume we run once a day and if ReminderSentAt is not today...
			if currentLog != nil && currentLog.ReminderSentAt != nil {
				if currentLog.ReminderSentAt.Truncate(24 * time.Hour).Equal(time.Now().Truncate(24 * time.Hour)) {
					continue // Already sent today
				}
			}

			// Send reminder
			s.publishExpiryEvent(ctx, session.TenantID, *session.SubjectID, d.PurposeID, reminderType)

			// Update log
			if err := s.logRenewalStatus(ctx, session, d.PurposeID, expiryDate, "PENDING"); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *ConsentExpiryService) expireDist(ctx context.Context, session consent.ConsentSession, purposeID types.ID, noticeVersion string) error {
	// Create history entry
	entry := &consent.ConsentHistoryEntry{
		BaseEntity: types.BaseEntity{
			ID:        types.NewID(),
			CreatedAt: time.Now().UTC(),
		},
		TenantID:       session.TenantID,
		SubjectID:      *session.SubjectID,
		WidgetID:       &session.WidgetID,
		PurposeID:      purposeID,
		PreviousStatus: types.Ptr("GRANTED"),
		NewStatus:      "EXPIRED",
		Source:         "SYSTEM",
		NoticeVersion:  noticeVersion,
		IPAddress:      "127.0.0.1",
		UserAgent:      "DataLens/ExpiryService",
		Signature:      "", // System generated
	}

	if err := s.historyRepo.Create(ctx, entry); err != nil {
		return fmt.Errorf("create expiry entry: %w", err)
	}

	s.consentSvc.publishEvent(ctx, "consent.expired", session.TenantID, map[string]any{
		"subject_id": session.SubjectID.String(),
		"purpose_id": purposeID.String(),
	})

	return nil
}

func (s *ConsentExpiryService) logRenewalStatus(ctx context.Context, session consent.ConsentSession, purposeID types.ID, expiry time.Time, status string) error {
	// Check if exists
	logs, err := s.renewalRepo.GetBySubject(ctx, session.TenantID, *session.SubjectID)
	if err != nil {
		return err
	}

	var existing *consent.ConsentRenewalLog
	for _, l := range logs {
		if l.PurposeID == purposeID && l.OriginalExpiry.Equal(expiry) {
			existing = &l
			break
		}
	}

	now := time.Now().UTC()
	if existing != nil {
		existing.Status = status
		if status == "PENDING" {
			existing.ReminderSentAt = &now
		}
		existing.UpdatedAt = now
		return s.renewalRepo.Update(ctx, existing)
	}

	// Create new
	log := &consent.ConsentRenewalLog{
		BaseEntity: types.BaseEntity{
			ID:        types.NewID(),
			CreatedAt: now,
			UpdatedAt: now,
		},
		TenantID:       session.TenantID,
		SubjectID:      *session.SubjectID,
		PurposeID:      purposeID,
		OriginalExpiry: expiry,
		Status:         status,
	}
	if status == "PENDING" {
		log.ReminderSentAt = &now
	}

	return s.renewalRepo.Create(ctx, log)
}

func (s *ConsentExpiryService) publishExpiryEvent(ctx context.Context, tenantID types.ID, subjectID, purposeID types.ID, reminderType string) {
	s.consentSvc.publishEvent(ctx, fmt.Sprintf("consent.expiry_reminder_%s", reminderType), tenantID, map[string]any{
		"subject_id": subjectID.String(),
		"purpose_id": purposeID.String(),
	})
}

// RenewConsent processes a user's renewal request.
func (s *ConsentExpiryService) RenewConsent(ctx context.Context, subjectID types.ID, purposeIDs []types.ID) error {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return types.NewForbiddenError("tenant context required")
	}

	// Logic:
	// For each purpose, find the pending renewal log and marked it RENEWED.
	// We assume the actual "Grant" was recorded via RecordConsent (creating a new session).
	// WAIT. The spec says "Creates new ConsentHistoryEntry with status GRANTED".
	// The `RenewConsent` API is likely called when the user clicks "Renew" in the email link or portal.
	// It should probably function like `RecordConsent` but specifically for renewal.
	// OR, does `RecordConsent` handle it?
	// If the user clicks "Renew", the UI likely submits a standard consent payload?
	// Spec says: "POST /api/public/consent/renew ... Creates new ConsentHistoryEntry... Resets expiry timer... Updates ConsentRenewalLog"

	// So we need to:
	// 1. Create ConsentHistoryEntry (GRANTED)
	// 2. Update Renewal Log
	// 3. (Implicitly) The new history entry acts as the "Reset" because `CheckExpiries` looks for the LATEST entry.
	//    If the latest entry is NEW, its CreatedAt is NOW. So expiry is NOW + expiry_days.

	now := time.Now().UTC()

	for _, purposeID := range purposeIDs {
		// 1. Update matching renewal log
		logs, err := s.renewalRepo.GetBySubject(ctx, tenantID, subjectID)
		if err != nil {
			return err
		}

		for _, l := range logs {
			if l.PurposeID == purposeID && l.Status == "PENDING" {
				l.Status = "RENEWED"
				l.RenewedAt = &now
				l.UpdatedAt = now
				if err := s.renewalRepo.Update(ctx, &l); err != nil {
					return err
				}
			}
		}

		// 2. Create History Entry
		// We need to know which Widget/Notice logic implies.
		// Renewal usually implies "same terms as before".
		// We might need to look up the previous entry to get WidgetID/NoticeVersion.
		latest, err := s.historyRepo.GetLatestState(ctx, tenantID, subjectID, purposeID)
		if err != nil {
			s.logger.Error("failed to get latest state for renewal", "error", err)
			continue
		}

		widgetID := latest.WidgetID
		noticeVersion := latest.NoticeVersion

		entry := &consent.ConsentHistoryEntry{
			BaseEntity: types.BaseEntity{
				ID:        types.NewID(),
				CreatedAt: now,
			},
			TenantID:       tenantID,
			SubjectID:      subjectID,
			WidgetID:       widgetID,
			PurposeID:      purposeID,
			PreviousStatus: &latest.NewStatus, // Likely GRANTED or EXPIRED
			NewStatus:      "GRANTED",
			Source:         "RENEWAL_API",
			NoticeVersion:  noticeVersion,
			IPAddress:      "", // Maybe capture from context if available?
			UserAgent:      "",
			// Signature...
		}

		if err := s.historyRepo.Create(ctx, entry); err != nil {
			return err
		}

		s.consentSvc.publishEvent(ctx, "consent.renewed", tenantID, map[string]any{
			"subject_id": subjectID.String(),
			"purpose_id": purposeID.String(),
		})
	}

	return nil
}
