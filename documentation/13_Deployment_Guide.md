# 13. Deployment Guide

## Overview

DataLens supports multiple deployment options to fit different organizational needs.

---

## Deployment Options

| Option | Best For | Complexity |
|--------|----------|------------|
| Docker Compose | Development, small deployments | Low |
| Kubernetes | Production, scalable | Medium |
| Managed Cloud | Serverless, minimal ops | Low |

---

## Docker Compose Deployment

### Prerequisites

- Docker 20.10+
- Docker Compose 2.0+
- 4GB RAM minimum
- 20GB disk space

### Quick Start

```bash
# Clone repository
git clone https://github.com/complyark/datalens-agent.git
cd datalens-agent

# Copy environment template
cp .env.example .env

# Edit configuration
nano .env

# Start services
docker-compose up -d

# Check status
docker-compose ps
```

### docker-compose.yml

```yaml
version: '3.8'

services:
  agent-backend:
    image: complyark/datalens-agent:latest
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgres://postgres:password@db:5432/datalens_agent
      - CONTROL CENTRE_ENDPOINT=${CONTROL CENTRE_ENDPOINT}
      - AGENT_API_KEY=${AGENT_API_KEY}
      - CLIENT_ID=${CLIENT_ID}
      - AGENT_ID=${AGENT_ID}
      - ENCRYPTION_KEY=${ENCRYPTION_KEY}
    depends_on:
      - db
      - nlp-service

  nlp-service:
    image: complyark/datalens-nlp:latest
    ports:
      - "5000:5000"

  db:
    image: postgres:14
    environment:
      - POSTGRES_PASSWORD=password
      - POSTGRES_DB=datalens_agent
    volumes:
      - pgdata:/var/lib/postgresql/data

  agent-frontend:
    image: complyark/datalens-agent-frontend:latest
    ports:
      - "3000:80"

volumes:
  pgdata:
```

---

## Kubernetes Deployment

### Prerequisites

- Kubernetes 1.24+
- kubectl configured
- Helm 3.0+ (optional)

### Helm Chart Installation

```bash
# Add repository
helm repo add complyark https://charts.complyark.com
helm repo update

# Install agent
helm install datalens-agent complyark/datalens-agent \
  --set config.CONTROL CENTREEndpoint="https://datalens.complyark.com" \
  --set config.clientId="your-client-id" \
  --set config.agentApiKey="your-api-key" \
  --namespace datalens \
  --create-namespace
```

### values.yaml

```yaml
replicaCount: 2

image:
  repository: complyark/datalens-agent
  tag: latest
  pullPolicy: IfNotPresent

config:
  CONTROL CENTREEndpoint: "https://datalens.complyark.com"
  clientId: ""
  agentApiKey: ""
  agentId: "agent-01"

database:
  enabled: true
  type: postgresql
  host: ""
  port: 5432
  name: datalens_agent
  existingSecret: datalens-db-secret

nlpService:
  enabled: true
  replicaCount: 1

resources:
  limits:
    cpu: 1000m
    memory: 1Gi
  requests:
    cpu: 250m
    memory: 256Mi

autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 10
  targetCPUUtilization: 70
```

### Manual Kubernetes Manifests

```yaml
# namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: datalens

---
# secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: datalens-secrets
  namespace: datalens
type: Opaque
stringData:
  agent-api-key: "your-api-key"
  encryption-key: "your-32-byte-key"
  database-url: "postgres://user:pass@host:5432/db"

---
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: datalens-agent
  namespace: datalens
spec:
  replicas: 2
  selector:
    matchLabels:
      app: datalens-agent
  template:
    metadata:
      labels:
        app: datalens-agent
    spec:
      containers:
      - name: agent
        image: complyark/datalens-agent:latest
        ports:
        - containerPort: 8080
        envFrom:
        - secretRef:
            name: datalens-secrets
        env:
        - name: CONTROL CENTRE_ENDPOINT
          value: "https://datalens.complyark.com"
        - name: CLIENT_ID
          value: "your-client-id"
        resources:
          limits:
            memory: "1Gi"
            cpu: "1000m"
          requests:
            memory: "256Mi"
            cpu: "250m"

---
# service.yaml
apiVersion: v1
kind: Service
metadata:
  name: datalens-agent
  namespace: datalens
spec:
  selector:
    app: datalens-agent
  ports:
  - port: 80
    targetPort: 8080
```

---

## Cloud-Specific Deployment

### AWS

