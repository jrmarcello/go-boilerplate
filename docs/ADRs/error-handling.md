# Decisão de Arquitetura: Tratamento de Erros

## Contexto

Em APIs Go, é comum misturar lógica de domínio com códigos HTTP (ex: retornar `400 Bad Request` de dentro de uma entidade), o que viola a Clean Architecture. Precisamos de um sistema que mantenha o domínio puro mas garanta respostas HTTP consistentes.

## Decisão

Adotamos um sistema de **Tratamento de Erros com Tradução em Camadas**, onde cada camada tem responsabilidades claras:

| Camada | Responsabilidade | Conhece HTTP? |
| ------ | ---------------- | ------------- |
| **Domain** | Erros semânticos puros | ❌ Não |
| **Use Case** | Erros de aplicação (`AppError`) | ❌ Não |
| **Handler** | Tradução para HTTP Status | ✅ Sim |

### Fluxo

1. **Domain**: Retorna erros como `ErrNotFound`, `ErrInvalidEmail`.
2. **Use Case**: Propaga ou enriquece erros com contexto.
3. **Handler**: Intercepta, traduz para HTTP e responde com formato padronizado.

## Justificativa

1. **Pureza do Domínio**: Entidades não dependem de bibliotecas HTTP.
2. **Consistência**: Toda resposta de erro segue o mesmo formato JSON.
3. **Simplicidade**: Handlers delegam tratamento para função única `HandleError`.

## Consequências

- **Positivas**:
  - Domínio 100% testável sem mocks HTTP.
  - Formato de erro padronizado com `trace_id` para debug.
  - Código de handler limpo e enxuto.

- **Negativas**:
  - Necessidade de manter o `translator` atualizado com novos erros.

## Implementação

### Erros de Domínio

```go
// domain/entity/errors.go
var (
    ErrNotFound     = errors.New("entity not found")
    ErrInvalidEmail = errors.New("invalid email format")
)
```

### Tradutor (Handler)

```go
// infrastructure/web/handler/error.go
func HandleError(c *gin.Context, span trace.Span, err error) {
    status, code, message := translateError(err)
    
    span.SetStatus(codes.Error, code)
    c.JSON(status, ErrorResponse{
        Error:   message,
        Code:    code,
        TraceID: extractTraceID(span),
    })
}

func translateError(err error) (int, string, string) {
    switch {
    case errors.Is(err, entity.ErrNotFound):
        return 404, "NOT_FOUND", "Entity not found"
    case errors.Is(err, entity.ErrInvalidEmail):
        return 400, "INVALID_EMAIL", "Invalid email format"
    default:
        return 500, "INTERNAL_ERROR", "Internal server error"
    }
}
```

### Formato de Resposta

```json
{
    "error": "Entity not found",
    "code": "NOT_FOUND",
    "trace_id": "abc123..."
}
```
