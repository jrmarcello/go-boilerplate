# Spec: Template CLI (`boilerplate`)

## Status: APPROVED

## Context

O go-boilerplate é um template para criar microserviços Go com Clean Architecture. Hoje, criar um novo serviço exige clonar o repo manualmente, fazer find-and-replace nos nomes, remover código de exemplo, ajustar `go.mod` — um processo propenso a erros que leva ~30min e frequentemente resulta em referências esquecidas ao template original.

A CLI `boilerplate` automatiza esse processo com dois comandos:
- `boilerplate new` — scaffold de um serviço novo a partir do template
- `boilerplate add domain` — scaffold de um novo domínio Clean Architecture dentro de um serviço existente

O engine de scaffold é compartilhado com o skill `/new-endpoint` do Claude Code, unificando a lógica de geração de código.

## Requirements

### `boilerplate new`

- [ ] REQ-1: Criar serviço a partir do template
  - **GIVEN** o usuário executa `boilerplate new my-service`
  - **WHEN** responde aos prompts interativos
  - **THEN** um diretório `my-service/` é criado com a estrutura completa do boilerplate, `go.mod` reescrito com o module path informado, e todos os imports internos atualizados

- [ ] REQ-2: Prompts interativos com bubbletea
  - **GIVEN** o usuário executa `boilerplate new`
  - **WHEN** o CLI inicia
  - **THEN** apresenta prompts sequenciais para:
    1. Nome do serviço (default: argumento do comando, se fornecido)
    2. Module path (ex: `github.com/appmax/payment-service`)
    3. Banco de dados: PostgreSQL / MySQL / SQLite3 / Outro (configuro depois)
    4. Protocolo: HTTP/REST (Gin) — *(gRPC: em breve)*
    5. Injeção de dependência: Manual — *(Uber Fx: em breve)*
    6. Cache Redis? [Y/n]
    7. Idempotência? [Y/n] — só aparece se Redis = sim
    8. Service Key Auth? [Y/n]
    9. Manter domínios de exemplo (user/role)? [Y/n]

- [ ] REQ-3: Flags para modo não-interativo
  - **GIVEN** o usuário passa flags (`--module`, `--db`, `--no-redis`, `--no-auth`, `--keep-examples`, `--protocol`, `--di`)
  - **WHEN** o CLI inicia
  - **THEN** pula os prompts correspondentes e usa os valores das flags

- [ ] REQ-4: Remoção condicional de código
  - **GIVEN** o usuário escolhe "sem Redis" nos prompts
  - **WHEN** o scaffold é gerado
  - **THEN** todo código relacionado a cache Redis é removido (config, imports, wiring em `server.go`, middleware de idempotência, arquivos `pkg/cache/`, `pkg/idempotency/`)

- [ ] REQ-5: Remoção de domínios de exemplo
  - **GIVEN** o usuário escolhe não manter domínios de exemplo
  - **WHEN** o scaffold é gerado
  - **THEN** os diretórios `internal/domain/user/`, `internal/domain/role/`, seus use cases, repositories, handlers, routers, migrations e wiring são removidos, deixando a estrutura limpa para o dev adicionar seus próprios domínios

- [ ] REQ-6: Suporte a múltiplos bancos de dados
  - **GIVEN** o usuário escolhe MySQL, SQLite3 ou "Outro"
  - **WHEN** o scaffold é gerado
  - **THEN** o driver correto é importado no código (`lib/pq` → `go-sql-driver/mysql` ou `modernc.org/sqlite`), e para "Outro" nenhum driver é importado (apenas `pkg/database` com `database/sql`)

- [ ] REQ-7: Git limpo
  - **GIVEN** o scaffold é gerado com sucesso
  - **WHEN** o processo finaliza
  - **THEN** o diretório `.git` do template é removido, `git init` é executado, e `go mod tidy` roda com sucesso

