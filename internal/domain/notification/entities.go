// Package notification defines the domain entities for multi-channel
// notification delivery including email, SMS, in-app, and webhooks.
package notification

import (
	"context"
	"time"

	"github.com/complyark/datalens/pkg/types"
)

// =============================================================================
// Notification — A message to be delivered
// =============================================================================

// Notification represents a message to be sent via one or more channels.
type Notification struct {
	types.TenantEntity
	Type       NotificationType `json:"type" db:"type"`
	Channel    Channel          `json:"channel" db:"channel"`
	Recipient  string           `json:"recipient" db:"recipient"`
	Subject    string           `json:"subject" db:"subject"`
	Body       string           `json:"body" db:"body"`
	Status     DeliveryStatus   `json:"status" db:"status"`
	SentAt     *time.Time       `json:"sent_at,omitempty" db:"sent_at"`
	Error      *string          `json:"error,omitempty" db:"error"`
	RetryCount int              `json:"retry_count" db:"retry_count"`
	Metadata   types.Metadata   `json:"metadata,omitempty" db:"metadata"`
}

// NotificationType classifies the notification.
type NotificationType string

const (
	NotificationDSRCreated      NotificationType = "DSR_CREATED"
	NotificationDSRCompleted    NotificationType = "DSR_COMPLETED"
	NotificationDSROverdue      NotificationType = "DSR_OVERDUE"
	NotificationConsentExpiring NotificationType = "CONSENT_EXPIRING"
	NotificationBreachDetected  NotificationType = "BREACH_DETECTED"
	NotificationPolicyViolation NotificationType = "POLICY_VIOLATION"
	NotificationScanCompleted   NotificationType = "SCAN_COMPLETED"
	NotificationSystemAlert     NotificationType = "SYSTEM_ALERT"
)

// Channel defines the delivery mechanism.
type Channel string

const (
	ChannelEmail   Channel = "EMAIL"
	ChannelSMS     Channel = "SMS"
	ChannelInApp   Channel = "IN_APP"
	ChannelWebhook Channel = "WEBHOOK"
	ChannelSlack   Channel = "SLACK"
	ChannelTeams   Channel = "TEAMS"
)

// DeliveryStatus tracks notification delivery.
type DeliveryStatus string

const (
	DeliveryPending   DeliveryStatus = "PENDING"
	DeliverySent      DeliveryStatus = "SENT"
	DeliveryDelivered DeliveryStatus = "DELIVERED"
	DeliveryFailed    DeliveryStatus = "FAILED"
)

// =============================================================================
// NotificationTemplate — Reusable message templates
// =============================================================================

// NotificationTemplate defines a reusable message format.
type NotificationTemplate struct {
	types.BaseEntity
	TenantID     *types.ID        `json:"tenant_id,omitempty" db:"tenant_id"` // nil = system template
	Name         string           `json:"name" db:"name"`
	Type         NotificationType `json:"type" db:"type"`
	Channel      Channel          `json:"channel" db:"channel"`
	Subject      string           `json:"subject" db:"subject"`
	BodyTemplate string           `json:"body_template" db:"body_template"`
	Language     string           `json:"language" db:"language"`
	IsSystem     bool             `json:"is_system" db:"is_system"`
}

// =============================================================================
// WebhookConfig — Outbound webhook configuration
// =============================================================================

// WebhookConfig defines an outbound webhook endpoint.
type WebhookConfig struct {
	types.TenantEntity
	Name       string   `json:"name" db:"name"`
	URL        string   `json:"url" db:"url"`
	Secret     string   `json:"-" db:"secret"`
	Events     []string `json:"events" db:"events"`
	IsActive   bool     `json:"is_active" db:"is_active"`
	MaxRetries int      `json:"max_retries" db:"max_retries"`
}

// =============================================================================
// Repository Interfaces
// =============================================================================

// NotificationRepository defines persistence for notifications.
type NotificationRepository interface {
	Create(ctx context.Context, n *Notification) error
	GetPending(ctx context.Context, limit int) ([]Notification, error)
	Update(ctx context.Context, n *Notification) error
}

// WebhookConfigRepository defines persistence for webhook configs.
type WebhookConfigRepository interface {
	Create(ctx context.Context, wh *WebhookConfig) error
	GetByTenant(ctx context.Context, tenantID types.ID) ([]WebhookConfig, error)
	GetByEvent(ctx context.Context, tenantID types.ID, eventType string) ([]WebhookConfig, error)
	Update(ctx context.Context, wh *WebhookConfig) error
	Delete(ctx context.Context, id types.ID) error
}
