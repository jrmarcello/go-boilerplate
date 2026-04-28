# SDD + Ralph Loop — Desenvolvimento Orientado a Especificação

## Visão Geral

O template inclui um fluxo de desenvolvimento integrado baseado em dois conceitos:

- **SDD (Specification-Driven Development)** — Metodologia onde especificações
  estruturadas guiam a implementação, em vez de codar diretamente.
- **Ralph Loop** — Execução autônoma da spec **em uma única passagem**, paralelizando
  tasks via worktrees, com self-review obrigatória antes de apresentar resultados.

Juntos, formam um workflow completo:
**debater → especificar → revisar spec → executar → revisar implementação → commitar**.

## Quick Start

```text
# 1. Debater o escopo (conversa)
# 2. Gerar uma especificação (com self-review automática)
/spec "Add audit logging to user write operations"

# 3. Revisar a spec apresentada, pedir ajustes se necessário, aprovar

# 4. Executar autonomamente — paraleliza onde der, faz self-review, apresenta
/ralph-loop .specs/user-audit-log.md

# 5. Revisar o resultado apresentado, pedir ajustes se necessário, aprovar commit

# 6. (Opcional) Auditoria independente contra os REQs
/spec-review .specs/user-audit-log.md
```

## Quando Usar

| Cenário | Recomendação |
| --- | --- |
| Feature nova com 3+ tasks | `/spec` + `/ralph-loop` |
| Bug fix simples | `/fix-issue` (workflow existente) |
| Novo endpoint isolado | `/new-endpoint` ou `/spec` (depende da complexidade) |
| Refactor cross-cutting | `/spec` + `/ralph-loop` (paraleliza por área) |
| Tarefa trivial (1-2 arquivos) | Execução direta, sem spec |

**Princípio: profundidade proporcional** — a spec deve ser tão detalhada quanto a
complexidade da task exige.

## O fluxo em detalhe

### Etapa 1 — `/spec`: três fases

```text
Author → Self-review (3 agents em paralelo) → Present
```

1. **Author** — `/spec` redige a spec a partir do `.specs/TEMPLATE.md`: Context,
   Requirements, Test Plan, Design, Tasks, Parallel Batches, Validation Criteria.
   Detecta arquivos shared-additive (e.g. `cmd/api/server.go`) e aplica o
   accumulator pattern já no momento da escrita.
2. **Self-review** — dispara três agents em paralelo numa única mensagem:
   - `spec-reviewer` — gaps, ambiguidade, DAG, accumulator pattern, layer rules
   - `test-reviewer` — coverage do Test Plan, boundary TCs, infra-failure TCs
   - `code-reviewer` — Design vs convenções (Clean Architecture, apperror, span
     classification, DI, idempotency)
   Trivial fixes são aplicados inline; pontos de julgamento viram "Pontos de atenção".
3. **Present** — spec é apresentada com resumo, stats do Test Plan, Parallel
   Batches, fixes aplicados e pontos de atenção. Aguarda aprovação.

Se você pedir mudanças após o present, **a self-review re-roda do zero** antes de
re-apresentar. Isso é intencional: protege contra regressões nas próprias correções.

### Etapa 2 — `/ralph-loop`: cinco fases

```text
Validate → Execute → Self-review → Present → Commit (após aprovação)
```

1. **Validate** — confere status da spec (`APPROVED`/`IN_PROGRESS`), presença de
   Test Plan e Parallel Batches.
2. **Execute** — para cada batch sequencialmente:
   - **1 task no batch:** executa inline (TDD se houver `tests:`, merge se for
     `TASK-MERGE-*`).
   - **2+ tasks no batch:** dispara N `Agent(isolation: "worktree")` numa única
     mensagem. Cada agent roda RED→GREEN→REFACTOR no seu worktree isolado.
   - **Auto-rollback:** se algum agent falhar, **NADA é mergeado silenciosamente**.
     A skill apresenta as opções (a) merge sucedidos + skip falho, (b) descartar
     tudo e re-rodar, (c) parar para investigação manual. Default = (c).
3. **Self-review** — três agents em paralelo (`code-reviewer`, `test-reviewer`,
   `security-reviewer`) auditam o diff. Trivial fixes inline; CRITICAL/HIGH e
   pontos de julgamento sobem como "Pontos de atenção".
4. **Present** — apresenta resumo da execução, diff stat, fixes aplicados, pontos
   de atenção, validação local. Aguarda aprovação.
5. **Commit** — só após "ok"/"pode commitar". Stage apenas os arquivos listados nas
   `files:` das tasks + a spec + os fragments mergeados.

Se você pedir mudanças após o present, **a self-review re-roda do zero** antes de
re-apresentar.

## Estrutura da Spec

As specs ficam em `.specs/` e seguem o template em `.specs/TEMPLATE.md`. Veja o
template para a estrutura completa; o resumo:

