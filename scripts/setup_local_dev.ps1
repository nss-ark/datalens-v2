# =============================================================================
# DataLens 2.0 — Local Development Setup
# =============================================================================
# Usage: .\scripts\setup_local_dev.ps1
# Requirements: Docker, Go 1.24+, Node 20+
# =============================================================================

$ErrorActionPreference = "Stop"

Write-Host "🚀 Starting DataLens Local Dev Setup..." -ForegroundColor Cyan

# --- 1. Environment Check ---
if (-not (Test-Path ".env")) {
    Write-Host "⚠️  .env not found. Creating from .env.example..." -ForegroundColor Yellow
    Copy-Item ".env.example" -Destination ".env"
    Write-Host "✅ Created .env" -ForegroundColor Green
}
else {
    Write-Host "✅ .env exists" -ForegroundColor Green
}

# --- 2. Docker Services ---
Write-Host "`n🐳 Starting App Infrastructure..." -ForegroundColor Cyan
docker compose -f docker-compose.dev.yml up -d
if ($LASTEXITCODE -ne 0) { Write-Error "Failed to start App Infrastructure" }

Write-Host "`n🐳 Starting Target Data Sources..." -ForegroundColor Cyan
docker compose -f docker-compose.sources.yml up -d
if ($LASTEXITCODE -ne 0) { Write-Error "Failed to start Target Data Sources" }

# --- 3. Health Check Wait ---
Write-Host "`n⏳ Waiting for database services to be ready (15s)..." -ForegroundColor Cyan
Start-Sleep -Seconds 15

# Optional: More robust health check loop
$maxRetries = 10
$retryCount = 0
$pgHealthy = $false

while (-not $pgHealthy -and $retryCount -lt $maxRetries) {
    # Check if main postgres is healthy
    $status = docker inspect --format='{{json .State.Health.Status}}' datalens-postgres | ConvertFrom-Json
    if ($status -eq "healthy") {
        $pgHealthy = $true
        Write-Host "✅ Postgres is healthy" -ForegroundColor Green
    }
    else {
        Write-Host "⏳ Waiting for Postgres... ($status)" -ForegroundColor Yellow
        Start-Sleep -Seconds 3
        $retryCount++
    }
}

if (-not $pgHealthy) {
    Write-Warning "Postgres might not be fully ready. Proceeding anyway..."
}


# --- 4. Migrations ---
if (Test-Path "cmd/migrate/main.go") {
    Write-Host "`nTesting Database Migrations..." -ForegroundColor Cyan
    # Note: Assuming 'go run' works. If not, build first.
    go run cmd/migrate/main.go up
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ Migrations completed" -ForegroundColor Green
    }
    else {
        Write-Error "Migrations failed"
    }
}
else {
    Write-Warning "cmd/migrate/main.go not found. Skipping migrations."
}

# --- 5. Seeding ---
if (Test-Path "cmd/seeder/main.go") {
    Write-Host "`n🌱 Seeding Database..." -ForegroundColor Cyan
    go run cmd/seeder/main.go
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ Seeding completed" -ForegroundColor Green
    }
    else {
        Write-Warning "Seeding failed (or partly failed). Proceeding..."
    }
}
else {
    Write-Warning "cmd/seeder/main.go not found. Skipping seeding."
}


# --- 6. Start Services ---
Write-Host "`n🚀 Launching Services..." -ForegroundColor Cyan

# Start Backend in new window
Write-Host "Starting Backend (go run cmd/api/main.go) in new window..."
Start-Process powershell -ArgumentList "-NoExit", "-Command", "go run cmd/api/main.go"

# Start Frontend in new window
Write-Host "Starting Frontend (npm run dev) in new window..."
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd frontend; npm run dev"


Write-Host "`n✅ Setup Complete!" -ForegroundColor Green
Write-Host "Backend API: http://localhost:8080"
Write-Host "Frontend:    http://localhost:5173"
Write-Host "Services running in background containers."
