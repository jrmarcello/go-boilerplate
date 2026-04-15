package interceptor

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// --- mock types for testing ---

// mockUnaryHandler simulates a gRPC unary handler for interceptor tests.
type mockUnaryHandler struct {
	called bool
	resp   any
	err    error
}

func (m *mockUnaryHandler) handle(_ context.Context, _ any) (any, error) {
	m.called = true
	return m.resp, m.err
}

// mockServerStream is a minimal grpc.ServerStream for testing stream interceptors.
type mockServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (m *mockServerStream) Context() context.Context {
	return m.ctx
}

// mockStreamHandler simulates a gRPC stream handler for interceptor tests.
type mockStreamHandler struct {
	called bool
	err    error
}

func (m *mockStreamHandler) handle(_ any, _ grpc.ServerStream) error {
	m.called = true
	return m.err
}

// --- TC-I-01: Request without service-key metadata -> codes.Unauthenticated ---

func TestAuthUnary_MissingMetadata(t *testing.T) {
	cfg := AuthConfig{
		Enabled: true,
		Keys:    map[string]string{"test-svc": "secret-key"},
	}
	interceptor := AuthUnary(cfg)

	handler := &mockUnaryHandler{resp: "ok"}
	info := &grpc.UnaryServerInfo{FullMethod: "/myapp.v1.UserService/GetUser"}

	// No metadata in context
	ctx := context.Background()
	resp, callErr := interceptor(ctx, nil, info, handler.handle)

	assert.Nil(t, resp)
	assert.False(t, handler.called)
	st, ok := status.FromError(callErr)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
	assert.Equal(t, "unauthorized", st.Message())
}

