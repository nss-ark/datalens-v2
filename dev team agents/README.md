# DataLens 2.0 — Multi-Agent Development Framework

## Overview

This folder contains the prompt files and coordination tools for the DataLens 2.0 multi-agent development system.

**Model**: Hub-and-Spoke with shared communication board
**Router**: Human (you) — copies task specs to agent chats and routes results back
**Last Updated**: February 13, 2026 (post-Batch 19)

---

## Agent Roster

| Agent | File | Role |
|-------|------|------|
| 🎯 **Orchestrator** | [`orchestrator-agent.md`](./orchestrator-agent.md) | Sprint planning, task decomposition, progress tracking |
| ⚙️ **Backend** | [`backend-agent.md`](./backend-agent.md) | Go 1.24 API, services, repositories, domain logic |
| 🎨 **Frontend** | [`frontend-agent.md`](./frontend-agent.md) | React 18 + TypeScript UI, pages, components |
| 🤖 **AI/ML** | [`ai-ml-agent.md`](./ai-ml-agent.md) | PII detection, LLM integration, purpose suggestion |
| 🧪 **Test** | [`test-agent.md`](./test-agent.md) | Unit tests, integration tests, E2E tests |
| 🚀 **DevOps** | [`devops-agent.md`](./devops-agent.md) | Docker, CI/CD, K8s, observability |
| 🎨 **UX Review** | [`ux-review-agent.md`](./ux-review-agent.md) | Screen-by-screen UI/UX review, accessibility, consistency |

## Communication

| File | Purpose |
|------|------------|
| 📋 [`AGENT_COMMS.md`](./AGENT_COMMS.md) | Shared message board for inter-agent communication |

---

## Current Sprint Progress

### ✅ Completed (Batches 1–19)
- Infrastructure: monorepo, Docker, CI/CD, NATS, PostgreSQL, Redis
- Auth: JWT + refresh tokens, API keys, RBAC, tenant isolation
- Data Discovery: 7 connectors (PostgreSQL, MySQL, MongoDB, S3, M365, Google, REST)
- AI/ML: AI Gateway (OpenAI, Anthropic, generic), PII detection, industry templates, dark pattern detection
- DSR Engine: state machine, executor (access/erasure/correction), auto-verification
- Consent Engine: widget CRUD, sessions, notices, expiry/renewal, public APIs, analytics
- Portal: OTP auth, DPR submission, consent history, guardian flow, grievances
- Governance: purpose mapping, policy engine, violations, data lineage
- Security: breach management (DPDPA §28), identity verification (DigiLocker), encryption, audit logging
- Superadmin: tenant/user management, cross-tenant DSR, platform stats
- Translation: IndicTrans2 pipeline, 22 languages
- Notifications: email/webhook channels, event-driven breach alerts
- Frontend: 36+ pages, 25+ components, 3 portals (Control Centre, Portal, Admin)
- Tests: 20+ test files, integration + unit + E2E

### 🔜 Next Up
| Batch | Focus | Key Deliverables |
|-------|-------|-----------------|
| **20** | Enterprise Scale | Vanilla JS Widget Bundle, Event Mesh Refactoring |
| **20A** | UI/UX Review Sprint | Screen-by-screen review of all 3 portals, fixes based on feedback |

---

## How It Works

```
                    ┌─────────────────┐
                    │   YOU (Human)   │
                    │   Router        │
                    └────────┬────────┘
                             │
              ┌──────────────┼──────────────┐
              ▼              ▼              ▼
    ┌─────────────┐ ┌──────────────┐ ┌─────────────┐
    │ Orchestrator │ │ AGENT_COMMS  │ │ TASK_TRACKER │
    │ (Plans work) │ │ (Message     │ │ (Progress)   │
    │              │ │  board)      │ │              │
    └──────┬──────┘ └──────────────┘ └──────────────┘
           │
    Task Specs flow through YOU to:
           │
    ┌──────┼──────┬──────────┬──────────┐
    ▼      ▼      ▼          ▼          ▼
 Backend Frontend AI/ML    Test     DevOps
 Agent   Agent   Agent    Agent    Agent
```

