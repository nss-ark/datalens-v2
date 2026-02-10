# 22. AI Integration Strategy

## Philosophy

> **AI Where It Matters**: Use AI to solve genuinely complex problems where pattern-based approaches fall short. Avoid AI for simple, deterministic tasks.

---

## AI Decision Framework

### When to Use AI

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    AI DECISION MATRIX                                        │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ✅ USE AI WHEN:                        ❌ DON'T USE AI WHEN:               │
│  ─────────────────                       ────────────────────                │
│  • Context is critical                   • Pattern is deterministic         │
│  • Patterns are ambiguous                • Speed is critical (<100ms)       │
│  • Learning improves accuracy            • Cost is a concern for high vol   │
│  • Human would struggle too              • Simple rule suffices             │
│  • Edge cases are numerous               • Regulatory requires explainabil. │
│                                                                              │
│  EXAMPLES:                               EXAMPLES:                           │
│  • "Is 'Rahul' a name or company?"       • "Is this a valid email format?"  │
│  • "What purpose does this data serve?" • "Does this match Aadhaar regex?" │
│  • "Is this DSR request legitimate?"     • "Calculate DSR deadline"         │
│  • "Summarize breach impact"             • "Check consent expiry"           │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## AI Use Cases in DataLens 2.0

### Tier 1: High-Value AI Applications

| Use Case | Problem | AI Solution | Fallback |
|----------|---------|-------------|----------|
| **Contextual PII Detection** | "John" could be name or product code | LLM analyzes context, column relationships | Heuristics + manual review |
| **Purpose Inference** | What purpose does `marketing_leads` serve? | LLM suggests based on table/column names | Manual assignment |
| **DSR Identity Verification** | Is this request legitimate? | LLM analyzes request patterns, history | Manual verification |
| **Breach Impact Assessment** | What's the severity of this breach? | LLM analyzes data types, volume, sensitivity | Rule-based scoring |
| **Multi-language Translation** | Translate consent notice to 22 languages | IndicTrans/LLM for accurate legal translation | Pre-approved templates |

### Tier 2: Medium-Value AI Applications

| Use Case | Problem | AI Solution | Fallback |
|----------|---------|-------------|----------|
| **Sector Detection** | What industry is this customer in? | LLM analyzes data patterns, company info | Manual selection |
| **Anomaly Detection** | Unusual access patterns | ML model for behavior analysis | Threshold-based alerts |
| **Smart Scheduling** | When to run scans for minimal impact | ML predicts optimal scan times | Fixed schedule |
| **Auto-categorization** | Map fields to PII categories | LLM suggests category based on patterns | Manual mapping |

### Tier 3: Future AI Applications

| Use Case | Problem | AI Solution |
|----------|---------|-------------|
| **Predictive Compliance** | Predict upcoming compliance issues | ML on historical violations |
| **Document Analysis** | Extract PII from unstructured docs | Vision + LLM for document understanding |
| **Voice/Video Processing** | PII in call recordings | Speech-to-text + NLP |
| **Automated Remediation** | Suggest fixes for policy violations | LLM recommends actions |

---

## AI Architecture