- [ ] REQ-8: Reescrita completa de imports
  - **GIVEN** o module path do template é `bitbucket.org/appmax-space/go-boilerplate`
  - **WHEN** o scaffold é gerado
  - **THEN** todas as ocorrências do module path em arquivos `.go`, `go.mod` e outros são substituídas pelo module path informado pelo usuário

### `boilerplate add domain`

- [ ] REQ-9: Scaffold de domínio completo
  - **GIVEN** o usuário executa `boilerplate add domain order` dentro de um projeto existente
  - **WHEN** o scaffold é gerado
  - **THEN** são criados:
    - `internal/domain/order/` — entity.go, errors.go, filter.go
    - `internal/usecases/order/` — create.go, get.go, list.go, update.go, delete.go
    - `internal/usecases/order/interfaces/` — repository.go
    - `internal/usecases/order/dto/` — create.go, get.go, list.go, update.go, delete.go
    - `internal/infrastructure/db/postgres/repository/order.go`
    - `internal/infrastructure/web/handler/order.go`
    - `internal/infrastructure/web/router/order.go`
    - Migration SQL com `CREATE TABLE orders`

- [ ] REQ-10: Wiring automático
  - **GIVEN** o scaffold de domínio é gerado
  - **WHEN** o dev verifica `cmd/api/server.go` e `internal/infrastructure/web/router/router.go`
  - **THEN** o novo domínio está registrado no DI wiring e nas rotas (ou instruções claras são exibidas sobre onde adicionar)

- [ ] REQ-11: Domínio personalizável
  - **GIVEN** o usuário executa `boilerplate add domain order`
  - **WHEN** o scaffold é gerado
  - **THEN** o nome do domínio é usado em: nomes de structs (`Order`, `OrderHandler`), tabela (`orders`), rotas (`/orders`), arquivos e packages

### Geral

- [ ] REQ-12: Instalação via `go install`
  - **GIVEN** o CLI vive em `cmd/cli/`
  - **WHEN** o dev executa `go install bitbucket.org/appmax-space/go-boilerplate/cmd/cli@latest`
  - **THEN** o binário `cli` (ou `boilerplate` via rename) é instalado no `$GOBIN`

- [ ] REQ-13: Mensagens de progresso claras
  - **GIVEN** qualquer comando é executado
  - **WHEN** cada etapa é concluída
  - **THEN** feedback visual é exibido (spinners, checkmarks, erros em vermelho)

- [ ] REQ-14: Documentação de uso
  - **GIVEN** o CLI é publicado
  - **WHEN** um dev quer aprender a usar
  - **THEN** `docs/guides/template-cli.md` explica instalação, comandos, exemplos, customização e troubleshooting

- [ ] REQ-15: Prompts "coming soon" para features não implementadas
  - **GIVEN** o usuário executa `boilerplate new`
  - **WHEN** chega aos prompts de Protocolo e Injeção de Dependência
  - **THEN** as opções disponíveis são exibidas com as futuras desabilitadas:
    - Protocolo: `HTTP/REST (Gin) ✓` | `gRPC (em breve)` | `HTTP + gRPC (em breve)`
    - DI: `Manual (padrão) ✓` | `Uber Fx (em breve)`
  - As opções "em breve" são visíveis mas não selecionáveis, sinalizando ao dev que a feature existe no roadmap

- [ ] REQ-16: Engine extensível para features futuras
  - **GIVEN** o engine de scaffold recebe uma `Config` struct
  - **WHEN** gRPC ou Uber Fx forem implementados no template (specs futuras)
  - **THEN** basta: (1) adicionar os templates `.tmpl` correspondentes, (2) adicionar a lógica condicional no engine, (3) habilitar a opção no prompt — sem reestruturar a CLI
  - A `Config` struct já deve incluir campos `Protocol` (enum: `http`, `grpc`, `both`) e `DI` (enum: `manual`, `fx`) com valores default `http` e `manual`

## Design

### Architecture Decisions

1. **Cobra + bubbletea**: Cobra para estrutura de comandos (subcomandos, flags, help, autocomplete). Bubbletea para prompts interativos (selects, confirms, inputs com estilo).

