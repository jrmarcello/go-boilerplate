# Spec: CLI New Commands + Bug Fixes

## Status: DRAFT

## Context

O CLI `gopherplate` tem 3 comandos (`new`, `add domain`, `version`) mas faltam comandos essenciais para o workflow completo. Alem disso, 2 bugs conhecidos quebram o build quando features sao desabilitadas.

**Comandos faltantes:**
1. **`remove domain`** — inverso do `add domain`. Sem ele o dev deleta 28+ arquivos manualmente.
2. **`doctor`** — diagnostico do projeto. Verifica tools, Docker, migrations, config. Tipo `flutter doctor`.
3. **`add endpoint`** — scaffolda um endpoint custom dentro de um dominio existente.
4. **`remove endpoint`** — inverso do `add endpoint`.
5. **`wiring`** — auto-gera server.go + router.go a partir dos dominios detectados.

**Bugs conhecidos:**
1. `--no-redis` quebra build — `FeatureFiles["redis"]` remove `pkg/cache/` inteiro (incluindo a Cache interface Go-pura que use cases importam via `interfaces/cache.go`)
2. `--no-auth` quebra build — `bootstrap/test_helpers.go` referencia `middleware.ServiceKeyAuth` que foi removido

**Seguranca:** Comandos destrutivos (`remove domain`, `remove endpoint`) devem pedir confirmacao explicita com lista de arquivos e default N, a menos que `--yes` seja passado.

## Requirements

### Bug Fixes

- [ ] REQ-1: **`--no-redis` nao quebra build**
  - GIVEN o usuario executa `gopherplate new my-svc --no-redis --yes`
  - WHEN o projeto e gerado
  - THEN `go build ./...` passa sem erros
  - AND `pkg/cache/redisclient/` nao existe (implementacao Redis removida)
  - AND `pkg/cache/cache.go` e `pkg/cache/singleflight.go` EXISTEM (interfaces Go-pura mantidas)
  - AND `go mod tidy` remove dependencia `go-redis` do go.mod

- [ ] REQ-2: **`--no-auth` nao quebra build**
  - GIVEN o usuario executa `gopherplate new my-svc --no-auth --yes`
  - WHEN o projeto e gerado
  - THEN `go build ./...` passa sem erros
  - AND `middleware/service_key.go` nao existe
  - AND `bootstrap/test_helpers.go` NAO contem `SetupTestRouterWithAuth` nem referencia `ServiceKeyAuth`
  - AND `SetupTestRouter` (sem auth) continua disponivel

### Remove Domain

- [ ] REQ-3: **`gopherplate remove domain <name>` remove todos os arquivos**
  - GIVEN um dominio `order` existe no projeto (adicionado via `add domain`)
  - WHEN o usuario executa `gopherplate remove domain order --yes`
  - THEN remove: `internal/domain/order/`, `internal/usecases/order/`, `internal/infrastructure/db/postgres/repository/order*.go`, `internal/infrastructure/web/handler/order.go`, `internal/infrastructure/web/router/order.go`
  - AND NAO remove migration files por default (risco de data loss)
  - AND imprime instrucoes de cleanup: remover wiring de server.go/router.go, considerar reverter migration

- [ ] REQ-4: **Remove domain pede confirmacao com lista de arquivos**
  - GIVEN o usuario executa `gopherplate remove domain order` (sem `--yes`)
  - WHEN o prompt aparece
  - THEN mostra: "The following files/directories will be deleted:\n  - internal/domain/order/\n  - ... (N items)\nAre you sure? [y/N]"
  - AND default e N (nao deletar)
  - AND se usuario responde N ou Enter: aborta sem deletar nada
  - AND com `--yes` pula a confirmacao

- [ ] REQ-5: **Remove domain valida existencia**
  - GIVEN o dominio `payment` NAO existe
  - WHEN o usuario executa `gopherplate remove domain payment`
  - THEN retorna erro: "domain 'payment' not found at internal/domain/payment/"

### Doctor

