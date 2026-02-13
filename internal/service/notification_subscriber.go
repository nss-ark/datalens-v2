package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/complyark/datalens/internal/domain/breach"
	"github.com/complyark/datalens/internal/domain/consent"
	"github.com/complyark/datalens/pkg/eventbus"
	"github.com/complyark/datalens/pkg/types"
)

type NotificationSubscriber struct {
	notificationService *NotificationService
	breachService       *BreachService
	eventBus            eventbus.EventBus
	logger              *slog.Logger
}

func NewNotificationSubscriber(
	notificationService *NotificationService,
	breachService *BreachService,
	eventBus eventbus.EventBus,
	logger *slog.Logger,
) *NotificationSubscriber {
	return &NotificationSubscriber{
		notificationService: notificationService,
		breachService:       breachService,
		eventBus:            eventBus,
		logger:              logger.With("service", "notification_subscriber"),
	}
}

// Start subscribes to relevant events
func (s *NotificationSubscriber) Start(ctx context.Context) error {
	topics := []string{
		"consent.granted",
		"consent.withdrawn",
		"consent.expiry_reminder", // Emitted by consent_expiry_service
		"consent.notice_published",
		// Add more as needed
	}

	for _, topic := range topics {
		if _, err := s.eventBus.Subscribe(ctx, topic, s.handleEvent); err != nil {
			return fmt.Errorf("subscribe %s: %w", topic, err)
		}
		s.logger.Info("subscribed to event", "topic", topic)
	}

	// Breach events
	breachTopics := []string{
		"breach.incident_created",
		"breach.incident_updated",
	}
	for _, topic := range breachTopics {
		if _, err := s.eventBus.Subscribe(ctx, topic, s.handleBreachEvent); err != nil {
			return fmt.Errorf("subscribe %s: %w", topic, err)
		}
		s.logger.Info("subscribed to breach event", "topic", topic)
	}

	return nil
}

func (s *NotificationSubscriber) handleEvent(ctx context.Context, event eventbus.Event) error {
	// 1. Parse payload
	// Payload in event.Data is any (likely map[string]any or struct)
	var payloadMap map[string]any

	// If it's already a map, use it. If it's a struct, we might need to marshal-unmarshal to satisfy map[string]any expectation of service
	switch v := event.Data.(type) {
	case map[string]any:
		payloadMap = v
	default:
		// Fallback: marshal/unmarshal
		dataBytes, err := json.Marshal(v)
		if err != nil {
			s.logger.Error("failed to marshal event data", "error", err)
			return err
		}
		if err := json.Unmarshal(dataBytes, &payloadMap); err != nil {
			s.logger.Error("failed to unmarshal payload to map", "error", err)
			return err
		}
	}

	// 2. Determine Recipient
	recipientType := consent.RecipientTypeDataPrincipal
	var recipientID string

	// Try to extract SubjectID/Email from payload
	// We need a way to reliably get the recipient.
	// For now, looking for "subject_id", "sub", or "email".
	if subID, ok := payloadMap["subject_id"]; ok {
		recipientID = fmt.Sprintf("%v", subID)
	} else if sub, ok := payloadMap["sub"]; ok {
		recipientID = fmt.Sprintf("%v", sub)
	} else if email, ok := payloadMap["email"]; ok {
		recipientID = fmt.Sprintf("%v", email)
	}

	if recipientID == "" {
		s.logger.Warn("skipping event: no recipient found in payload", "event_type", event.Type)
		return nil
	}

	// 3. Dispatch
	// Use event.TenantID if present, otherwise try payload
	tenantID := event.TenantID
	if tenantID == (types.ID{}) {
		if tID, ok := payloadMap["tenant_id"]; ok {
			parsedID, err := types.ParseID(fmt.Sprintf("%v", tID))
			if err == nil {
				tenantID = parsedID
			}
		}
	}

	if tenantID == (types.ID{}) {
		s.logger.Warn("skipping event: no tenant_id found", "event_type", event.Type)
		return nil
	}

	// Inject TenantID into context for the service
	// context.Context is immutable, so we create a new one.
	// However, DispatchNotification takes tenantID as argument too.
	// But it calls store.Create which might need context if using row-level security or similar.
	// We'll pass it explicitly.

	// Note: DispatchNotification logic currently takes tenantID as argument.

	err := s.notificationService.DispatchNotification(
		ctx,
		event.Type,
		tenantID,
		recipientType,
		recipientID,
		payloadMap,
	)

	if err != nil {
		s.logger.Error("failed to dispatch notification", "error", err, "event", event.Type)
		return err
	}

	return nil
}

func (s *NotificationSubscriber) handleBreachEvent(ctx context.Context, event eventbus.Event) error {
	// Parse incident
	var incident breach.BreachIncident
	dataBytes, err := json.Marshal(event.Data)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(dataBytes, &incident); err != nil {
		return err
	}

	// Check severity
	shouldNotify := false
	if incident.Severity == breach.SeverityHigh || incident.Severity == breach.SeverityCritical {
		shouldNotify = true
	}

	if event.Type == "breach.incident_created" {
		if shouldNotify {
			s.logger.Info("triggering breach notification for created incident", "id", incident.ID)
			return s.breachService.NotifyDataPrincipals(ctx, incident.ID)
		}
	} else if event.Type == "breach.incident_updated" {
		if shouldNotify {
			s.logger.Info("triggering breach notification for updated incident (potential escalation)", "id", incident.ID)
			return s.breachService.NotifyDataPrincipals(ctx, incident.ID)
		}
	}

	return nil
}
