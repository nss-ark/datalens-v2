// DataLens 2.0 — Development Seed Data
//
// Populates a freshly migrated database with demo data for local development.
// Usage: go run ./scripts/seed.go
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"

	"github.com/complyark/datalens/internal/config"
	"github.com/complyark/datalens/pkg/logging"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	log := logging.New(cfg.App.Env, cfg.App.LogLevel).
		WithComponent("seed")

	connStr := cfg.DB.DSN()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		log.Error("Failed to connect", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Error("Ping failed", "error", err)
		os.Exit(1)
	}
	log.Info("Connected to database")

	// ── Seed Tenant ──
	log.Info("Seeding demo tenant...")
	var tenantID string
	err = pool.QueryRow(ctx, `
		INSERT INTO tenants (name, domain, industry, country, plan, status, settings)
		VALUES ('Acme Corp', 'acme.local', 'FINTECH', 'IN', 'PROFESSIONAL', 'ACTIVE',
			'{"default_regulation":"DPDPA","enabled_regulations":["DPDPA","GDPR"],"retention_days":365,"enable_ai":false}')
		ON CONFLICT (domain) DO UPDATE SET name = EXCLUDED.name
		RETURNING id
	`).Scan(&tenantID)
	if err != nil {
		log.Error("Failed to seed tenant", "error", err)
		os.Exit(1)
	}
	log.Info("Tenant seeded", "id", tenantID)

	// ── Seed Admin User ──
	log.Info("Seeding admin user...")
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	var userID string
	err = pool.QueryRow(ctx, `
		INSERT INTO users (tenant_id, email, name, password, status)
		VALUES ($1, 'admin@acme.local', 'Acme Admin', $2, 'ACTIVE')
		ON CONFLICT (tenant_id, email) DO UPDATE SET name = EXCLUDED.name
		RETURNING id
	`, tenantID, string(hash)).Scan(&userID)
	if err != nil {
		log.Error("Failed to seed admin", "error", err)
		os.Exit(1)
	}
	log.Info("Admin user seeded", "id", userID, "email", "admin@acme.local")

	// ── Assign ADMIN role ──
	_, _ = pool.Exec(ctx, `
		INSERT INTO user_roles (user_id, role_id)
		SELECT $1, r.id FROM roles r WHERE r.name = 'ADMIN' AND r.is_system = TRUE
		ON CONFLICT DO NOTHING
	`, userID)

	// ── Seed Data Sources ──
	log.Info("Seeding data sources...")
	sources := []struct {
		name, typ, host, db string
		port                int
	}{
		{"Production PostgreSQL", "POSTGRESQL", "db.acme.internal", "acme_prod", 5432},
		{"Analytics MySQL", "MYSQL", "analytics.acme.internal", "analytics", 3306},
		{"Customer MongoDB", "MONGODB", "mongo.acme.internal", "customers", 27017},
	}
	for _, s := range sources {
		_, err := pool.Exec(ctx, `
			INSERT INTO data_sources (tenant_id, name, type, host, port, database_name, status)
			VALUES ($1, $2, $3, $4, $5, $6, 'DISCONNECTED')
			ON CONFLICT DO NOTHING
		`, tenantID, s.name, s.typ, s.host, s.port, s.db)
		if err != nil {
			log.Error("Failed to seed data source", "name", s.name, "error", err)
		}
	}
	log.Info("Data sources seeded", "count", len(sources))

	// ── Seed Purposes ──
	log.Info("Seeding purposes...")
	purposes := []struct {
		code, name, basis string
		retention         int
		consent           bool
	}{
		{"MARKETING", "Marketing Analytics", "CONSENT", 180, true},
		{"SERVICE_DELIVERY", "Service Delivery", "CONTRACT", 365, false},
		{"LEGAL_COMPLIANCE", "Legal & Regulatory", "LEGAL_OBLIGATION", 2555, false},
		{"HR_MANAGEMENT", "Employee Data Management", "CONTRACT", 730, false},
	}
	for _, p := range purposes {
		_, err := pool.Exec(ctx, `
			INSERT INTO purposes (tenant_id, code, name, legal_basis, retention_days, requires_consent, is_active)
			VALUES ($1, $2, $3, $4, $5, $6, TRUE)
			ON CONFLICT (tenant_id, code) DO NOTHING
		`, tenantID, p.code, p.name, p.basis, p.retention, p.consent)
		if err != nil {
			log.Error("Failed to seed purpose", "code", p.code, "error", err)
		}
	}
	log.Info("Purposes seeded", "count", len(purposes))

	fmt.Println("\n  ✅ Seed data applied successfully!")
	fmt.Println("  ─────────────────────────────────")
	fmt.Printf("  Tenant:   Acme Corp (%s)\n", tenantID)
	fmt.Printf("  Admin:    admin@acme.local / password123\n")
	fmt.Printf("  Sources:  %d data sources\n", len(sources))
	fmt.Printf("  Purposes: %d purposes\n", len(purposes))
	fmt.Println()
}