- [ ] REQ-6: **`gopherplate doctor` verifica pre-requisitos**
  - GIVEN o usuario executa `gopherplate doctor`
  - WHEN o diagnostico roda
  - THEN verifica e reporta com formato checkmark/X:
    - Go (versao)
    - Docker daemon (rodando ou nao)
    - golangci-lint, swag, goose, air, k6, kind, kubectl (instalados ou nao)
  - AND para cada tool ausente: mostra instrucao de instalacao (ex: `brew install golangci-lint`)
  - AND verifica se `go.mod` existe (projeto Go valido)
  - AND verifica se Docker containers do projeto estao rodando (postgres, redis)

### Add Endpoint

- [ ] REQ-7: **`gopherplate add endpoint <domain> <name>` scaffolda endpoint custom**
  - GIVEN um dominio `order` existe no projeto
  - WHEN o usuario executa `gopherplate add endpoint order cancel`
  - THEN cria 3 arquivos:
    - `internal/usecases/order/cancel.go` — UC com ClassifyError, toAppError, SpanFromContext
    - `internal/usecases/order/dto/cancel.go` — Input/Output DTOs
    - `internal/usecases/order/cancel_test.go` — testes com mockRepository + AppError assertions
  - AND imprime instrucoes de wiring manual (handler method + rota)

- [ ] REQ-8: **Add endpoint valida dominio e nome**
  - GIVEN o dominio `payment` NAO existe
  - WHEN o usuario executa `gopherplate add endpoint payment refund`
  - THEN retorna erro: "domain 'payment' not found"
  - GIVEN o endpoint `create` ja existe no dominio `order`
  - WHEN o usuario executa `gopherplate add endpoint order create`
  - THEN retorna erro: "endpoint 'create' already exists in domain 'order'"
  - AND nome do endpoint segue mesmas regras de validacao do domain (snake_case, starts with letter)

### Remove Endpoint

- [ ] REQ-9: **`gopherplate remove endpoint <domain> <name>` remove arquivos**
  - GIVEN o endpoint `cancel` existe no dominio `order`
  - WHEN o usuario executa `gopherplate remove endpoint order cancel --yes`
  - THEN remove: `cancel.go`, `dto/cancel.go`, `cancel_test.go`
  - AND NAO remove se endpoint e um dos 5 CRUD padrao (create, get, update, delete, list) — protege contra remocao acidental da estrutura base

- [ ] REQ-10: **Remove endpoint pede confirmacao**
  - GIVEN o usuario executa `gopherplate remove endpoint order cancel` (sem `--yes`)
  - WHEN o prompt aparece
  - THEN mostra lista de arquivos e "Are you sure? [y/N]"
  - AND default e N

### Wiring

- [ ] REQ-11: **`gopherplate wiring` auto-gera server.go + router.go**
  - GIVEN o projeto tem dominios `user`, `role`, `order` (detectados em `internal/domain/`)
  - WHEN o usuario executa `gopherplate wiring`
  - THEN detecta todos os dominios pela presenca de `internal/domain/<name>/`
  - AND verifica que para cada dominio existem: repository, handler, router files
  - AND regenera `cmd/api/server.go` com `bootstrap.New()` e todos os handlers
  - AND regenera `internal/infrastructure/web/router/router.go` com `Register<Domain>Routes()` para cada dominio
  - AND regenera `internal/bootstrap/container.go` com structs e wiring para todos os dominios
  - AND `go build ./...` passa apos regeneracao

## Test Plan

### Unit Tests

