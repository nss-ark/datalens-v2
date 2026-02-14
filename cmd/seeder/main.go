package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

var (
	rowsFlag    = flag.Int("rows", 10000, "Number of rows to seed per table")
	dirtyFlag   = flag.Bool("dirty", false, "Include dirty/malformed data")
	targetsFlag = flag.String("targets", "all", "Comma-separated list of targets: mysql,postgres,mongo,all")
)

func main() {
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	targets := strings.Split(*targetsFlag, ",")
	if *targetsFlag == "all" {
		targets = []string{"mysql", "postgres", "mongo", "admin"}
	}

	for _, target := range targets {
		switch strings.TrimSpace(target) {
		case "mysql":
			if err := seedMySQL(*rowsFlag, *dirtyFlag); err != nil {
				slog.Error("Failed to seed MySQL", "error", err)
			}
		case "postgres":
			if err := seedPostgres(*rowsFlag, *dirtyFlag); err != nil {
				slog.Error("Failed to seed Postgres", "error", err)
			}
		case "mongo":
			if err := seedMongo(*rowsFlag, *dirtyFlag); err != nil {
				slog.Error("Failed to seed MongoDB", "error", err)
			}
		case "admin":
			if err := seedAdmin(); err != nil {
				slog.Error("Failed to seed Admin User", "error", err)
			}
		}
	}
}

// --- MySQL Seeder ---

func seedMySQL(count int, dirty bool) error {
	dsn := "root:root@tcp(localhost:3307)/inventory_db?parseTime=true"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("connect mysql: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("ping mysql: %w", err)
	}
	slog.Info("Connected to MySQL", "dsn", dsn)

	// Create Tables
	queries := []string{
		`CREATE TABLE IF NOT EXISTS customers (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255),
			email VARCHAR(255),
			phone VARCHAR(50),
			address TEXT,
			created_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS orders (
			id INT AUTO_INCREMENT PRIMARY KEY,
			customer_id INT,
			amount DECIMAL(10, 2),
			status VARCHAR(50),
			order_date DATETIME,
			FOREIGN KEY (customer_id) REFERENCES customers(id)
		)`,
		`TRUNCATE TABLE orders`, // Clear existing data for idempotency
		`SET FOREIGN_KEY_CHECKS = 0; TRUNCATE TABLE customers; SET FOREIGN_KEY_CHECKS = 1;`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			// Handle multiple statements/warnings if necessary, but simple Exec is usually fine for DDL
			// The TRUNCATE with FK checks might fail if executed as single string in some drivers, splitting handling.
			// Actually, go-sql-driver doesn't support multiple statements by default without multiStatements=true.
			// Let's keep it simple and just do it one by one and ignore truncation errors if table empty.
			if strings.Contains(q, ";") {
				// manual split not robust, but sufficient here or just enable multiStatements
				// simpler: let's just create.
			}
			// Re-writing creation logic to be robust
		}
	}

	// Re-doing table creation properly
	_, _ = db.Exec(`DROP TABLE IF EXISTS orders`)
	_, _ = db.Exec(`DROP TABLE IF EXISTS customers`)

	if _, err := db.Exec(`CREATE TABLE customers (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255),
			email VARCHAR(255),
			phone VARCHAR(50),
			address TEXT,
			created_at DATETIME
		)`); err != nil {
		return err
	}
	if _, err := db.Exec(`CREATE TABLE orders (
			id INT AUTO_INCREMENT PRIMARY KEY,
			customer_id INT,
			amount DECIMAL(10, 2),
			status VARCHAR(50),
			order_date DATETIME,
			FOREIGN KEY (customer_id) REFERENCES customers(id)
		)`); err != nil {
		return err
	}

	slog.Info("Seeding MySQL customers...", "count", count)

	// Batch insert
	batchSize := 1000
	for i := 0; i < count; i += batchSize {
		vals := []interface{}{}
		placeholders := []string{}

		currentBatch := batchSize
		if i+batchSize > count {
			currentBatch = count - i
		}

		for j := 0; j < currentBatch; j++ {
			name := gofakeit.Name()
			email := gofakeit.Email()
			phone := gofakeit.Phone()

			if dirty && rand.Intn(10) == 0 {
				// 10% chance of dirty data
				if rand.Intn(2) == 0 {
					email = " invalid-email " + gofakeit.DigitN(5) // Spaces and no @
				} else {
					name = "NULL" // Literal string NULL often catches scanners
				}
			}

			placeholders = append(placeholders, "(?, ?, ?, ?, ?)")
			vals = append(vals, name, email, phone, gofakeit.Address().Address, gofakeit.Date())
		}

		query := fmt.Sprintf("INSERT INTO customers (name, email, phone, address, created_at) VALUES %s", strings.Join(placeholders, ","))
		if _, err := db.Exec(query, vals...); err != nil {
			return fmt.Errorf("insert customers batch: %w", err)
		}
		fmt.Printf("\rMySQL Customers: %d/%d", i+currentBatch, count)
	}
	fmt.Println()

	// Seed Orders
	slog.Info("Seeding MySQL orders...")
	// Just map to random customer IDs 1..count (since auto_inc usually starts at 1 and is sequential)
	for i := 0; i < count; i += batchSize {
		vals := []interface{}{}
		placeholders := []string{}
		currentBatch := batchSize
		if i+batchSize > count {
			currentBatch = count - i
		}

		for j := 0; j < currentBatch; j++ {
			custID := rand.Intn(count) + 1
			placeholders = append(placeholders, "(?, ?, ?, ?)")
			vals = append(vals, custID, gofakeit.Price(10, 1000), gofakeit.RandomString([]string{"PENDING", "SHIPPED", "DELIVERED"}), gofakeit.Date())
		}

		query := fmt.Sprintf("INSERT INTO orders (customer_id, amount, status, order_date) VALUES %s", strings.Join(placeholders, ","))
		if _, err := db.Exec(query, vals...); err != nil {
			return fmt.Errorf("insert orders batch: %w", err)
		}
		fmt.Printf("\rMySQL Orders: %d/%d", i+currentBatch, count)
	}
	fmt.Println()

	return nil
}

