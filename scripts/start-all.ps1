# =============================================================================
# DataLens 2.0 - Unified Startup Script
# =============================================================================
# Usage: .\scripts\start-all.ps1
# This script manages dependencies, infrastructure, and services.
# =============================================================================

$ErrorActionPreference = "Stop"

# Ensure HOME is set for tools that require it (like Playwright)
if (-not $env:HOME) {
    $env:HOME = $env:USERPROFILE
    Write-Host "Set HOME to $env:HOME" -ForegroundColor Gray
}

# ---- Helper Functions ----

function Check-Command {
    param (
        [string]$Name,
        [string]$Command,
        [string]$HelpUrl
    )
    Write-Host "Checking for $Name..." -NoNewline
    if (Get-Command $Command -ErrorAction SilentlyContinue) {
        Write-Host " OK" -ForegroundColor Green
        return $true
    }
    Write-Host " MISSING" -ForegroundColor Red
    Write-Warning "Please install $Name. See: $HelpUrl"
    return $false
}

function Kill-ProcessByPort {
    param ([int]$Port)
    $connections = Get-NetTCPConnection -LocalPort $Port -ErrorAction SilentlyContinue
    if ($connections) {
        foreach ($conn in $connections) {
            $pidToKill = $conn.OwningProcess
            if ($pidToKill -gt 0) {
                $process = Get-Process -Id $pidToKill -ErrorAction SilentlyContinue
                if ($process) {
                    Write-Host "  [WARN] Stopping PID $pidToKill ($($process.ProcessName)) on port $Port..." -ForegroundColor Yellow
                    Stop-Process -Id $pidToKill -Force -ErrorAction SilentlyContinue
                }
            }
        }
        Start-Sleep -Milliseconds 500
    }
}

# ---- Main Script ----

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  DataLens 2.0 - Local Dev Startup" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# --- 1. Dependency Checks ---
Write-Host "--- Dependency Checks ---" -ForegroundColor Cyan

if (-not (Check-Command -Name "Docker" -Command "docker" -HelpUrl "https://docs.docker.com/get-docker/")) {
    throw "Docker is required. Please install it and try again."
}
if (-not (Check-Command -Name "Go" -Command "go" -HelpUrl "https://go.dev/dl/")) {
    throw "Go is required. Please install it and try again."
}
if (-not (Check-Command -Name "Node.js" -Command "node" -HelpUrl "https://nodejs.org/")) {
    throw "Node.js is required. Please install it and try again."
}

Write-Host "Checking for Tesseract OCR..." -NoNewline
if (Get-Command "tesseract" -ErrorAction SilentlyContinue) {
    Write-Host " OK" -ForegroundColor Green
}
else {
    Write-Host " MISSING (optional)" -ForegroundColor Yellow
    Write-Host "  OCR will be unavailable. Install from: https://github.com/UB-Mannheim/tesseract/wiki" -ForegroundColor Gray
}

Write-Host ""

# --- 2. Environment Setup ---
Write-Host "--- Environment ---" -ForegroundColor Cyan
if (-not (Test-Path ".env")) {
    Write-Host "  [WARN] .env not found. Creating from .env.example..." -ForegroundColor Yellow
    Copy-Item ".env.example" -Destination ".env"
    Write-Host "  [OK] Created .env" -ForegroundColor Green
}
else {
    Write-Host "  [OK] .env exists" -ForegroundColor Green
}

Write-Host ""

# --- 3. Infrastructure (Docker) ---
Write-Host "--- Docker Infrastructure ---" -ForegroundColor Cyan

Write-Host "  Starting App Infrastructure..."
docker compose -f docker-compose.dev.yml up -d
if ($LASTEXITCODE -ne 0) { Write-Warning "Failed to start App Infrastructure. Proceeding..." }

Write-Host "  Starting Target Data Sources..."
docker compose -f docker-compose.sources.yml up -d
if ($LASTEXITCODE -ne 0) { Write-Warning "Failed to start Target Data Sources. Proceeding..." }

Write-Host ""

# --- 4. Health Checks ---
Write-Host "--- Health Checks ---" -ForegroundColor Cyan
Write-Host "  Waiting for database services (15s)..."
Start-Sleep -Seconds 15

$maxRetries = 10
$retryCount = 0
$pgHealthy = $false