| TC | REQ | Category | Description | Expected |
|----|-----|----------|-------------|----------|
| TC-U-01 | REQ-1 | happy | new --no-redis generates buildable project | `go build ./...` passes |
| TC-U-02 | REQ-1 | edge | no-redis keeps pkg/cache/cache.go (interface) | file exists |
| TC-U-03 | REQ-1 | edge | no-redis removes pkg/cache/redisclient/ | dir not exists |
| TC-U-04 | REQ-2 | happy | new --no-auth generates buildable project | `go build ./...` passes |
| TC-U-05 | REQ-2 | edge | no-auth test_helpers has no ServiceKeyAuth | grep returns 0 |
| TC-U-06 | REQ-2 | edge | no-auth test_helpers still has SetupTestRouter | function exists |
| TC-U-07 | REQ-3 | happy | remove domain --yes deletes all domain files | dirs/files gone |
| TC-U-08 | REQ-3 | edge | remove domain preserves migration files | migration exists |
| TC-U-09 | REQ-3 | edge | remove domain prints cleanup instructions | output contains "server.go" |
| TC-U-10 | REQ-4 | happy | remove domain without --yes shows file list | output contains file paths |
| TC-U-11 | REQ-4 | edge | remove domain user answers N — nothing deleted | files still exist |
| TC-U-12 | REQ-5 | edge | remove nonexistent domain returns error | "not found" in error |
| TC-U-13 | REQ-6 | happy | doctor detects installed Go | checkmark + version |
| TC-U-14 | REQ-6 | edge | doctor reports missing tool with install hint | "not found" + brew command |
| TC-U-15 | REQ-6 | edge | doctor reports go.mod missing | "not a Go project" |
| TC-U-16 | REQ-7 | happy | add endpoint creates 3 files | UC + DTO + test exist |
| TC-U-17 | REQ-7 | happy | generated UC has ClassifyError + toAppError | content check |
| TC-U-18 | REQ-8 | edge | add endpoint on nonexistent domain fails | "not found" error |
| TC-U-19 | REQ-8 | edge | add endpoint duplicate name fails | "already exists" error |
| TC-U-20 | REQ-8 | edge | add endpoint invalid name fails | validation error |
| TC-U-21 | REQ-9 | happy | remove endpoint --yes deletes 3 files | files gone |
| TC-U-22 | REQ-9 | edge | remove endpoint on CRUD name (create) refused | "protected" error |
| TC-U-23 | REQ-10 | edge | remove endpoint user answers N — nothing deleted | files still exist |
| TC-U-24 | REQ-11 | happy | wiring detects 3 domains and generates files | server.go has all 3 |
| TC-U-25 | REQ-11 | edge | wiring with only examples (user/role) | generates correctly |
| TC-U-26 | REQ-11 | edge | wiring when no domains exist | minimal server.go, compiles |

**Rigor check:** 11 happy + 15 edge = 26 TCs. Edge/error > happy. Todos os REQs cobertos (1-11).

## Design

### Architecture Decisions

**Bug fix --no-redis (abordagem correta):**
O problema: `FeatureFiles["redis"]` remove `pkg/cache/` inteiro mas `pkg/cache/cache.go` e `pkg/cache/singleflight.go` sao interfaces Go-pura (zero dependencia Redis). Solucao: mudar para remover apenas `pkg/cache/redisclient/` (implementacao Redis concreta). `go mod tidy` remove `go-redis` do go.mod. Use cases mantem `.WithCache()` como pattern util (pode usar in-memory cache futuro).

**Bug fix --no-auth (abordagem):**
Adicionar template condicional no `CleanupWiring` para `internal/bootstrap/test_helpers.go`. Com `Auth=false`: gera apenas `NewForTest` + `SetupTestRouter` (sem auth). Com `Auth=true`: gera tudo (incluindo `SetupTestRouterWithAuth` com imports de `middleware.ServiceKeyAuth`). Segue o mesmo pattern que ja existe para server.go e router.go — template Go com `{{if .Auth}}`.

**Remove domain — deteccao de arquivos:**
Escaneia paths por pattern (nao hardcoded): `internal/domain/<name>/`, `internal/usecases/<name>/`, `internal/infrastructure/db/postgres/repository/<name>*.go`, `internal/infrastructure/web/handler/<name>.go`, `internal/infrastructure/web/router/<name>.go`. Migrations (`*_create_<plural>*.sql`) sao LISTADAS na confirmacao mas NAO removidas por default (flag `--include-migration` explicita).

**Remove domain/endpoint — confirmacao:**
Output: lista de arquivos com path completo, contagem, prompt "Are you sure? [y/N]". Default N. Usa `promptConfirm` existente de `new.go` (extrair para package `commands` shared se necessario). Flag `--yes` pula prompt.

**Doctor — output formatado:**
```
gopherplate doctor

  [OK] Go 1.25.9
  [OK] Docker Desktop (running)
  [OK] golangci-lint 2.11.4
  [!!] swag — not found (install: go install github.com/swaggo/swag/cmd/swag@latest)
  [OK] goose v3.24.3
  ...
  
  Project: go.mod found (github.com/test/my-service)
  Docker:  gopherplate-db (running), gopherplate-redis (running)
```

