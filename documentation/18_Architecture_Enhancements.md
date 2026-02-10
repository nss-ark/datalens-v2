# 18. Architecture Enhancements

## Overview

This document proposes architectural enhancements for DataLens 2.0 to support scalability, performance, and new capabilities.

---

## Current vs. Proposed Architecture

### Current Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        CURRENT ARCHITECTURE                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  AGENT (On-Premise)                    CONTROL CENTRE (Cloud)                         │
│  ┌─────────────────┐                   ┌─────────────────┐                  │
│  │   Monolithic    │    HTTP/REST      │   Monolithic    │                  │
│  │   Go Backend    │◄────────────────►│   Go Backend    │                  │
│  │                 │                   │                 │                  │
│  │ • Handlers      │                   │ • Handlers      │                  │
│  │ • Services      │                   │ • Services      │                  │
│  │ • DB Access     │                   │ • DB Access     │                  │
│  └────────┬────────┘                   └────────┬────────┘                  │
│           │                                     │                            │
│           ▼                                     ▼                            │
│  ┌─────────────────┐                   ┌─────────────────┐                  │
│  │   PostgreSQL    │                   │   PostgreSQL    │                  │
│  └─────────────────┘                   └─────────────────┘                  │
│           │                                                                  │
│           ▼                                                                  │
│  ┌─────────────────┐                                                        │
│  │   NLP Service   │                                                        │
│  │   (Python)      │                                                        │
│  └─────────────────┘                                                        │
│                                                                              │
│  LIMITATIONS:                                                                │
│  • Synchronous processing                                                   │
│  • No message queue                                                         │
│  • Limited horizontal scaling                                               │
│  • Tight coupling                                                           │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Proposed Architecture (2.0)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                      PROPOSED ARCHITECTURE (2.0)                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌───────────────────────────────────────────────────────────────────────┐  │
│  │                         AGENT (On-Premise)                             │  │
│  │                                                                         │  │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐   │  │
│  │  │ Scanner     │  │  LLM        │  │   DSR       │  │   Sync      │   │  │
│  │  │ Workers     │  │  Analyzer   │  │   Executor  │  │   Manager   │   │  │
│  │  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘   │  │
│  │         │                │                │                │           │  │
│  │         └────────────────┴────────────────┴────────────────┘           │  │
│  │                               │                                         │  │
│  │                               ▼                                         │  │
│  │  ┌─────────────────────────────────────────────────────────────────┐   │  │
│  │  │                    LOCAL MESSAGE QUEUE                           │   │  │
│  │  │                    (NATS / RabbitMQ)                            │   │  │
│  │  └─────────────────────────────────────────────────────────────────┘   │  │
│  │         │                │                │                             │  │
│  │         ▼                ▼                ▼                             │  │
│  │  ┌───────────┐    ┌───────────┐    ┌───────────┐                       │  │
│  │  │PostgreSQL │    │   Redis   │    │  LLM API  │                       │  │
│  │  │ (Local)   │    │  (Cache)  │    │  Client   │                       │  │
│  │  └───────────┘    └───────────┘    └───────────┘                       │  │
│  └───────────────────────────────────────────────────────────────────────┘  │
│                                    │                                         │
│                           gRPC / HTTPS                                       │
│                                    │                                         │
│  ┌───────────────────────────────────────────────────────────────────────┐  │
│  │                          CONTROL CENTRE (Cloud)                                  │  │
│  │                                                                         │  │
│  │  ┌─────────────────────────────────────────────────────────────────┐   │  │
│  │  │                     API GATEWAY                                  │   │  │
│  │  │                   (Rate Limiting, Auth, Routing)                │   │  │
│  │  └────────────────────────────┬────────────────────────────────────┘   │  │
│  │                               │                                         │  │
│  │    ┌──────────────────────────┼──────────────────────────┐             │  │
│  │    │                          │                          │              │  │
│  │    ▼                          ▼                          ▼              │  │
│  │  ┌────────────┐        ┌────────────┐        ┌────────────┐            │  │
│  │  │ Core API   │        │ Compliance │        │  Webhook   │            │  │
│  │  │ Service    │        │  Service   │        │  Service   │            │  │
│  │  └─────┬──────┘        └─────┬──────┘        └─────┬──────┘            │  │
│  │        │                     │                     │                    │  │
│  │        └──────────────┬──────┴─────────────────────┘                    │  │
│  │                       │                                                  │  │
│  │                       ▼                                                  │  │
│  │  ┌─────────────────────────────────────────────────────────────────┐   │  │
│  │  │                   MESSAGE QUEUE (NATS/Kafka)                     │   │  │
│  │  └─────────────────────────────────────────────────────────────────┘   │  │
│  │        │                     │                     │                    │  │
│  │        ▼                     ▼                     ▼                    │  │
│  │  ┌───────────┐        ┌───────────┐        ┌───────────┐               │  │
│  │  │PostgreSQL │        │   Redis   │        │   S3      │               │  │
│  │  │ (Primary) │        │  (Cache)  │        │  (Files)  │               │  │
│  │  └───────────┘        └───────────┘        └───────────┘               │  │
│  │                                                                         │  │
│  └───────────────────────────────────────────────────────────────────────┘  │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Key Architectural Changes

