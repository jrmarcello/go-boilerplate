---
name: api-patterns
description: REST API patterns for Gin handlers — response format, routing, pagination, validation, and Swagger
---

# API Patterns

## Response Format

All JSON responses follow a unified wrapper from `pkg/httputil`:

```go
// Success
httputil.SendSuccess(c, http.StatusOK, data)
// → {"data": {...}}

// Success with metadata (lists)
httputil.SendSuccessWithMeta(c, http.StatusOK, items, meta, links)
// → {"data": [...], "meta": {...}, "links": {...}}

// Error
httputil.SendError(c, http.StatusBadRequest, "invalid input")
// → {"errors": {"message": "invalid input"}}

// Error with code
httputil.SendErrorWithCode(c, http.StatusConflict, "DUPLICATE", "already exists")
// → {"errors": {"message": "already exists", "code": "DUPLICATE"}}
```

Never use raw `c.JSON()` — always use the `pkg/httputil` wrappers.

## Handler Pattern

Every handler follows this flow:

1. Parse and validate request (bind JSON / read params)
2. Call use case
3. Translate errors via `HandleError`
4. Send response with `SendSuccess`

```go
func (h *EntityHandler) Create(c *gin.Context) {
    var req dto.CreateInput
    if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
        httputil.SendError(c, http.StatusBadRequest, "invalid request body: "+bindErr.Error())
        return
    }

    res, execErr := h.CreateUC.Execute(c.Request.Context(), req)
    if execErr != nil {
        HandleError(c, execErr)
        return
    }

    httputil.SendSuccess(c, http.StatusCreated, res)
}
```

## Pagination (List)

```go
func (h *EntityHandler) List(c *gin.Context) {
    filter := parseFilter(c)

    res, execErr := h.ListUC.Execute(c.Request.Context(), filter)
    if execErr != nil {
        HandleError(c, execErr)
        return
    }

    httputil.SendSuccessWithMeta(c, http.StatusOK, res.Data, res.Pagination, nil)
}
```

## Error Translation

Domain errors are pure. Handlers translate via centralized `HandleError`:

```go
func HandleError(c *gin.Context, err error) {
    var appErr *apperror.AppError
    if errors.As(err, &appErr) {
        httputil.SendErrorWithCode(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
        return
    }

    // Fallback: domain errors
    switch {
    case errors.Is(err, entity.ErrNotFound):
        httputil.SendError(c, http.StatusNotFound, err.Error())
    default:
        httputil.SendError(c, http.StatusInternalServerError, "internal server error")
    }
}
```

## Swagger

```bash
# Generate docs
swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal

# Toggle via env
SWAGGER_ENABLED=true  # show /swagger/* routes
```

## Route Organization

Routes registered in `internal/infrastructure/web/router/router.go`. Middleware wired at router level.
