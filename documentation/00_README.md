# DataLens v2.0 Documentation

> **Comprehensive technical and business documentation for the DataLens DPDPA Compliance Platform**

## Documentation Index

| Document | Description | Audience |
|----------|-------------|----------|
| [01_Executive_Summary](./01_Executive_Summary.md) | High-level overview, business case, key features | Business, All |
| [02_Architecture_Overview](./02_Architecture_Overview.md) | System architecture, component relationships | Technical, Business |
| [03_DataLens_Agent_v2](./03_DataLens_Agent_v2.md) | Agent component deep dive | Technical |
| [04_DataLens_CONTROL CENTRE_Application](./04_DataLens_CONTROL CENTRE_Application.md) | CONTROL CENTRE platform features and modules | Technical, Business |
| [05_PII_Detection_Engine](./05_PII_Detection_Engine.md) | PII detection capabilities and patterns | Technical |
| [06_Data_Source_Scanners](./06_Data_Source_Scanners.md) | Database, file, CRM, email scanners | Technical |
| [07_DSR_Management](./07_DSR_Management.md) | Data Subject Rights workflow | Business, Technical |
| [08_Consent_Management](./08_Consent_Management.md) | Consent capture and tracking | Business, Technical |
| [09_Database_Schema](./09_Database_Schema.md) | Database tables and relationships | Technical |
| [10_API_Reference](./10_API_Reference.md) | REST API endpoints | Technical |
| [11_Frontend_Components](./11_Frontend_Components.md) | UI pages and components | Technical |
| [12_Security_Compliance](./12_Security_Compliance.md) | Security features, DPDPA alignment | Security, Compliance |
| [13_Deployment_Guide](./13_Deployment_Guide.md) | Installation and deployment options | DevOps |
# DataLens Documentation

## Quick Start

DataLens is a comprehensive DPDPA (Digital Personal Data Protection Act) compliance platform consisting of:

- **DataLens Agent (v2)**: On-premise component for PII discovery and DSR execution
- **DataLens CONTROL CENTRE**: Cloud-hosted compliance management platform

## Table of Contents

| # | Document | Description |
|---|----------|-------------|
| 01 | [Executive Summary](./01_Executive_Summary.md) | Business overview, key features, DPDPA alignment |
| 02 | [Architecture Overview](./02_Architecture_Overview.md) | System architecture, Zero-PII design, data flow |
| 03 | [DataLens Agent v2](./03_DataLens_Agent_v2.md) | Agent architecture, handlers, services, multi-agent |
| 04 | [DataLens Control Centre](./04_DataLens_CONTROL CENTRE_Application.md) | CONTROL CENTRE modules, pages, API structure |
| 05 | [PII Detection Engine](./05_PII_Detection_Engine.md) | Regex patterns, heuristics, NLP, confidence scoring |
| 06 | [Data Source Scanners](./06_Data_Source_Scanners.md) | PostgreSQL, MySQL, File, S3, Salesforce, IMAP, MongoDB |
| 07 | [DSR Management](./07_DSR_Management.md) | DSR workflow, executor, SLA tracking |
| 08 | [Consent Management](./08_Consent_Management.md) | Consent capture, tracking, DPDPA compliance |
| 09 | [Database Schema](./09_Database_Schema.md) | Complete CONTROL CENTRE and Agent schema documentation |
| 10 | [API Reference](./10_API_Reference.md) | REST API endpoints for CONTROL CENTRE and Agent |
| 11 | [Frontend Components](./11_Frontend_Components.md) | React pages, components, patterns |
| 12 | [Security & Compliance](./12_Security_Compliance.md) | Encryption, RBAC, audit, DPDPA features |
| 13 | [Deployment Guide](./13_Deployment_Guide.md) | Docker, Kubernetes, cloud deployment |
| 14 | [Technology Stack](./14_Technology_Stack.md) | Languages, frameworks, tools, versions |

## Documentation Statistics

- **24 documentation files** covering the entire DataLens platform
- **Architecture**: Zero-PII principle, multi-agent communication
- **PII Detection**: 9 regex patterns, 77+ column heuristics, NLP integration
- **Data Sources**: 7 scanner types (PostgreSQL, MySQL, File, S3, Salesforce, IMAP, MongoDB)
- **Compliance**: DPDPA Sections 5, 6, 8, 11-14 coverage

---

## DataLens 2.0 Foundation

| # | Document | Description |
|---|----------|-------------|
| 15 | [Gap Analysis](./15_Gap_Analysis.md) | Current state assessment, technical debt, opportunities |
| 16 | [Improvement Recommendations](./16_Improvement_Recommendations.md) | AI detection, automation, performance, integrations |
| 17 | [V2 Feature Roadmap](./17_V2_Feature_Roadmap.md) | Quarterly phases with effort estimates |
| 18 | [Architecture Enhancements](./18_Architecture_Enhancements.md) | Message queues, microservices, caching, observability |
| 19 | [User Feedback Suggestions](./19_User_Feedback_Suggestions.md) | Legal expert feedback, priority refinements |

## Phase 0: Strategic Architecture (Pre-Development)

| # | Document | Description |
|---|----------|-------------|
| 20 | [Strategic Architecture](./20_Strategic_Architecture.md) | Modular, pluggable, regulation-agnostic design |
| 21 | [Domain Model](./21_Domain_Model.md) | Bounded contexts, entities, DDD patterns |
| 22 | [AI Integration Strategy](./22_AI_Integration_Strategy.md) | Multi-provider AI gateway, fallbacks, caching |
| 23 | [AGILE Development Plan](./23_AGILE_Development_Plan.md) | Sprint breakdown, team structure, milestones |

### Architecture at a Glance

```
┌─────────────────────────────────┐
│     YOUR ORGANIZATION           │
│  ┌───────────────────────────┐  │
│  │    DataLens Agent (v2)    │  │  ← Scans data sources
│  │  • Database Scanners      │  │  ← Detects PII patterns
│  │  • File Scanners          │  │  ← Executes DSR locally
│  │  • Email/CRM Scanners     │  │
│  └─────────────┬─────────────┘  │
└────────────────┼────────────────┘
                 │ HTTPS (metadata only)
                 ▼
┌─────────────────────────────────┐
│     DataLens CONTROL CENTRE (Cloud)       │
│  • PII Verification             │  ← Compliance team works here
│  • Consent Management           │
│  • DSR Orchestration            │
│  • Reporting & Analytics        │
└─────────────────────────────────┘
```

---

## Document Generation

*Last generated: February 10, 2026*  
*Based on analysis of DataLens application codebase*
