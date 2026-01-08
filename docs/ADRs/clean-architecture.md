# Decisão de Arquitetura: Clean Architecture

## Contexto

Aplicações complexas tendem a se tornar difíceis de manter, testar e evoluir quando as regras de negócio estão acopladas a detalhes de implementação (frameworks, banco de dados, UI). Buscamos uma arquitetura que garanta longevidade ao projeto e facilite a manutenção.

## Decisão

Adotamos a **Clean Architecture** proposta por Robert C. Martin, focando em seus **pilares fundamentais**:

1. **Regra da Dependência**: Dependências de código fonte apontam **apenas para dentro** (camadas internas nunca conhecem camadas externas).
2. **Entidades**: Objetos de domínio que encapsulam regras de negócio corporativas.
3. **Casos de Uso**: Orquestram o fluxo de dados e aplicam regras de negócio específicas da aplicação.
4. **Inversão de Dependência**: Camadas internas definem **interfaces**; camadas externas as implementam.

### Camadas

| Camada | Responsabilidade | Exemplo |
|--------|------------------|---------|
| **Domain** | Entidades e Value Objects puros | `Entity`, `ID`, `Email` |
| **Usecases** | Lógica de aplicação, DTOs, interfaces de repositório | `CreateUseCase`, `Repository` (interface) |
| **Infrastructure** | Implementações concretas (DB, Web, Cache) | `PostgresRepository`, `GinHandler` |

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

## Implementação

### Estrutura de Pastas

```
internal/
├── domain/              # 🟢 Camada mais interna (sem dependências externas)
│   └── entity/
│       ├── entity.go         # Entidade de domínio
│       ├── errors.go         # Erros de domínio
│       └── filter.go         # Filtros de busca
│
├── usecases/            # 🟡 Orquestração (depende apenas do Domain)
│   └── entity/
│       ├── interfaces/
│       │   └── repository.go  # Interface do repositório (definida aqui!)
│       ├── create.go          # Caso de uso de criação
│       ├── get.go
│       └── dto/               # Data Transfer Objects
│
└── infrastructure/      # 🔴 Camada externa (implementa interfaces)
    ├── db/postgres/
    │   └── repository/
    │       └── entity.go      # Implementação concreta do Repository
    └── web/
        ├── handler/
        │   └── entity.go      # Handler HTTP (Gin)
        └── router/
```

### Inversão de Dependência (DI)

O **Use Case** define a interface do repositório. A **Infrastructure** a implementa.

```go
// usecases/entity/interfaces/repository.go (Camada Interna)
type Repository interface {
    Save(ctx context.Context, e *entity.Entity) error
    FindByID(ctx context.Context, id vo.ID) (*entity.Entity, error)
}
```

```go
// infrastructure/db/postgres/repository/entity.go (Camada Externa)
type EntityRepository struct {
    db *sqlx.DB
}

func (r *EntityRepository) Save(ctx context.Context, e *entity.Entity) error {
    // Implementação concreta usando sqlx
}
```

### Composição (Bootstrap)

No `main.go` ou `server.go`, injetamos as dependências concretas:

```go
// cmd/api/server.go
func Run(cfg *config.Config) {
    db := postgres.Connect(cfg.DB.DSN)

    // Injeção de Dependência
    repo := repository.NewEntityRepository(db)   // Implementação concreta
    createUC := entityuc.NewCreateUseCase(repo)  // Use Case recebe a interface
    handler := handler.NewEntityHandler(createUC)

    router.Setup(handler)
}
```

O Use Case (`createUC`) **não sabe** que está falando com Postgres. Ele só conhece a interface `Repository`. Isso permite trocar a implementação (ex: para MongoDB ou um mock em testes) sem alterar o caso de uso.
