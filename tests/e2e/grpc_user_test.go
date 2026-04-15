package e2e

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	userv1 "github.com/jrmarcello/gopherplate/gen/proto/appmax/user/v1"
)

// =============================================================================
// TC-E2E-G-06: Full CRUD cycle via gRPC (create -> get -> update -> delete)
// =============================================================================

func TestGRPC_UserFullCycle(t *testing.T) {
	require.NoError(t, CleanupUsers())
	clients, cleanup := setupGRPCServer(t)
	defer cleanup()

	ctx := context.Background()

	// 1. Create
	createResp, createErr := clients.user.CreateUser(ctx, &userv1.CreateUserRequest{
		Name:  "gRPC Cycle User",
		Email: "grpc-cycle@example.com",
	})
	require.NoError(t, createErr)
	assert.NotEmpty(t, createResp.GetId())
	assert.NotEmpty(t, createResp.GetCreatedAt())

	userID := createResp.GetId()

	// 2. Get
	getResp, getErr := clients.user.GetUser(ctx, &userv1.GetUserRequest{Id: userID})
	require.NoError(t, getErr)
	require.NotNil(t, getResp.GetUser())
	assert.Equal(t, userID, getResp.GetUser().GetId())
	assert.Equal(t, "gRPC Cycle User", getResp.GetUser().GetName())
	assert.Equal(t, "grpc-cycle@example.com", getResp.GetUser().GetEmail())
	assert.True(t, getResp.GetUser().GetActive())

	// 3. Update
	newName := "gRPC Updated User"
	updateResp, updateErr := clients.user.UpdateUser(ctx, &userv1.UpdateUserRequest{
		Id:   userID,
		Name: &newName,
	})
	require.NoError(t, updateErr)
	require.NotNil(t, updateResp.GetUser())
	assert.Equal(t, "gRPC Updated User", updateResp.GetUser().GetName())
	assert.Equal(t, "grpc-cycle@example.com", updateResp.GetUser().GetEmail())

	// 4. Verify update via Get
	getResp2, getErr2 := clients.user.GetUser(ctx, &userv1.GetUserRequest{Id: userID})
	require.NoError(t, getErr2)
	assert.Equal(t, "gRPC Updated User", getResp2.GetUser().GetName())

	// 5. Delete
	deleteResp, deleteErr := clients.user.DeleteUser(ctx, &userv1.DeleteUserRequest{Id: userID})
	require.NoError(t, deleteErr)
	assert.Equal(t, userID, deleteResp.GetId())

	// 6. Verify soft-delete: get should still work but user should be inactive
	getResp3, getErr3 := clients.user.GetUser(ctx, &userv1.GetUserRequest{Id: userID})
	if getErr3 != nil {
		// Some implementations return NotFound after soft-delete
		st, ok := status.FromError(getErr3)
		require.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())
	} else {
		assert.False(t, getResp3.GetUser().GetActive(), "User should be inactive after delete")
	}
}

// =============================================================================
// ERROR SCENARIOS
// =============================================================================

func TestGRPC_CreateUser_InvalidEmail(t *testing.T) {
	clients, cleanup := setupGRPCServer(t)
	defer cleanup()

	_, createErr := clients.user.CreateUser(context.Background(), &userv1.CreateUserRequest{
		Name:  "Bad Email User",
		Email: "not-an-email",
	})
	require.Error(t, createErr)

	st, ok := status.FromError(createErr)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestGRPC_GetUser_NotFound(t *testing.T) {
	clients, cleanup := setupGRPCServer(t)
	defer cleanup()

	// Valid UUID v7 format but non-existent
	_, getErr := clients.user.GetUser(context.Background(), &userv1.GetUserRequest{
		Id: "018e4a2c-6b4d-7000-9410-abcdef123456",
	})
	require.Error(t, getErr)

	st, ok := status.FromError(getErr)
	require.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
}

func TestGRPC_CreateUser_DuplicateEmail(t *testing.T) {
	require.NoError(t, CleanupUsers())
	clients, cleanup := setupGRPCServer(t)
	defer cleanup()

	ctx := context.Background()

	// First create succeeds
	_, firstErr := clients.user.CreateUser(ctx, &userv1.CreateUserRequest{
		Name:  "First User",
		Email: "duplicate-grpc@example.com",
	})
	require.NoError(t, firstErr)

	// Second create with same email should fail
	_, secondErr := clients.user.CreateUser(ctx, &userv1.CreateUserRequest{
		Name:  "Second User",
		Email: "duplicate-grpc@example.com",
	})
	require.Error(t, secondErr)

	st, ok := status.FromError(secondErr)
	require.True(t, ok)
	assert.Equal(t, codes.AlreadyExists, st.Code())
}

func TestGRPC_UpdateUser_NotFound(t *testing.T) {
	clients, cleanup := setupGRPCServer(t)
	defer cleanup()

	newName := "Ghost"
	_, updateErr := clients.user.UpdateUser(context.Background(), &userv1.UpdateUserRequest{
		Id:   "018e4a2c-6b4d-7000-9410-abcdef123456",
		Name: &newName,
	})
	require.Error(t, updateErr)

	st, ok := status.FromError(updateErr)
	require.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
}

func TestGRPC_DeleteUser_NotFound(t *testing.T) {
	clients, cleanup := setupGRPCServer(t)
	defer cleanup()

	_, deleteErr := clients.user.DeleteUser(context.Background(), &userv1.DeleteUserRequest{
		Id: "018e4a2c-6b4d-7000-9410-abcdef123456",
	})
	require.Error(t, deleteErr)

	st, ok := status.FromError(deleteErr)
	require.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
}

func TestGRPC_ListUsers(t *testing.T) {
	require.NoError(t, CleanupUsers())
	clients, cleanup := setupGRPCServer(t)
	defer cleanup()

	ctx := context.Background()

	// Create 3 users
	for i, email := range []string{"a@example.com", "b@example.com", "c@example.com"} {
		_, createErr := clients.user.CreateUser(ctx, &userv1.CreateUserRequest{
			Name:  "User " + string(rune('A'+i)),
			Email: email,
		})
		require.NoError(t, createErr)
	}

	// List with pagination
	listResp, listErr := clients.user.ListUsers(ctx, &userv1.ListUsersRequest{
		Page:  1,
		Limit: 2,
	})
	require.NoError(t, listErr)
	assert.Len(t, listResp.GetUsers(), 2)
	require.NotNil(t, listResp.GetPagination())
	assert.Equal(t, int32(3), listResp.GetPagination().GetTotal())
	assert.Equal(t, int32(1), listResp.GetPagination().GetPage())
	assert.Equal(t, int32(2), listResp.GetPagination().GetLimit())
}