```hcl
# terraform/aws/main.tf

resource "aws_ecs_task_definition" "agent" {
  family = "datalens-agent"
  
  container_definitions = jsonencode([
    {
      name  = "agent"
      image = "complyark/datalens-agent:latest"
      portMappings = [
        {
          containerPort = 8080
          hostPort      = 8080
        }
      ]
      environment = [
        {
          name  = "CONTROL CENTRE_ENDPOINT"
          value = var.CONTROL CENTRE_endpoint
        }
      ]
      secrets = [
        {
          name      = "AGENT_API_KEY"
          valueFrom = aws_secretsmanager_secret.agent_key.arn
        }
      ]
    }
  ])
}
```

### Azure

```hcl
# terraform/azure/main.tf

resource "azurerm_container_group" "agent" {
  name                = "datalens-agent"
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name
  os_type             = "Linux"

  container {
    name   = "agent"
    image  = "complyark/datalens-agent:latest"
    cpu    = "1"
    memory = "1"

    ports {
      port     = 8080
      protocol = "TCP"
    }
  }
}
```

### GCP

```hcl
# terraform/gcp/main.tf

resource "google_cloud_run_service" "agent" {
  name     = "datalens-agent"
  location = var.region

  template {
    spec {
      containers {
        image = "complyark/datalens-agent:latest"
        
        env {
          name  = "CONTROL CENTRE_ENDPOINT"
          value = var.CONTROL CENTRE_endpoint
        }
      }
    }
  }
}
```

---

## Configuration

### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `CONTROL CENTRE_ENDPOINT` | Yes | DataLens CONTROL CENTRE URL |
| `AGENT_API_KEY` | Yes | Agent authentication key |
| `CLIENT_ID` | Yes | Client/tenant identifier |
| `AGENT_ID` | Yes | Unique agent identifier |
| `DATABASE_URL` | Yes | PostgreSQL connection string |
| `ENCRYPTION_KEY` | Yes | 32-byte encryption key |
| `NLP_SERVICE_URL` | No | NLP microservice URL |
| `LLM_API_KEY` | No | OpenAI-compatible API key |
| `LOG_LEVEL` | No | debug, info, warn, error |

### agent-config.yaml

```yaml
agent:
  id: "agent-001"
  name: "Production Agent"
  
CONTROL CENTRE:
  endpoint: "https://datalens.complyark.com"
  sync_interval: "5m"
  timeout: "30s"
  
database:
  host: "localhost"
  port: 5432
  name: "datalens_agent"
  ssl_mode: "require"
  
scanning:
  sample_size: 100
  max_concurrent_scans: 3
  timeout_per_table: "10m"
  
nlp:
  enabled: true
  service_url: "http://nlp-service:5000"
  
logging:
  level: "info"
  format: "json"
```

---

## Post-Deployment Steps

### 1. Verify Connection

```bash
# Check agent status
curl http://localhost:8080/api/status

# Expected response
{
  "status": "healthy",
  "CONTROL CENTRE_connected": true,
  "database_connected": true,
  "version": "2.1.0"
}
```

### 2. Run Database Migrations

```bash
# Docker
docker exec -it datalens-agent ./migrate up

# Kubernetes
kubectl exec -it deployment/datalens-agent -n datalens -- ./migrate up
```

### 3. Configure Data Sources

```bash
# Add a PostgreSQL data source
curl -X POST http://localhost:8080/api/datasources \
  -H "Content-Type: application/json" \
  -d '{
    "name": "HR Database",
    "type": "postgresql",
    "connection_details": {
      "host": "hr-db.internal",
      "port": 5432,
      "database": "hr_production",
      "username": "datalens_reader",
      "password": "secure-password",
      "ssl_mode": "require"
    }
  }'
```

### 4. Test Scan

```bash
# Trigger a scan
curl -X POST http://localhost:8080/api/datasources/1/scan
```

---

## Troubleshooting

### Common Issues

| Issue | Cause | Solution |
|-------|-------|----------|
| Connection refused | CONTROL CENTRE not reachable | Check firewall, DNS |
| Authentication failed | Invalid API key | Verify key in CONTROL CENTRE |
| Database error | Wrong credentials | Check connection string |
| NLP service unavailable | Service not running | Check nlp-service container |

### Logs

```bash
# Docker
docker logs datalens-agent-backend

# Kubernetes
kubectl logs -f deployment/datalens-agent -n datalens
```

### Health Check

```bash
curl http://localhost:8080/api/health
```
