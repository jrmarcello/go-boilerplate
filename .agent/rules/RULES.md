# AI_RULES.md — Regras para Agentes de IA

> Este arquivo define como **qualquer agente de IA** (GitHub Copilot, Claude, Gemini, ChatGPT, etc.) deve se comportar neste workspace. Não é específico para nenhuma ferramenta.

---

## Governança Principal

Os arquivos de governança do projeto são:

| Arquivo | Função |
| --- | --- |
| `AGENTS.md` | Regras obrigatórias, princípios arquiteturais, checklist |
| `CLAUDE.md` | Comandos úteis, padrões de código, convenções |
| `.agent/ARCHITECTURE.md` | Índice de agentes, skills e workflows |

**Regra:** Ler `AGENTS.md` e `CLAUDE.md` antes de qualquer implementação. As regras desses arquivos têm prioridade sobre este.

---

## Classificação de Requisições

Antes de qualquer ação, classificar a requisição:

| Tipo | Gatilhos | Ação |
| --- | --- | --- |
| **PERGUNTA** | "o que é", "como funciona", "explique" | Resposta em texto, sem edição de código |
| **CÓDIGO SIMPLES** | "corrigir", "adicionar", "alterar" (arquivo único) | Edição direta |
| **CÓDIGO COMPLEXO** | "implementar", "criar", "refatorar" (múltiplos arquivos) | Planejar antes de implementar |
| **INVESTIGAÇÃO** | "analisar", "listar", "overview" | Pesquisa e análise, sem edição |

---

## Roteamento de Agentes

Para tarefas de código, selecionar o agente mais adequado em `.agent/agents/`:

| Domínio | Agente |
| --- | --- |
| API, lógica de negócio, Go | `backend-specialist` |
| Banco de dados, queries, migrations | `database-architect` |
| Deploy, K8s, CI/CD, Docker | `devops-engineer` |
| Segurança, vulnerabilidades | `security-auditor` |
| Pentest, ataques | `penetration-tester` |
| Testes unitários, E2E, mocks | `test-engineer` |
| Debugging, troubleshooting | `debugger` |
| Performance, profiling | `performance-optimizer` |
| Documentação | `documentation-writer` |
| Análise de código legado | `code-archaeologist` |
| Exploração e pesquisa | `explorer-agent` |
| Orquestração multi-domínio | `orchestrator` |
| Produto e roadmap | `product-manager` |
| Planejamento de tarefas | `project-planner` |

**Protocolo:**

1. Identificar o domínio da requisição
2. Consultar o arquivo do agente em `.agent/agents/{agente}.md`
3. Carregar skills relevantes do frontmatter do agente
4. Aplicar as regras do agente na resposta

Para tarefas multi-domínio, usar `orchestrator`.

---

## Regras de Código Go

### Qualidade Obrigatória

- Seguir Clean Architecture: `domain` ← `usecases` ← `infrastructure`
- Nunca importar camadas externas de camadas internas
- Usar Value Objects para validação (`vo.ID`, `vo.Email`)
- Retornar erros de domínio puros (`entity.ErrNotFound`), sem referências HTTP
- Injetar dependências via construtor (interfaces, não implementações)
- Nomear variáveis de erro de forma única (evitar shadowing): `parseErr`, `saveErr`

### Convenções de Código

- Comentários e variáveis em inglês
- Respostas e comunicação em português brasileiro
- Formato de commit: `type(scope): description`
- Respostas JSON seguem wrapper padrão: `{"data": ...}` ou `{"errors": {"message": ...}}`
- Usar `httputil.SendSuccess()` e `httputil.SendError()` de `pkg/httputil` para respostas HTTP

### Validação com Make

Antes de considerar qualquer tarefa concluída:

```bash
make lint          # go vet + gofmt
make test          # Todos os testes
make test-unit     # Testes unitários
```

---

## Carregamento de Skills

Antes de implementar, consultar o skill relevante em `.agent/skills/`:

| Skill | Quando usar |
| --- | --- |
| `clean-code` | Sempre (qualquer código) |
| `api-patterns` | Endpoints, handlers, DTOs |
| `database-design` | Schema, queries, migrations |
| `testing-patterns` | Testes unitários, mocks |
| `tdd-workflow` | Desenvolvimento guiado por testes |
| `go-patterns` | Padrões idiomáticos Go |
| `go-performance` | Otimização de performance |
| `systematic-debugging` | Investigação de bugs |
| `vulnerability-scanner` | Análise de segurança |
| `k8s-argocd-deploy` | Deploy e operações |
| `code-review-checklist` | Revisão de código |

**Protocolo:** Ler o `SKILL.md` (índice) do skill, depois apenas as seções relevantes à tarefa.

---

## Ciclo Obrigatório de Execução

Toda implementação não trivial deve seguir:

1. **Planejar** — Definir escopo, arquivos afetados, riscos e estratégia de validação
2. **Implementar** — Aplicar mudanças seguindo arquitetura e convenções do projeto
3. **Testar** — Criar/ajustar testes e executar `make test`
4. **Validar** — Confirmar que tudo funciona (`make lint` + evidência de teste)

**Não encerrar tarefa sem evidência de validação.**

---

## Regras Universais

### Idioma

- Prompt do usuário em português → responder em português
- Código, variáveis e comentários → sempre em inglês

### Tarefas Complexas

- Para requisições vagas, perguntar antes de implementar
- Para mudanças estruturais, propor plano antes de executar
- Para múltiplas abordagens válidas, apresentar opções

### Proibições

- Nunca usar `--no-verify` em commits
- Nunca adicionar dependências sem perguntar
- Nunca colocar lógica de negócio em handlers HTTP
- Nunca retornar HTTP status codes do domínio
- Nunca deixar código comentado
- Nunca assumir que o usuário quer solução complexa quando uma simples resolve
