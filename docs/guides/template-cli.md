# Template CLI

O **boilerplate CLI** é uma ferramenta de linha de comando que gera novos microsserviços a partir deste template. Em vez de clonar o repositório e fazer find-replace manual, um único comando cria um projeto configurado com a arquitetura correta, dependências selecionadas e código pronto para rodar.

---

## Sumário

- [Instalação](#instalação)
- [Quick Start](#quick-start)
- [Comandos](#comandos)
  - [`boilerplate new`](#boilerplate-new-service-name)
  - [`boilerplate add domain`](#boilerplate-add-domain-name)
  - [`boilerplate version`](#boilerplate-version)
- [Flags (modo não-interativo)](#flags-modo-não-interativo)
- [O que cada opção faz](#o-que-cada-opção-faz)
- [Customização dos templates](#customização-dos-templates)
- [Em breve (Roadmap)](#em-breve-roadmap)
- [Troubleshooting](#troubleshooting)
- [FAQ](#faq)

---

## Instalação

### Pré-requisitos

- **Go 1.25+** instalado e configurado
- `$GOBIN` (ou `$GOPATH/bin`) presente no `$PATH`

### Instalando

```bash
go install bitbucket.org/appmax-space/go-boilerplate/cmd/cli@latest
```

### Verificando

```bash
boilerplate version
# boilerplate v0.1.0
```

---

## Quick Start

O fluxo mais comum: criar um novo serviço, responder aos prompts e começar a desenvolver.

```bash
boilerplate new payment-service
# Responda aos prompts interativos...

cd payment-service
make setup    # Instala ferramentas + sobe Docker + roda migrations
make dev      # Inicia o servidor com hot reload
```

Em poucos minutos você tem um microsserviço rodando com Clean Architecture, observabilidade e infraestrutura configurada.

---

## Comandos

### `boilerplate new [service-name]`

Cria um novo projeto completo a partir do template. O comando gera toda a estrutura de diretórios, configura dependências e deixa o projeto pronto para `make setup && make dev`.

#### Prompts interativos

Ao executar `boilerplate new`, o CLI guia você por uma série de perguntas:

| # | Prompt | Opções | Descrição |
|---|--------|--------|-----------|
| 1 | Nome do serviço | texto livre | Nome do diretório e referência interna (ex: `payment-service`) |
| 2 | Module path | texto livre | Go module path completo (ex: `bitbucket.org/appmax-space/payment-service`) |
| 3 | Banco de dados | PostgreSQL / MySQL / SQLite3 / Outro | Driver de banco de dados que será configurado no projeto |
| 4 | Protocolo | HTTP/REST (Gin) / ~~gRPC~~ | Protocolo de comunicação da API (gRPC em breve) |
| 5 | Injeção de dependência | Manual / ~~Uber Fx~~ | Estratégia de DI (Uber Fx em breve) |
| 6 | Cache Redis? | sim / não | Habilita cache com Redis (pkg/cache) |
| 7 | Idempotência? | sim / não | Habilita middleware de idempotência (só aparece se Redis = sim) |
| 8 | Service Key Auth? | sim / não | Habilita autenticação service-to-service via headers |
| 9 | Manter domínios de exemplo? | sim / não | Mantém os domínios `user` e `role` como referência |

#### Exemplo completo

```bash
$ boilerplate new payment-service

  Nome do serviço: payment-service
  Module path: bitbucket.org/appmax-space/payment-service
  Banco de dados: PostgreSQL
  Protocolo: HTTP/REST (Gin)
  Injeção de dependência: Manual
  Cache Redis? sim
  Idempotência? sim
  Service Key Auth? sim
  Manter domínios de exemplo? não

  ✔ Projeto criado em ./payment-service

  Próximos passos:
    cd payment-service
    make setup
    make dev
```

#### Estrutura gerada

```text
payment-service/
├── cmd/
│   ├── api/
│   │   ├── main.go              # Entrypoint da aplicação
│   │   ├── server.go            # DI manual e inicialização
│   │   └── doc.go               # Swagger metadata
│   └── migrate/
│       └── main.go              # CLI de migrations
├── config/
│   └── config.go                # Carregamento de configuração (env vars)
├── internal/
│   ├── domain/                  # Camada de domínio (zero dependências externas)
│   ├── usecases/                # Camada de aplicação (use cases + DTOs)
│   └── infrastructure/
│       ├── db/postgres/
│       │   ├── repository/      # Implementações de repositório
│       │   └── migration/       # Migrations SQL (Goose)
│       ├── web/
│       │   ├── handler/         # HTTP handlers (Gin)
│       │   ├── middleware/       # Logger, metrics, idempotency, auth
│       │   └── router/          # Registro de rotas
│       └── telemetry/           # Métricas de negócio (OpenTelemetry)
├── pkg/                         # Pacotes reutilizáveis
│   ├── apperror/                # Erros estruturados com HTTP status
│   ├── cache/                   # Interface de cache + implementação Redis
│   ├── database/                # Conexão DB com Writer/Reader cluster
│   ├── httputil/                # Helpers de resposta HTTP padronizada
│   ├── idempotency/             # Store de idempotência (Redis)
│   ├── logutil/                 # Logging estruturado com mascaramento PII
│   └── telemetry/               # Setup OpenTelemetry (traces + métricas)
├── tests/
│   └── e2e/                     # Testes E2E com TestContainers
├── docs/                        # Documentação
├── deploy/                      # Kustomize overlays (staging/production)
├── .env.example                 # Template de variáveis de ambiente
├── docker-compose.yml           # Infraestrutura local (Postgres, Redis)
├── Makefile                     # Comandos de desenvolvimento
├── Dockerfile                   # Build multi-stage
├── go.mod
└── go.sum
```

> **Nota:** Se você respondeu "não" para Redis, os diretórios `pkg/cache/`, `pkg/idempotency/` e o middleware de idempotência não são incluídos. O mesmo vale para Service Key Auth e o middleware correspondente.

---

### `boilerplate add domain [name]`

Adiciona um novo domínio a um projeto existente. Gera todas as camadas da Clean Architecture para o domínio especificado: entity, use cases, repository, handler, router e migration.

#### Uso

```bash
cd payment-service
boilerplate add domain order
```

#### Arquivos gerados

```text
internal/
├── domain/order/
│   ├── entity.go                # Aggregate Order com factory NewOrder()
│   ├── errors.go                # Erros de domínio (ErrNotFound, etc.)
│   └── filter.go                # Filtros de listagem
│
├── usecases/order/
│   ├── create.go                # CreateUseCase
│   ├── get.go                   # GetUseCase
│   ├── list.go                  # ListUseCase
│   ├── update.go                # UpdateUseCase
│   ├── delete.go                # DeleteUseCase
│   ├── dto/                     # Input/Output DTOs
│   │   ├── create.go
│   │   ├── get.go
│   │   ├── list.go
│   │   ├── update.go
│   │   └── delete.go
│   └── interfaces/
│       └── repository.go        # Interface do repositório
│
└── infrastructure/
    ├── db/postgres/
    │   ├── repository/
    │   │   └── order.go         # Implementação do repositório (sqlx)
    │   └── migration/
    │       └── 20260329120000_create_orders.sql
    ├── web/
    │   ├── handler/
    │   │   └── order.go         # HTTP handlers
    │   └── router/
    │       └── order.go         # Registro de rotas
    └── telemetry/
        └── metrics.go           # (atualizado com métricas do novo domínio)
```

#### Próximos passos após `add domain`

1. **Wiring**: Registre as dependências do novo domínio em `cmd/api/server.go:buildDependencies()`
2. **Migration**: Execute `make migrate-up` para criar a tabela no banco
3. **Customização**: Edite a entity, value objects e use cases conforme sua regra de negócio
4. **Testes**: Os arquivos de teste base (`*_test.go`) são gerados junto com os mocks

```go
// cmd/api/server.go — exemplo de wiring manual
orderRepo := repository.NewOrderRepository(dbCluster)
createOrderUC := order.NewCreateUseCase(orderRepo)
// ...
orderHandler := handler.NewOrderHandler(createOrderUC, getOrderUC, listOrderUC, updateOrderUC, deleteOrderUC)
router.RegisterOrderRoutes(r, orderHandler)
```

---

### `boilerplate version`

Exibe a versão instalada do CLI.

```bash
boilerplate version
# boilerplate v0.1.0
```

---

## Flags (modo não-interativo)

Para uso em CI/CD ou scripts, todas as opções podem ser passadas como flags, eliminando os prompts interativos.

### Referência de flags para `boilerplate new`

| Flag | Tipo | Default | Descrição |
|------|------|---------|-----------|
| `--module` | string | — | Go module path (ex: `bitbucket.org/org/svc`) |
| `--db` | string | — | Driver de banco: `postgres`, `mysql`, `sqlite`, `other` |
| `--protocol` | string | `http` | Protocolo da API: `http` (gRPC em breve) |
| `--di` | string | `manual` | Estratégia de DI: `manual` (Uber Fx em breve) |
| `--no-redis` | bool | `false` | Desabilita cache Redis e pacotes relacionados |
| `--no-auth` | bool | `false` | Desabilita Service Key Auth |
| `--no-idempotency` | bool | `false` | Desabilita middleware de idempotência |
| `--keep-examples` | bool | `false` | Mantém os domínios de exemplo (`user` e `role`) |

### Exemplo em CI/scripting

```bash
boilerplate new my-svc \
  --module bitbucket.org/appmax-space/my-svc \
  --db postgres \
  --no-redis \
  --no-auth
```

Neste exemplo, o projeto é gerado com PostgreSQL, sem Redis (e consequentemente sem idempotência), e sem Service Key Auth. Os domínios de exemplo são removidos por padrão.

---

## O que cada opção faz

### Banco de dados

| Opção | Driver | Pacote Go | Descrição |
|-------|--------|-----------|-----------|
| **PostgreSQL** | `pgx` | `github.com/jackc/pgx/v5/stdlib` | Driver recomendado. Migrations via Goose, repositórios com sqlx. |
| **MySQL** | `mysql` | `github.com/go-sql-driver/mysql` | Configurado com `pkg/database.DBCluster`. Migrations adaptadas para MySQL. |
| **SQLite3** | `sqlite3` | `github.com/mattn/go-sqlite3` | Ideal para testes e prototipagem. Arquivo local, sem infraestrutura externa. |
| **Outro** | — | — | Gera o projeto com `pkg/database` configurado mas sem driver específico. Você adiciona o driver desejado. |

> **Todos os drivers** usam a abstração `database/sql` via `pkg/database.DBCluster`, que suporta split Writer/Reader. Consulte o guia [Multi-Database](multi-database.md) para detalhes.

### Cache Redis

| Estado | O que inclui | Quando usar |
|--------|-------------|-------------|
| **Habilitado** (padrão) | `pkg/cache/` com implementação Redis, configuração de pool, TTL, health check. Use cases gerados com `.WithCache()` builder. | Serviços com leitura frequente e tolerância a dados levemente desatualizados. |
| **Desabilitado** (`--no-redis`) | Remove `pkg/cache/`, `pkg/idempotency/`, middleware de idempotência e todas as referências ao Redis no `docker-compose.yml` e configuração. | Serviços simples, batch jobs, ou quando o cache é gerenciado externamente. |

### Idempotência

| Estado | O que inclui | Quando usar |
|--------|-------------|-------------|
| **Habilitada** (padrão, requer Redis) | `pkg/idempotency/` com Store Redis, middleware que intercepta requests com `X-Idempotency-Key`. Usa SHA-256 fingerprint + lock/unlock. | Endpoints de escrita (POST, PUT) onde retry seguro é necessário. |
| **Desabilitada** (`--no-idempotency`) | Remove `pkg/idempotency/` e o middleware de idempotência. Redis continua disponível para cache. | Quando idempotência é tratada pelo API Gateway ou não é necessária. |

### Service Key Auth

| Estado | O que inclui | Quando usar |
|--------|-------------|-------------|
| **Habilitada** (padrão) | Middleware que valida `X-Service-Name` + `X-Service-Key` headers. Configuração via env vars `SERVICE_KEYS`. | Comunicação service-to-service em ambientes sem API Gateway com auth. |
| **Desabilitada** (`--no-auth`) | Remove o middleware de Service Key e as configurações relacionadas. | Quando a autenticação é feita pelo API Gateway, ou em serviços internos sem exposição externa. |

### Domínios de exemplo

| Estado | O que inclui | Quando usar |
|--------|-------------|-------------|
| **Mantidos** (`--keep-examples`) | Domínios `user` (CRUD completo com cache, singleflight, idempotência) e `role` (exemplo simples de multi-domain DI). Incluem testes unitários e E2E. | Primeiro contato com o template. Use como referência para entender os padrões. |
| **Removidos** (padrão) | Remove `internal/domain/user/`, `internal/domain/role/`, use cases, handlers, routers, repositories e migrations dos domínios de exemplo. | Projetos reais. Crie seus próprios domínios com `boilerplate add domain`. |

---

## Customização dos templates

Os templates usados pelo CLI estão embarcados no binário via Go `embed.FS`. Isso significa que o CLI funciona como um único executável, sem dependências externas de arquivos.

### Estrutura dos templates

```text
cmd/cli/
├── main.go                      # Entrypoint do CLI
└── templates/                   # Templates embarcados (embed.FS)
    ├── project/                 # Template do projeto completo (boilerplate new)
    │   ├── cmd/
    │   ├── config/
    │   ├── internal/
    │   ├── pkg/
    │   └── ...
    └── domain/                  # Template de domínio (boilerplate add domain)
        ├── entity.go.tmpl
        ├── usecase.go.tmpl
        ├── repository.go.tmpl
        ├── handler.go.tmpl
        └── ...
```

### Como customizar

1. **Fork** o repositório do boilerplate
2. **Edite** os templates em `cmd/cli/templates/`
3. **Rebuild** o CLI:

```bash
go build -o boilerplate ./cmd/cli/
```

4. **Instale** localmente:

```bash
go install ./cmd/cli/
```

### Engine de scaffold

A lógica de geração de código está em `cmd/cli/scaffold/`. Para customizações avançadas -- como adicionar novos prompts, alterar a lógica de remoção condicional de código, ou integrar novos protocolos -- este é o ponto de extensão.

---

## Em breve (Roadmap)

Duas opções aparecem nos prompts como desabilitadas, sinalizando o roadmap do template:

### gRPC

Atualmente o único protocolo disponível é **HTTP/REST (Gin)**. O suporte a gRPC adicionará:

- Definição de `.proto` files com protobuf
- Servidor gRPC com interceptors (logging, metrics, tracing)
- Handlers gRPC como alternativa aos HTTP handlers
- Opção de rodar ambos os protocolos simultaneamente (gRPC + HTTP gateway)

### Uber Fx

Atualmente a única estratégia de DI é **Manual** (wiring em `server.go:buildDependencies()`). O suporte a Uber Fx substituirá o wiring manual por:

- `fx.Module` para cada domínio (providers agrupados)
- `fx.Lifecycle` para gerenciamento de ciclo de vida (graceful shutdown)
- Autowiring automático via tipos de interface
- Redução significativa de código boilerplate em `server.go`

> Para entender como Uber Fx funciona com este projeto, consulte o guia [Uber Fx para Injeção de Dependência](fx-dependency-injection.md).

---

## Troubleshooting

### `command not found: boilerplate`

O binário do Go não está no `$PATH`. Verifique:

```bash
# Onde o Go instala binários
go env GOBIN
go env GOPATH

# Adicione ao seu ~/.zshrc ou ~/.bashrc
export PATH="$PATH:$(go env GOPATH)/bin"
```

### `go mod tidy` falha após gerar o projeto

- Verifique se o module path é válido e acessível
- Confirme que você tem acesso à rede (para baixar dependências)
- Para módulos privados (Bitbucket), configure `GOPRIVATE`:

```bash
export GOPRIVATE=bitbucket.org/appmax-space/*
```

### `permission denied` ao criar o projeto

O CLI precisa de permissão de escrita no diretório atual:

```bash
ls -la .
# Verifique se o usuário tem permissão de escrita
```

### `domain already exists` ao usar `add domain`

O CLI não sobrescreve domínios existentes para evitar perda de código. Se você precisa recriá-lo:

1. Remova manualmente os diretórios do domínio (`domain/`, `usecases/`, `infrastructure/` do domínio)
2. Execute `boilerplate add domain` novamente

---

## FAQ

### Posso usar em projetos existentes?

O comando `boilerplate add domain` funciona em projetos existentes que seguem a estrutura deste template. Já o comando `boilerplate new` cria um projeto do zero -- não é indicado para projetos já iniciados.

### Como atualizo o CLI?

```bash
go install bitbucket.org/appmax-space/go-boilerplate/cmd/cli@latest
```

### Funciona no Windows?

Sim. O CLI é escrito em Go, que compila nativamente para Windows, macOS e Linux. Os templates gerados também são compatíveis com todos os sistemas operacionais.

### Posso adicionar meus próprios templates?

Sim, via fork. Faça fork do repositório, edite os templates em `cmd/cli/templates/`, e rebuilde o binário. Veja a seção [Customização dos templates](#customização-dos-templates).

### O CLI precisa de conexão com internet?

Não para gerar o projeto. Os templates estão embarcados no binário. Porém, após a geração, `go mod tidy` e `make setup` precisam de internet para baixar dependências.

### Posso gerar um projeto sem nenhuma feature opcional?

Sim. O modo mais enxuto possível:

```bash
boilerplate new minimal-svc \
  --module bitbucket.org/appmax-space/minimal-svc \
  --db postgres \
  --no-redis \
  --no-auth
```

Isso gera um projeto apenas com Clean Architecture, PostgreSQL e OpenTelemetry -- sem cache, idempotência ou autenticação de serviço.

---

## Referências

- [Clean Architecture - Guia de Arquitetura](architecture.md)
- [Cache Strategy - Guia de Cache](cache.md)
- [Multi-Database - Guia de Banco de Dados](multi-database.md)
- [Uber Fx - Guia de DI](fx-dependency-injection.md)
- [Go embed package](https://pkg.go.dev/embed)