### Multi-Provider Strategy

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    AI GATEWAY (Abstraction Layer)                            │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  Application Code                                                            │
│       │                                                                      │
│       ▼                                                                      │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                      AI Gateway Service                              │    │
│  │                                                                       │    │
│  │  • Unified API for all AI operations                                 │    │
│  │  • Provider selection logic                                          │    │
│  │  • Fallback orchestration                                            │    │
│  │  • Rate limiting & cost tracking                                     │    │
│  │  • Response caching                                                   │    │
│  │  • Metrics & logging                                                  │    │
│  │                                                                       │    │
│  └────────────────────────────┬────────────────────────────────────────┘    │
│                               │                                              │
│         ┌─────────────────────┼─────────────────────┐                       │
│         │                     │                     │                        │
│         ▼                     ▼                     ▼                        │
│  ┌─────────────┐       ┌─────────────┐       ┌─────────────┐                │
│  │   OpenAI    │       │  Anthropic  │       │  Local LLM  │                │
│  │   Provider  │       │   Provider  │       │   Provider  │                │
│  │             │       │             │       │             │                │
│  │ • GPT-4     │       │ • Claude 3  │       │ • Ollama    │                │
│  │ • GPT-4o    │       │ • Claude 3.5│       │ • LLaMA 3   │                │
│  │ • Embeddings│       │             │       │ • Mistral   │                │
│  └─────────────┘       └─────────────┘       └─────────────┘                │
│                                                                              │
│  ┌─────────────┐       ┌─────────────┐       ┌─────────────┐                │
│  │  IndicTrans │       │   Google    │       │   Azure     │                │
│  │  (Indic)    │       │   (Gemini)  │       │  (OpenAI)   │                │
│  │             │       │             │       │             │                │
│  │ • 22 Indian │       │ • Gemini    │       │ • Enterprise│                │
│  │   languages │       │ • PaLM      │       │   Backup    │                │
│  └─────────────┘       └─────────────┘       └─────────────┘                │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Provider Selection Logic

```go
type AIGateway interface {
    // Analyze content for PII
    DetectPII(ctx context.Context, input PIIDetectionInput) (*PIIDetectionResult, error)
    
    // Suggest purposes based on context
    SuggestPurposes(ctx context.Context, input PurposeSuggestionInput) ([]PurposeSuggestion, error)
    
    // Translate text
    Translate(ctx context.Context, text string, targetLang string) (string, error)
    
    // Assess breach impact
    AssessBreachImpact(ctx context.Context, breach BreachDetails) (*ImpactAssessment, error)
    
    // Generic completion
    Complete(ctx context.Context, prompt string, options CompletionOptions) (string, error)
}

type AIGatewayImpl struct {
    providers       map[string]AIProvider
    selector        ProviderSelector
    cache          AICache
    rateLimiter    RateLimiter
    costTracker    CostTracker
    fallbackChain  []string  // e.g., ["openai", "anthropic", "local"]
}

func (g *AIGatewayImpl) Complete(ctx context.Context, prompt string, opts CompletionOptions) (string, error) {
    // Check cache first
    if cached := g.cache.Get(prompt); cached != nil {
        return cached.Response, nil
    }
    
    // Select provider based on use case
    provider := g.selector.Select(opts.UseCase, opts.Priority)
    
    // Try with fallback
    var lastErr error
    for _, providerName := range g.getProviderChain(provider) {
        p := g.providers[providerName]
        
        // Check rate limits
        if !g.rateLimiter.Allow(providerName) {
            continue
        }
        
        result, err := p.Complete(ctx, prompt, opts)
        if err == nil {
            // Track cost
            g.costTracker.Track(providerName, result.TokensUsed)
            
            // Cache result
            g.cache.Set(prompt, result.Response, opts.CacheTTL)
            
            return result.Response, nil
        }
        lastErr = err
    }
    
    return "", fmt.Errorf("all providers failed: %w", lastErr)
}
```

### Provider Configuration

```yaml
# config/ai_providers.yaml
providers:
  openai:
    api_key: ${OPENAI_API_KEY}
    models:
      default: gpt-4o
      fast: gpt-4o-mini
      embedding: text-embedding-3-small
    rate_limits:
      requests_per_minute: 60
      tokens_per_minute: 100000
    cost_per_1k_tokens: 0.01
    
  anthropic:
    api_key: ${ANTHROPIC_API_KEY}
    models:
      default: claude-3-5-sonnet-20241022
      fast: claude-3-haiku-20240307
    rate_limits:
      requests_per_minute: 50
      tokens_per_minute: 80000
    cost_per_1k_tokens: 0.015
    
  local:
    endpoint: http://localhost:11434
    models:
      default: llama3.2
      fast: mistral
    cost_per_1k_tokens: 0  # No API cost
    
  indictrans:
    endpoint: ${INDICTRANS_ENDPOINT}
    supported_languages:
      - hi  # Hindi
      - bn  # Bengali
      - ta  # Tamil
      - te  # Telugu
      - mr  # Marathi
      # ... 17 more

provider_selection:
  pii_detection:
    primary: openai
    fallback: [anthropic, local]
  translation:
    indic_languages: indictrans
    other: openai
  purpose_suggestion:
    primary: anthropic
    fallback: [openai, local]
```

