package e2e

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	rolev1 "github.com/jrmarcello/gopherplate/gen/proto/appmax/role/v1"
)

// =============================================================================
// SUCCESS SCENARIOS
// =============================================================================

func TestGRPC_RoleFullCycle(t *testing.T) {
	require.NoError(t, cleanupRoles())
	clients, cleanup := setupGRPCServer(t)
	defer cleanup()

	ctx := context.Background()

	// 1. Create role
	createResp, createErr := clients.role.CreateRole(ctx, &rolev1.CreateRoleRequest{
		Name: "grpc-admin",
	})
	require.NoError(t, createErr)
	assert.NotEmpty(t, createResp.GetId())
	assert.NotEmpty(t, createResp.GetCreatedAt())

	roleID := createResp.GetId()

	// 2. List roles - verify it appears
	listResp, listErr := clients.role.ListRoles(ctx, &rolev1.ListRolesRequest{
		Page:  1,
		Limit: 10,
	})
	require.NoError(t, listErr)
	assert.Len(t, listResp.GetRoles(), 1)

	firstRole := listResp.GetRoles()[0]
	assert.Equal(t, roleID, firstRole.GetId())
	assert.Equal(t, "grpc-admin", firstRole.GetName())

	// 3. Delete role
	deleteResp, deleteErr := clients.role.DeleteRole(ctx, &rolev1.DeleteRoleRequest{
		Id: roleID,
	})
	require.NoError(t, deleteErr)
	assert.Equal(t, roleID, deleteResp.GetId())

	// 4. List again - verify gone
	listResp2, listErr2 := clients.role.ListRoles(ctx, &rolev1.ListRolesRequest{
		Page:  1,
		Limit: 10,
	})
	require.NoError(t, listErr2)
	assert.Empty(t, listResp2.GetRoles())
}

// =============================================================================
// ERROR SCENARIOS
// =============================================================================

func TestGRPC_DeleteRole_NotFound(t *testing.T) {
	clients, cleanup := setupGRPCServer(t)
	defer cleanup()

	_, deleteErr := clients.role.DeleteRole(context.Background(), &rolev1.DeleteRoleRequest{
		Id: "018e4a2c-6b4d-7000-9410-abcdef123456",
	})
	require.Error(t, deleteErr)

	st, ok := status.FromError(deleteErr)
	require.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
}

func TestGRPC_ListRoles_Empty(t *testing.T) {
	require.NoError(t, cleanupRoles())
	clients, cleanup := setupGRPCServer(t)
	defer cleanup()

	listResp, listErr := clients.role.ListRoles(context.Background(), &rolev1.ListRolesRequest{
		Page:  1,
		Limit: 10,
	})
	require.NoError(t, listErr)
	assert.Empty(t, listResp.GetRoles())
}
