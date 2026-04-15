package interceptor

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// LoggingUnary returns a unary interceptor that logs method, duration, and status.
func LoggingUnary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		resp, handlerErr := handler(ctx, req)
		duration := time.Since(start)

		st, _ := status.FromError(handlerErr)
		level := slog.LevelInfo
		if handlerErr != nil {
			level = slog.LevelError
		}

		slog.Log(ctx, level, "gRPC request",
			"grpc.method", info.FullMethod,
			"grpc.code", st.Code().String(),
			"duration", duration.String(),
		)
		return resp, handlerErr
	}
}

// LoggingStream returns a stream interceptor that logs method, duration, and status.
func LoggingStream() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()
		handlerErr := handler(srv, ss)
		duration := time.Since(start)

		st, _ := status.FromError(handlerErr)
		level := slog.LevelInfo
		if handlerErr != nil {
			level = slog.LevelError
		}

		slog.Log(ss.Context(), level, "gRPC stream",
			"grpc.method", info.FullMethod,
			"grpc.code", st.Code().String(),
			"duration", duration.String(),
		)
		return handlerErr
	}
}