---

## Fallback Strategy

### Multi-Level Fallback

```
Level 1: Primary AI Provider (OpenAI/Anthropic)
    ↓ (Rate limit/Error/Timeout)
Level 2: Secondary AI Provider (Alternative cloud)
    ↓ (All cloud providers fail)
Level 3: Local LLM (Ollama with LLaMA/Mistral)
    ↓ (AI completely unavailable)
Level 4: Deterministic Fallback (Regex/Rules/Manual)
```

### Graceful Degradation

```go
type PIIDetectionWithFallback struct {
    aiDetector      AIDetector
    regexDetector   RegexDetector
    heuristicDetector HeuristicDetector
}

func (d *PIIDetectionWithFallback) Detect(ctx context.Context, input DetectionInput) ([]Detection, error) {
    // Try AI first (best accuracy)
    aiResult, err := d.aiDetector.Detect(ctx, input)
    if err == nil {
        return d.enrichWithValidation(aiResult), nil
    }
    
    // Log AI failure, continue with fallback
    log.Warn("AI detection failed, using fallback", "error", err)
    
    // Fallback to regex + heuristics
    regexResult := d.regexDetector.Detect(input)
    heuristicResult := d.heuristicDetector.Detect(input)
    
    // Merge results with lower confidence
    combined := d.mergeWithLowerConfidence(regexResult, heuristicResult, 0.7)
    
    // Mark for manual review
    for i := range combined {
        combined[i].RequiresManualReview = true
        combined[i].FallbackReason = err.Error()
    }
    
    return combined, nil
}
```

---

## Prompt Engineering

### PII Detection Prompt

```go
const PIIDetectionPrompt = `You are an expert PII (Personal Identifiable Information) detector.

CONTEXT:
- Table name: {{.TableName}}
- Column name: {{.ColumnName}}
- Column type: {{.DataType}}
- Sample values (anonymized): {{.Samples}}
- Adjacent columns: {{.AdjacentColumns}}
- Industry: {{.Industry}}

TASK:
Determine if this column contains PII. If yes, classify it.

RULES:
1. Consider column name AND sample values together
2. Consider context from adjacent columns (e.g., first_name + last_name)
3. Be conservative: if unsure, mark for human review
4. "John" alone is ambiguous; "John" next to "email" column is likely a name

RESPOND IN JSON:
{
  "is_pii": boolean,
  "category": "IDENTITY|CONTACT|FINANCIAL|HEALTH|...",
  "type": "NAME|EMAIL|PHONE|AADHAAR|...",
  "confidence": 0.0-1.0,
  "reasoning": "brief explanation",
  "requires_review": boolean
}`
```

### Purpose Suggestion Prompt

```go
const PurposeSuggestionPrompt = `You are a data governance expert.

CONTEXT:
- Data source type: {{.DataSourceType}}
- Table/Entity: {{.EntityName}}
- Column: {{.ColumnName}}  
- PII Type: {{.PIIType}}
- Industry: {{.Industry}}

TASK:
Suggest the most likely purpose(s) for collecting this data.

AVAILABLE PURPOSES:
{{range .AvailablePurposes}}
- {{.Code}}: {{.Description}}
{{end}}

RESPOND IN JSON:
{
  "suggested_purposes": [
    {"code": "PURPOSE_CODE", "confidence": 0.0-1.0, "reasoning": "..."}
  ],
  "legal_basis": "CONSENT|CONTRACT|LEGAL_OBLIGATION|...",
  "requires_explicit_consent": boolean
}`
```