2. **Engine de scaffold compartilhado**: O pacote `internal/scaffold/` contém a lógica de geração independente da interface (CLI ou skill). Recebe uma struct de configuração e executa as transformações. O skill `/new-endpoint` será atualizado para consumir este engine.

3. **Templates embedded via `embed.FS`**: Templates `.tmpl` ficam em `cmd/cli/templates/` e são embedded no binário. Usa `text/template` do stdlib com helpers para pluralização e case conversion.

4. **Estratégia de `new`**: Copia o repo inteiro para o destino, depois aplica transformações (rewrite imports, remove código condicional, remove exemplos). Não usa `git clone` — os arquivos são embedded no binário.

5. **Estratégia de `add domain`**: Renderiza templates `.tmpl` com dados do domínio (nome, fields) e escreve nos paths corretos. Para wiring automático, faz append/insert em `server.go` e `router.go` usando AST parsing ou regex seguro.

6. **Nomeação do binário**: O package em `cmd/cli/main.go` produz binário `cli` por padrão. O Makefile terá `make build-cli` que compila para `bin/boilerplate`. Para `go install`, o dev pode usar alias ou renomearemos o diretório para `cmd/cli/`.

7. **Extensibilidade para features futuras (gRPC, Fx)**: A `Config` struct do engine já inclui campos `Protocol` e `DI` com valores default. Os prompts exibem as opções futuras como desabilitadas ("em breve"). Quando gRPC ou Fx forem implementados no template:
   - Criar templates `.tmpl` específicos (ex: `cmd/cli/templates/grpc/`, `cmd/cli/templates/fx/`)
   - Adicionar lógica condicional no engine: `{{if eq .Protocol "grpc"}}...{{end}}`
   - Habilitar a opção no prompt (remover o "em breve")
   - Nenhuma mudança estrutural na CLI é necessária

### Files to Create

```
cmd/cli/
  main.go                          # Entrypoint, Cobra root command
  commands/
    new.go                         # `boilerplate new` command + prompts
    add.go                         # `boilerplate add` parent command
    add_domain.go                  # `boilerplate add domain` command + prompts
    version.go                     # `boilerplate version`
  templates/
    boilerplate/                   # Snapshot completo do template (para `new`)
      (embedded copy of project structure, templatized)
    domain/                        # Templates para `add domain`
      entity.go.tmpl
      errors.go.tmpl
      filter.go.tmpl
      create_usecase.go.tmpl
      get_usecase.go.tmpl
      list_usecase.go.tmpl
      update_usecase.go.tmpl
      delete_usecase.go.tmpl
      repository_interface.go.tmpl
      dto_create.go.tmpl
      dto_get.go.tmpl
      dto_list.go.tmpl
      dto_update.go.tmpl
      dto_delete.go.tmpl
      repository_postgres.go.tmpl
      handler.go.tmpl
      router.go.tmpl
      migration.sql.tmpl

internal/scaffold/
  scaffold.go                      # Engine principal: Config → executa transformações
  rewriter.go                      # Reescrita de module paths e imports
  remover.go                       # Remoção condicional de código/arquivos
  renderer.go                      # Renderização de templates com text/template
  helpers.go                       # Template funcs: plural, camelCase, snakeCase, etc.

docs/guides/template-cli.md       # Guia completo de uso
```

### Files to Modify

```
Makefile                           # Adicionar build-cli target
go.mod                             # Adicionar dependências (cobra, bubbletea, lipgloss)
.claude/skills/new-endpoint/SKILL.md  # Atualizar para referenciar engine compartilhado
```

### Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/spf13/cobra` | Framework de CLI (commands, flags, help) |
| `github.com/charmbracelet/bubbletea` | TUI framework para prompts interativos |
| `github.com/charmbracelet/lipgloss` | Estilização de output (cores, borders) |
| `github.com/charmbracelet/huh` | Form components de alto nível sobre bubbletea (selects, confirms, inputs) |

