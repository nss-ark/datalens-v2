# Agent Communication Board

> **Purpose**: This file is the shared message board for inter-agent communication. All agents read and write to this file to stay aligned. The human router facilitates by copying relevant messages between agent sessions.

---

## How To Use

### Posting a Message
When you need to communicate with another agent, add an entry under the appropriate section below using this format:

```markdown
### [TIMESTAMP] [FROM: Agent Name] → [TO: Agent Name or ALL]
**Subject**: Brief topic
**Type**: INFO | REQUEST | HANDOFF | BLOCKER | QUESTION

[Your message here — be specific and concise]

**Action Required**: [What the target agent needs to do, or "None — FYI only"]
```

### Reading Messages
At the start of every task, check this file for:
1. Messages addressed to **YOU** or **ALL**
2. Any **BLOCKER** type messages
3. Recent **HANDOFF** messages that affect your work

### Message Lifecycle
- Messages stay in this file for the current sprint
- The Orchestrator clears resolved messages at sprint boundaries
- Mark resolved messages with ~~strikethrough~~ and add resolution notes

---

## Active Messages

_No active messages yet. Messages will appear here as agents communicate._

---

## Message Types Reference

| Type | When to Use | Example |
|------|------------|---------|
| **INFO** | Sharing context another agent needs | "Backend: new `/api/v2/agents` endpoint is live with these response fields..." |
| **REQUEST** | Asking another agent to do something | "Frontend → Backend: Need a `GET /api/v2/dashboard/stats` endpoint" |
| **HANDOFF** | Passing completed work for the next agent | "Backend → Test: PII verification service is complete, needs unit tests" |
| **BLOCKER** | Something is preventing your task | "Frontend: Cannot proceed — `GET /api/pii/inventory` returns 500" |
| **QUESTION** | Need clarification from another agent | "AI/ML → Backend: Should detection results be cached per-tenant or globally?" |

---

## Contract Definitions

> When a Backend agent creates an API endpoint, or an AI/ML agent defines an interface, they should document the contract here so the Frontend and Test agents can work against it immediately.

### Active API Contracts

_Document new API contracts here as they're created:_

```markdown
### [Endpoint Name]
**Created by**: [Agent]
**Date**: [Date]
**Method**: GET | POST | PUT | DELETE
**Path**: `/api/v2/...`
**Request Body**: (if applicable)
**Response Body**: (JSON structure)
**Status**: Draft | Implemented | Tested
```

### Active Interface Contracts

_Document Go interfaces that cross agent boundaries:_

```markdown
### [Interface Name]
**Created by**: [Agent]
**File**: `path/to/file.go`
**Consumers**: [Which agents need to know about this]
**Notes**: [Any important details]
```

---

## Sprint Alignment

> At the start of each sprint, the Orchestrator posts the sprint goals here so all agents share context.

### Current Sprint Goals

_To be updated by the Orchestrator at each sprint start._

---

## Resolved Messages Archive

_Resolved messages are moved here with resolution notes._
