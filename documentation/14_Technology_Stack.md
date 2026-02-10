# 14. Technology Stack

## Overview

DataLens uses modern, production-grade technologies for reliability, performance, and maintainability.

---

## Technology Summary

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          TECHNOLOGY STACK                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  FRONTEND                                                                    │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │  React 18 │ TypeScript │ Vite │ Zustand │ Recharts │ Material-UI   │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
│  BACKEND                                                                     │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │  Go (Golang) │ Gin │ GORM │ JWT │ gRPC                              │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
│  NLP/AI                                                                      │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │  Python 3.11 │ Flask │ spaCy │ OpenAI API │ Tesseract OCR          │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
│  DATA                                                                        │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │  PostgreSQL 14+ │ Redis (optional)                                  │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
│  INFRASTRUCTURE                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │  Docker │ Kubernetes │ Terraform │ Helm                             │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Frontend Technologies

### Core

| Technology | Version | Purpose |
|------------|---------|---------|
| **React** | 18.x | UI framework |
| **TypeScript** | 5.x | Type-safe development |
| **Vite** | 5.x | Build tool |

### State & Data

| Technology | Purpose |
|------------|---------|
| **Zustand** | State management |
| **React Query** | Server state/caching |
| **Axios** | HTTP client |

### UI Components

| Technology | Purpose |
|------------|---------|
| **Material-UI** style | Component library |
| **Recharts** | Chart visualizations |
| **React Flow** | Data lineage graphs |
| **Lucide React** | Icons |

### Development

| Technology | Purpose |
|------------|---------|
| **ESLint** | Code linting |
| **Prettier** | Code formatting |
| **Vitest** | Unit testing |

---

## Backend Technologies

### Core (Go)

| Technology | Version | Purpose |
|------------|---------|---------|
| **Go** | 1.21+ | Primary language |
| **Gin** | 1.9.x | HTTP framework |
| **GORM** | 1.25.x | ORM |

### Authentication

| Technology | Purpose |
|------------|---------|
| **JWT** | Token-based auth |
| **bcrypt** | Password hashing |
| **mTLS** | Certificate auth |

### Database Drivers

| Technology | Purpose |
|------------|---------|
| **lib/pq** | PostgreSQL driver |
| **mysql** | MySQL driver |
| **mongo-driver** | MongoDB driver |

### Communication

| Technology | Purpose |
|------------|---------|
| **gRPC** | Agent-to-agent communication |
| **WebSocket** | Real-time updates |

### Utilities

| Technology | Purpose |
|------------|---------|
| **godotenv** | Environment loading |
| **uuid** | UUID generation |
| **zap/logrus** | Structured logging |
| **viper** | Configuration |

---

## NLP/AI Technologies

### Core (Python)

| Technology | Version | Purpose |
|------------|---------|---------|
| **Python** | 3.11+ | NLP runtime |
| **Flask** | 3.x | API framework |
| **spaCy** | 3.x | NLP processing |

### ML/AI

| Technology | Purpose |
|------------|---------|
| **spaCy NER** | Named Entity Recognition |
| **OpenAI API** | LLM integration |
| **Transformers** | Hugging Face models |

### OCR

| Technology | Purpose |
|------------|---------|
| **Tesseract** | Image text extraction |
| **PyMuPDF** | PDF processing |
| **python-docx** | Word document processing |

---

## Database Technologies

### Primary Database

| Technology | Version | Purpose |
|------------|---------|---------|
| **PostgreSQL** | 14+ | Primary data store |

### Features Used

| Feature | Purpose |
|---------|---------|
| JSONB | Flexible metadata storage |
| Full-text search | PII content search |
| Row-level security | Multi-tenant isolation |
| Partitioning | Large table performance |

### Caching (Optional)

| Technology | Purpose |
|------------|---------|
| **Redis** | Session storage, caching |

---

## Infrastructure Technologies

### Containerization

| Technology | Purpose |
|------------|---------|
| **Docker** | Container runtime |
| **Docker Compose** | Local development |

### Orchestration

| Technology | Purpose |
|------------|---------|
| **Kubernetes** | Container orchestration |
| **Helm** | K8s package management |

### Infrastructure as Code

| Technology | Purpose |
|------------|---------|
| **Terraform** | Cloud provisioning |

### Cloud Platforms

| Platform | Services Used |
|----------|---------------|
| **AWS** | ECS, RDS, S3, Secrets Manager |
| **Azure** | Container Instances, PostgreSQL, Blob Storage |
| **GCP** | Cloud Run, Cloud SQL, Cloud Storage |

---

## Security Technologies

| Technology | Purpose |
|------------|---------|
| **TLS 1.3** | Transport encryption |
| **AES-256-GCM** | Data encryption |
| **bcrypt** | Password hashing |
| **HMAC-SHA256** | API key validation |

---

## Development Tools

### Version Control

| Tool | Purpose |
|------|---------|
| **Git** | Source control |
| **GitHub/GitLab** | Repository hosting |

### CI/CD

| Tool | Purpose |
|------|---------|
| **GitHub Actions** | CI/CD pipelines |
| **Docker Hub** | Container registry |

### Testing

| Tool | Purpose |
|------|---------|
| **Go test** | Backend unit tests |
| **Vitest** | Frontend unit tests |
| **Playwright** | E2E testing |
| **k6** | Load testing |

### Documentation

| Tool | Purpose |
|------|---------|
| **Swagger/OpenAPI** | API documentation |
| **Markdown** | Technical docs |

---

## Third-Party Integrations

### Data Sources

| Integration | Purpose |
|-------------|---------|
| PostgreSQL | Database scanning |
| MySQL | Database scanning |
| MongoDB | Document scanning |
| AWS S3 | Cloud storage scanning |
| Salesforce | CRM scanning |
| IMAP | Email scanning |

### External Services

| Integration | Purpose |
|-------------|---------|
| OpenAI API | LLM-powered detection |
| SMTP | Email notifications |
| Webhook | External integrations |

---

## Version Requirements

### Minimum Versions

| Component | Minimum Version |
|-----------|-----------------|
| Go | 1.21 |
| Node.js | 18.x |
| Python | 3.11 |
| PostgreSQL | 14 |
| Docker | 20.10 |
| Kubernetes | 1.24 |

### Recommended Versions

| Component | Recommended Version |
|-----------|---------------------|
| Go | 1.22 |
| Node.js | 20.x |
| Python | 3.12 |
| PostgreSQL | 16 |
| Docker | 24.x |
| Kubernetes | 1.28 |
