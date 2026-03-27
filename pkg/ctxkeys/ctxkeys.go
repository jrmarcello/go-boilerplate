package ctxkeys

// Gin context keys used across middleware and handlers.
// Plain string constants so they work with gin.Context.Set/Get (which require string keys).
const (
	// RequestID stores the unique request ID.
	RequestID = "request_id"

	// CallerService stores the caller service name.
	CallerService = "caller_service"
)
