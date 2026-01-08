# Decisão de Arquitetura: Tratamento de Erros Centralizado

## Contexto

Em APIs Go, é comum misturar lógica de domínio com códigos HTTP (ex: retornar `400 Bad Request` de dentro de uma entidade), o que viola a Clean Architecture. Precisamos de um sistema que mantenha o domínio puro mas garanta respostas HTTP consistentes.

## Decisão

Adotamos um sistema de **Tratamento de Erros com Tradução em Camadas**.

### Fluxo de Erros

1. **Domain (Puro)**: Retorna erros semânticos (`ErrInvalidCPF`). **Ignora HTTP**.
2. **Use Case (Application)**: Pode retornar erros de aplicação (`AppError`) com códigos de erro (`INVALID_CPF`).
3. **Handler/Translator (Infra)**: Intercepta erros e decide o Status Code.

### Tradução

Implementamos um `HandleError` centralizado na camada de infraestrutura que:

1. Verifica se é um `AppError` (já traduzido).
2. Verifica se é um erro de domínio conhecido e traduz para `AppError` + HTTP Status correspondente.
3. Caso contrário, retorna **500 Internal Server Error**.

## Justificativa

- **Pureza do Domínio**: Entidades não dependem de bibliotecas HTTP ou frameworks.
- **Consistência**: Toda resposta de erro segue o mesmo formato JSON (`error`, `code`, `trace_id`).
- **Simplicidade**: Handlers delegam o tratamento para uma função única, removendo switches repetitivos.

## Exemplo de Implementação

```go
// Domain
var ErrInvalidCPF = errors.New("CPF inválido")

// Handler (Translator)
func translateDomainError(err error) *usecases.AppError {
    switch {
    case errors.Is(err, vo.ErrInvalidCPF):
        return personuc.ErrInvalidCPF // Mapeia para 400
    default:
        return nil
    }
}
```