### 1. Message Queue Introduction

**Problem**: Current synchronous processing creates bottlenecks and limits scalability.

**Solution**: Introduce NATS or Kafka for async processing.

```go
// Current: Synchronous
func (h *ScanHandler) TriggerScan(c *gin.Context) {
    result, err := h.scanner.ScanDataSource(id) // Blocks for minutes
    c.JSON(200, result)
}

// Proposed: Async with queue
func (h *ScanHandler) TriggerScan(c *gin.Context) {
    jobID := uuid.New()
    h.queue.Publish("scan.start", ScanJob{ID: jobID, DataSourceID: id})
    c.JSON(202, gin.H{"job_id": jobID, "status": "queued"})
}

// Worker processes async
func (w *ScanWorker) ProcessScanJob(job ScanJob) {
    result := w.scanner.ScanDataSource(job.DataSourceID)
    w.queue.Publish("scan.complete", result)
}
```

**Benefits**:
- Non-blocking API responses
- Retry failed jobs automatically
- Horizontal worker scaling
- Better observability

### 2. Service Decomposition (CONTROL CENTRE)

**Current**: Monolithic with 34 handlers, 35 services

**Proposed**: Domain-based microservices

```
┌─────────────────────────────────────────────────────────────────┐
│                    MICROSERVICES BREAKDOWN                       │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌────────────────┐  ┌────────────────┐  ┌────────────────┐     │
│  │  AGENT-MGMT    │  │  PII-SERVICE   │  │  DSR-SERVICE   │     │
│  │  • Registration │  │  • Discovery   │  │  • Requests    │     │
│  │  • Health       │  │  • Verification│  │  • Execution   │     │
│  │  • Config       │  │  • Inventory   │  │  • SLA         │     │
│  └────────────────┘  └────────────────┘  └────────────────┘     │
│                                                                  │
│  ┌────────────────┐  ┌────────────────┐  ┌────────────────┐     │
│  │  CONSENT-SVC   │  │ COMPLIANCE-SVC │  │  WEBHOOK-SVC   │     │
│  │  • Records      │  │  • Breach      │  │  • Sub mgmt    │     │
│  │  • Notices      │  │  • Grievance   │  │  • Delivery    │     │
│  │  • Analytics    │  │  • RoPA        │  │  • Retry       │     │
│  └────────────────┘  └────────────────┘  └────────────────┘     │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

**Note**: This is a gradual transition, not a complete rewrite.

### 3. LLM Integration Layer

**New Component**: Abstracted LLM client supporting multiple providers

```go
// llm/client.go
type LLMClient interface {
    Analyze(ctx context.Context, prompt string) (*AnalysisResult, error)
    ClassifyPII(ctx context.Context, data SampleData) ([]PIIClassification, error)
    SuggestPurpose(ctx context.Context, context ContextInfo) ([]Purpose, error)
}

