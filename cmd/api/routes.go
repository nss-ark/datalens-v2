// Package main — route mounting functions for mode-based process splitting.
//
// These functions are called from main.go based on the --mode flag.
// Each function mounts a subset of routes with the appropriate middleware.
package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/complyark/datalens/internal/domain/consent"
	"github.com/complyark/datalens/internal/domain/identity"
	"github.com/complyark/datalens/internal/handler"
	mw "github.com/complyark/datalens/internal/middleware"
	"github.com/complyark/datalens/internal/service"
)

// mountSharedRoutes mounts routes that are common to CC and Admin modes.
// This includes the /api/v2/auth endpoints and /api/v2/ping.
func mountSharedRoutes(
	r chi.Router,
	authHandler *handler.AuthHandler,
) {
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"pong":true}`))
	})
	r.Mount("/auth", authHandler.Routes())
}

// mountCCRoutes mounts the Control Centre's protected API routes.
// These require JWT auth + tenant isolation + rate limiting.
func mountCCRoutes(
	r chi.Router,
	authSvc *service.AuthService,
	apiKeySvc *service.APIKeyService,
	rateLimiter *mw.RateLimiter,
	dsHandler *handler.DataSourceHandler,
	purposeHandler *handler.PurposeHandler,
	authHandler *handler.AuthHandler,
	discoveryHandler *handler.DiscoveryHandler,
	feedbackHandler *handler.FeedbackHandler,
	dashboardHandler *handler.DashboardHandler,
	dsrHandler *handler.DSRHandler,
	consentHandler *handler.ConsentHandler,
	noticeHandler *handler.NoticeHandler,
	analyticsHandler *handler.AnalyticsHandler,
	governanceHandler *handler.GovernanceHandler,
	breachHandler *handler.BreachHandler,
	m365Handler *handler.M365Handler,
	googleHandler *handler.GoogleHandler,
	identityHandler *handler.IdentityHandler,
	grievanceHandler *handler.GrievanceHandler,
	notificationHandler *handler.NotificationHandler,
	dpoHandler *handler.DPOHandler,
	auditHandler *handler.AuditHandler,
) {
	// Protected routes (auth + tenant isolation + rate limiting)
	r.Group(func(r chi.Router) {
		r.Use(mw.Auth(authSvc, apiKeySvc))
		r.Use(mw.TenantIsolation())
		r.Use(rateLimiter.Middleware())

		// Data Sources
		r.Mount("/data-sources", dsHandler.Routes())

		// Purposes
		r.Mount("/purposes", purposeHandler.Routes())

		// Auth (protected: /me)
		r.Mount("/users", authHandler.ProtectedRoutes())

		// OAuth2 Connectors
		r.Mount("/auth/m365", m365Handler.Routes())
		r.Mount("/auth/google", googleHandler.Routes())

		// Discovery (inventories, entities, fields)
		r.Mount("/discovery", discoveryHandler.Routes())

		// Detection Feedback (verify/correct/reject PII classifications)
		r.Mount("/discovery/feedback", feedbackHandler.Routes())

		// PII Classifications
		// r.Route("/classifications", func(r chi.Router) {
		// 	// TODO: Wire PII classification handlers (Sprint 2)
		// })

		// DSR
		r.Mount("/dsr", dsrHandler.Routes())

		// Dashboard
		r.Mount("/dashboard", dashboardHandler.Routes())

		// Consent
		r.Route("/consent", func(r chi.Router) {
			r.Mount("/", consentHandler.Routes())
			r.Mount("/notices", noticeHandler.Routes())
		})

		// Audit Logs
		r.Mount("/audit-logs", auditHandler.Routes())

		// Governance
		r.Mount("/governance", governanceHandler.Routes())

		// Breach
		r.Mount("/breach", breachHandler.Routes())

		// Identity
		r.Mount("/identity", identityHandler.Routes())

		// Grievances
		r.Mount("/grievances", grievanceHandler.Routes())

		// Notifications
		r.Mount("/notifications", notificationHandler.Routes())

		// Analytics
		r.Mount("/analytics", analyticsHandler.Routes())

		// Compliance (DPO, etc.)
		r.Mount("/compliance/dpo", dpoHandler.Routes())
	})
}

// mountAdminRoutes mounts the Super Admin routes.
// These require JWT auth + PlatformAdmin role.
func mountAdminRoutes(
	r chi.Router,
	authSvc *service.AuthService,
	apiKeySvc *service.APIKeyService,
	rateLimiter *mw.RateLimiter,
	adminHandler *handler.AdminHandler,
) {
	// Public Routes (Login)
	r.Mount("/superadmin", adminHandler.PublicRoutes())

	// Protected Routes (Platform Admin)
	r.Route("/admin", func(r chi.Router) {
		r.Use(mw.Auth(authSvc, apiKeySvc))
		r.Use(mw.RequireRole(identity.RolePlatformAdmin))
		r.Use(rateLimiter.Middleware())
		r.Mount("/", adminHandler.Routes())
	})
}

// mountPortalRoutes mounts the Data Principal Portal and Consent Widget public APIs.
// Portal uses OTP + short-lived JWT. Consent uses widget API keys.
func mountPortalRoutes(
	r chi.Router,
	consentHandler *handler.ConsentHandler,
	portalHandler *handler.PortalHandler,
	dpoHandler *handler.DPOHandler,
	consentWidgetRepo consent.ConsentWidgetRepository,
) {
	r.Route("/api/v2/public", func(r chi.Router) {
		// Consent Widget API (Public, Widget Key Auth)
		r.Route("/consent", func(r chi.Router) {
			r.Use(mw.WidgetAuthMiddleware(consentWidgetRepo))
			r.Use(mw.WidgetCORSMiddleware())
			r.Mount("/", consentHandler.PublicRoutes())
		})

		// Portal API (Public + Portal JWT Auth)
		// All portal routes including grievances are handled by portalHandler.Routes()
		r.Mount("/portal", portalHandler.Routes())

		// Public DPO Contact
		r.Mount("/compliance/dpo", dpoHandler.PublicRoutes())
	})
}
