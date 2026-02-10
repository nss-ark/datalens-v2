// DataLens 2.0 — API Server
//
// This is the main Control Centre API server entrypoint. It serves the REST API
// for the web dashboard, handles authentication, and orchestrates
// compliance operations.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"

	"github.com/complyark/datalens/internal/config"
	"github.com/complyark/datalens/internal/handler"
	mw "github.com/complyark/datalens/internal/middleware"
	"github.com/complyark/datalens/internal/repository"
	"github.com/complyark/datalens/internal/service"
	"github.com/complyark/datalens/internal/subscriber"
	"github.com/complyark/datalens/pkg/database"
	"github.com/complyark/datalens/pkg/eventbus"
	"github.com/complyark/datalens/pkg/logging"
)

func main() {
	// Load .env in development
	_ = godotenv.Load()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log := logging.New(cfg.App.Env, cfg.App.LogLevel).
		WithComponent("api")

	log.Info("Starting DataLens API server",
		"env", cfg.App.Env,
		"port", cfg.App.Port,
	)

	// =========================================================================
	// Initialize Infrastructure
	// =========================================================================

	// Database connection pool
	dbPool, err := database.New(cfg.DB)
	if err != nil {
		log.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()
	log.Info("Database connected")

	// NATS event bus
	eb, err := eventbus.NewNATSEventBus(cfg.NATS.URL, slog.Default())
	if err != nil {
		log.Error("Failed to connect to NATS", "error", err)
		os.Exit(1)
	}
	defer eb.Close()

	// =========================================================================
	// Initialize Repositories
	// =========================================================================

	dsRepo := repository.NewDataSourceRepo(dbPool)
	purposeRepo := repository.NewPurposeRepo(dbPool)
	tenantRepo := repository.NewTenantRepo(dbPool)
	userRepo := repository.NewUserRepo(dbPool)
	roleRepo := repository.NewRoleRepo(dbPool)

	// =========================================================================
	// Initialize Domain Services
	// =========================================================================

	dsSvc := service.NewDataSourceService(dsRepo, eb, slog.Default())
	purposeSvc := service.NewPurposeService(purposeRepo, eb, slog.Default())
	authSvc := service.NewAuthService(
		userRepo,
		cfg.JWT.Secret,
		cfg.JWT.AccessTokenExpiry,
		cfg.JWT.RefreshTokenExpiry,
		slog.Default(),
	)
	tenantSvc := service.NewTenantService(tenantRepo, userRepo, roleRepo, authSvc, slog.Default())

	// =========================================================================
	// Initialize Event Subscribers
	// =========================================================================

	auditSub := subscriber.NewAuditSubscriber(dbPool, slog.Default())
	if _, err := auditSub.Register(context.Background(), eb); err != nil {
		log.Error("Failed to register audit subscriber", "error", err)
		os.Exit(1)
	}
	log.Info("Audit subscriber registered")

	// =========================================================================
	// Initialize API Handlers
	// =========================================================================

	dsHandler := handler.NewDataSourceHandler(dsSvc)
	purposeHandler := handler.NewPurposeHandler(purposeSvc)
	authHandler := handler.NewAuthHandler(authSvc, tenantSvc)

	// =========================================================================
	// Rate Limiter
	// =========================================================================

	rateLimiter := mw.NewRateLimiter(100, time.Minute, 200)

	// =========================================================================
	// HTTP Router
	// =========================================================================

	r := chi.NewRouter()

	// --- Global Middleware ---
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(30 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // Restrict in production
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Tenant-ID"},
		ExposedHeaders:   []string{"Link", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// --- Health Check ---
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","version":"2.0.0-alpha"}`))
	})

	// --- API Routes ---
	r.Route("/api/v2", func(r chi.Router) {

		// Public routes (no auth required)
		r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"pong":true}`))
		})
		r.Mount("/auth", authHandler.Routes())

		// Protected routes (auth + tenant isolation + rate limiting)
		r.Group(func(r chi.Router) {
			r.Use(mw.Auth(authSvc))
			r.Use(mw.TenantIsolation())
			r.Use(rateLimiter.Middleware())

			// Data Sources
			r.Mount("/data-sources", dsHandler.Routes())

			// Purposes
			r.Mount("/purposes", purposeHandler.Routes())

			// Auth (protected: /me)
			r.Mount("/users", authHandler.ProtectedRoutes())

			// PII Classifications
			r.Route("/classifications", func(r chi.Router) {
				// TODO: Wire PII classification handlers (Sprint 2)
			})

			// DSR
			r.Route("/dsr", func(r chi.Router) {
				// TODO: Wire DSR handlers (Sprint 2)
			})

			// Consent
			r.Route("/consent", func(r chi.Router) {
				// TODO: Wire consent handlers (Sprint 3)
			})

			// Audit
			r.Route("/audit", func(r chi.Router) {
				// TODO: Wire audit log handlers (Sprint 2)
			})
		})
	})

	// =========================================================================
	// Start Server
	// =========================================================================

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.App.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigChan

		log.Info("Shutdown signal received", "signal", sig.String())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Close event bus
		if err := eb.Close(); err != nil {
			log.Error("Event bus close error", "error", err)
		}

		// Database pool auto-deferred above

		if err := srv.Shutdown(ctx); err != nil {
			log.Error("Server shutdown error", "error", err)
		}
	}()

	log.Info("API server listening", "addr", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error("Server failed", "error", err)
		os.Exit(1)
	}

	log.Info("Server stopped gracefully")
}