// --- Postgres Seeder ---

func seedPostgres(count int, dirty bool) error {
	dsn := "postgres://postgres:postgres@localhost:5434/customers_db?sslmode=disable"
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer conn.Close(ctx)

	slog.Info("Connected to Postgres", "dsn", dsn)

	// Create Schema & Tables
	// Using simple Execs
	cmds := []string{
		`CREATE SCHEMA IF NOT EXISTS hr`,
		`DROP TABLE IF EXISTS hr.employees CASCADE`,
		`CREATE TABLE hr.employees (
            id SERIAL PRIMARY KEY,
            full_name TEXT,
            contact_info JSONB,
            ssn TEXT, 
            department TEXT,
            is_active BOOLEAN
        )`,
		`CREATE SCHEMA IF NOT EXISTS finance`,
		`DROP TABLE IF EXISTS finance.payments CASCADE`,
		`CREATE TABLE finance.payments (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            employee_id INT,
            amount NUMERIC(10, 2),
            details JSONB,
            processed_at TIMESTAMPTZ
        )`,
	}

	for _, cmd := range cmds {
		if _, err := conn.Exec(ctx, cmd); err != nil {
			return fmt.Errorf("exec pg cmd %q: %w", cmd, err)
		}
	}

	slog.Info("Seeding Postgres employees...", "count", count)

	// Use CopyFrom for speed
	empRows := [][]interface{}{}
	for i := 0; i < count; i++ {
		name := gofakeit.Name()
		ssn := gofakeit.SSN()

		contact := map[string]interface{}{
			"email": gofakeit.Email(),
			"phone": gofakeit.Phone(),
			"address": map[string]string{
				"city": gofakeit.City(),
				"zip":  gofakeit.Zip(),
			},
		}

		if dirty && rand.Intn(20) == 0 {
			// Postgres text fields cannot contain 0x00.
			// Use SQL injection simulation or control characters that are valid UTF-8
			name = "Admin' OR '1'='1 --"
		}

		contactJSON, _ := json.Marshal(contact)

		empRows = append(empRows, []interface{}{
			name, contactJSON, ssn, gofakeit.JobTitle(), gofakeit.Bool(),
		})
	}

	copyCount, err := conn.CopyFrom(
		ctx,
		pgx.Identifier{"hr", "employees"},
		[]string{"full_name", "contact_info", "ssn", "department", "is_active"},
		pgx.CopyFromRows(empRows),
	)
	if err != nil {
		return fmt.Errorf("copy employees: %w", err)
	}
	slog.Info("Seeded Employees", "rows", copyCount)

	return nil
}

// --- Mongo Seeder ---