// Implementations
type OpenAIClient struct { ... }
type AnthropicClient struct { ... }
type LocalLLMClient struct { ... }  // Ollama, vLLM

// Factory with fallback
func NewLLMClient(config LLMConfig) LLMClient {
    clients := []LLMClient{}
    
    if config.OpenAI.Enabled {
        clients = append(clients, NewOpenAIClient(config.OpenAI))
    }
    if config.Anthropic.Enabled {
        clients = append(clients, NewAnthropicClient(config.Anthropic))
    }
    if config.Local.Enabled {
        clients = append(clients, NewLocalLLMClient(config.Local))
    }
    
    return NewFallbackClient(clients) // Try in order, fallback on failure
}
```

### 4. Caching Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     CACHING LAYERS                               │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  L1: In-Memory (Go sync.Map)        TTL: 1 min                  │
│  └── Hot path: session, auth, rate limits                       │
│                                                                  │
│  L2: Redis                          TTL: 5 min - 1 hour         │
│  └── API responses, dashboard metrics, LLM results              │
│                                                                  │
│  L3: Database                       Persistent                   │
│  └── Verified PII, scan results, audit logs                     │
│                                                                  │
│  Cache Invalidation Strategy:                                    │
│  • Write-through for critical data                               │
│  • TTL-based for read-heavy data                                 │
│  • Event-based invalidation for updates                          │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 5. Agent Worker Pool Architecture

```go
// Current: Fixed concurrent scans = 3
// Proposed: Dynamic worker pool

type WorkerPool struct {
    scanWorkers    chan *ScanWorker
    analysisWorkers chan *AnalysisWorker
    dsrWorkers     chan *DSRWorker
}

func NewWorkerPool(config WorkerConfig) *WorkerPool {
    pool := &WorkerPool{
        scanWorkers:     make(chan *ScanWorker, config.MaxScanWorkers),
        analysisWorkers: make(chan *AnalysisWorker, config.MaxAnalysisWorkers),
        dsrWorkers:      make(chan *DSRWorker, config.MaxDSRWorkers),
    }
    
    // Auto-scale based on load
    go pool.autoScale()
    
    return pool
}

func (p *WorkerPool) autoScale() {
    for {
        metrics := p.getMetrics()
        
        if metrics.ScanQueueDepth > 10 && len(p.scanWorkers) < p.maxScanWorkers {
            p.spawnScanWorker()
        }
        
        time.Sleep(10 * time.Second)
    }
}
```

---

## Data Architecture Enhancements

### 1. Schema Optimizations

```sql
-- Current: Single table for all PII discoveries
-- Proposed: Partitioned by date for better query performance

CREATE TABLE pii_discoveries (
    id UUID PRIMARY KEY,
    client_id UUID,
    discovered_at DATE,
    ...
) PARTITION BY RANGE (discovered_at);

CREATE TABLE pii_discoveries_2026_q1 
    PARTITION OF pii_discoveries 
    FOR VALUES FROM ('2026-01-01') TO ('2026-04-01');

-- Index for common queries
CREATE INDEX idx_pii_disc_client_status 
    ON pii_discoveries (client_id, status) 
    INCLUDE (pii_category, confidence);
```

### 2. Read Replicas

```
┌─────────────────────────────────────────────────────────────────┐
│                   DATABASE TOPOLOGY                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│              ┌─────────────┐                                    │
│              │   Primary   │                                    │
│              │  (Writes)   │                                    │
│              └──────┬──────┘                                    │
│                     │                                            │
│         ┌───────────┼───────────┐                               │
│         │           │           │                                │
│         ▼           ▼           ▼                                │
│  ┌───────────┐ ┌───────────┐ ┌───────────┐                      │
│  │ Read      │ │ Read      │ │ Analytics │                      │
│  │ Replica 1 │ │ Replica 2 │ │ Replica   │                      │
│  │ (API)     │ │ (API)     │ │ (Reports) │                      │
│  └───────────┘ └───────────┘ └───────────┘                      │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 3. Event Sourcing for Audit

