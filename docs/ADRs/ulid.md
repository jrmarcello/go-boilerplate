# Decisão de Arquitetura: Uso de ULID

## Contexto

A escolha do formato de Identificador Único é crítica para sistemas distribuídos e bancos de dados. Precisamos de IDs que sejam únicos, performáticos e práticos de usar.

## Decisão

Utilizar **ULID (Universally Unique Lexicographically Sortable Identifier)** para chaves primárias de entidades.

### Formato

```text
01ARZ3NDEKTSV4RRFFQ69G5FAV
└──────┬──────┘└────┬─────┘
   Timestamp     Randomness
   (48 bits)     (80 bits)
```

## Justificativa

1. **Ordenação Lexicográfica**: ULIDs são ordenáveis por tempo, melhorando performance de inserção em índices B-Tree.
2. **Legibilidade**: Base32 (Crockford's) resulta em 26 caracteres URL-safe.
3. **Compatibilidade**: 128-bit compatíveis com colunas `UUID` no Postgres.
4. **Timestamp Embutido**: Permite saber quando o registro foi criado pelo ID.

## Consequências

- **Positivas**:
  - Menor fragmentação de índices comparado a UUID v4.
  - IDs mais curtos e legíveis em logs e URLs.
  - Geração descentralizada (não precisa de sequência no DB).

- **Negativas**:
  - Biblioteca adicional (`oklog/ulid`).
  - Menos comum que UUID (curva de aprendizado para equipe).

## Comparação

| Característica | UUID v4 | UUID v7 | ULID |
| -------------- | ------- | ------- | ---- |
| Ordenável | ❌ Não | ✅ Sim | ✅ Sim |
| Colisão | Rara | Rara | Rara |
| Tamanho (String) | 36 chars | 36 chars | 26 chars |
| Indexação DB | Ruim | Ótima | Ótima |
| URL Safe | ❌ Não | ❌ Não | ✅ Sim |

## Implementação

### Biblioteca

```go
import "github.com/oklog/ulid/v2"
```

### Value Object

```go
// domain/entity/vo/id.go
type ID string

func NewID() ID {
    entropy := ulid.Monotonic(rand.Reader, 0)
    return ID(ulid.MustNew(ulid.Timestamp(time.Now()), entropy).String())
}

func ParseID(s string) (ID, error) {
    _, err := ulid.Parse(s)
    if err != nil {
        return "", ErrInvalidID
    }
    return ID(s), nil
}
```

### Armazenamento (Postgres)

```sql
CREATE TABLE entities (
    id CHAR(26) PRIMARY KEY,  -- ULID como string
    ...
);
```
