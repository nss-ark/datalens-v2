# =============================================================================
# DataLens 2.0 — Production Dockerfile
# =============================================================================
# Multi-stage build for minimal production image
# =============================================================================

# --- Build Stage ---
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /bin/api ./cmd/api

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /bin/agent ./cmd/agent

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /bin/migrate ./cmd/migrate

# --- Production Stage ---
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

# Non-root user
RUN addgroup -S datalens && adduser -S datalens -G datalens

WORKDIR /app

COPY --from=builder /bin/api /app/api
COPY --from=builder /bin/agent /app/agent
COPY --from=builder /bin/migrate /app/migrate
COPY --from=builder /app/migrations /app/migrations
COPY --from=builder /app/config /app/config

USER datalens

EXPOSE 8080

ENTRYPOINT ["/app/api"]
