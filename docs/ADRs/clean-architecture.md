# Decisão de Arquitetura: Clean Architecture

## Contexto

Aplicações complexas tendem a se tornar difíceis de manter, testar e evoluir quando as regras de negócio estão acopladas a detalhes de implementação (frameworks, banco de dados, UI). Buscamos uma arquitetura que garanta longevidade ao projeto e facilite a manutenção.

## Decisão

Adotamos a **Clean Architecture** (Arquitetura Limpa), combinando conceitos de **Portas e Adaptadores** (Arquitetura Hexagonal) e **Inversão de Dependência**.

### Regra Fundamental

**Dependências de código fonte devem apontar apenas para dentro**, em direção às políticas de alto nível.

- **Entidades (Domain)**: Núcleo da aplicação. Regras de negócio puras.
- **Casos de Uso (Usecase)**: Orquestração de fluxo de dados.
- **Adaptadores (Infrastructure)**: Implementações concretas (DB, Web, etc).

## Justificativa

1. **Independência de Frameworks**: Frameworks são ferramentas, não o centro da aplicação.
2. **Testabilidade**: Regras de negócio testáveis sem UI, DB ou Web Server.
3. **Independência de Banco de Dados**: O DB é um detalhe. Podemos trocar Postgres por Mongo ou In-Memory sem tocar nas regras de negócio.
4. **Independência de Interface**: A UI (Web, CLI, Mobile) pode mudar sem afetar o core.

## Consequências

- **Positivas**:
  - Padronização do projeto.
  - Testes unitários triviais (mocks fáceis).
  - Evolução flexível (ex: começar com repositório em memória).

- **Negativas**:
  - Setup inicial mais verboso (mais arquivos e camadas).
  - Curva de aprendizado inicial para quem vem de MVC tradicional.
