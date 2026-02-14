// DataLens 2.0 — API Server
//
// This is the main API server entrypoint. It supports mode-based process splitting
// via the --mode flag:
//   - all     (default) — serves everything on one port (development)
//   - cc      — Control Centre API only
//   - admin   — Super Admin API only
//   - portal  — Data Principal Portal + Consent Widget public APIs only
//
// In production, run 3 instances with different modes on different ports
// for full process isolation, independent scaling, and crash domain separation.
package main

import (
	"context"
	"flag"
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
	"github.com/complyark/datalens/internal/domain/discovery"
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

	"github.com/complyark/datalens/internal/infrastructure/cache"
	"github.com/complyark/datalens/internal/infrastructure/connector"
	"github.com/complyark/datalens/internal/infrastructure/queue"
	"github.com/complyark/datalens/internal/service/ai"
	"github.com/complyark/datalens/internal/service/detection"
	govService "github.com/complyark/datalens/internal/service/governance"

	"github.com/complyark/datalens/internal/service/analytics"

	// Identity
	"github.com/complyark/datalens/internal/domain/identity"
	identityProvider "github.com/complyark/datalens/internal/infrastructure/identity/provider"
)

func main() {
	// =========================================================================
	// Command-Line Flags
	// =========================================================================

	mode := flag.String("mode", "all", "Server mode: all, cc, admin, portal")
	portOverride := flag.Int("port", 0, "Override listen port (default: from config)")
	flag.Parse()

	// Validate mode
	validModes := map[string]bool{"all": true, "cc": true, "admin": true, "portal": true}
	if !validModes[*mode] {
		fmt.Fprintf(os.Stderr, "Invalid mode %q: must be one of: all, cc, admin, portal\n", *mode)
		os.Exit(1)
	}

	// Helper: check if a component should be initialized for the current mode
	shouldInit := func(modes ...string) bool {
		for _, m := range modes {
			if *mode == m || *mode == "all" {
				return true
			}
		}
		return false
	}

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
		"mode", *mode,
	)

	// =========================================================================
	// Initialize Infrastructure (always needed)
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
	// Initialize Repositories (always needed — cheap struct wrappers)
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
	dprRepo := repository.NewDPRRequestRepo(dbPool)
	profileRepo := repository.NewDataPrincipalProfileRepo(dbPool)
	consentWidgetRepo := repository.NewConsentWidgetRepo(dbPool)
	consentNoticeRepo := repository.NewPostgresNoticeRepository(dbPool)
	consentSessionRepo := repository.NewConsentSessionRepo(dbPool)
	consentHistoryRepo := repository.NewConsentHistoryRepo(dbPool)
	consentRenewalRepo := repository.NewConsentRenewalRepo(dbPool)
	policyRepo := repository.NewPostgresPolicyRepository(dbPool)

	violationRepo := repository.NewPostgresViolationRepository(dbPool)
	mappingRepo := repository.NewPostgresDataMappingRepository(dbPool)
	auditRepo := repository.NewPostgresAuditRepository(dbPool)
	breachRepo := repository.NewPostgresBreachRepository(dbPool)
	translationRepo := repository.NewPostgresConsentNoticeTranslationRepository(dbPool)
	identityProfileRepo := repository.NewIdentityProfileRepo(dbPool)
	grievanceRepo := repository.NewPostgresGrievanceRepository(dbPool)
	notificationRepo := repository.NewPostgresNotificationRepository(dbPool)
	notificationTemplateRepo := repository.NewPostgresNotificationTemplateRepository(dbPool)

	// Cache
	var consentCache cache.ConsentCache
	if rdb != nil {
		consentCache = cache.NewRedisConsentCache(rdb)
		log.Info("Redis consent cache initialized")
	} else {
		log.Warn("Redis unavailable — consent cache disabled")
	}

	// =========================================================================
	// Initialize Domain Services (conditional based on mode)
	// =========================================================================

	// --- Shared service variables (declared before conditional blocks) ---
	var authSvc *service.AuthService
	var apiKeySvc *service.APIKeyService
	var authHandler *handler.AuthHandler
	var rateLimiter *mw.RateLimiter

	// CC-mode services & handlers
	var dsHandler *handler.DataSourceHandler
	var purposeHandler *handler.PurposeHandler
	var discoveryHandler *handler.DiscoveryHandler
	var feedbackHandler *handler.FeedbackHandler
	var dashboardHandler *handler.DashboardHandler
	var dsrHandler *handler.DSRHandler
	var consentSvc *service.ConsentService
	var consentHandler *handler.ConsentHandler
	var noticeHandler *handler.NoticeHandler
	var analyticsHandler *handler.AnalyticsHandler
	var governanceHandler *handler.GovernanceHandler
	var breachHandler *handler.BreachHandler
	var m365Handler *handler.M365Handler
	var googleHandler *handler.GoogleHandler
	var identityHandler *handler.IdentityHandler
	var grievanceSvc *service.GrievanceService
	var grievanceHandler *handler.GrievanceHandler
	var notificationHandler *handler.NotificationHandler

	// Admin-mode
	var adminHandler *handler.AdminHandler

	// Portal-mode
	var portalHandler *handler.PortalHandler

	// =========================================================================
	// Auth + APIKey (CC + Admin modes)
	// =========================================================================

	if shouldInit("cc", "admin") {
		auditSvc := service.NewAuditService(auditRepo, slog.Default())

		authSvc = service.NewAuthService(
			userRepo,
			roleRepo,
			cfg.JWT.Secret,
			cfg.JWT.AccessTokenExpiry,
			cfg.JWT.RefreshTokenExpiry,
			slog.Default(),
			auditSvc,
		)
		tenantSvc := service.NewTenantService(tenantRepo, userRepo, roleRepo, authSvc, slog.Default())
		apiKeySvc = service.NewAPIKeyService(dbPool, slog.Default())
		authHandler = handler.NewAuthHandler(authSvc, tenantSvc)
		rateLimiter = mw.NewRateLimiter(100, time.Minute, 200)

		log.Info("Auth services initialized", "mode", *mode)
	}

	// =========================================================================
	// CC Mode — Full Control Centre services, scanners, workers, subscribers
	// =========================================================================

	if shouldInit("cc") {
		// Audit Service for CC (may already exist from cc+admin block — create fresh for CC-only deps)
		auditSvc := service.NewAuditService(auditRepo, slog.Default())

		purposeSvc := service.NewPurposeService(purposeRepo, eb, slog.Default())
		feedbackSvc := service.NewFeedbackService(feedbackRepo, piiRepo, eb, slog.Default())
		m365AuthSvc := service.NewM365AuthService(cfg, dsRepo, eb, slog.Default())
		googleAuthSvc := service.NewGoogleAuthService(cfg, dsRepo, eb, slog.Default())

		consentSvc = service.NewConsentService(
			consentWidgetRepo,
			consentSessionRepo,
			consentHistoryRepo,
			eb,
			consentCache,
			cfg.Consent.SigningKey,
			slog.Default(),
			cfg.Consent.CacheTTL,
		)

		consentExpirySvc := service.NewConsentExpiryService(
			consentSessionRepo,
			consentRenewalRepo,
			consentHistoryRepo,
			consentWidgetRepo,
			eb,
			slog.Default(),
			consentSvc,
		)

		noticeSvc := service.NewNoticeService(consentNoticeRepo, consentWidgetRepo, eb, slog.Default())
		translationSvc := service.NewTranslationService(translationRepo, consentNoticeRepo, eb, "", "")
		dashboardSvc := service.NewDashboardService(dsRepo, piiRepo, scanRunRepo, slog.Default())
		analyticsSvc := analytics.NewConsentAnalyticsService(consentSessionRepo)

		// --- AI Gateway Wiring ---
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

		// Hugging Face (Generic HTTP)
		if cfg.AI.HuggingFace.APIKey != "" {
			aiProviders = append(aiProviders, ai.ProviderConfig{
				Name:                "huggingface",
				Type:                ai.ProviderTypeGenericHTTP,
				APIKey:              cfg.AI.HuggingFace.APIKey,
				Endpoint:            cfg.AI.HuggingFace.Endpoint + "/" + cfg.AI.HuggingFace.Model,
				RequestBodyTemplate: `{"inputs": "{{.Prompt}}", "parameters": {"max_new_tokens": {{.MaxTokens}}, "temperature": {{.Temperature}}}}`,
				ResponseContentPath: "0.generated_text",
				DefaultModel:        cfg.AI.HuggingFace.Model,
				RequestsPerMinute:   100,
				TokensPerMinute:     10000,
				Timeout:             30 * time.Second,
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

		aiRegistry, err := ai.BuildRegistryFromConfig(aiProviders)
		if err != nil {
			log.Error("Failed to build AI registry", "error", err)
			os.Exit(1)
		}

		fallbackChain := []string{cfg.AI.DefaultProvider, "openai", "anthropic", "local"}
		aiSelector := ai.NewSelector(aiRegistry, fallbackChain, slog.Default())

		var aiGateway ai.Gateway = ai.NewDefaultGateway(aiSelector, slog.Default())

		if rdb != nil {
			aiGateway = ai.NewCachedGateway(aiGateway, rdb, slog.Default(), cfg.AI)
			log.Info("AI Gateway: Caching and Budgeting enabled")
		} else {
			log.Warn("AI Gateway: Caching disabled (Redis unavailable)")
		}

		// Parsing Service (for File Uploads / OCR)
		parsingSvc := ai.NewParsingService(slog.Default())
		defer func() {
			if p, ok := parsingSvc.(interface{ Close() error }); ok {
				p.Close()
			}
		}()

		// Detector (Strategy Composer)
		detector := detection.NewDefaultDetector(aiGateway)

		// Connector Registry
		connRegistry := connector.NewConnectorRegistry(cfg, detector, parsingSvc)
		connRegistry.Register(types.DataSourceMicrosoft365, func() discovery.Connector {
			return connector.NewM365Connector(detector)
		})
		log.Info("Connector registry initialized", "supported_types", connRegistry.SupportedTypes())

		dsSvc := service.NewDataSourceService(dsRepo, connRegistry, eb, slog.Default())

		// Discovery Service
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

		// Governance Context Engine
		templateLoader, err := templates.NewLoader()
		if err != nil {
			log.Error("Failed to initialize template loader", "error", err)
			os.Exit(1)
		}
		contextEngine := govService.NewContextEngine(templateLoader, aiGateway, slog.Default())

		// Policy Engine
		policySvc := service.NewPolicyService(
			policyRepo,
			violationRepo,
			mappingRepo,
			dsRepo,
			piiRepo,
			eb,
			auditSvc,
			slog.Default(),
		)

		// Lineage Engine
		lineageRepo := repository.NewPostgresLineageRepository(dbPool)
		lineageSvc := service.NewLineageService(lineageRepo, dsRepo, eb, slog.Default())

		// Identity Architecture
		digiLockerProvider := identityProvider.NewDigiLockerProvider(
			cfg.Identity.DigiLocker.ClientID,
			cfg.Identity.DigiLocker.ClientSecret,
			cfg.Identity.DigiLocker.RedirectURI,
		)
		identitySvc := service.NewIdentityService(
			identityProfileRepo,
			[]identity.IdentityProvider{digiLockerProvider},
			slog.Default(),
		)

		// Grievance Redressal
		grievanceSvc = service.NewGrievanceService(grievanceRepo, eb, slog.Default())

		// Notification System
		clientRepo := service.NewPostgresClientRepository(dbPool)
		notificationSvc := service.NewNotificationService(notificationRepo, notificationTemplateRepo, clientRepo, slog.Default())

		// Breach Management
		breachSvc := service.NewBreachService(breachRepo, profileRepo, notificationSvc, auditSvc, eb, slog.Default())

		// --- Scan Orchestrator ---
		scanQueue, err := queue.NewNATSScanQueue(natsConn, slog.Default())
		if err != nil {
			log.Error("Failed to initialize scan queue", "error", err)
			os.Exit(1)
		}

		scanSvc := service.NewScanService(scanRunRepo, dsRepo, scanQueue, discoverySvc, slog.Default())

		// Start Scan Worker
		go func() {
			if err := scanSvc.StartWorker(context.Background()); err != nil {
				log.Error("Scan worker failed", "error", err)
			}
		}()

		// Scan Scheduler
		schedulerSvc := service.NewSchedulerService(dsRepo, tenantRepo, policySvc, scanSvc, consentExpirySvc, slog.Default())
		if err := schedulerSvc.Start(context.Background()); err != nil {
			log.Error("Failed to start scan scheduler", "error", err)
		}
		log.Info("Scan scheduler started")

		// --- DSR Execution Queue ---
		dsrQueue, err := queue.NewNATSDSRQueue(natsConn, slog.Default())
		if err != nil {
			log.Error("Failed to initialize DSR queue", "error", err)
			os.Exit(1)
		}

		dsrSvc := service.NewDSRService(dsrRepo, dsRepo, dsrQueue, dprRepo, eb, auditSvc, slog.Default())

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

		// --- Event Subscribers ---
		auditSub := subscriber.NewAuditSubscriber(dbPool, slog.Default())
		if _, err := auditSub.Register(context.Background(), eb); err != nil {
			log.Error("Failed to register audit subscriber", "error", err)
			os.Exit(1)
		}
		log.Info("Audit subscriber registered")

		notificationSub := service.NewNotificationSubscriber(notificationSvc, breachSvc, eb, slog.Default())
		if err := notificationSub.Start(context.Background()); err != nil {
			log.Error("Failed to start notification subscriber", "error", err)
			os.Exit(1)
		}
		log.Info("Notification subscriber started")

		if consentCache != nil {
			consentCacheSub := service.NewConsentCacheSubscriber(consentCache, eb, slog.Default(), cfg.Consent.CacheTTL)
			if err := consentCacheSub.Start(context.Background()); err != nil {
				log.Error("Failed to start consent cache subscriber", "error", err)
			} else {
				log.Info("Consent cache subscriber started")
			}
		}

		// --- CC Handlers ---
		dsHandler = handler.NewDataSourceHandler(dsSvc, scanSvc)
		purposeHandler = handler.NewPurposeHandler(purposeSvc)
		discoveryHandler = handler.NewDiscoveryHandler(discoverySvc, scanSvc, inventoryRepo, entityRepo, fieldRepo)
		feedbackHandler = handler.NewFeedbackHandler(feedbackSvc)
		dashboardHandler = handler.NewDashboardHandler(dashboardSvc)
		dsrHandler = handler.NewDSRHandler(dsrSvc, dsrExecutor)
		consentHandler = handler.NewConsentHandler(consentSvc, consentExpirySvc)
		noticeHandler = handler.NewNoticeHandler(noticeSvc, translationSvc)
		analyticsHandler = handler.NewAnalyticsHandler(analyticsSvc)
		governanceHandler = handler.NewGovernanceHandler(contextEngine, policySvc, lineageSvc)
		breachHandler = handler.NewBreachHandler(breachSvc)
		m365Handler = handler.NewM365Handler(m365AuthSvc)
		googleHandler = handler.NewGoogleHandler(googleAuthSvc)
		identityHandler = handler.NewIdentityHandler(identitySvc)
		grievanceHandler = handler.NewGrievanceHandler(grievanceSvc)
		notificationHandler = handler.NewNotificationHandler(notificationSvc)

		log.Info("CC services and handlers initialized")
	}

	// =========================================================================
	// Admin Mode — Cross-tenant admin operations
	// =========================================================================

	if shouldInit("admin") {
		adminSvc := service.NewAdminService(tenantRepo, userRepo, roleRepo, dsrRepo, service.NewTenantService(tenantRepo, userRepo, roleRepo, authSvc, slog.Default()), slog.Default())
		adminHandler = handler.NewAdminHandler(adminSvc)

		log.Info("Admin services initialized")
	}

	// =========================================================================
	// Portal Mode — Data Principal Portal + Consent Widget public APIs
	// =========================================================================

	if shouldInit("portal") {
		portalAuthSvc := service.NewPortalAuthService(
			profileRepo,
			rdb,
			cfg.Portal.JWTSecret,
			cfg.Portal.JWTExpiry,
			slog.Default(),
		)
		dataPrincipalSvc := service.NewDataPrincipalService(
			profileRepo,
			dprRepo,
			dsrRepo,
			consentHistoryRepo,
			eb,
			rdb,
			slog.Default(),
		)
		portalHandler = handler.NewPortalHandler(portalAuthSvc, dataPrincipalSvc)

		// Portal also needs ConsentService for widget APIs (if not already initialized by CC mode)
		if consentSvc == nil {
			consentSvc = service.NewConsentService(
				consentWidgetRepo,
				consentSessionRepo,
				consentHistoryRepo,
				eb,
				consentCache,
				cfg.Consent.SigningKey,
				slog.Default(),
				cfg.Consent.CacheTTL,
			)
		}
		if consentHandler == nil {
			consentHandler = handler.NewConsentHandler(consentSvc, nil)
		}

		// Portal also needs GrievanceService for portal grievances (if not already initialized by CC mode)
		if grievanceSvc == nil {
			grievanceSvc = service.NewGrievanceService(grievanceRepo, eb, slog.Default())
		}
		if grievanceHandler == nil {
			grievanceHandler = handler.NewGrievanceHandler(grievanceSvc)
		}

		log.Info("Portal services initialized")
	}

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
		AllowedOrigins:   cfg.CORS.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Tenant-ID", "X-API-Key"},
		ExposedHeaders:   []string{"Link", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// --- Health Check (always mounted) ---
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","version":"2.0.0-alpha","mode":"` + *mode + `"}`))
	})

	// --- Mount Routes Based on Mode ---

	// Shared routes (CC + Admin need auth endpoints)
	if shouldInit("cc", "admin") {
		mountSharedRoutes(r, authHandler)
	}

	// CC routes
	if shouldInit("cc") {
		mountCCRoutes(r, authSvc, apiKeySvc, rateLimiter,
			dsHandler, purposeHandler, authHandler,
			discoveryHandler, feedbackHandler, dashboardHandler,
			dsrHandler, consentHandler, noticeHandler,
			analyticsHandler, governanceHandler, breachHandler,
			m365Handler, googleHandler, identityHandler,
			grievanceHandler, notificationHandler,
		)
	}

	// Admin routes
	if shouldInit("admin") {
		mountAdminRoutes(r, authSvc, apiKeySvc, rateLimiter, adminHandler)
	}

	// Portal routes
	if shouldInit("portal") {
		mountPortalRoutes(r, consentHandler, portalHandler, grievanceHandler, consentWidgetRepo)
	}

	// =========================================================================
	// Start Server
	// =========================================================================

	listenPort := cfg.App.Port
	if *portOverride > 0 {
		listenPort = *portOverride
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", listenPort),
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

	log.Info("API server listening", "addr", srv.Addr, "mode", *mode)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error("Server failed", "error", err)
		os.Exit(1)
	}

	log.Info("Server stopped gracefully")
}
