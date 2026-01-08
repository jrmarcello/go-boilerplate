# People Service Registry

[![Go Version](https://img.shields.io/badge/Go-1.24-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Tests](https://img.shields.io/badge/Tests-Passing-success)](tests/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker)](docker/Dockerfile)

Microserviço de cadastro de pessoas com arquitetura Clean, cache Redis e deploy Kubernetes.

## 🚀 Quick Start

```bash
# Clone e setup
git clone <repo> && cd people-service-registry
make setup

# Desenvolvimento local (Docker)
make docker-up
make dev

# Desenvolvimento local (Kubernetes/Kind)
make kind-up
make kind-deploy
curl http://people.localhost/health
```

## 📋 Pré-requisitos

- Go 1.24+
- Docker
- Kind (opcional, para Kubernetes local)
- K6 (opcional, para testes de carga)

## 🛠️ Comandos

```bash
make help              # Lista todos os comandos

# Desenvolvimento
make setup             # Setup completo (tools + docker + migrations)
make dev               # Servidor com hot reload
make build             # Compila binário

# Testes
make test              # Todos os testes
make test-unit         # Testes unitários
make test-e2e          # Testes e2e com Postgres + Redis
make test-coverage     # Gera relatório HTML

# Docker
make docker-up         # Sobe Postgres + Redis
make docker-down       # Para containers
make docker-build      # Build da imagem

# Kubernetes (Kind)
make kind-up           # Cria cluster local + Ingress + Postgres + Redis
make kind-deploy       # Build + deploy + migrations
make kind-logs         # Ver logs do serviço
make kind-down         # Remove cluster

# Banco de dados
make migrate-up        # Roda migrations
make migrate-down      # Reverte última migration
make migrate-status    # Status das migrations
```

## 📁 Estrutura

```text
people-service-registry/
├── cmd/api/                          # Entrypoint (main.go, server.go)
├── config/                           # Configuração (env vars)
├── deploy/                           # Kubernetes manifests
│   ├── base/                         # Manifests base (Kustomize)
│   └── overlays/
│       ├── dev-local/                # Kind (local)
│       └── homologacao/              # AWS EKS staging
├── docker/                           # Dockerfile, docker-compose
├── docs/                             # Swagger, documentação
├── internal/
│   ├── domain/person/                # Entidades, Value Objects, Erros
│   │   └── vo/                       # ID, CPF, Email, Phone, Address
│   ├── infrastructure/
│   │   ├── cache/                    # Redis client
│   │   ├── db/postgres/              # Conexão, migrations, repository
│   │   ├── telemetry/                # OpenTelemetry
│   │   └── web/                      # HTTP Handlers, Middlewares, Router
│   ├── pkg/apperror/                 # Erros de aplicação
│   └── usecases/person/              # Casos de uso (Create, Get, Update, Delete)
└── tests/
    ├── e2e/                          # Testes e2e (TestContainers)
    └── load/                         # Testes de carga (k6)
```

## ⚙️ Configuração

### Docker Compose (`.env`)

Para desenvolvimento local com Docker, crie um arquivo `.env`:

```bash
SERVER_PORT=8080
POSTGRES_USER=user
POSTGRES_PASSWORD=password
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DB=people
REDIS_URL=redis://localhost:6379
REDIS_TTL=5m
REDIS_ENABLED=true
OTEL_SERVICE_NAME=people-service-registry
OTEL_COLLECTOR_URL=localhost:4317
```

### Kubernetes (ConfigMap)

Para Kubernetes (Kind/EKS), as variáveis ficam em:

- **dev-local**: `deploy/overlays/dev-local/configmap.yaml`
- **homologação**: `deploy/overlays/homologacao/configmap.yaml`

## 🔌 API

```http
### Health Check
GET /health

### Readiness (verifica DB)
GET /ready

### Criar Pessoa
POST /people
Content-Type: application/json

{
  "name": "João Silva",
  "document": "52998224725",
  "phone": "11999999999",
  "email": "joao@example.com"
}

### Buscar por ID (com cache)
GET /people/:id

### Atualizar Pessoa
PUT /people/:id

### Deletar Pessoa (soft delete)
DELETE /people/:id
```

📚 Swagger: `http://localhost:8080/swagger/index.html`

Veja [api.http](api.http) para mais exemplos.

## 🧪 Testes

| Tipo | Comando | Descrição |
|------|---------|-----------|
| Unit | `make test-unit` | Domínio, VOs, UseCases |
| E2E | `make test-e2e` | API + Postgres + Redis (TestContainers) |
| Coverage | `make test-coverage` | Gera relatório HTML |
| Load | `make load-smoke` | Teste de carga básico (k6) |

## 🐳 Deploy

### Docker Compose

```bash
docker compose -f docker/docker-compose.yml up -d
```

### Kubernetes (Kind - local)

```bash
# Setup inicial (uma vez)
make kind-up

# Deploy (repetir a cada mudança)
make kind-deploy

# Acessar
curl http://people.localhost/health
```

### Kubernetes (EKS - produção)

Os manifests estão em `deploy/overlays/homologacao/`. O deploy é feito via Bitbucket Pipelines + ArgoCD/Kustomize.

## 📊 Arquitetura

```text
                    ┌─────────────────┐
                    │    Ingress      │
                    │   (NGINX)       │
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │   API Service   │
                    │   (Go 1.24)     │
                    └────────┬────────┘
                             │
              ┌──────────────┼──────────────┐
              │              │              │
     ┌────────▼────────┐ ┌───▼───┐ ┌───────▼───────┐
     │   PostgreSQL    │ │ Redis │ │ OTel Collector│
     │   (Dados)       │ │(Cache)│ │ (Telemetria)  │
     └─────────────────┘ └───────┘ └───────────────┘
```