**Add endpoint — template minimo:**
3 arquivos: UC (com ClassifyError/toAppError/span), DTO (Input/Output), test (success + error). NAO modifica handler nem router (instruções impressas). Usa mesmos templates do `add domain` como base mas simplificado (sem entity, filter, repo, etc).

**Wiring — regeneracao completa:**
Escaneia `internal/domain/*/` para detectar dominios. Para cada, verifica existencia de: handler, router, repository. Gera 3 arquivos via templates Go condicionais:
1. `cmd/api/server.go` — imports de bootstrap, health, idempotency, config
2. `internal/infrastructure/web/router/router.go` — imports de handler por dominio, `Register<Domain>Routes()`
3. `internal/bootstrap/container.go` — Repos, UseCases, Handlers structs com campos por dominio

**Comando pai `remove`:**
Criar `var removeCmd = &cobra.Command{Use: "remove"}` com subcommands `domain` e `endpoint`. Registrar `rootCmd.AddCommand(removeCmd)`.

### Files to Create

- `cmd/cli/commands/remove_domain.go` + `_test.go`
- `cmd/cli/commands/doctor.go` + `_test.go`
- `cmd/cli/commands/add_endpoint.go` + `_test.go`
- `cmd/cli/commands/remove_endpoint.go` + `_test.go`
- `cmd/cli/commands/wiring.go` + `_test.go`
- `cmd/cli/templates/domain/endpoint_usecase.go.tmpl`
- `cmd/cli/templates/domain/endpoint_dto.go.tmpl`
- `cmd/cli/templates/domain/endpoint_usecase_test.go.tmpl`

### Files to Modify

- `cmd/cli/commands/root.go` — registrar `removeCmd`, `doctorCmd`, `wiringCmd`
- `cmd/cli/scaffold/remover.go` — fix `FeatureFiles["redis"]` para `pkg/cache/redisclient`
- `cmd/cli/scaffold/wiring.go` — adicionar template condicional de `bootstrap/test_helpers.go`

### Dependencies

- Nenhuma dependencia externa nova

## Tasks

- [ ] TASK-1: Fix bug --no-redis (FeatureFiles granular)
  - Mudar `FeatureFiles["redis"]` de `"pkg/cache"` para `"pkg/cache/redisclient"`
  - Manter `"pkg/idempotency"` como esta (idempotency depende 100% de Redis)
  - Resultado: `pkg/cache/cache.go` (interface) e `pkg/cache/singleflight.go` permanecem; apenas `redisclient/` removido
  - Verificar: `gopherplate new test --no-redis --yes` → `go build ./...` passa; `go mod tidy` remove go-redis
  - files: `cmd/cli/scaffold/remover.go`, `cmd/cli/scaffold/remover_test.go`
  - tests: TC-U-01, TC-U-02, TC-U-03

- [ ] TASK-2: Fix bug --no-auth (bootstrap test_helpers condicional)
  - Adicionar em `CleanupWiring` (wiring.go) geracao de `internal/bootstrap/test_helpers.go` via template Go:
    - `Auth=true`: gera `NewForTest`, `SetupTestRouter`, `SetupTestRouterWithAuth`, `newTestEngine`, `registerTestHealthRoutes` (identico ao atual)
    - `Auth=false`: gera tudo EXCETO `SetupTestRouterWithAuth` — remove imports de `middleware.ServiceKeyConfig`, `middleware.ParseServiceKeys`, `middleware.ServiceKeyAuth`
  - O template tem ~100 linhas — adicionar como `testHelpersGoTemplate` const em wiring.go (mesmo pattern de `serverGoTemplate`)
  - Tambem gerar `internal/bootstrap/container.go` via template para que o bootstrap exista em projetos novos sem examples
  - Verificar: `gopherplate new test --no-auth --yes` → `go build ./...` passa
  - files: `cmd/cli/scaffold/wiring.go`
  - tests: TC-U-04, TC-U-05, TC-U-06

