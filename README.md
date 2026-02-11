# DataLens 2.0

[![CI](https://github.com/complyark/datalens/actions/workflows/ci.yml/badge.svg)](https://github.com/complyark/datalens/actions/workflows/ci.yml)

> The world's most automated, reliable, and evidence-ready privacy compliance platform.

## Quick Start

### Prerequisites

- [Go 1.23+](https://go.dev/dl/)
- [Docker & Docker Compose](https://docs.docker.com/get-docker/)
- [golangci-lint](https://golangci-lint.run/usage/install/) (optional, for linting)

### Setup

#### Linux / macOS
```bash
# 1. Clone and configure
cp .env.example .env

# 2. Start infrastructure (PostgreSQL, Redis, NATS)
make docker-up

# 3. Run database migrations
make migrate

# 4. Start the API server
make dev
```

#### Windows (PowerShell)
```powershell
# 1. Configure
Copy-Item .env.example .env

# 2. Start infrastructure
.\scripts\build.ps1 -Target docker-up

# 3. Run migrations
.\scripts\build.ps1 -Target migrate

# 4. Start API
.\scripts\build.ps1 -Target dev
```

### Available Commands

| Linux (Make) | Windows (PS) | Description |
|--------------|--------------|-------------|
| `make build` | `build.ps1 -Target build` | Build binaries |
| `make test` | `build.ps1 -Target test` | Run tests |
| `make dev` | `build.ps1 -Target dev` | Start dev server |


## Architecture

```
Datalens v2.0/
├── cmd/                 # Application entrypoints
│   ├── api/             # Control Centre API server
│   ├── agent/           # On-premise agent
│   └── migrate/         # Migration runner
├── internal/            # Private application code
│   ├── domain/          # Core entities (regulation-agnostic)
│   ├── service/         # Application services
│   ├── repository/      # Data access layer
│   ├── handler/         # HTTP handlers + middleware
│   ├── connector/       # Data source connectors
│   ├── adapter/         # Compliance regulation adapters
│   └── config/          # Configuration
├── pkg/                 # Shared, importable packages
│   ├── types/           # Universal types & enums
│   ├── eventbus/        # Event bus interface
│   ├── logging/         # Structured logging
│   ├── crypto/          # Encryption & hashing
│   └── httputil/        # HTTP utilities
├── migrations/          # SQL migration files
├── config/              # Configuration files (YAML)
├── test/                # Integration & E2E tests
└── documentation/       # 24 architecture & design docs
```

## Design Principles

1. **Regulation-Agnostic Core** — Core engines know nothing about specific regulations
2. **Plugin Architecture** — Data sources, AI providers, and regulations are pluggable
3. **Event-Driven** — Every action emits events; components react
4. **Evidence-First** — Every operation creates immutable, legally admissible evidence
5. **AI Where It Matters** — LLMs for complex classification, rules for deterministic tasks

## Documentation

See [documentation/00_README.md](./documentation/00_README.md) for the complete documentation index (24 documents).

---

## Deployment

### Production Deployment

DataLens uses Docker containers for production deployment.

**Docker Images** (published to GitHub Container Registry):
- `ghcr.io/complyark/datalens-api:latest` — Backend API server
- `ghcr.io/complyark/datalens-frontend:latest` — Frontend static files (nginx)

**Quick Start:**

```bash
# 1. Configure environment
cp .env.example .env
# Edit .env with production values

# 2. Start production stack
docker compose -f docker-compose.prod.yml up -d

# 3. Run migrations
docker exec datalens-api-prod ./migrate up

# 4. Access application
# Frontend: http://localhost:3000
# API: http://localhost:8080
```

**Building Locally:**

```bash
# Build Docker images
make build-docker-prod

# Start with local images
docker compose -f docker-compose.prod.yml up -d
```

**Required Environment Variables:**

| Variable | Description | Example |
|----------|-------------|---------|
| `DB_PASSWORD` | PostgreSQL password | `secure_password` |
| `JWT_SECRET` | JWT signing secret | `random32bytestring` |
| `AI_OPENAI_API_KEY` | OpenAI API key (optional) | `sk-...` |

See [.env.example](./.env.example) for complete list.

---

## License

Proprietary — © 2026 Comply Ark