---

## Caching Strategy

### What to Cache

| Data Type | Cache TTL | Cache Key |
|-----------|-----------|-----------|
| PII detection results | 24 hours | hash(column_name + samples) |
| Purpose suggestions | 7 days | hash(entity + column + industry) |
| Translations | 30 days | hash(text + target_lang) |
| Embeddings | 90 days | hash(text) |

### Cache Implementation

```go
type AICache struct {
    redis  *redis.Client
    ttl    map[string]time.Duration
}

func (c *AICache) Get(key string) *CachedResponse {
    data, err := c.redis.Get(ctx, "ai:"+key).Result()
    if err != nil {
        return nil
    }
    
    var response CachedResponse
    json.Unmarshal([]byte(data), &response)
    return &response
}

func (c *AICache) Set(key string, response string, ttl time.Duration) {
    cached := CachedResponse{
        Response:  response,
        CachedAt:  time.Now(),
    }
    data, _ := json.Marshal(cached)
    c.redis.Set(ctx, "ai:"+key, data, ttl)
}

// Invalidation on learning
func (c *AICache) InvalidateForColumn(columnPattern string) {
    keys, _ := c.redis.Keys(ctx, "ai:pii:*"+columnPattern+"*").Result()
    if len(keys) > 0 {
        c.redis.Del(ctx, keys...)
    }
}
```

---

## Learning & Feedback Loop

### Continuous Improvement

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    FEEDBACK LEARNING LOOP                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌─────────────┐                                                            │
│  │ AI Detection│ ──► Detection Result ──► User Reviews ──► Feedback        │
│  └─────────────┘           │                   │              │             │
│                            ▼                   ▼              ▼             │
│                    ┌─────────────┐     ┌─────────────┐  ┌─────────────┐    │
│                    │  Verified   │     │  Corrected  │  │  Rejected   │    │
│                    │  (Correct)  │     │ (Partially) │  │  (Wrong)    │    │
│                    └──────┬──────┘     └──────┬──────┘  └──────┬──────┘    │
│                           │                   │                 │           │
│                           └───────────────────┼─────────────────┘           │
│                                               ▼                             │
│                                    ┌─────────────────────┐                  │
│                                    │   Learning Service  │                  │
│                                    │                     │                  │
│                                    │ • Extract patterns  │                  │
│                                    │ • Update rules      │                  │
│                                    │ • Fine-tune prompts │                  │
│                                    │ • Cache corrections │                  │
│                                    └─────────────────────┘                  │
│                                               │                             │
│                                               ▼                             │
│                                    ┌─────────────────────┐                  │
│                                    │  Improved Detection │                  │
│                                    │  (Next Iteration)   │                  │
│                                    └─────────────────────┘                  │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Feedback Storage

```go
type DetectionFeedback struct {
    ID              UUID
    ClassificationID UUID
    OriginalResult  PIIClassification
    CorrectedResult *PIIClassification  // nil if correct
    FeedbackType    FeedbackType        // VERIFIED, CORRECTED, REJECTED
    CorrectedBy     UUID
    CorrectedAt     time.Time
    Notes           string
}

// Learning service analyzes feedback
type LearningService struct {
    feedbackRepo    FeedbackRepository
    ruleEngine      RuleEngine
    promptOptimizer PromptOptimizer
}

func (s *LearningService) LearnFromFeedback(ctx context.Context, feedback DetectionFeedback) error {
    // If column name pattern is frequently corrected, add rule
    if s.shouldCreateRule(feedback) {
        rule := s.extractRule(feedback)
        s.ruleEngine.AddRule(rule)
    }
    
    // Invalidate related cache
    s.cache.InvalidateForColumn(feedback.OriginalResult.ColumnName)
    
    // Track for prompt optimization (weekly batch)
    s.promptOptimizer.AddSample(feedback)
    
    return nil
}
```

---

## Cost Management

### Token Budget System