- [ ] TASK-3: Criar comando pai `remove` + implementar `remove domain`
  - Criar `var removeCmd = &cobra.Command{Use: "remove", Short: "Remove project components"}`
  - Em root.go: `rootCmd.AddCommand(removeCmd)`
  - Criar `cmd/cli/commands/remove_domain.go`:
    - `removeCmd.AddCommand(removeDomainCmd)`
    - `removeDomainCmd` com `Args: cobra.ExactArgs(1)`, flag `--yes`
    - Fluxo: detectar module path → validar dominio existe (`internal/domain/<name>/`) → listar arquivos (domain/, usecases/, repo, handler, router) → se nao `--yes`: mostrar lista + promptConfirm → remover com `os.RemoveAll` → imprimir instrucoes cleanup (server.go, router.go, migrations)
    - Migrations: listar na msg mas NAO deletar (imprimir: "Migration files were NOT deleted. Review: <path>")
    - Extrair `promptConfirm` para funcao compartilhada em commands package (atualmente privada em new.go)
  - Testes: dominio existe (happy), dominio nao existe (error), --yes skips confirm
  - files: `cmd/cli/commands/remove_domain.go`, `cmd/cli/commands/remove_domain_test.go`, `cmd/cli/commands/root.go`, `cmd/cli/commands/prompts.go`
  - tests: TC-U-07, TC-U-08, TC-U-09, TC-U-10, TC-U-11, TC-U-12

- [ ] TASK-4: Implementar `gopherplate doctor`
  - Criar `cmd/cli/commands/doctor.go`:
    - Struct `toolCheck{name, command, installHint}` para cada tool
    - Tools: Go (`go version`), Docker (`docker info`), golangci-lint, swag, goose, air, k6, kind, kubectl
    - Para cada: `exec.LookPath` + run version command → parse version → print [OK] ou [!!]
    - Project checks: `go.mod` existe, Docker containers rodando (`docker ps --filter name=<service>`)
    - Output formatado com [OK]/[!!] prefix
  - Em root.go: `rootCmd.AddCommand(doctorCmd)`
  - Testes: mock de `exec.LookPath` via interface ou test com known-installed Go
  - files: `cmd/cli/commands/doctor.go`, `cmd/cli/commands/doctor_test.go`, `cmd/cli/commands/root.go`
  - tests: TC-U-13, TC-U-14, TC-U-15

- [ ] TASK-5: Implementar `gopherplate add endpoint`
  - Criar `cmd/cli/commands/add_endpoint.go`:
    - `addCmd.AddCommand(addEndpointCmd)` (subcomando de `add`, junto com `domain`)
    - Args: `<domain> <endpoint-name>` (ex: `order cancel`)
    - Validacoes: domain existe (`internal/domain/<name>/`), endpoint nao existe (`internal/usecases/<domain>/<endpoint>.go`), nome valido (snake_case)
    - Gera 3 arquivos via templates: `endpoint_usecase.go.tmpl`, `endpoint_dto.go.tmpl`, `endpoint_usecase_test.go.tmpl`
    - Templates usam TemplateData expandido com campo `EndpointName` (PascalCase, camelCase, snake_case)
    - UC template: segue mesmo padrao dos CRUD (ClassifyError, toAppError, SpanFromContext)
    - Imprime instrucoes de wiring: "Add handler method to handler/<domain>.go:\n  func (h *<Domain>Handler) Cancel(c *gin.Context) { ... }\nAdd route to router/<domain>.go:\n  rg.POST(\"/<plural>/:<id>/cancel\", h.Cancel)"
  - Criar 3 templates em `cmd/cli/templates/domain/`
  - Testes: happy (3 files created), domain not found, endpoint exists, invalid name
  - files: `cmd/cli/commands/add_endpoint.go`, `cmd/cli/commands/add_endpoint_test.go`, `cmd/cli/commands/root.go`, `cmd/cli/templates/domain/endpoint_usecase.go.tmpl`, `cmd/cli/templates/domain/endpoint_dto.go.tmpl`, `cmd/cli/templates/domain/endpoint_usecase_test.go.tmpl`
  - tests: TC-U-16, TC-U-17, TC-U-18, TC-U-19, TC-U-20

