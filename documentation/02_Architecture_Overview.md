# 02. Architecture Overview

## System Architecture

DataLens follows a **two-tier architecture** with clear separation of responsibilities:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           CLIENT INFRASTRUCTURE                              │
│                                                                              │
│  ┌────────────────────────────────────────────────────────────────────────┐ │
│  │                    DATALENS AGENT CLUSTER                               │ │
│  │                                                                         │ │
│  │  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐                 │ │
│  │  │  Agent #1   │◄──►│  Agent #2   │◄──►│  Agent #3   │   ...           │ │
│  │  │  (HR Dept)  │    │ (Marketing) │    │  (Finance)  │                 │ │
│  │  │  AWS EC2    │    │  Azure VM   │    │   GCP CE    │                 │ │
│  │  └──────┬──────┘    └──────┬──────┘    └──────┬──────┘                 │ │
│  │         │                  │                  │                         │ │
│  │         ▼                  ▼                  ▼                         │ │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐                  │ │
│  │  │ PostgreSQL   │  │  Salesforce  │  │  S3 Bucket   │                  │ │
│  │  │ MySQL        │  │  HubSpot     │  │  File System │                  │ │
│  │  │ MongoDB      │  │  Email IMAP  │  │  Documents   │                  │ │
│  │  └──────────────┘  └──────────────┘  └──────────────┘                  │ │
│  └────────────────────────────────────────────────────────────────────────┘ │
│                                    │                                         │
└────────────────────────────────────┼─────────────────────────────────────────┘
                                     │ HTTPS/gRPC (Metadata Only)
                                     ▼
                    ┌────────────────────────────────┐
                    │      DATALENS CONTROL CENTRE (Cloud)      │
                    │                                 │
                    │  ┌──────────────────────────┐  │
                    │  │   Verification Engine    │  │
                    │  │   (PII Review Queue)     │  │
                    │  └──────────────────────────┘  │
                    │                                 │
                    │  ┌──────────────────────────┐  │
                    │  │   Compliance Management  │  │
                    │  │   - Consent              │  │
                    │  │   - DSR                  │  │
                    │  │   - Grievances           │  │
                    │  │   - Reporting            │  │
                    │  └──────────────────────────┘  │
                    │                                 │
                    └────────────────────────────────┘
```

---

## Component Responsibilities

### DataLens Agent (v2)

| Responsibility | Description |
|----------------|-------------|
| **PII Discovery** | Scan databases, files, emails, CRMs for personal data |
| **Data Subject Identification** | Find and track individuals across systems |
| **Metadata Upload** | Send discovery results to CONTROL CENTRE (never actual PII) |
| **DSR Execution** | Delete/modify data locally when instructed by CONTROL CENTRE |
| **Peer Communication** | Sync with other agents to avoid duplicates |
| **Data Lineage** | Track how data flows between systems |

### DataLens Control Centre

| Responsibility | Description |
|----------------|-------------|
| **PII Verification** | Human review of discovered PII |
| **Purpose Management** | Define why data is processed |
| **Consent Management** | Track consent records |
| **DSR Orchestration** | Manage data subject requests |
| **Grievance Handling** | Track and resolve complaints |
| **Compliance Reporting** | Generate required documentation |
| **User Management** | Role-based access control |

---

## Zero-PII Architecture

> **Critical Design Principle**: Actual personal data NEVER leaves the client's infrastructure

### What Gets Sent to CONTROL CENTRE

| Sent ✅ | NOT Sent ❌ |
|---------|------------|
| Table/column names | Actual data values |
| PII category detected | Real emails, phones, names |
| Confidence scores | Actual Aadhaar/PAN numbers |
| File paths | File contents |
| Statistics | Personal information |

### Example

```
Agent SENDS:
{
  "object_identifier": "hr_db.public.employees.email",
  "pii_category": "EMAIL_ADDRESS",
  "confidence": 0.95,
  "sample_anonymized": "j***@e***.com",  // Masked
  "record_count": 1500
}