### Step-by-Step Execution

1. **Start Orchestrator chat** → paste `orchestrator-agent.md` as system prompt
2. Orchestrator reads `TASK_TRACKER.md` and produces **Task Specs**
3. **You** copy each task spec into the appropriate agent's chat
4. Each agent reads `AGENT_COMMS.md`, does the work, posts results back to `AGENT_COMMS.md`
5. **You** copy agent results back to the Orchestrator
6. Orchestrator updates `TASK_TRACKER.md` and plans the next batch
7. Repeat!

### Inter-Agent Communication Flow

```
Backend Agent                          Frontend Agent
  │                                        │
  │ Creates new API endpoint               │
  │ Posts to AGENT_COMMS.md:               │
  │ "INFO → Frontend: GET /api/v2/agents   │
  │  is live, response: {id, name, ...}"   │
  │                                        │
  └──── YOU copy message ────────────────►│
                                           │ Reads AGENT_COMMS.md
                                           │ Builds UI against contract
                                           │ Posts: "HANDOFF → Test: Page done"
                                           │
                        YOU copy ──────────┘
                        │
                        ▼
                    Test Agent
                    (writes E2E tests)
```

---

## Quick Start

1. Open 2+ chat windows (Orchestrator + at least one agent)
2. Paste the agent's `.md` file as the system prompt
3. Start with: "Read TASK_TRACKER.md and plan the next batch of work"
4. Route the task specs to the appropriate agent chats
5. Ensure each agent checks `AGENT_COMMS.md` at the start of every task

---

## What Each Prompt Contains

Every agent prompt has been comprehensively written to include:
- **Role clarity** — exactly what they do and don't do
- **Completed work inventory** — what's already built so they don't recreate it
- **Real code patterns** — actual examples from the codebase (handler pattern, service pattern, API unwrapping, etc.)
- **Correct technical details** — Go 1.24, correct directory paths, module path
- **Reference documentation** — which docs to read and when
- **Upcoming work context** — what's expected in Batches 5–8
- **Cross-cutting concerns** — public APIs, consent widget CORS, portal OTP, digital signatures
- **Inter-agent communication** — what to check and post in AGENT_COMMS.md
- **Known gotchas** — bugs and patterns that have caused issues before (context keys, ApiResponse unwrapping)

---

## Documentation References

All agents reference documentation from:
```
documentation/
├── 00_README.md                    # Documentation index
├── 02_Architecture_Overview.md     # System topology
├── 03_DataLens_Agent_v2.md        # Agent component
├── 04_DataLens_SaaS_Application.md # Control Centre
├── 05_PII_Detection_Engine.md     # Detection pipeline
├── 06_Data_Source_Scanners.md     # Connectors
├── 07_DSR_Management.md          # DSR workflow
├── 08_Consent_Management.md      # Consent engine
├── 09_Database_Schema.md         # DB structure
├── 10_API_Reference.md           # API specs
├── 11_Frontend_Components.md     # UI patterns
├── 12_Security_Compliance.md     # Auth & security
├── 13_Deployment_Guide.md        # Deployment
├── 14_Technology_Stack.md        # Tech decisions
├── 15_Gap_Analysis.md            # Current gaps
├── 16_Improvement_Recommendations.md
├── 17_V2_Feature_Roadmap.md
├── 18_Architecture_Enhancements.md
├── 19_User_Feedback_Suggestions.md
├── 20_Strategic_Architecture.md  # Design patterns
├── 21_Domain_Model.md            # DDD entities
├── 22_AI_Integration_Strategy.md # AI integration
├── 23_AGILE_Development_Plan.md  # Sprint methodology
├── 24_DigiLocker_Integration.md  # DigiLocker OAuth + identity verification (NEW)
└── 25_Notice_Management.md       # Notice lifecycle + HuggingFace translation (NEW)
```
