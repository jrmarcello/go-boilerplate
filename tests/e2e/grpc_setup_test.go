package e2e

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/test/bufconn"

	rolev1 "github.com/jrmarcello/gopherplate/gen/proto/appmax/role/v1"
	userv1 "github.com/jrmarcello/gopherplate/gen/proto/appmax/user/v1"
	"github.com/jrmarcello/gopherplate/internal/bootstrap"
	appgrpc "github.com/jrmarcello/gopherplate/internal/infrastructure/grpc"
	"github.com/jrmarcello/gopherplate/internal/infrastructure/grpc/interceptor"
)

const bufSize = 1024 * 1024

// grpcClients bundles all gRPC service clients returned by setupGRPCServer.
type grpcClients struct {
	user   userv1.UserServiceClient
	role   rolev1.RoleServiceClient
	health healthpb.HealthClient
}

// setupGRPCServer creates an in-process gRPC server using bufconn for testing.
// Returns a grpcClients struct and a cleanup function that must be deferred.
func setupGRPCServer(t *testing.T) (grpcClients, func()) {
	t.Helper()

	c := bootstrap.NewForTest(t, GetTestDB(), GetTestCache())

	srv := appgrpc.NewServer(appgrpc.Config{
		ReflectionEnabled: true,
		AuthConfig: interceptor.AuthConfig{
			Enabled: false,
		},
	}, c.GRPCHandlers.User, c.GRPCHandlers.Role)

	lis := bufconn.Listen(bufSize)
	go func() {
		if serveErr := srv.Serve(lis); serveErr != nil {
			t.Logf("gRPC server error: %v", serveErr)
		}
	}()

	conn, dialErr := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if dialErr != nil {
		t.Fatalf("Failed to dial bufnet: %v", dialErr)
	}

	cleanup := func() {
		conn.Close()
		srv.GracefulStop()
	}

	clients := grpcClients{
		user:   userv1.NewUserServiceClient(conn),
		role:   rolev1.NewRoleServiceClient(conn),
		health: healthpb.NewHealthClient(conn),
	}

	return clients, cleanup
}

// =============================================================================
// TC-E2E-G-01: gRPC server starts on configured port (via bufconn)
// =============================================================================

func TestGRPC_ServerStarts(t *testing.T) {
	clients, cleanup := setupGRPCServer(t)
	defer cleanup()

	// Verify the server is reachable by making a health check call.
	resp, checkErr := clients.health.Check(context.Background(), &healthpb.HealthCheckRequest{})
	require.NoError(t, checkErr)
	assert.Equal(t, healthpb.HealthCheckResponse_SERVING, resp.GetStatus())
}

// =============================================================================
// TC-E2E-G-02: Health check returns SERVING
// =============================================================================

func TestGRPC_HealthCheck_Serving(t *testing.T) {
	clients, cleanup := setupGRPCServer(t)
	defer cleanup()

	tests := []struct {
		name    string
		service string
	}{
		{name: "overall", service: ""},
		{name: "user service", service: "appmax.user.v1.UserService"},
		{name: "role service", service: "appmax.role.v1.RoleService"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp, checkErr := clients.health.Check(context.Background(), &healthpb.HealthCheckRequest{
				Service: tc.service,
			})
			require.NoError(t, checkErr)
			assert.Equal(t, healthpb.HealthCheckResponse_SERVING, resp.GetStatus())
		})
	}
}