func TestAuthUnary_EmptyHeaders(t *testing.T) {
	cfg := AuthConfig{
		Enabled: true,
		Keys:    map[string]string{"test-svc": "secret-key"},
	}
	interceptor := AuthUnary(cfg)

	handler := &mockUnaryHandler{resp: "ok"}
	info := &grpc.UnaryServerInfo{FullMethod: "/myapp.v1.UserService/GetUser"}

	// Metadata present but empty service name/key
	md := metadata.New(map[string]string{})
	ctx := metadata.NewIncomingContext(context.Background(), md)
	resp, callErr := interceptor(ctx, nil, info, handler.handle)

	assert.Nil(t, resp)
	assert.False(t, handler.called)
	st, ok := status.FromError(callErr)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

// --- TC-I-02: Request with invalid service key -> codes.Unauthenticated ---

func TestAuthUnary_InvalidKey(t *testing.T) {
	cfg := AuthConfig{
		Enabled: true,
		Keys:    map[string]string{"test-svc": "correct-key"},
	}
	interceptor := AuthUnary(cfg)

	handler := &mockUnaryHandler{resp: "ok"}
	info := &grpc.UnaryServerInfo{FullMethod: "/myapp.v1.UserService/GetUser"}

	md := metadata.Pairs("x-service-name", "test-svc", "x-service-key", "wrong-key")
	ctx := metadata.NewIncomingContext(context.Background(), md)
	resp, callErr := interceptor(ctx, nil, info, handler.handle)

	assert.Nil(t, resp)
	assert.False(t, handler.called)
	st, ok := status.FromError(callErr)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
	assert.Equal(t, "unauthorized", st.Message())
}

// --- TC-I-03: Request with unknown service name -> codes.Unauthenticated ---

func TestAuthUnary_UnknownService(t *testing.T) {
	cfg := AuthConfig{
		Enabled: true,
		Keys:    map[string]string{"known-svc": "the-key"},
	}
	interceptor := AuthUnary(cfg)

	handler := &mockUnaryHandler{resp: "ok"}
	info := &grpc.UnaryServerInfo{FullMethod: "/myapp.v1.UserService/GetUser"}

	md := metadata.Pairs("x-service-name", "unknown-svc", "x-service-key", "some-key")
	ctx := metadata.NewIncomingContext(context.Background(), md)
	resp, callErr := interceptor(ctx, nil, info, handler.handle)

	assert.Nil(t, resp)
	assert.False(t, handler.called)
	st, ok := status.FromError(callErr)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

// --- TC-I-04: Request with valid service-key metadata -> OK ---

func TestAuthUnary_ValidKey(t *testing.T) {
	cfg := AuthConfig{
		Enabled: true,
		Keys:    map[string]string{"test-svc": "correct-key"},
	}
	interceptor := AuthUnary(cfg)

	handler := &mockUnaryHandler{resp: "ok"}
	info := &grpc.UnaryServerInfo{FullMethod: "/myapp.v1.UserService/GetUser"}

	md := metadata.Pairs("x-service-name", "test-svc", "x-service-key", "correct-key")
	ctx := metadata.NewIncomingContext(context.Background(), md)
	resp, callErr := interceptor(ctx, nil, info, handler.handle)

	assert.Equal(t, "ok", resp)
	assert.NoError(t, callErr)
	assert.True(t, handler.called)
}

// --- TC-I-05: Health check RPC bypasses auth -> OK without metadata ---

func TestAuthUnary_HealthCheckBypass(t *testing.T) {
	cfg := AuthConfig{
		Enabled: true,
		Keys:    map[string]string{"test-svc": "secret-key"},
	}
	interceptor := AuthUnary(cfg)

	handler := &mockUnaryHandler{resp: "healthy"}
	info := &grpc.UnaryServerInfo{FullMethod: "/grpc.health.v1.Health/Check"}

	// No metadata at all
	ctx := context.Background()
	resp, callErr := interceptor(ctx, nil, info, handler.handle)

	assert.Equal(t, "healthy", resp)
	assert.NoError(t, callErr)
	assert.True(t, handler.called)
}

func TestAuthUnary_HealthWatchBypass(t *testing.T) {
	cfg := AuthConfig{
		Enabled: true,
		Keys:    map[string]string{"test-svc": "secret-key"},
	}
	interceptor := AuthUnary(cfg)

	handler := &mockUnaryHandler{resp: "watching"}
	info := &grpc.UnaryServerInfo{FullMethod: "/grpc.health.v1.Health/Watch"}

	ctx := context.Background()
	resp, callErr := interceptor(ctx, nil, info, handler.handle)

	assert.Equal(t, "watching", resp)
	assert.NoError(t, callErr)
	assert.True(t, handler.called)
}

// --- TC-I-06: Handler panics — recovery interceptor -> codes.Internal ---

func TestRecoveryUnary_PanicRecovery(t *testing.T) {
	interceptor := RecoveryUnary()

	panicHandler := func(_ context.Context, _ any) (any, error) {
		panic("test panic")
	}
	info := &grpc.UnaryServerInfo{FullMethod: "/myapp.v1.UserService/GetUser"}

	resp, callErr := interceptor(context.Background(), nil, info, panicHandler)

	assert.Nil(t, resp)
	st, ok := status.FromError(callErr)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
	assert.Equal(t, "internal server error", st.Message())
}

func TestRecoveryUnary_NoPanic(t *testing.T) {
	interceptor := RecoveryUnary()

	handler := &mockUnaryHandler{resp: "ok"}
	info := &grpc.UnaryServerInfo{FullMethod: "/myapp.v1.UserService/GetUser"}

	resp, callErr := interceptor(context.Background(), nil, info, handler.handle)

	assert.Equal(t, "ok", resp)
	assert.NoError(t, callErr)
	assert.True(t, handler.called)
}

func TestRecoveryStream_PanicRecovery(t *testing.T) {
	interceptor := RecoveryStream()

	panicHandler := func(_ any, _ grpc.ServerStream) error {
		panic("stream panic")
	}
	info := &grpc.StreamServerInfo{FullMethod: "/myapp.v1.UserService/ListUsers"}
	stream := &mockServerStream{ctx: context.Background()}

	callErr := interceptor(nil, stream, info, panicHandler)

	st, ok := status.FromError(callErr)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
	assert.Equal(t, "internal server error", st.Message())
}

// --- TC-I-07: Logging interceptor records method + duration ---

func TestLoggingUnary_LogsMethodAndDuration(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	interceptor := LoggingUnary()

	handler := &mockUnaryHandler{resp: "ok"}
	info := &grpc.UnaryServerInfo{FullMethod: "/myapp.v1.UserService/GetUser"}

	resp, callErr := interceptor(context.Background(), nil, info, handler.handle)

	assert.Equal(t, "ok", resp)
	assert.NoError(t, callErr)

	logOutput := buf.String()
	assert.True(t, strings.Contains(logOutput, "gRPC request"), "should contain 'gRPC request'")
	assert.True(t, strings.Contains(logOutput, "/myapp.v1.UserService/GetUser"), "should contain method name")
	assert.True(t, strings.Contains(logOutput, "duration"), "should contain duration")
	assert.True(t, strings.Contains(logOutput, "OK"), "should contain OK code")
}

func TestLoggingUnary_LogsErrorLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	interceptor := LoggingUnary()

	handler := &mockUnaryHandler{err: status.Error(codes.NotFound, "not found")}
	info := &grpc.UnaryServerInfo{FullMethod: "/myapp.v1.UserService/GetUser"}

	_, callErr := interceptor(context.Background(), nil, info, handler.handle)

	assert.Error(t, callErr)

	logOutput := buf.String()
	assert.True(t, strings.Contains(logOutput, "gRPC request"), "should contain 'gRPC request'")
	assert.True(t, strings.Contains(logOutput, "ERROR"), "should log at error level for errors")
	assert.True(t, strings.Contains(logOutput, "NotFound"), "should contain NotFound code")
}

func TestLoggingStream_LogsMethodAndDuration(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	interceptor := LoggingStream()

	handler := &mockStreamHandler{}
	info := &grpc.StreamServerInfo{FullMethod: "/myapp.v1.UserService/ListUsers"}
	stream := &mockServerStream{ctx: context.Background()}

	callErr := interceptor(nil, stream, info, handler.handle)

	assert.NoError(t, callErr)
	assert.True(t, handler.called)

	logOutput := buf.String()
	assert.True(t, strings.Contains(logOutput, "gRPC stream"), "should contain 'gRPC stream'")
	assert.True(t, strings.Contains(logOutput, "/myapp.v1.UserService/ListUsers"), "should contain method name")
	assert.True(t, strings.Contains(logOutput, "duration"), "should contain duration")
}

// --- TC-I-08: Auth disabled — all requests pass ---

func TestAuthUnary_Disabled(t *testing.T) {
	cfg := AuthConfig{
		Enabled: false,
		Keys:    map[string]string{"test-svc": "key"},
	}
	interceptor := AuthUnary(cfg)

	handler := &mockUnaryHandler{resp: "ok"}
	info := &grpc.UnaryServerInfo{FullMethod: "/myapp.v1.UserService/GetUser"}

	// No metadata at all — should still pass because auth is disabled
	ctx := context.Background()
	resp, callErr := interceptor(ctx, nil, info, handler.handle)

	assert.Equal(t, "ok", resp)
	assert.NoError(t, callErr)
	assert.True(t, handler.called)
}

func TestAuthStream_Disabled(t *testing.T) {
	cfg := AuthConfig{
		Enabled: false,
		Keys:    map[string]string{"test-svc": "key"},
	}
	interceptor := AuthStream(cfg)

	handler := &mockStreamHandler{}
	info := &grpc.StreamServerInfo{FullMethod: "/myapp.v1.UserService/ListUsers"}
	stream := &mockServerStream{ctx: context.Background()}

	callErr := interceptor(nil, stream, info, handler.handle)

	assert.NoError(t, callErr)
	assert.True(t, handler.called)
}

// --- Auth stream tests (mirror unary coverage) ---

func TestAuthStream_MissingMetadata(t *testing.T) {
	cfg := AuthConfig{
		Enabled: true,
		Keys:    map[string]string{"test-svc": "secret-key"},
	}
	interceptor := AuthStream(cfg)

	handler := &mockStreamHandler{}
	info := &grpc.StreamServerInfo{FullMethod: "/myapp.v1.UserService/ListUsers"}
	stream := &mockServerStream{ctx: context.Background()}

	callErr := interceptor(nil, stream, info, handler.handle)

	assert.False(t, handler.called)
	st, ok := status.FromError(callErr)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

func TestAuthStream_ValidKey(t *testing.T) {
	cfg := AuthConfig{
		Enabled: true,
		Keys:    map[string]string{"test-svc": "correct-key"},
	}
	interceptor := AuthStream(cfg)

	handler := &mockStreamHandler{}
	info := &grpc.StreamServerInfo{FullMethod: "/myapp.v1.UserService/ListUsers"}
	md := metadata.Pairs("x-service-name", "test-svc", "x-service-key", "correct-key")
	ctx := metadata.NewIncomingContext(context.Background(), md)
	stream := &mockServerStream{ctx: ctx}

	callErr := interceptor(nil, stream, info, handler.handle)

	assert.NoError(t, callErr)
	assert.True(t, handler.called)
}

// --- Auth fail-closed: enabled but no keys ---

func TestAuthUnary_EnabledNoKeys(t *testing.T) {
	cfg := AuthConfig{
		Enabled: true,
		Keys:    map[string]string{},
	}
	interceptor := AuthUnary(cfg)

	handler := &mockUnaryHandler{resp: "ok"}
	info := &grpc.UnaryServerInfo{FullMethod: "/myapp.v1.UserService/GetUser"}

	md := metadata.Pairs("x-service-name", "test-svc", "x-service-key", "any-key")
	ctx := metadata.NewIncomingContext(context.Background(), md)
	resp, callErr := interceptor(ctx, nil, info, handler.handle)

	assert.Nil(t, resp)
	assert.False(t, handler.called)
	st, ok := status.FromError(callErr)
	require.True(t, ok)
	assert.Equal(t, codes.Unavailable, st.Code())
	assert.Equal(t, "service authentication not configured", st.Message())
}
