package types

import "context"

// ContextKey is a type for context keys to avoid collisions.
type ContextKey string

const (
	ContextKeyUserID    ContextKey = "user_id"
	ContextKeyTenantID  ContextKey = "tenant_id"
	ContextKeyEmail     ContextKey = "email"
	ContextKeyName      ContextKey = "name"
	ContextKeyRoles     ContextKey = "roles"
	ContextKeyWidgetID  ContextKey = "widget_id"
	ContextKeyIP        ContextKey = "ip"
	ContextKeyUserAgent ContextKey = "user_agent"
)

// UserIDFromContext extracts the user ID from the request context.
func UserIDFromContext(ctx context.Context) (ID, bool) {
	id, ok := ctx.Value(ContextKeyUserID).(ID)
	return id, ok
}

// TenantIDFromContext extracts the tenant ID from the request context.
func TenantIDFromContext(ctx context.Context) (ID, bool) {
	id, ok := ctx.Value(ContextKeyTenantID).(ID)
	return id, ok
}
