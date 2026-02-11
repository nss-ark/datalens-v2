#!/bin/bash
# =============================================================================
# DataLens 2.0 — Local Development Setup
# =============================================================================
# Usage: ./scripts/setup_local_dev.sh
# Requirements: Docker, Go 1.24+, Node 20+
# =============================================================================

set -e

echo -e "\033[0;36m🚀 Starting DataLens Local Dev Setup...\033[0m"

# --- 1. Environment Check ---
if [ ! -f .env ]; then
    echo -e "\033[0;33m⚠️  .env not found. Creating from .env.example...\033[0m"
    cp .env.example .env
    echo -e "\033[0;32m✅ Created .env\033[0m"
else
    echo -e "\033[0;32m✅ .env exists\033[0m"
fi

# --- 2. Docker Services ---
echo -e "\n\033[0;36m🐳 Starting App Infrastructure...\033[0m"
docker compose -f docker-compose.dev.yml up -d

echo -e "\n\033[0;36m🐳 Starting Target Data Sources...\033[0m"
docker compose -f docker-compose.sources.yml up -d

# --- 3. Health Check Wait ---
echo -e "\n\033[0;36m⏳ Waiting for database services to be ready (15s)...\033[0m"
sleep 15

# --- 4. Migrations ---
if [ -f "cmd/migrate/main.go" ]; then
    echo -e "\n\033[0;36mTesting Database Migrations...\033[0m"
    go run cmd/migrate/main.go up
    echo -e "\033[0;32m✅ Migrations completed\033[0m"
else
    echo -e "\033[0;33m⚠️  cmd/migrate/main.go not found. Skipping migrations.\033[0m"
fi

# --- 5. Seeding ---
if [ -f "cmd/seeder/main.go" ]; then
    echo -e "\n\033[0;36m🌱 Seeding Database...\033[0m"
    go run cmd/seeder/main.go
    echo -e "\033[0;32m✅ Seeding completed\033[0m"
else
    echo -e "\033[0;33m⚠️  cmd/seeder/main.go not found. Skipping seeding.\033[0m"
fi

# --- 6. Start Services ---
echo -e "\n\033[0;36m🚀 Launching Services...\033[0m"

# Cross-platform terminal opening is hard. 
# We'll just run them in background and show how to kill them.

echo "Starting Backend (background)..."
go run cmd/api/main.go > backend.log 2>&1 &
BACKEND_PID=$!
echo "Backend running (PID: $BACKEND_PID, logs: backend.log)"

echo "Starting Frontend (background)..."
(cd frontend && npm run dev > ../frontend.log 2>&1 &)
FRONTEND_PID=$!
echo "Frontend running (PID: $FRONTEND_PID, logs: frontend.log)"

echo -e "\n\033[0;32m✅ Setup Complete!\033[0m"
echo "Backend API: http://localhost:8080"
echo "Frontend:    http://localhost:5173"
echo "To stop services: kill $BACKEND_PID $FRONTEND_PID"
