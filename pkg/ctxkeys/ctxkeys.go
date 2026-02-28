package ctxkeys

// contextKey is a private type for HTTP context keys.
// Prevents collisions with other packages using context.Value.
type contextKey string

const (
	// IPAddress stores the client IP.
	IPAddress contextKey = "ip_address"

	// UserID stores the authenticated user ID.
	UserID contextKey = "user_id"

	// ServiceKey stores the authenticated service key.
	ServiceKey contextKey = "service_key"

	// RequestID stores the unique request ID.
	RequestID contextKey = "request_id"

	// CallerService stores the caller service name.
	CallerService contextKey = "caller_service"
)
