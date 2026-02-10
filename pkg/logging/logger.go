// Package logging provides structured logging for DataLens.
//
// It wraps Go's slog package with convenience methods and
// consistent field naming conventions.
package logging

import (
	"context"
	"log/slog"
	"os"
)

// Logger wraps slog.Logger with DataLens-specific conveniences.
type Logger struct {
	*slog.Logger
}

// New creates a new Logger based on the environment.
func New(env, level string) *Logger {
	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level: parseLevel(level),
	}

	switch env {
	case "production":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	default:
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return &Logger{
		Logger: slog.New(handler),
	}
}

// WithContext returns a logger with request-scoped fields.
func (l *Logger) WithContext(ctx context.Context) *Logger {
	logger := l.Logger

	if tenantID, ok := ctx.Value(ctxKeyTenantID).(string); ok {
		logger = logger.With("tenant_id", tenantID)
	}
	if requestID, ok := ctx.Value(ctxKeyRequestID).(string); ok {
		logger = logger.With("request_id", requestID)
	}
	if userID, ok := ctx.Value(ctxKeyUserID).(string); ok {
		logger = logger.With("user_id", userID)
	}

	return &Logger{Logger: logger}
}

// WithComponent returns a logger tagged with a component name.
func (l *Logger) WithComponent(name string) *Logger {
	return &Logger{
		Logger: l.Logger.With("component", name),
	}
}

// WithError returns a logger with an error field.
func (l *Logger) WithError(err error) *Logger {
	return &Logger{
		Logger: l.Logger.With("error", err.Error()),
	}
}

// =============================================================================
// Context Keys
// =============================================================================

type contextKey string

const (
	ctxKeyTenantID  contextKey = "tenant_id"
	ctxKeyRequestID contextKey = "request_id"
	ctxKeyUserID    contextKey = "user_id"
)

// ContextWithTenantID attaches a tenant ID to the context.
func ContextWithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, ctxKeyTenantID, tenantID)
}

// ContextWithRequestID attaches a request ID to the context.
func ContextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, ctxKeyRequestID, requestID)
}

// ContextWithUserID attaches a user ID to the context.
func ContextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, ctxKeyUserID, userID)
}

// =============================================================================
// Helpers
// =============================================================================

func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