- **Status** state machine: `DRAFT` → `APPROVED` → `IN_PROGRESS` → `DONE` | `FAILED`
- **Requirements** em GIVEN/WHEN/THEN
- **Test Plan** agrupado por camada (TC-D-NN, TC-UC-NN, TC-E2E-NN, TC-S-NN)
- **Tasks** com `files:`, `tests:`, `depends:` — formam o DAG que dirige o paralelismo
- **Parallel Batches** auto-gerada por `/spec`
- **Validation Criteria** concretos (comandos, estados observáveis)
- Items incertos marcados `[NEEDS CLARIFICATION]`

## Paralelismo

### Detecção de Batches

`/spec` analisa as tasks e gera a seção **Parallel Batches** automaticamente. Duas
tasks **não podem** rodar em paralelo se:

1. **Dependência explícita** — uma aparece no `depends:` da outra
2. **Overlap de arquivos** — mesma entrada em `files:`

### Classificação de arquivos compartilhados

| Classificação | Definição | Estratégia |
| --- | --- | --- |
| **Exclusive** | Só uma task toca o arquivo | Paralelo direto |
| **Shared-additive** | Múltiplas tasks adicionam código (ex: DI wiring em `cmd/api/server.go`, rotas em `cmd/api/router.go`) | Accumulator pattern — fragments em `.specs/wiring/<spec-slug>/` + `TASK-MERGE-<TARGET>` |
| **Shared-mutative** | Múltiplas tasks modificam código existente no mesmo arquivo | Serializar — batches diferentes |

### Accumulator pattern (detalhe)

Para shared-additive, cada task paralela:

1. **Remove** o arquivo compartilhado do seu próprio `files:`.
2. **Ganha** um fragment em
   `.specs/wiring/<spec-slug>/<task-id>.<target-slug>.fragment.md` listado no `files:`.
3. Uma `TASK-MERGE-<TARGET>` no batch seguinte declara o arquivo compartilhado em
   seu `files:`, lista todos os fragments, e tem `depends:` cobrindo todos os
   produtores.

`/ralph-loop` executa a `TASK-MERGE` inline (sequencial), aplica os fragments em
ordem alfabética por `<task-id>`, e roda `gofmt -w` + `go build ./...` no final.
Conflito entre dois fragments no mesmo anchor → para a merge, surface ao usuário.

Anchors registrados (canonical para este projeto): ver
[`.claude/rules/sdd.md`](../../.claude/rules/sdd.md) §Registered anchors.

### Auto-rollback em batches paralelos

Se 1 agent num batch paralelo falhar, `/ralph-loop` **nunca** mergeia silenciosamente
os sucedidos. Em vez disso, apresenta:

```text
⚠️ Batch [TASK-3, TASK-4, TASK-5] — partial failure.

✅ TASK-3: <summary>
✅ TASK-4: <summary>
❌ TASK-5: <one-line failure cause>

Nothing has been merged into main. Choose:
  (a) merge X and Y, leave Z for me to fix manually
  (b) discard everything, rerun the batch with adjustments
  (c) stop here so I can investigate
```

Default sem resposta = (c). Veja `.claude/rules/sdd.md` §Auto-rollback semantics.

### Execução inter-spec (paralelismo entre specs)

Specs independentes rodam em paralelo abrindo múltiplos terminais — cada um com um
Claude executando `/ralph-loop` numa spec diferente, em seu próprio worktree.

## Integração com workflows existentes

| Workflow | Integração |
| --- | --- |
| `/new-endpoint` | Pode ser usado como referência para tasks na spec |
| `/fix-issue` | Para bugs simples, continue usando — mais direto |
| `/validate` | `/ralph-loop` já roda fmt/vet/lint/build/test ao fim do Phase 2; `/validate` adiciona Kind + smoke |
| `/review`, `/full-review-team` | Use após `/ralph-loop` para review mais profundo |
| `/spec-review` | Auditoria independente da implementação contra os REQs |
| Lefthook (pre-commit) | Continua rodando normalmente |
| `lint-go-file.sh` (PostToolUse) | Roda a cada edit, inclusive dentro de worktrees paralelos |

## Referências

- [Specification-Driven Development (Thoughtworks/Martin Fowler)](https://martinfowler.com/articles/exploring-gen-ai/sdd-3-tools.html)
- [Ralph Wiggum Technique (Geoffrey Huntley)](https://ghuntley.com/loop/)
- [`.claude/rules/sdd.md`](../../.claude/rules/sdd.md) — regras formais (TC-IDs, fragment format, anchors, auto-rollback)
- [`.claude/skills/spec/SKILL.md`](../../.claude/skills/spec/SKILL.md) — definição do `/spec`
- [`.claude/skills/ralph-loop/SKILL.md`](../../.claude/skills/ralph-loop/SKILL.md) — definição do `/ralph-loop`
