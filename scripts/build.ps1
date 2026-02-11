<#
.SYNOPSIS
    Build and management script for DataLens 2.0 on Windows.
    Equivalent to Makefile but for PowerShell.

.EXAMPLE
    .\scripts\build.ps1 -Target help
    .\scripts\build.ps1 -Target build
    .\scripts\build.ps1 -Target dev
#>

param (
    [string]$Target = "help"
)

function Show-Help {
    Write-Host "DataLens 2.0 - Windows Build Script" -ForegroundColor Cyan
    Write-Host "Usage: .\scripts\build.ps1 -Target <target>"
    Write-Host ""
    Write-Host "Targets:"
    Write-Host "  build         Build all binaries (api, agent, migrate)"
    Write-Host "  test          Run all tests"
    Write-Host "  lint          Run linter (requires golangci-lint)"
    Write-Host "  dev           Start API server (requires DB running)"
    Write-Host "  docker-up     Start dev infrastructure (docker-compose)"
    Write-Host "  docker-down   Stop dev infrastructure"
    Write-Host "  docker-prod   Build and start production stack"
    Write-Host "  migrate       Run database migrations"
}

function Build-Binaries {
    Write-Host "Building binaries..." -ForegroundColor Green
    if (-not (Test-Path "bin")) { New-Item -ItemType Directory -Path "bin" | Out-Null }
    
    $env:GOOS = "windows"
    $env:GOARCH = "amd64"
    
    Write-Host "Building API..."
    go build -o bin/api.exe ./cmd/api
    
    Write-Host "Building Agent..."
    go build -o bin/agent.exe ./cmd/agent
    
    Write-Host "Building Migrate..."
    go build -o bin/migrate.exe ./cmd/migrate
    
    Write-Host "Build complete." -ForegroundColor Green
}

function Run-Tests {
    Write-Host "Running tests..." -ForegroundColor Green
    go test ./... -v -race -count=1
}

function Run-Lint {
    Write-Host "Running linter..." -ForegroundColor Green
    golangci-lint run ./...
}

function Start-Dev {
    Write-Host "Starting API server..." -ForegroundColor Green
    # Load .env if exists (simple parsing)
    if (Test-Path ".env") {
        Get-Content ".env" | ForEach-Object {
            if ($_ -match "^([^#=]+)=(.*)$") {
                [Environment]::SetEnvironmentVariable($matches[1], $matches[2])
            }
        }
    }
    go run ./cmd/api
}

function Docker-Up {
    Write-Host "Starting dev infrastructure..." -ForegroundColor Green
    docker compose -f docker-compose.dev.yml up -d
}

function Docker-Down {
    Write-Host "Stopping dev infrastructure..." -ForegroundColor Green
    docker compose -f docker-compose.dev.yml down
}

function Docker-Prod {
    Write-Host "Building and starting production stack..." -ForegroundColor Green
    docker build -t datalens-api:local -f Dockerfile .
    docker build -t datalens-frontend:local -f frontend/Dockerfile ./frontend
    docker compose -f docker-compose.prod.yml up -d
}

function Run-Migrate {
    Write-Host "Running migrations..." -ForegroundColor Green
    go run ./cmd/migrate up
}

# Main Dispatch
switch ($Target) {
    "build"       { Build-Binaries }
    "test"        { Run-Tests }
    "lint"        { Run-Lint }
    "dev"         { Start-Dev }
    "docker-up"   { Docker-Up }
    "docker-down" { Docker-Down }
    "docker-prod" { Docker-Prod }
    "migrate"     { Run-Migrate }
    "help"        { Show-Help }
    Default       { Show-Help }
}