> **Nota sobre huh**: O `charmbracelet/huh` é uma camada de abstração sobre bubbletea específica para forms/prompts. Simplifica bastante vs. bubbletea puro para o nosso caso de uso (selects, confirms, text inputs). Avaliar se a complexidade dos prompts justifica usar `huh` ou se bubbletea puro é suficiente. [NEEDS CLARIFICATION — avaliar durante TASK-2]

## Tasks

### Fase 1: Fundação

- [ ] TASK-1: Setup do CLI com Cobra
  - Criar `cmd/cli/main.go` com root command
  - Configurar subcomandos: `new`, `add domain`, `version`
  - Adicionar Cobra como dependência no `go.mod`
  - Adicionar `make build-cli` no Makefile (`go build -o bin/boilerplate ./cmd/cli`)
  - Verificar: `go build ./cmd/cli` compila, `bin/boilerplate --help` exibe ajuda
  - Referência: usar `cmd/api/main.go` e `cmd/migrate/main.go` como padrão de entrypoints existentes

- [ ] TASK-2: Prompts interativos para `new`
  - Implementar prompts em `cmd/cli/commands/new.go` usando bubbletea (ou huh — avaliar)
  - Prompts: service name, module path, DB choice, protocol, DI, Redis, idempotency, auth, keep examples
  - Prompts de Protocol e DI exibem opções futuras como desabilitadas: `gRPC (em breve)`, `Uber Fx (em breve)`
  - Lógica condicional: idempotency só aparece se Redis = sim
  - Suporte a flags para modo não-interativo (`--module`, `--db postgres`, `--protocol http`, `--di manual`, `--no-redis`, etc.)
  - Verificar: `bin/boilerplate new` exibe prompts e coleta respostas corretamente; opções "em breve" são visíveis mas não selecionáveis

- [ ] TASK-3: Engine de scaffold (`internal/scaffold/`)
  - Criar `scaffold.go` com struct `Config` (ServiceName, ModulePath, DB, Protocol, DI, features...) e método `Execute()`
  - `Config.Protocol` é enum (`http`, `grpc`, `both`) com default `http`; `Config.DI` é enum (`manual`, `fx`) com default `manual`
  - Na fase atual, `Execute()` ignora Protocol/DI (sempre usa http + manual), mas os campos existem para extensibilidade
  - Criar `renderer.go` com renderização de templates via `text/template`
  - Criar `helpers.go` com template funcs: `plural`, `singular`, `camelCase`, `pascalCase`, `snakeCase`, `kebabCase`
  - Criar `rewriter.go` com lógica de reescrita de module path em arquivos `.go` e `go.mod`
  - Criar `remover.go` com lógica de remoção condicional de arquivos/diretórios
  - Verificar: testes unitários para helpers e rewriter passam

### Fase 2: Comando `add domain`

- [ ] TASK-4: Templates de domínio (`.tmpl`)
  - Criar todos os templates em `cmd/cli/templates/domain/` baseados nos padrões existentes:
    - Domain: entity, errors, filter (baseado em `internal/domain/user/` e `internal/domain/role/`)
    - Usecases: create, get, list, update, delete, interfaces/repository, DTOs (baseado em `internal/usecases/user/`)
    - Infrastructure: repository postgres, handler, router (baseado em `internal/infrastructure/`)
    - Migration SQL
  - Usar `text/template` com variáveis: `{{.DomainName}}`, `{{.DomainNamePlural}}`, `{{.ModulePath}}`, etc.
  - Verificar: templates renderizam sem erro com dados mock

- [ ] TASK-5: Implementar `boilerplate add domain`
  - Implementar `cmd/cli/commands/add_domain.go`
  - Detectar module path do projeto atual via `go.mod`
  - Renderizar templates do TASK-4 nos paths corretos
  - Gerar migration SQL via `goose create` ou template com timestamp
  - Exibir instruções de wiring manual (quais linhas adicionar em `server.go` e `router.go`)
  - Verificar: `bin/boilerplate add domain order` gera todos os arquivos esperados, `go build ./...` passa

