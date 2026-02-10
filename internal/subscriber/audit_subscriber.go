// Package subscriber provides event bus subscribers for cross-cutting concerns.
package subscriber

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/complyark/datalens/pkg/eventbus"
)

// AuditSubscriber logs all domain events to the audit_events table.
type AuditSubscriber struct {
	pool         *pgxpool.Pool
	logger       *slog.Logger
	mu           sync.Mutex
	previousHash string
}

// NewAuditSubscriber creates and returns an audit subscriber.
func NewAuditSubscriber(pool *pgxpool.Pool, logger *slog.Logger) *AuditSubscriber {
	return &AuditSubscriber{
		pool:         pool,
		logger:       logger.With("subscriber", "audit"),
		previousHash: "",
	}
}

// Register subscribes to all events and logs them.
func (s *AuditSubscriber) Register(ctx context.Context, bus eventbus.EventBus) (eventbus.Subscription, error) {
	return bus.Subscribe(ctx, "*", s.handleEvent)
}

func (s *AuditSubscriber) handleEvent(ctx context.Context, event eventbus.Event) error {
	// Derive entity type and action from event type (e.g., "datasource.created" → "datasource", "created")
	entityType := "unknown"
	action := event.Type
	if idx := strings.IndexByte(event.Type, '.'); idx >= 0 {
		entityType = event.Type[:idx]
		action = event.Type[idx+1:]
	}

	// Extract entity ID and actor from event data
	entityID := uuid.Nil
	actorID := uuid.Nil
	var metadata []byte

	if data, ok := event.Data.(map[string]any); ok {
		if id, ok := data["id"].(string); ok {
			if parsed, err := uuid.Parse(id); err == nil {
				entityID = parsed
			}
		}
		if actor, ok := data["actor_id"].(string); ok {
			if parsed, err := uuid.Parse(actor); err == nil {
				actorID = parsed
			}
		}
		metadata, _ = json.Marshal(data)
	} else {
		metadata = []byte("{}")
	}

	// If no actor in event data, use a system actor
	if actorID == uuid.Nil {
		actorID = uuid.Nil // system action
	}

	// Compute hash chain for tamper detection
	s.mu.Lock()
	previousHash := s.previousHash
	hashInput := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s",
		event.ID, event.TenantID, event.Type, entityType, entityID, previousHash, event.Timestamp)
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(hashInput)))
	s.previousHash = hash
	s.mu.Unlock()

	query := `
		INSERT INTO audit_events (
			id, tenant_id, event_type, actor_id, actor_type,
			resource_type, resource_id, action,
			metadata, previous_hash, hash, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	_, err := s.pool.Exec(ctx, query,
		event.ID, event.TenantID, event.Type, actorID, "SYSTEM",
		entityType, entityID, action,
		metadata, previousHash, hash, event.Timestamp,
	)
	if err != nil {
		s.logger.Error("failed to write audit event", "error", err, "event_type", event.Type)
		return fmt.Errorf("write audit event: %w", err)
	}

	s.logger.Debug("audit event written", "event_type", event.Type, "entity_type", entityType)
	return nil
}
