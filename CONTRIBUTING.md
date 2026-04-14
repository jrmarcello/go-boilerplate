# Guia de Contribuicao

Obrigado pelo interesse em contribuir com o **Gopherplate**!

## Como comecar

1. Clone o projeto: `git clone https://github.com/jrmarcello/gopherplate`
2. Crie uma branch: `git checkout -b feat/minha-feature`
3. Setup: `make setup` (instala tools, sobe Docker, roda migrations)

## Propor features e reportar bugs

Use o **Issues** do GitHub para sugerir melhorias ou reportar problemas:

1. Acesse [Issues](https://github.com/jrmarcello/gopherplate/issues)
2. Verifique se ja existe uma issue similar
3. Crie uma nova issue com o tipo adequado:
   - **Bug**: algo nao funciona como esperado (inclua steps to reproduce)
   - **Enhancement**: nova funcionalidade ou melhoria (descreva o problema, nao so a solucao)
   - **Task**: melhoria tecnica, refactoring, docs

Se quiser implementar a feature, comente na issue antes de comecar para alinhar a abordagem.

## Desenvolvimento

Ferramentas necessarias: `Go 1.26`, `Docker`, `Make`.

```bash
make setup     # Setup completo (tools + Docker + migrations)
make dev       # Servidor com hot reload
make test      # Todos os testes
make lint      # golangci-lint + gofmt
```

Ferramentas opcionais (o Makefile mostra como instalar se faltarem):

- `k6` para load tests (`make load-smoke`)
- `kind` + `kubectl` para Kubernetes local (`make kind-setup`)

## Commits

Seguimos **Conventional Commits** (enforced por Lefthook):

```text
feat(scope): nova funcionalidade
fix(scope): correcao de bug
docs(scope): documentacao
refactor(scope): mudanca sem alterar comportamento
test(scope): testes
chore(scope): configuracao, dependencias
```

Exemplo: `feat(api): add pagination to list endpoint`

O scope e opcional mas recomendado: `api`, `config`, `cache`, `db`, `auth`, `ci`, `dx`, `docs`.

## Pull requests

Ao abrir um PR:

1. Descreva claramente o que foi feito e por que
2. Garanta que `make lint` e `make test` passam
3. Se mudou a API, regenere o Swagger: `make swagger`
4. **Nao edite `CHANGELOG.md` manualmente** — ele e gerado automaticamente a partir dos commits durante `make release` (via `git-cliff`). Para ver o changelog previsto a qualquer momento, use `make changelog`.

O pipeline roda automaticamente: lint, vulncheck, unit tests e E2E tests em paralelo.

## Releases

> Esta secao e para mantenedores. Contribuidores externos nao precisam rodar `make release`.

### Quando lancar uma nova versao

Rode `make release` quando houver pelo menos **um commit `feat:` ou `fix:` user-visible** acumulado desde a ultima tag. Commits apenas de `chore`, `docs`, `ci`, `refactor`, `test`, `style` **nao** justificam uma release — isso e ruido pra quem clona o template.

Verifique o que entraria:

```bash
git log v<ultima-tag>..HEAD --oneline
```

### Escolhendo a versao (semver pre-1.0)

O projeto esta em `0.x.x`. A regra pratica:

| Situacao desde a ultima tag | Bump | Exemplo |
| --- | --- | --- |
| Apenas `fix:` (bug em feature ja lancada) | PATCH | `0.12.0` -> `0.12.1` |
| Qualquer `feat:` (novo comando, endpoint, env var) | MINOR | `0.12.0` -> `0.13.0` |
| Breaking change | MINOR + secao "BREAKING CHANGES" | `0.12.0` -> `0.13.0` |

Quem decide MAJOR/MINOR/PATCH e voce no `VERSION=...`. O `git-cliff` so gera o changelog a partir dos commits.

### Fluxo

```bash
make release VERSION=0.13.0
```

O Makefile faz, em ordem:

1. Valida working tree limpa
2. `git-cliff --tag v0.13.0 --output CHANGELOG.md`
3. `git commit -m "chore(release): v0.13.0 [skip ci]"`
4. `git tag v0.13.0`
5. Pergunta `Push para origin/main + tag? [y/N]` — responda `y` para publicar
6. `git push origin main --follow-tags`

Apos o push, o workflow `.github/workflows/release.yml` dispara automaticamente e publica a **GitHub Release** em `~30s`, com notes geradas pelo `git-cliff` (escopo: so commits desde a tag anterior). Acompanhe com `gh run watch` ou na aba Actions.

### Checklist mental antes de rodar

- [ ] `git log v<ultima-tag>..HEAD --oneline` contem algum `feat:` ou `fix:`?
- [ ] `make test` e `make lint` passam?
- [ ] Estou na `main` atualizada (`git pull`)?
- [ ] Bump correto (`feat` -> MINOR, `fix` -> PATCH)?

## Testes

O CI exige **60% de coverage** minimo (pacotes com logica, excluindo handler/router/telemetry).
Coverage atual: ~89%. Use `make test-coverage` para verificar localmente.

Novas funcionalidades devem incluir:

- **Testes unitarios** para domain e usecases (hand-written mocks em `mocks_test.go`)
- **Testes de repositorio** com go-sqlmock
- **Testes de pkg/** com miniredis (cache, idempotency) ou sqlmock (database)
- **Testes E2E** com TestContainers para mudancas criticas
- **Smoke tests** com k6 (`make load-smoke`) para validacao funcional de endpoints
- Cobrir tanto **happy path** quanto **todos os error paths** possiveis

## SDD Workflow (features complexas)

Para features nao-triviais, use o fluxo Specification-Driven Development:

1. **Spec**: crie uma especificacao com `/spec "descricao"` — gera requisitos, test plan, tasks e analise de paralelismo em `.specs/`
2. **Review**: revise a spec, ajuste o que precisar, aprove (status APPROVED)
3. **Execute**: rode `/ralph-loop .specs/<nome>.md` para execucao autonoma task-by-task com TDD
4. **Validate**: `/spec-review .specs/<nome>.md` para revisao formal contra os requisitos

Detalhes em `docs/guides/sdd-ralph-loop.md` e `.claude/rules/sdd.md`.

## Error Handling

Erros seguem o padrao de 3 camadas (ADR-009):

- **Domain**: sentinels puros (`user.ErrNotFound`, `role.ErrDuplicateRoleName`)
- **Use Case**: mapeia via `toAppError()` + classifica span via `ClassifyError()`
- **Handler**: resolve generico via `errors.As()` + `codeToStatus` map — zero imports de dominio

Guia pratico: `docs/guides/error-handling.md`.

## Load Tests

Estrutura modular em `tests/load/`:

- `helpers.js` — HTTP client, assertions, UUID, headers
- `users.js` / `roles.js` — operacoes e smoke groups por dominio
- `main.js` — orquestrador de cenarios (smoke, load, stress, spike)

```bash
make load-smoke   # Smoke: 1 VU, 1 iteracao, validacao funcional
make load-test    # Load: ramping ate 50 VUs
make load-stress  # Stress: ate 200 VUs
make load-spike   # Spike: burst de 100 VUs
```

## Arquitetura

Antes de criar ou modificar arquivos, consulte:

- `CLAUDE.md` — visao geral da arquitetura e padroes
- `docs/adr/` — decisoes arquiteturais (Clean Architecture, IDs, config, errors, auth, migrations, pkg/)
- `docs/guides/error-handling.md` — guia pratico de error handling