### Fase 3: Comando `new`

- [ ] TASK-6: Templates do boilerplate completo para `new`
  - Criar snapshot templatizado do projeto em `cmd/cli/templates/boilerplate/`
  - Substituir valores hardcoded por template vars: module path, service name, DB driver
  - Marcar seções condicionais: `{{if .Redis}}...{{end}}`, `{{if .Auth}}...{{end}}`, `{{if .KeepExamples}}...{{end}}`
  - Embed via `embed.FS` em `cmd/cli/templates/embed.go`
  - Verificar: `embed.FS` carrega todos os templates sem erro

- [ ] TASK-7: Implementar `boilerplate new`
  - Conectar prompts do TASK-2 ao engine do TASK-3
  - Implementar fluxo completo: renderizar templates → escrever no diretório destino → reescrever imports → remover condicionais → `git init` → `go mod tidy`
  - Tratar DB "Outro": gerar sem driver específico, com comentário indicando onde adicionar
  - Exibir resumo final: o que foi criado, próximos passos
  - Verificar: `bin/boilerplate new test-service --module github.com/test/test-service --db postgres --no-redis --no-auth --no-examples` gera projeto funcional, `cd test-service && go build ./cmd/api` compila

### Fase 4: Integração e documentação

- [ ] TASK-8: Integrar engine com skill `/new-endpoint`
  - Atualizar `.claude/skills/new-endpoint/SKILL.md` para referenciar `internal/scaffold/` como engine
  - Garantir que o skill pode invocar `go run ./cmd/cli add domain ...` ou usar o engine diretamente
  - Verificar: `/new-endpoint` continua funcionando corretamente

- [ ] TASK-9: Testes do engine de scaffold
  - Testes unitários para `internal/scaffold/`: helpers, rewriter, remover, renderer
  - Teste de integração: `add domain` gera arquivos que compilam
  - Teste de integração: `new` gera projeto que compila (pode usar `t.TempDir()`)
  - Verificar: `go test ./internal/scaffold/... -v` passa, `go test ./cmd/cli/... -v` passa

- [ ] TASK-10: Documentação (`docs/guides/template-cli.md`)
  - Seções: O que é, Instalação, Quick Start, Comandos (`new` e `add domain`), Exemplos, Flags, Customização de templates, Roadmap de features (gRPC, Fx), Troubleshooting, FAQ
  - Exemplos reais com output esperado
  - Seção sobre como customizar os templates embedded (fork + rebuild)
  - Seção "Em breve" explicando gRPC e Fx: o que são, por que estão no roadmap, e que os prompts já estão preparados
  - Verificar: documento cobre todos os REQs, links internos funcionam

- [ ] TASK-11: Makefile + README
  - Adicionar targets no Makefile: `build-cli`, `install-cli`
  - Adicionar seção no README sobre a CLI (breve, apontando para o guia)
  - Atualizar roadmap.md marcando o item como concluído
  - Verificar: `make build-cli` compila para `bin/boilerplate`, `make install-cli` instala em `$GOBIN`

## Validation Criteria

- [ ] `go build ./cmd/cli` compila sem erro
- [ ] `make build-cli` gera `bin/boilerplate`
- [ ] `bin/boilerplate --help` exibe ajuda formatada
- [ ] `bin/boilerplate new test-svc` (com flags) gera projeto que compila (`go build ./cmd/api`)
- [ ] `bin/boilerplate add domain order` gera arquivos corretos e `go build ./...` passa
- [ ] `make lint` passa (sem novos warnings)
- [ ] `make test` passa (testes do engine)
- [ ] `go vet ./...` passa
- [ ] Templates renderizam corretamente para todos os cenários de DB (postgres, mysql, sqlite, outro)
- [ ] Documentação em `docs/guides/template-cli.md` cobre instalação, uso e troubleshooting

## Execution Log

<!-- Ralph Loop appends here automatically — do not edit manually -->