while (-not $pgHealthy -and $retryCount -lt $maxRetries) {
    try {
        $rawStatus = docker inspect --format='{{json .State.Health.Status}}' datalens-postgres 2>$null
        if ($rawStatus) {
            $status = $rawStatus | ConvertFrom-Json
            if ($status -eq "healthy") {
                $pgHealthy = $true
                Write-Host "  [OK] Postgres is healthy" -ForegroundColor Green
            }
            else {
                Write-Host "  Waiting for Postgres... ($status)" -ForegroundColor Yellow
                Start-Sleep -Seconds 3
                $retryCount++
            }
        }
        else {
            Write-Host "  Waiting for Postgres container..." -ForegroundColor Yellow
            Start-Sleep -Seconds 3
            $retryCount++
        }
    }
    catch {
        Write-Host "  Waiting for Postgres container..." -ForegroundColor Yellow
        Start-Sleep -Seconds 3
        $retryCount++
    }
}

if (-not $pgHealthy) {
    Write-Warning "Postgres might not be fully ready. Proceeding anyway..."
}

Write-Host ""

# --- 5. Migrations and Seeding ---
Write-Host "--- Database Setup ---" -ForegroundColor Cyan

if (Test-Path "cmd/migrate/main.go") {
    Write-Host "  Running Migrations..."
    go run cmd/migrate/main.go up
    if ($LASTEXITCODE -eq 0) {
        Write-Host "  [OK] Migrations completed" -ForegroundColor Green
    }
    else {
        Write-Warning "Migrations failed. Proceeding..."
    }
}
else {
    Write-Host "  Skipping migrations (cmd/migrate/main.go not found)" -ForegroundColor Gray
}

if (Test-Path "cmd/seeder/main.go") {
    Write-Host "  Seeding Database..."
    go run cmd/seeder/main.go
    if ($LASTEXITCODE -eq 0) {
        Write-Host "  [OK] Seeding completed" -ForegroundColor Green
    }
    else {
        Write-Warning "Seeding failed. Proceeding..."
    }
}
else {
    Write-Host "  Skipping seeding (cmd/seeder/main.go not found)" -ForegroundColor Gray
}

Write-Host ""

# --- 6. Start Services ---
Write-Host "--- Launching Services ---" -ForegroundColor Cyan

$backendCwd = (Get-Location).Path
$frontendCwd = Join-Path $backendCwd "frontend"

# Backend (Port 8080, all mode)
Write-Host "  Clearing port 8080..."
Kill-ProcessByPort 8080
Write-Host "  Starting Backend API (mode=all)..."
Start-Process powershell -ArgumentList "-NoExit", "-Command", "Set-Location '$backendCwd'; `$env:APP_PORT='8080'; go run cmd/api/main.go --mode=all --port=8080"

# Frontend - Control Centre (Port 3000)
Write-Host "  Clearing port 3000..."
Kill-ProcessByPort 3000
Write-Host "  Starting Control Centre frontend..."
Start-Process powershell -ArgumentList "-NoExit", "-Command", "Set-Location '$frontendCwd'; npm run dev -w @datalens/control-centre"

# Frontend - Admin (Port 3001)
Write-Host "  Clearing port 3001..."
Kill-ProcessByPort 3001
Write-Host "  Starting Admin frontend..."
Start-Process powershell -ArgumentList "-NoExit", "-Command", "Set-Location '$frontendCwd'; npm run dev -w @datalens/admin"

# Frontend - Portal (Port 3002)
Write-Host "  Clearing port 3002..."
Kill-ProcessByPort 3002
Write-Host "  Starting Portal frontend..."
Start-Process powershell -ArgumentList "-NoExit", "-Command", "Set-Location '$frontendCwd'; npm run dev -w @datalens/portal"

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "  [DONE] Setup Complete!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "  --- Via Reverse Proxy (recommended) ---" -ForegroundColor White
Write-Host "  Control Centre: http://cc.localhost:8000" -ForegroundColor White
Write-Host "  Admin Panel:    http://admin.localhost:8000" -ForegroundColor White
Write-Host "  Portal:         http://portal.localhost:8000" -ForegroundColor White
Write-Host "  API:            http://api.localhost:8000" -ForegroundColor White
Write-Host ""
Write-Host "  --- Direct Access (no proxy) ---" -ForegroundColor Gray
Write-Host "  Backend API:    http://localhost:8080" -ForegroundColor Gray
Write-Host "  CC Frontend:    http://localhost:3000" -ForegroundColor Gray
Write-Host "  Admin Frontend: http://localhost:3001" -ForegroundColor Gray
Write-Host "  Portal Frontend:http://localhost:3002" -ForegroundColor Gray
Write-Host ""
Write-Host "  Services are running in separate windows." -ForegroundColor Gray
Write-Host "  Nginx proxy is running via Docker on :8000" -ForegroundColor Gray
Write-Host ""
