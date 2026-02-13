package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/complyark/datalens/internal/infrastructure/cache"
	"github.com/complyark/datalens/pkg/eventbus"
	"github.com/complyark/datalens/pkg/types"
)

// ConsentCacheSubscriber subscribes to consent events to keep the cache warm/consistent.
type ConsentCacheSubscriber struct {
	cache    cache.ConsentCache
	eventBus eventbus.EventBus
	logger   *slog.Logger
	ttl      time.Duration
}

// NewConsentCacheSubscriber creates a new ConsentCacheSubscriber.
func NewConsentCacheSubscriber(
	c cache.ConsentCache,
	eventBus eventbus.EventBus,
	logger *slog.Logger,
	ttl time.Duration,
) *ConsentCacheSubscriber {
	return &ConsentCacheSubscriber{
		cache:    c,
		eventBus: eventBus,
		logger:   logger.With("service", "consent_cache_subscriber"),
		ttl:      ttl,
	}
}

// Start subscribes to relevant events.
func (s *ConsentCacheSubscriber) Start(ctx context.Context) error {
	topics := []string{
		eventbus.EventConsentGranted,
		eventbus.EventConsentWithdrawn,
		"consent.expired", // Hardcoded as it might not be in eventbus package yet
		"consent.renewed",
	}

	for _, topic := range topics {
		if _, err := s.eventBus.Subscribe(ctx, topic, s.handleEvent); err != nil {
			return fmt.Errorf("subscribe %s: %w", topic, err)
		}
		s.logger.Info("subscribed to event", "topic", topic)
	}
	return nil
}

func (s *ConsentCacheSubscriber) handleEvent(ctx context.Context, event eventbus.Event) error {
	// Parse payload
	var payload map[string]any
	switch v := event.Data.(type) {
	case map[string]any:
		payload = v
	default:
		dataBytes, err := json.Marshal(v)
		if err != nil {
			s.logger.Error("failed to marshal event data", "error", err)
			return err
		}
		if err := json.Unmarshal(dataBytes, &payload); err != nil {
			s.logger.Error("failed to unmarshal payload", "error", err)
			return err
		}
	}

	// Extract IDs
	tenantID := event.TenantID
	if tenantID == (types.ID{}) {
		// Try to find in payload
		if t, ok := payload["tenant_id"].(string); ok {
			if parsed, err := types.ParseID(t); err == nil {
				tenantID = parsed
			}
		}
	}
	if tenantID == (types.ID{}) {
		return fmt.Errorf("missing tenant_id")
	}

	subjectIDStr, _ := payload["subject_id"].(string)
	purposeIDStr, _ := payload["purpose_id"].(string)

	if subjectIDStr == "" {
		return nil // Cannot cache without subject ID
	}

	subjectID, err := types.ParseID(subjectIDStr)
	if err != nil {
		return fmt.Errorf("invalid subject_id: %w", err)
	}

	// Handle by Event Type
	switch event.Type {
	case eventbus.EventConsentGranted, "consent.renewed":
		if purposeIDStr == "" {
			return fmt.Errorf("missing purpose_id for grant/renew")
		}
		purposeID, err := types.ParseID(purposeIDStr)
		if err != nil {
			return fmt.Errorf("invalid purpose_id: %w", err)
		}
		if err := s.cache.SetConsentStatus(ctx, tenantID, subjectID, purposeID, true, s.ttl); err != nil {
			s.logger.Error("failed to cache consent grant", "error", err)
		}

	case eventbus.EventConsentWithdrawn:
		// Withdrawn could be specific purpose or all (if purpose_id is missing/empty?)
		// Spec says "upsert cache entry as false (OR invalidate)"
		if purposeIDStr != "" {
			purposeID, err := types.ParseID(purposeIDStr)
			if err == nil {
				// Set to false
				if err := s.cache.SetConsentStatus(ctx, tenantID, subjectID, purposeID, false, s.ttl); err != nil {
					s.logger.Error("failed to cache consent withdrawal", "error", err)
				}
				return nil
			}
		}
		// Fallback: Invalidate subject if purpose is missing or invalid
		if err := s.cache.InvalidateSubject(ctx, tenantID, subjectID); err != nil {
			s.logger.Error("failed to invalidate subject cache", "error", err)
		}

	case "consent.expired":
		// Expiry might be specific purpose
		if purposeIDStr != "" {
			purposeID, err := types.ParseID(purposeIDStr)
			if err == nil {
				// Set to false (expired = no consent)
				if err := s.cache.SetConsentStatus(ctx, tenantID, subjectID, purposeID, false, s.ttl); err != nil {
					s.logger.Error("failed to cache consent expiry", "error", err)
				}
				return nil
			}
		}
		if err := s.cache.InvalidateSubject(ctx, tenantID, subjectID); err != nil {
			s.logger.Error("failed to invalidate subject cache (expiry)", "error", err)
		}
	}

	return nil
}