```go
type TokenBudget struct {
    TenantID       UUID
    MonthlyLimit   int64        // Total tokens per month
    UsedThisMonth  int64
    AlertThreshold float64      // Alert at 80%
    HardLimit      bool         // Block at 100% or just alert
}

type CostTracker struct {
    budgets  map[UUID]*TokenBudget
    costs    map[string]float64  // Provider -> cost per 1k tokens
}

func (t *CostTracker) CanUseTokens(tenantID UUID, estimatedTokens int) bool {
    budget := t.budgets[tenantID]
    if budget == nil {
        return true  // No budget = unlimited
    }
    
    newUsage := budget.UsedThisMonth + int64(estimatedTokens)
    
    // Check alert threshold
    if float64(newUsage)/float64(budget.MonthlyLimit) >= budget.AlertThreshold {
        t.sendBudgetAlert(tenantID, newUsage, budget.MonthlyLimit)
    }
    
    // Check hard limit
    if budget.HardLimit && newUsage > budget.MonthlyLimit {
        return false
    }
    
    return true
}
```

### Cost Optimization Strategies

| Strategy | Implementation |
|----------|----------------|
| **Caching** | Cache LLM responses for identical/similar inputs |
| **Batching** | Batch multiple detections into single API call |
| **Model tiering** | Use cheaper models for simple tasks |
| **Local fallback** | Use local LLM when cloud budget exhausted |
| **Smart sampling** | Sample subset for large tables, extrapolate |
| **Progressive detection** | Start with regex, use AI only for ambiguous |

---

## Security & Privacy

### Data Handling

```go
// NEVER send actual PII to external AI providers

type PIISanitizer struct{}

func (s *PIISanitizer) SanitizeForAI(samples []string) []string {
    sanitized := make([]string, len(samples))
    for i, sample := range samples {
        // Replace actual values with pattern indicators
        sanitized[i] = s.replaceWithPattern(sample)
    }
    return sanitized
}

// Examples:
// "Rahul Kumar" -> "[NAME: 2 words]"
// "rahul@email.com" -> "[EMAIL: format valid]"
// "1234-5678-9012-3456" -> "[CARD: 16 digits, Luhn valid]"
// "123456789012" -> "[AADHAAR: 12 digits, checksum valid]"
```

### Audit Trail for AI Decisions

```go
type AIAuditLog struct {
    ID              UUID
    Timestamp       time.Time
    Provider        string
    Operation       string        // DETECT, SUGGEST, TRANSLATE
    InputHash       string        // SHA-256 of input (not actual input)
    OutputSummary   string        // Brief summary of output
    TokensUsed      int
    DurationMs      int
    Status          string        // SUCCESS, FALLBACK, FAILED
    FallbackReason  *string
}
```

---

## Monitoring & Metrics

### AI-Specific Metrics

| Metric | Description | Alert Threshold |
|--------|-------------|-----------------|
| `ai_request_latency_ms` | Response time by provider | >5000ms |
| `ai_error_rate` | Failures by provider | >5% |
| `ai_fallback_rate` | How often fallback is used | >10% |
| `ai_tokens_used` | Token consumption | Budget threshold |
| `ai_cache_hit_rate` | Cache effectiveness | <50% |
| `ai_detection_accuracy` | Based on human corrections | <85% |
| `ai_cost_per_operation` | Running cost tracking | Budget threshold |

---

## Implementation Priority

| Phase | AI Feature | Priority |
|-------|------------|----------|
| **P1** | Contextual PII Detection | P0 |
| **P1** | Purpose Suggestion | P0 |
| **P1** | Provider Fallback | P0 |
| **P2** | Multi-language Translation | P1 |
| **P2** | Caching Layer | P1 |
| **P3** | Feedback Learning | P1 |
| **P3** | Cost Management | P1 |
| **P4** | Anomaly Detection | P2 |
| **P4** | Document Analysis | P2 |

---

## Next Document

➡️ See [23_AGILE_Development_Plan.md](./23_AGILE_Development_Plan.md) for sprint breakdown and milestones.