```go
// Store all changes as immutable events
type AuditEvent struct {
    ID          string
    Timestamp   time.Time
    EventType   string
    AggregateID string
    ActorID     string
    Payload     json.RawMessage
}

// Reconstruct state from events
func (s *AuditService) GetStateAt(resourceID string, at time.Time) (*ResourceState, error) {
    events := s.repo.GetEventsUntil(resourceID, at)
    return s.replay(events)
}
```

---

## Security Architecture

### 1. Zero Trust Model

```
┌─────────────────────────────────────────────────────────────────┐
│                   ZERO TRUST ARCHITECTURE                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Every Request:                                                  │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  1. Authenticate (JWT/mTLS)                               │   │
│  │  2. Authorize (RBAC check)                                │   │
│  │  3. Validate tenant context                               │   │
│  │  4. Rate limit check                                      │   │
│  │  5. Audit log                                             │   │
│  │  6. Execute                                               │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                  │
│  Agent Communication:                                            │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  • mTLS certificate validation                            │   │
│  │  • API key verification                                   │   │
│  │  • Client ID matching                                     │   │
│  │  • Request signing (HMAC)                                 │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 2. Secrets Management

```go
// Migrate from environment variables to vault

type SecretsProvider interface {
    GetSecret(ctx context.Context, key string) (string, error)
    RotateSecret(ctx context.Context, key string) error
}

// Implementations
type VaultProvider struct { ... }      // HashiCorp Vault
type AWSSecretsProvider struct { ... } // AWS Secrets Manager
type EnvProvider struct { ... }        // Legacy fallback
```

---

## Observability Architecture

### 1. Structured Logging

```go
// Migrate to structured logging with correlation IDs
type Logger struct {
    *zap.Logger
}

func (l *Logger) WithContext(ctx context.Context) *Logger {
    return &Logger{
        l.With(
            zap.String("trace_id", GetTraceID(ctx)),
            zap.String("client_id", GetClientID(ctx)),
            zap.String("user_id", GetUserID(ctx)),
        ),
    }
}
```

### 2. Metrics & Tracing

```
┌─────────────────────────────────────────────────────────────────┐
│                   OBSERVABILITY STACK                            │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Metrics: Prometheus                                             │
│  └── Scan duration, API latency, error rates, queue depth       │
│                                                                  │
│  Tracing: Jaeger/OpenTelemetry                                  │
│  └── Request flow across services                               │
│                                                                  │
│  Logging: Loki/ELK                                              │
│  └── Structured logs with correlation                           │
│                                                                  │
│  Alerting: Grafana/PagerDuty                                    │
│  └── SLA breaches, error spikes, resource exhaustion           │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Migration Strategy

### Phase 1: Add, Don't Replace
- Add message queue alongside existing sync
- Add caching layer
- Add LLM client

### Phase 2: Migrate Gradually
- Move scan jobs to queue
- Move DSR processing to queue
- Split first microservice

### Phase 3: Optimize
- Enable read replicas
- Add partitioning
- Full observability

```
Week 1-4      Week 5-8      Week 9-12     Week 13+
────────      ────────      ─────────     ────────
 Add Queue  ─► Migrate    ─► Optimize  ─► Monitor
 Add Cache      Scans        Split        Tune
 Add LLM        DSR          Services     Scale
```

---

## Technology Recommendations

| Component | Current | Recommended |
|-----------|---------|-------------|
| Message Queue | None | NATS JetStream or Kafka |
| Cache | None | Redis Cluster |
| LLM | OpenAI only | Multi-provider + local |
| Observability | Basic logs | Prometheus + Jaeger + Loki |
| Secrets | Env vars | HashiCorp Vault |
| API Gateway | None | Kong or custom |
| CI/CD | Basic | GitHub Actions + ArgoCD |

---

## Summary

The architectural enhancements focus on:

1. **Async Processing**: Message queues for non-blocking operations
2. **Scalability**: Worker pools, read replicas, partitioning
3. **AI Integration**: Abstracted LLM layer with fallbacks
4. **Security**: Zero trust, secrets management
5. **Observability**: Metrics, tracing, structured logging

These changes enable DataLens 2.0 to handle enterprise-scale deployments while maintaining performance and reliability.