- [ ] TASK-6: Implementar `gopherplate remove endpoint`
  - Criar `cmd/cli/commands/remove_endpoint.go`:
    - `removeCmd.AddCommand(removeEndpointCmd)` (subcomando de `remove`)
    - Args: `<domain> <endpoint-name>`, flag `--yes`
    - Validacoes: domain existe, endpoint existe, endpoint NAO e CRUD padrao (create/get/update/delete/list)
    - CRUD protection: se nome e um dos 5 CRUD, retorna erro: "Cannot remove standard CRUD endpoint 'create'. Use 'gopherplate remove domain order' to remove the entire domain."
    - Lista: `<name>.go`, `dto/<name>.go`, `<name>_test.go` → confirmacao → remove
  - Testes: happy (--yes), CRUD protected, user cancels (N)
  - files: `cmd/cli/commands/remove_endpoint.go`, `cmd/cli/commands/remove_endpoint_test.go`, `cmd/cli/commands/root.go`
  - tests: TC-U-21, TC-U-22, TC-U-23
  - depends: TASK-3, TASK-5

- [ ] TASK-7: Implementar `gopherplate wiring`
  - Criar `cmd/cli/commands/wiring.go`:
    - `rootCmd.AddCommand(wiringCmd)`
    - Fluxo: detectar module path → escanear `internal/domain/` → para cada subdir: verificar handler/router/repo existem → coletar lista de dominios → gerar 3 arquivos via templates
    - Templates (adicionados em wiring.go como consts, mesmo pattern do serverGoTemplate existente):
      - `serverGoTemplate` (ja existe mas precisa ser expandido para N dominios dinamicos)
      - `routerGoTemplate` (idem)
      - `bootstrapContainerGoTemplate` (novo — gera Container com N dominios)
    - Cada dominio no template tem: import de UC/repo/handler, struct field, wiring in buildRepos/buildUseCases/buildHandlers, Register<Domain>Routes
    - Deteccao: `filepath.WalkDir("internal/domain", ...)` → filtra subdirs → converte para PascalCase/camelCase/plural
    - Pede confirmacao antes de sobrescrever: "This will regenerate server.go, router.go and bootstrap/container.go. Continue? [Y/n]" (default Y neste caso)
  - Testes: 3 dominios detectados (happy), 0 dominios (edge), arquivos gerados compilam
  - files: `cmd/cli/commands/wiring.go`, `cmd/cli/commands/wiring_test.go`, `cmd/cli/commands/root.go`
  - tests: TC-U-24, TC-U-25, TC-U-26

## Parallel Batches

```text
Batch 1: [TASK-1, TASK-2]                — bug fixes (exclusive files: remover.go vs wiring.go)
Batch 2: [TASK-3, TASK-4, TASK-5]        — new commands (root.go shared-additive, command files exclusive)
Batch 3: [TASK-6, TASK-7]                — remove endpoint + wiring (TASK-6 depends TASK-3+5, TASK-7 independent)
```

File overlap analysis:
- `cmd/cli/scaffold/remover.go`: TASK-1 only → exclusive
- `cmd/cli/scaffold/wiring.go`: TASK-2 only → exclusive
- `cmd/cli/commands/root.go`: TASK-3, TASK-4, TASK-5, TASK-6, TASK-7 → shared-additive (each adds `AddCommand`)
- `cmd/cli/commands/prompts.go`: TASK-3 creates, TASK-6 uses → exclusive to TASK-3
- All new command files: exclusive to their task
- Note: root.go shared-additive across batches — each task adds one line, no conflict risk

## Validation Criteria

- [ ] `go build ./...` passa
- [ ] `make lint` passa
- [ ] `go test ./cmd/cli/...` passa
- [ ] `gopherplate new test --no-redis --yes` → `go build ./...` no projeto gerado
- [ ] `gopherplate new test --no-auth --yes` → `go build ./...` no projeto gerado
- [ ] `gopherplate remove domain order --yes` remove todos os arquivos (exceto migrations)
- [ ] `gopherplate remove domain order` (sem --yes) mostra lista + pede confirmacao
- [ ] `gopherplate doctor` lista tools formatado com [OK]/[!!]
- [ ] `gopherplate add endpoint order cancel` cria 3 arquivos com ClassifyError
- [ ] `gopherplate remove endpoint order cancel --yes` remove 3 arquivos
- [ ] `gopherplate remove endpoint order create` recusa (CRUD protegido)
- [ ] `gopherplate wiring` regenera server.go + router.go + container.go para todos os dominios detectados
- [ ] Zero referencias residuais ao bitbucket/go-boilerplate em qualquer arquivo gerado

## Execution Log

<!-- Ralph Loop appends here automatically — do not edit manually -->
