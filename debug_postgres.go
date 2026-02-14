package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Simulate the inputs from the UI/Service
	host := "localhost"
	port := 5434
	database := "customers_db"
	// format: username:password
	credentials := "postgres:postgres"

	fmt.Printf("Testing connection to %s:%d/%s with creds %s\n", host, port, database, credentials)

	// Logic from internal/infrastructure/connector/postgres.go
	dsn := fmt.Sprintf("postgres://%s@%s:%d/%s?sslmode=disable",
		credentials, host, port, database)

	fmt.Printf("Constructed DSN: %s\n", dsn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing config: %v\n", err)
		os.Exit(1)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating pool: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error pinging database: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Successfully connected to Postgres!")
}
