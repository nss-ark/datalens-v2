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
	"github.com/complyark/datalens/internal/domain/governance/templates"
	"github.com/complyark/datalens/internal/handler"
	mw "github.com/complyark/datalens/internal/middleware"
	"github.com/complyark/datalens/internal/repository"
	"github.com/complyark/datalens/internal/service"
	"github.com/complyark/datalens/internal/subscriber"
	"github.com/complyark/datalens/pkg/database"
	"github.com/complyark/datalens/pkg/eventbus"
	"github.com/complyark/datalens/pkg/logging"
	"github.com/complyark/datalens/pkg/types"

	"github.com/complyark/datalens/internal/infrastructure/connector"
	"github.com/complyark/datalens/internal/infrastructure/queue"
	"github.com/complyark/datalens/internal/service/ai"
	"github.com/complyark/datalens/internal/service/detection"
	govService "github.com/complyark/datalens/internal/service/governance"
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

	// NATS Connection
	natsConn, err := eventbus.Connect(cfg.NATS.URL, slog.Default())
	if err != nil {
		log.Error("Failed to connect to NATS", "error", err)
		os.Exit(1)
	}
	defer natsConn.Close()

	// Event Bus
	eb, err := eventbus.NewNATSEventBus(natsConn, slog.Default())
	if err != nil {
		log.Error("Failed to initialize event bus", "error", err)
		os.Exit(1)
	}
	defer eb.Close()

	// Redis client
	rdb, err := database.NewRedis(cfg.Redis)
	if err != nil {
		log.Error("Failed to connect to Redis", "error", err)
		// We don't exit here; caching is optional/resilient
	} else {
		defer rdb.Close()
		log.Info("Redis connected")
	}

	// =========================================================================
	// Initialize Repositories
	// =========================================================================

	dsRepo := repository.NewDataSourceRepo(dbPool)
	purposeRepo := repository.NewPurposeRepo(dbPool)
	tenantRepo := repository.NewTenantRepo(dbPool)
	userRepo := repository.NewUserRepo(dbPool)
	roleRepo := repository.NewRoleRepo(dbPool)
	inventoryRepo := repository.NewDataInventoryRepo(dbPool)
	entityRepo := repository.NewDataEntityRepo(dbPool)
	fieldRepo := repository.NewDataFieldRepo(dbPool)
	piiRepo := repository.NewPIIClassificationRepo(dbPool)
	feedbackRepo := repository.NewDetectionFeedbackRepo(dbPool)
	scanRunRepo := repository.NewScanRunRepo(dbPool)
	dsrRepo := repository.NewDSRRepo(dbPool)
	consentWidgetRepo := repository.NewConsentWidgetRepo(dbPool)
	consentSessionRepo := repository.NewConsentSessionRepo(dbPool)
	consentHistoryRepo := repository.NewConsentHistoryRepo(dbPool)
	policyRepo := repository.NewPostgresPolicyRepository(dbPool)
	violationRepo := repository.NewPostgresViolationRepository(dbPool)
	mappingRepo := repository.NewPostgresDataMappingRepository(dbPool)

	// =========================================================================
	// Initialize Domain Services
	// =========================================================================

	dsSvc := service.NewDataSourceService(dsRepo, eb, slog.Default())
	purposeSvc := service.NewPurposeService(purposeRepo, eb, slog.Default())
	authSvc := service.NewAuthService(
		userRepo,
		roleRepo,
		cfg.JWT.Secret,
		cfg.JWT.AccessTokenExpiry,
		cfg.JWT.RefreshTokenExpiry,
		slog.Default(),
	)
	tenantSvc := service.NewTenantService(tenantRepo, userRepo, roleRepo, authSvc, slog.Default())
	apiKeySvc := service.NewAPIKeyService(dbPool, slog.Default())
	feedbackSvc := service.NewFeedbackService(feedbackRepo, piiRepo, eb, slog.Default())
	var dsrSvc *service.DSRService // Will be initialized after DSR queue is created
	consentSvc := service.NewConsentService(
		consentWidgetRepo,
		consentSessionRepo,
		consentHistoryRepo,
		eb,
		cfg.Consent.SigningKey,
		slog.Default(),
	)
	dashboardSvc := service.NewDashboardService(dsRepo, piiRepo, scanRunRepo, slog.Default())

	// Connector Registry (Postgres + MySQL built-in)
	connRegistry := connector.NewConnectorRegistry()
	log.Info("Connector registry initialized", "supported_types", connRegistry.SupportedTypes())

	// --- AI Gateway Wiring ---
	// 1. Build Provider Configs
	var aiProviders []ai.ProviderConfig

	// OpenAI
	if cfg.AI.OpenAI.APIKey != "" {
		aiProviders = append(aiProviders, ai.ProviderConfig{
			Name:              "openai",
			Type:              ai.ProviderTypeOpenAICompatible,
			APIKey:            cfg.AI.OpenAI.APIKey,
			Endpoint:          "https://api.openai.com/v1",
			DefaultModel:      cfg.AI.OpenAI.Model,
			RequestsPerMinute: 500,
			TokensPerMinute:   100000,
		})
	}

	// Anthropic
	if cfg.AI.Anthropic.APIKey != "" {
		aiProviders = append(aiProviders, ai.ProviderConfig{
			Name:              "anthropic",
			Type:              ai.ProviderTypeAnthropic,
			APIKey:            cfg.AI.Anthropic.APIKey,
			DefaultModel:      cfg.AI.Anthropic.Model,
			RequestsPerMinute: 500,
			TokensPerMinute:   100000,
		})
	}

	// Local LLM (Ollama)
	aiProviders = append(aiProviders, ai.ProviderConfig{
		Name:              "local",
		Type:              ai.ProviderTypeOpenAICompatible,
		Endpoint:          cfg.AI.LocalLLM.Endpoint + "/v1",
		DefaultModel:      cfg.AI.LocalLLM.Model,
		RequestsPerMinute: 1000,
		TokensPerMinute:   1000000,
	})

	// 2. Build Registry
	aiRegistry, err := ai.BuildRegistryFromConfig(aiProviders)
	if err != nil {
		log.Error("Failed to build AI registry", "error", err)
		os.Exit(1)
	}

	// 3. Build Selector
	fallbackChain := []string{cfg.AI.DefaultProvider, "openai", "anthropic", "local"}
	aiSelector := ai.NewSelector(aiRegistry, fallbackChain, slog.Default())

	// 4. Build Core Gateway
	var aiGateway ai.Gateway = ai.NewDefaultGateway(aiSelector, slog.Default())

	// 5. Wrap with Caching & Budgeting (if Redis is available)
	if rdb != nil {
		aiGateway = ai.NewCachedGateway(aiGateway, rdb, slog.Default(), cfg.AI)
		log.Info("AI Gateway: Caching and Budgeting enabled")
	} else {
		log.Warn("AI Gateway: Caching disabled (Redis unavailable)")
	}

	// 6. Build Detector (Strategy Composer)
	// We use the "Default" detector which includes Pattern + Heuristic.
	// We add AI strategy if gateway is available.
	detector := detection.NewDefaultDetector(aiGateway) // aiGateway is ai.Gateway interface (CachedGateway implements it)

	// 7. Discovery Service
	discoverySvc := service.NewDiscoveryService(
		dsRepo,
		inventoryRepo,
		entityRepo,
		fieldRepo,
		piiRepo,
		scanRunRepo,
		connRegistry,
		detector,
		eb,
		slog.Default(),
	)

	// 7b. Governance Context Engine
	templateLoader, err := templates.NewLoader()
	if err != nil {
		log.Error("Failed to initialize template loader", "error", err)
		os.Exit(1)
	}
	contextEngine := govService.NewContextEngine(templateLoader, aiGateway, slog.Default())

	// 7c. Policy Engine
	policySvc := service.NewPolicyService(
		policyRepo,
		violationRepo,
		mappingRepo,
		dsRepo,
		piiRepo,
		eb,
		slog.Default(),
	)

	// 8. Scan Orchestrator
	// Initialize Scan Queue (NATS)
	scanQueue, err := queue.NewNATSScanQueue(natsConn, slog.Default())
	if err != nil {
		log.Error("Failed to initialize scan queue", "error", err)
		os.Exit(1)
	}

	// Initialize Scan Service
	scanSvc := service.NewScanService(scanRunRepo, dsRepo, scanQueue, discoverySvc, slog.Default())

	// Start Scan Worker
	go func() {
		if err := scanSvc.StartWorker(context.Background()); err != nil {
			log.Error("Scan worker failed", "error", err)
		}
	}()

	// 8b. Scan Scheduler
	schedulerSvc := service.NewSchedulerService(dsRepo, tenantRepo, policySvc, scanSvc, slog.Default())
	if err := schedulerSvc.Start(context.Background()); err != nil {
		log.Error("Failed to start scan scheduler", "error", err)
	}
	log.Info("Scan scheduler started")

	// 9. DSR Execution Queue & Service
	// Initialize DSR Queue (NATS)
	dsrQueue, err := queue.NewNATSDSRQueue(natsConn, slog.Default())
	if err != nil {
		log.Error("Failed to initialize DSR queue", "error", err)
		os.Exit(1)
	}

	// Update DSR Service with queue
	dsrSvc = service.NewDSRService(dsrRepo, dsRepo, dsrQueue, eb, slog.Default())

	// Initialize DSR Executor
	dsrExecutor := service.NewDSRExecutor(dsrRepo, dsRepo, piiRepo, connRegistry, eb, slog.Default())

	// Start DSR Worker
	go func() {
		if err := dsrQueue.Subscribe(context.Background(), func(ctx context.Context, dsrID string) error {
			id, parseErr := types.ParseID(dsrID)
			if parseErr != nil {
				return fmt.Errorf("parse dsr id: %w", parseErr)
			}
			return dsrExecutor.ExecuteDSR(ctx, id)
		}); err != nil {
			log.Error("DSR worker failed", "error", err)
		}
	}()

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
	discoveryHandler := handler.NewDiscoveryHandler(discoverySvc, scanSvc, inventoryRepo, entityRepo, fieldRepo)
	feedbackHandler := handler.NewFeedbackHandler(feedbackSvc)
	dashboardHandler := handler.NewDashboardHandler(dashboardSvc)
	dsrHandler := handler.NewDSRHandler(dsrSvc, dsrExecutor) // dsrExecutor was created earlier
	consentHandler := handler.NewConsentHandler(consentSvc)
	governanceHandler := handler.NewGovernanceHandler(contextEngine, policySvc)

	// Portal Services
	portalAuthSvc := service.NewPortalAuthService(
		repository.NewDataPrincipalProfileRepo(dbPool),
		rdb,
		cfg.Portal.JWTSecret,
		cfg.Portal.JWTExpiry,
		slog.Default(),
	)
	dataPrincipalSvc := service.NewDataPrincipalService(
		repository.NewDataPrincipalProfileRepo(dbPool),
		repository.NewDPRRequestRepo(dbPool),
		dsrRepo,
		consentHistoryRepo,
		eb,
		slog.Default(),
	)
	portalHandler := handler.NewPortalHandler(portalAuthSvc, dataPrincipalSvc)

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
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Tenant-ID", "X-API-Key"},
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
	r.Route("/api/public", func(r chi.Router) {
		// Consent Widget API (Public, Widget Key Auth)
		r.Route("/consent", func(r chi.Router) {
			r.Use(mw.WidgetAuthMiddleware(consentWidgetRepo))
			r.Use(mw.WidgetCORSMiddleware())
			r.Mount("/", consentHandler.PublicRoutes())
		})

		// Portal API (Public + Portal JWT Auth)
		r.Mount("/portal", portalHandler.Routes())
	})

	r.Route("/api/v2", func(r chi.Router) {

		// Public routes (no auth required)
		r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"pong":true}`))
		})
		r.Mount("/auth", authHandler.Routes())

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

			// Discovery (inventories, entities, fields)
			r.Mount("/discovery", discoveryHandler.Routes())

			// Detection Feedback (verify/correct/reject PII classifications)
			r.Mount("/discovery/feedback", feedbackHandler.Routes())

			// PII Classifications
			r.Route("/classifications", func(r chi.Router) {
				// TODO: Wire PII classification handlers (Sprint 2)
			})

			// DSR
			r.Mount("/dsr", dsrHandler.Routes())

			// Dashboard
			r.Mount("/dashboard", dashboardHandler.Routes())

			// Consent
			r.Mount("/consent", consentHandler.Routes())

			// Audit
			r.Route("/audit", func(r chi.Router) {
				// TODO: Wire audit log handlers (Sprint 2)
			})

			// Governance
			r.Mount("/governance", governanceHandler.Routes())

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
