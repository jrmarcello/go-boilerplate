# Decisão de Arquitetura: Clean Architecture

## Contexto

Aplicações complexas tendem a se tornar difíceis de manter, testar e evoluir quando as regras de negócio estão acopladas a detalhes de implementação (frameworks, banco de dados, UI).## Decisão

Adotamos a **Clean Architecture** baseada estritamente em seus pilares fundamentais, priorizando a separação de preocupações e a testabilidade, sem necessariamente adotar a terminologia de "Portas e Adaptadores" (Hexagonal).

### Regra Fundamental (The Dependency Rule)

A regra suprema que seguimos é: **dependências de código fonte devem apontar apenas para dentro**. As camadas internas (Domínio) não devem saber nada sobre as camadas externas (Web, DB).

- **Entidades (Domain)**: Objetos de negócio puros. Não dependem de nada.
- **Casos de Uso (Usecases)**: Regras de aplicação. Dependem apenas do Domínio.
- **Adapters/Infra (Infrastructure)**: Implementações técnicas. Dependem dos Casos de Uso e Domínio.

## Justificativa

1. **Independência de Frameworks**: Frameworks são ferramentas e devem ser mantidos nas bordas (Infraestrutura).
2. **Testabilidade**: A lógica de negócio pode ser testada unitariamente sem subir banco de dados ou servidor web.
3. **Independência de Banco de Dados**: O banco de dados é um detalhe de infraestrutura. A regra de negócio interage com interfaces (repositórios).
4. **Independência de UI**: A API Web é apenas um mecanismo de entrega.

## Consequências

- **Positivas**:
  - Código desacoplado e fácil de testar.
  - Facilidade para trocar tecnologias de ponta (ex: biblioteca de log, banco de dados).
  - Manutenção simplificada a longo prazo.

- **Negativas**:
  - Boilerplate inicial (criação de arquivos em camadas separadas).
  - Exige disciplina para interfaces e DI.

## Implementação

### Estrutura de Pastas

Organizamos o projeto para refletir as camadas da arquitetura:

```text
internal/
├── domain/            # (Inner Layer) Regras de Negócio Corporativas
│   └── entity/        # Structs, Value Objects e Interfaces de Erros Puros
│
├── usecases/          # (Application Layer) Regras de Aplicação
│   └── entity/        # Interactors que orquestram o fluxo
│       ├── create.go  # Lógica específica
│       └── interfaces.go # Define interfaces que a Infra deve implementar (ex: Repository)
│
└── infrastructure/    # (Outer Layer) Detalhes Técnicos
    ├── db/            # Implementação real do banco (Postgres)
    └── web/           # Controllers, Routers (Gin)
```

### Injeção de Dependência (DI)

Seguimos o **Princípio de Inversão de Dependência (DIP)**.

1. **Definição**: O Use Case precisa salvar dados, mas não pode depender do banco de dados (Infra).
2. **Abstração**: O Use Case define uma interface `Repository` na camada `usecases` (ou `domain`).
3. **Implementação**: A camada `infrastructure` implementa essa interface.
4. **Injeção**: No `main.go` (ou `router`), injetamos a implementação concreta dentro do Use Case.

```go
// Camada Use Case (Define o contrato)
type Repository interface {
    Create(ctx context.Context, e *entity.Entity) error
}

// Camada Infrastructure (Cumpre o contrato)
type PostgresRepo struct { db *sqlx.DB }
func (r *PostgresRepo) Create(ctx context.Context, e *entity.Entity) error { ... }

// Main (Cola tudo)
repo := postgres.NewRepo(dbConn)       // Infra
useCase := entityuc.NewCreate(repo)    // Use Case recebendo Infra como interface
handler := web.NewHandler(useCase)     // Web recebendo Use Case
```