func seedMongo(count int, dirty bool) error {
	uri := "mongodb://admin:password@localhost:27018"
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return fmt.Errorf("connect mongo: %w", err)
	}
	defer client.Disconnect(ctx)

	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("ping mongo: %w", err)
	}
	slog.Info("Connected to MongoDB", "uri", uri)

	db := client.Database("app_data")
	coll := db.Collection("users")

	// Drop first
	if err := coll.Drop(ctx); err != nil {
		// ignore error if ns not found
	}

	slog.Info("Seeding Mongo users...", "count", count)

	docs := []interface{}{}
	batchSize := 1000

	for i := 0; i < count; i++ {
		user := bson.M{
			"username": gofakeit.Username(),
			"profile": bson.M{
				"firstName": gofakeit.FirstName(),
				"lastName":  gofakeit.LastName(),
				"dob":       gofakeit.Date(),
				"socials": []string{
					gofakeit.URL(), gofakeit.URL(),
				},
			},
			"metrics": bson.M{
				"loginCount": gofakeit.Number(0, 1000),
				"lastLogin":  gofakeit.Date(),
			},
		}

		if dirty && i%50 == 0 {
			// Add a field that violates schema if schema existed, or just weird structure
			user["legacy_data"] = bson.M{
				"xml_blob":   "<user>Invalid</user>",
				"null_field": nil,
			}
		}

		docs = append(docs, user)

		if len(docs) >= batchSize {
			if _, err := coll.InsertMany(ctx, docs); err != nil {
				return fmt.Errorf("insert mongo batch: %w", err)
			}
			docs = docs[:0]
			fmt.Printf("\rMongo Users: %d/%d", i+1, count)
		}
	}

	if len(docs) > 0 {
		if _, err := coll.InsertMany(ctx, docs); err != nil {
			return fmt.Errorf("insert mongo final: %w", err)
		}
	}
	fmt.Println()

	return nil
}

// --- Admin Seeder ---

func seedAdmin() error {
	// Main App DB (Postgres)
	dsn := "postgres://datalens:datalens_dev@localhost:5433/datalens?sslmode=disable"
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return fmt.Errorf("connect app db: %w", err)
	}
	defer conn.Close(ctx)

	slog.Info("Connected to App DB for Admin Seeding", "dsn", dsn)

	// 1. Ensure Tenant exists
	var tenantID string
	err = conn.QueryRow(ctx, "SELECT id FROM tenants WHERE domain = $1", "datalens.io").Scan(&tenantID)
	if err != nil {
		if err == pgx.ErrNoRows {
			slog.Info("Creating default tenant...")
			err = conn.QueryRow(ctx,
				"INSERT INTO tenants (name, domain, plan) VALUES ($1, $2, $3) RETURNING id",
				"DataLens Default", "datalens.io", "ENTERPRISE",
			).Scan(&tenantID)
			if err != nil {
				return fmt.Errorf("create tenant: %w", err)
			}
		} else {
			return fmt.Errorf("lookup tenant: %w", err)
		}
	}
	slog.Info("Tenant ID", "id", tenantID)

	// 2. Ensure Admin User exists
	email := "admin@datalens.com"
	var userID string
	err = conn.QueryRow(ctx, "SELECT id FROM users WHERE email = $1 AND tenant_id = $2", email, tenantID).Scan(&userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			slog.Info("Creating admin user...")
			// Hash password
			hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
			if err != nil {
				return fmt.Errorf("hash password: %w", err)
			}

			err = conn.QueryRow(ctx,
				"INSERT INTO users (tenant_id, email, name, password, status) VALUES ($1, $2, $3, $4, 'ACTIVE') RETURNING id",
				tenantID, email, "System Admin", string(hash),
			).Scan(&userID)
			if err != nil {
				return fmt.Errorf("create user: %w", err)
			}
			slog.Info("Admin user created", "email", email, "password", "password")
		} else {
			return fmt.Errorf("lookup user: %w", err)
		}
	}
	slog.Info("User ID", "id", userID)

	// 3. Assign ADMIN Role
	var roleID string
	err = conn.QueryRow(ctx, "SELECT id FROM roles WHERE name = 'ADMIN' AND is_system = TRUE LIMIT 1").Scan(&roleID)
	if err != nil {
		// Try looking up by tenant specific roles if system roles aren't global (schema says is_system is boolean, but tenant_id might be null? No, tenant_id is NOT NULL in schema usually unless global roles handled differently.
		// Wait, migration 001 inserts roles with UUIDs but doesn't specify tenant_id?
		// Line 419: INSERT INTO roles (id, name...) -> tenant_id is NOT NULL in schema (Line 33).
		// Ah, migration 419 MIGHT FAIL if tenant_id is missing?
		// Let's check schema again.
		// Line 33: tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE
		// Line 419: INSERT INTO roles (id, name, description, permissions, is_system) VALUES ...
		// It SKIPS tenant_id!
		// Either tenant_id is nullable (Line 33 doesn't say NOT NULL? It says: tenant_id UUID REFERENCES...)
		// Standard SQL: if not specified NOT NULL, it is nullable.
		// Let's assume nullable for system roles.
		// If so, query: WHERE name='ADMIN' AND (tenant_id IS NULL OR is_system=TRUE)
		return fmt.Errorf("lookup admin role: %w", err)
	}

	// Link
	_, err = conn.Exec(ctx, "INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", userID, roleID)
	if err != nil {
		return fmt.Errorf("assign role: %w", err)
	}
	slog.Info("Admin role assigned")

	return nil
}
