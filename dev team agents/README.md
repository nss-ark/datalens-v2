# DataLens 2.0 — Multi-Agent Development Framework

## Overview

This folder contains the prompt files and coordination tools for the DataLens 2.0 multi-agent development system.

**Model**: Hub-and-Spoke with shared communication board
**Router**: Human (you) — copies task specs to agent chats and routes results back

---

## Agent Roster

| Agent | File | Role |
|-------|------|------|
| 🎯 **Orchestrator** | [`orchestrator-agent.md`](./orchestrator-agent.md) | Sprint planning, task decomposition, progress tracking |
| ⚙️ **Backend** | [`backend-agent.md`](./backend-agent.md) | Go API, services, repositories, domain logic |
| 🎨 **Frontend** | [`frontend-agent.md`](./frontend-agent.md) | React + TypeScript UI, pages, components |
| 🤖 **AI/ML** | [`ai-ml-agent.md`](./ai-ml-agent.md) | PII detection, LLM integration, AI gateway |
| 🧪 **Test** | [`test-agent.md`](./test-agent.md) | Unit tests, integration tests, E2E tests |
| 🚀 **DevOps** | [`devops-agent.md`](./devops-agent.md) | Docker, CI/CD, K8s, observability |

## Communication

| File | Purpose |
|------|---------|
| 📋 [`AGENT_COMMS.md`](./AGENT_COMMS.md) | Shared message board for inter-agent communication |

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
└── 23_AGILE_Development_Plan.md  # Sprint methodology
```