Agent DOES NOT SEND:
- "john.smith@example.com"
- Actual employee records
```

---

## Multi-Agent Architecture

### Why Multiple Agents?

Organizations have data in:
- Multiple departments (HR, Marketing, Finance)
- Multiple clouds (AWS, Azure, GCP)
- Multiple locations (US, EU, India)
- Multiple security zones (DMZ, internal)

Each agent scans its designated scope, but they cooperate to:
1. **Avoid duplicate detections**
2. **Build unified data subject profiles**
3. **Maintain consistent lineage**

### Agent Communication

```
┌─────────────────────────────────────────────────────────────────┐
│                      AGENT COMMUNICATION                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Agent #1                    Agent #2                           │
│  ┌─────────┐    "Found email     ┌─────────┐                   │
│  │         │  ────────────────►  │         │                   │
│  │   HR    │  john@example.com   │   CRM   │                   │
│  │         │                     │         │                   │
│  │         │  ◄────────────────  │         │                   │
│  │         │    "Same person,    │         │                   │
│  └─────────┘     link them"      └─────────┘                   │
│                                                                  │
│  Messages Exchanged:                                             │
│  • HEARTBEAT - "I'm alive"                                       │
│  • PII_DISCOVERED - "Found this PII, seen before?"              │
│  • DATA_SUBJECT_FOUND - "Found person with these identifiers"   │
│  • LINEAGE_UPDATE - "Subject X has data in source Y"            │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Technology Stack

### Agent (v2)

| Component | Technology |
|-----------|------------|
| Backend | Go (Golang) 1.21+ |
| Database | PostgreSQL 14+ |
| NLP Service | Python 3.11+, Flask, spaCy |
| AI/LLM | OpenAI-compatible API |
| OCR | Native Go libraries |
| Container | Docker |
| Orchestration | Kubernetes, Terraform |

### Control Centre

| Component | Technology |
|-----------|------------|
| Backend | Go (Golang) |
| Database | PostgreSQL |
| Frontend | React 18+, TypeScript, Vite |
| State Management | Zustand/Redux |
| UI Components | Material-UI style |
| Charts | Recharts |

---

## Data Flow

### 1. Discovery Flow

```
[Data Source] → [Agent Scanner] → [PII Detection] → [Local DB] → [CONTROL CENTRE Upload]
                                          ↓
                               [NLP Service] (optional)
```

### 2. Verification Flow

```
[CONTROL CENTRE] receives metadata → [PII Review Queue] → [Human Verification]
        ↓                                               ↓
   [Dashboard]                              [Verified PII Inventory]
```

### 3. DSR Flow

```
[Data Subject Portal] → [Control Centre DSR Manager] → [Agent Task Queue] → [Local Execution]
         ↓                     ↓                                         ↓
    [Submit Request]    [Track SLA]                              [Delete/Retrieve Data]
                              ↓                                         ↓
                     [Completion Report] ← ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ [Report Back]
```

---

## Deployment Topology

### Single-Agent Deployment

```
┌─────────────────────────────────┐
│     Small Organization          │
│  ┌─────────────────────────┐   │
│  │      Single Agent       │   │
│  │   (scans everything)    │   │
│  └─────────────────────────┘   │
└─────────────────────────────────┘
         ↓
    [DataLens CONTROL CENTRE]
```

### Multi-Agent Deployment

```
┌─────────────────────────────────────────────────────────────┐
│                    Large Enterprise                          │
│                                                              │
│  ┌───────────┐   ┌───────────┐   ┌───────────┐             │
│  │ Agent-HR  │   │Agent-Sales│   │Agent-Fin  │             │
│  │   AWS     │   │   Azure   │   │   GCP     │             │
│  └───────────┘   └───────────┘   └───────────┘             │
│        ↑               ↑               ↑                    │
│        └───────────────┼───────────────┘                    │
│                        │                                     │
│              [Peer-to-Peer Sync]                            │
│                                                              │
└─────────────────────────────────────────────────────────────┘
                          ↓
                   [DataLens CONTROL CENTRE]
```

---

## Security Architecture

### Transport Security

| Connection | Security |
|------------|----------|
| Agent ↔ CONTROL CENTRE | TLS 1.3 / mTLS |
| Agent ↔ Data Sources | Encrypted connections |
| Agent ↔ Agent | gRPC over TLS |
| Browser ↔ CONTROL CENTRE | HTTPS |

### Data Security

| Data | Protection |
|------|------------|
| Connection credentials | AES-256 encryption at rest |
| API keys | Hashed storage |
| Audit logs | Immutable storage |
| User passwords | bcrypt hashing |
