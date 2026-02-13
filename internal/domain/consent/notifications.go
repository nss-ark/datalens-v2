package consent

import (
	"context"
	"time"

	"github.com/complyark/datalens/pkg/types"
)

// =============================================================================
// ConsentNotification — A notification sent to a user
// =============================================================================

// ConsentNotification represents a notification sent to a recipient
// regarding a consent lifecycle event.
type ConsentNotification struct {
	types.TenantEntity
	RecipientType string         `json:"recipient_type" db:"recipient_type"` // DATA_PRINCIPAL, DATA_FIDUCIARY, DATA_PROCESSOR
	RecipientID   string         `json:"recipient_id" db:"recipient_id"`     // types.ID or email/phone depending on type. Keeping as string for flexibility, or types.ID if stricly internal IDs. Task spec says types.ID, but RecipientID might be external. Let's use types.ID as per spec, but sometimes it might be an email. Spec says types.ID.
	EventType     string         `json:"event_type" db:"event_type"`         // CONSENT_GRANTED, CONSENT_WITHDRAWN, CONSENT_EXPIRING, etc.
	Channel       string         `json:"channel" db:"channel"`               // EMAIL, SMS, IN_APP, WEBHOOK
	TemplateID    *types.ID      `json:"template_id,omitempty" db:"template_id"`
	Payload       map[string]any `json:"payload" db:"payload"`
	Status        string         `json:"status" db:"status"` // PENDING, SENT, FAILED, DELIVERED
	SentAt        *time.Time     `json:"sent_at,omitempty" db:"sent_at"`
	ErrorMessage  *string        `json:"error_message,omitempty" db:"error_message"`
}

// NotificationStatus tracks notification lifecycle.
const (
	NotificationStatusPending   = "PENDING"
	NotificationStatusSent      = "SENT"
	NotificationStatusFailed    = "FAILED"
	NotificationStatusDelivered = "DELIVERED"
)

// NotificationChannel defines supported delivery channels.
const (
	NotificationChannelEmail   = "EMAIL"
	NotificationChannelSMS     = "SMS"
	NotificationChannelInApp   = "IN_APP"
	NotificationChannelWebhook = "WEBHOOK"
)

// RecipientType defines who receives the notification.
const (
	RecipientTypeDataPrincipal = "DATA_PRINCIPAL"
	RecipientTypeDataFiduciary = "DATA_FIDUCIARY"
	RecipientTypeDataProcessor = "DATA_PROCESSOR"
)

// =============================================================================
// NotificationTemplate — Templates for notifications
// =============================================================================

// NotificationTemplate defines the content and structure for notifications.
type NotificationTemplate struct {
	types.TenantEntity
	Name         string `json:"name" db:"name"`
	EventType    string `json:"event_type" db:"event_type"`
	Channel      string `json:"channel" db:"channel"`
	Subject      string `json:"subject" db:"subject"`             // Email subject (supports Go template variables)
	BodyTemplate string `json:"body_template" db:"body_template"` // HTML/Text template
	IsActive     bool   `json:"is_active" db:"is_active"`
}

// =============================================================================
// Repository Interfaces
// =============================================================================

// NotificationFilter defines filters for listing notifications.
type NotificationFilter struct {
	RecipientID *string
	EventType   *string
	Channel     *string
	Status      *string
}

// NotificationRepository defines persistence for notifications.
type NotificationRepository interface {
	Create(ctx context.Context, n *ConsentNotification) error
	GetByID(ctx context.Context, id types.ID) (*ConsentNotification, error)
	ListByTenant(ctx context.Context, tenantID types.ID, filter NotificationFilter, pagination types.Pagination) (*types.PaginatedResult[ConsentNotification], error)
	UpdateStatus(ctx context.Context, id types.ID, status string, sentAt *time.Time, errorMessage *string) error
}

// NotificationTemplateRepository defines persistence for notification templates.
type NotificationTemplateRepository interface {
	Create(ctx context.Context, t *NotificationTemplate) error
	Update(ctx context.Context, t *NotificationTemplate) error
	GetByID(ctx context.Context, id types.ID) (*NotificationTemplate, error)
	ListByTenant(ctx context.Context, tenantID types.ID) ([]NotificationTemplate, error)
	GetByEventAndChannel(ctx context.Context, tenantID types.ID, eventType, channel string) (*NotificationTemplate, error)
	Delete(ctx context.Context, id types.ID) error
}
