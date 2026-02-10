# DataLens 2.0

> The world's most automated, reliable, and evidence-ready privacy compliance platform.

## Quick Start

### Prerequisites

- [Go 1.23+](https://go.dev/dl/)
- [Docker & Docker Compose](https://docs.docker.com/get-docker/)
- [golangci-lint](https://golangci-lint.run/usage/install/) (optional, for linting)

### Setup

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

### Available Commands

```bash
make help          # Show all commands
make build         # Build all binaries
make test          # Run all tests
make lint          # Run linter
make check         # Run all checks (fmt + vet + lint + test)
make dev           # Start dev server
make docker-up     # Start infrastructure
make docker-down   # Stop infrastructure
make migrate       # Run migrations
make seed          # Seed dev data
```

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

## License

Proprietary — © 2026 Comply Ark
