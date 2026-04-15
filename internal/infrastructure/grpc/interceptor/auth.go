package interceptor

import (
	"context"
	"crypto/subtle"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const dummyServiceKey = "unknown-service-constant-time-dummy"

// AuthConfig holds service key authentication config for gRPC.
type AuthConfig struct {
	Enabled bool
	Keys    map[string]string // service-name -> key
}

// AuthUnary returns a unary interceptor that validates service key auth from metadata.
func AuthUnary(cfg AuthConfig) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if authErr := validateAuth(ctx, cfg, info.FullMethod); authErr != nil {
			return nil, authErr
		}
		return handler(ctx, req)
	}
}

// AuthStream returns a stream interceptor that validates service key auth from metadata.
func AuthStream(cfg AuthConfig) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if authErr := validateAuth(ss.Context(), cfg, info.FullMethod); authErr != nil {
			return authErr
		}
		return handler(srv, ss)
	}
}

func validateAuth(ctx context.Context, cfg AuthConfig, fullMethod string) error {
	// Skip auth for health check RPCs.
	if strings.HasPrefix(fullMethod, "/grpc.health.v1.Health/") {
		return nil
	}

	// Not enabled — development mode, allow all.
	if !cfg.Enabled {
		return nil
	}

	// Fail-closed: enabled but no keys configured.
	if len(cfg.Keys) == 0 {
		return status.Error(codes.Unavailable, "service authentication not configured")
	}

	// Extract metadata.
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "unauthorized")
	}

	serviceNames := md.Get("x-service-name")
	serviceKeys := md.Get("x-service-key")
	if len(serviceNames) == 0 || len(serviceKeys) == 0 {
		return status.Error(codes.Unauthenticated, "unauthorized")
	}

	serviceName := serviceNames[0]
	serviceKey := serviceKeys[0]

	// Always run constant-time compare, even for unknown services (CWE-203 mitigation).
	expectedKey, exists := cfg.Keys[serviceName]
	if !exists {
		expectedKey = dummyServiceKey
	}
	keyMatch := subtle.ConstantTimeCompare([]byte(expectedKey), []byte(serviceKey)) == 1
	if !exists || !keyMatch {
		return status.Error(codes.Unauthenticated, "unauthorized")
	}

	return nil
}
