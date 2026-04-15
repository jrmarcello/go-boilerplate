package handler

import (
	"context"

	commonv1 "github.com/jrmarcello/gopherplate/gen/proto/appmax/common/v1"
	rolev1 "github.com/jrmarcello/gopherplate/gen/proto/appmax/role/v1"
	roleuc "github.com/jrmarcello/gopherplate/internal/usecases/role"
	"github.com/jrmarcello/gopherplate/internal/usecases/role/dto"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// RoleHandler implements rolev1.RoleServiceServer using role use cases.
type RoleHandler struct {
	rolev1.UnimplementedRoleServiceServer
	createUC *roleuc.CreateUseCase
	listUC   *roleuc.ListUseCase
	deleteUC *roleuc.DeleteUseCase
}

// NewRoleHandler creates a new RoleHandler with the required use cases.
func NewRoleHandler(
	createUC *roleuc.CreateUseCase,
	listUC *roleuc.ListUseCase,
	deleteUC *roleuc.DeleteUseCase,
) *RoleHandler {
	return &RoleHandler{
		createUC: createUC,
		listUC:   listUC,
		deleteUC: deleteUC,
	}
}

// CreateRole handles the gRPC CreateRole request.
func (h *RoleHandler) CreateRole(ctx context.Context, req *rolev1.CreateRoleRequest) (*rolev1.CreateRoleResponse, error) {
	ctx, span := otel.Tracer("grpc-handler").Start(ctx, "RoleHandler.CreateRole")
	defer span.End()

	input := dto.CreateInput{
		Name:        req.GetName(),
		Description: "",
	}

	output, execErr := h.createUC.Execute(ctx, input)
	if execErr != nil {
		return nil, toGRPCStatus(execErr)
	}

	span.SetAttributes(attribute.String("role.id", output.ID))

	return &rolev1.CreateRoleResponse{
		Id:        output.ID,
		CreatedAt: output.CreatedAt,
	}, nil
}

// ListRoles handles the gRPC ListRoles request.
func (h *RoleHandler) ListRoles(ctx context.Context, req *rolev1.ListRolesRequest) (*rolev1.ListRolesResponse, error) {
	ctx, span := otel.Tracer("grpc-handler").Start(ctx, "RoleHandler.ListRoles")
	defer span.End()

	input := dto.ListInput{
		Page:  int(req.GetPage()),
		Limit: int(req.GetLimit()),
	}

	output, execErr := h.listUC.Execute(ctx, input)
	if execErr != nil {
		return nil, toGRPCStatus(execErr)
	}

	roles := make([]*rolev1.Role, 0, len(output.Data))
	for _, r := range output.Data {
		roles = append(roles, &rolev1.Role{
			Id:        r.ID,
			Name:      r.Name,
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
		})
	}

	span.SetAttributes(attribute.Int("role.count", len(roles)))

	return &rolev1.ListRolesResponse{
		Roles: roles,
		Pagination: &commonv1.Pagination{
			Page:       toInt32(output.Pagination.Page),
			Limit:      toInt32(output.Pagination.Limit),
			Total:      toInt32(output.Pagination.Total),
			TotalPages: toInt32(output.Pagination.TotalPages),
		},
	}, nil
}

// DeleteRole handles the gRPC DeleteRole request.
func (h *RoleHandler) DeleteRole(ctx context.Context, req *rolev1.DeleteRoleRequest) (*rolev1.DeleteRoleResponse, error) {
	ctx, span := otel.Tracer("grpc-handler").Start(ctx, "RoleHandler.DeleteRole")
	defer span.End()

	input := dto.DeleteInput{
		ID: req.GetId(),
	}

	output, execErr := h.deleteUC.Execute(ctx, input)
	if execErr != nil {
		return nil, toGRPCStatus(execErr)
	}

	span.SetAttributes(attribute.String("role.id", output.ID))

	return &rolev1.DeleteRoleResponse{
		Id: output.ID,
	}, nil
}
