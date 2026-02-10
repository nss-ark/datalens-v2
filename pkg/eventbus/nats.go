// Package eventbus provides a NATS-based implementation of the EventBus interface.
package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/nats-io/nats.go"
)

// NATSEventBus implements EventBus using NATS as the messaging transport.
type NATSEventBus struct {
	conn   *nats.Conn
	logger *slog.Logger
	mu     sync.Mutex
	subs   []*nats.Subscription
}

// NewNATSEventBus creates a new NATS-backed event bus.
func NewNATSEventBus(url string, logger *slog.Logger) (*NATSEventBus, error) {
	conn, err := nats.Connect(url,
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(10),
		nats.ReconnectBufSize(8*1024*1024),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			if err != nil {
				logger.Warn("NATS disconnected", "error", err)
			}
		}),
		nats.ReconnectHandler(func(_ *nats.Conn) {
			logger.Info("NATS reconnected")
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("connect to NATS: %w", err)
	}

	logger.Info("NATS event bus connected", "url", url)

	return &NATSEventBus{
		conn:   conn,
		logger: logger.With("component", "eventbus"),
	}, nil
}

// Publish sends an event to all subscribers of the event type.
func (b *NATSEventBus) Publish(_ context.Context, event Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	// Use event type as subject: "datasource.created" → "datalens.datasource.created"
	subject := "datalens." + event.Type

	if err := b.conn.Publish(subject, data); err != nil {
		return fmt.Errorf("publish event: %w", err)
	}

	b.logger.Debug("event published", "type", event.Type, "id", event.ID, "tenant_id", event.TenantID)
	return nil
}

// natsSubscription wraps a NATS subscription to implement the Subscription interface.
type natsSubscription struct {
	sub *nats.Subscription
}

func (s *natsSubscription) Unsubscribe() error {
	return s.sub.Unsubscribe()
}

// Subscribe registers a handler for events matching the pattern.
// Patterns use "." as separator: "datasource.*" matches all datasource events.
func (b *NATSEventBus) Subscribe(_ context.Context, pattern string, handler EventHandler) (Subscription, error) {
	// Convert our pattern format to NATS subject format
	subject := "datalens." + pattern

	// Replace "*" wildcards with NATS ">" for suffix wildcard
	if strings.HasSuffix(subject, ".*") {
		subject = strings.TrimSuffix(subject, ".*") + ".>"
	}

	sub, err := b.conn.Subscribe(subject, func(msg *nats.Msg) {
		var event Event
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			b.logger.Error("unmarshal event", "error", err, "subject", msg.Subject)
			return
		}

		if err := handler(context.Background(), event); err != nil {
			b.logger.Error("handle event", "error", err, "type", event.Type, "id", event.ID)
		}
	})
	if err != nil {
		return nil, fmt.Errorf("subscribe to %s: %w", pattern, err)
	}

	b.mu.Lock()
	b.subs = append(b.subs, sub)
	b.mu.Unlock()

	b.logger.Info("subscribed to events", "pattern", pattern, "subject", subject)
	return &natsSubscription{sub: sub}, nil
}

// Close drains and closes the NATS connection.
func (b *NATSEventBus) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, sub := range b.subs {
		_ = sub.Unsubscribe()
	}
	b.subs = nil

	if err := b.conn.Drain(); err != nil {
		b.conn.Close()
		return fmt.Errorf("drain NATS: %w", err)
	}

	b.logger.Info("NATS event bus closed")
	return nil
}

// Compile-time check.
var _ EventBus = (*NATSEventBus)(nil)
