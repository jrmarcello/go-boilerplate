package grpc

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	rolev1 "github.com/jrmarcello/gopherplate/gen/proto/appmax/role/v1"
	userv1 "github.com/jrmarcello/gopherplate/gen/proto/appmax/user/v1"
	grpchandler "github.com/jrmarcello/gopherplate/internal/infrastructure/grpc/handler"
	"github.com/jrmarcello/gopherplate/internal/infrastructure/grpc/interceptor"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
)

// Config holds gRPC server configuration.
type Config struct {
	ReflectionEnabled bool
	AuthConfig        interceptor.AuthConfig
}

// NewServer creates a gRPC server with interceptors, health check, and service registrations.
func NewServer(cfg Config, userHandler *grpchandler.UserHandler, roleHandler *grpchandler.RoleHandler) *grpc.Server {
	srv := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			interceptor.RecoveryUnary(),
			interceptor.LoggingUnary(),
			interceptor.AuthUnary(cfg.AuthConfig),
		),
		grpc.ChainStreamInterceptor(
			interceptor.RecoveryStream(),
			interceptor.LoggingStream(),
			interceptor.AuthStream(cfg.AuthConfig),
		),
	)

	// Register domain services
	userv1.RegisterUserServiceServer(srv, userHandler)
	rolev1.RegisterRoleServiceServer(srv, roleHandler)

	// Health check (grpc.health.v1.Health protocol)
	healthSrv := health.NewServer()
	healthpb.RegisterHealthServer(srv, healthSrv)
	healthSrv.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthSrv.SetServingStatus("appmax.user.v1.UserService", healthpb.HealthCheckResponse_SERVING)
	healthSrv.SetServingStatus("appmax.role.v1.RoleService", healthpb.HealthCheckResponse_SERVING)

	// Reflection (for grpcurl/grpcui — configurable per environment)
	if cfg.ReflectionEnabled {
		reflection.Register(srv)
	}

	return srv
}
