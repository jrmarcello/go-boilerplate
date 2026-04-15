package handler

import (
	"context"

	commonv1 "github.com/jrmarcello/gopherplate/gen/proto/appmax/common/v1"
	userv1 "github.com/jrmarcello/gopherplate/gen/proto/appmax/user/v1"
	infratelemetry "github.com/jrmarcello/gopherplate/internal/infrastructure/telemetry"
	useruc "github.com/jrmarcello/gopherplate/internal/usecases/user"
	"github.com/jrmarcello/gopherplate/internal/usecases/user/dto"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// UserHandler implements userv1.UserServiceServer by delegating to use cases.
// Mirrors the HTTP handler pattern: convert proto -> DTO, call use case, convert output -> proto.
type UserHandler struct {
	userv1.UnimplementedUserServiceServer
	createUC *useruc.CreateUseCase
	getUC    *useruc.GetUseCase
	listUC   *useruc.ListUseCase
	updateUC *useruc.UpdateUseCase
	deleteUC *useruc.DeleteUseCase
	metrics  *infratelemetry.Metrics
}

// NewUserHandler creates a new gRPC UserHandler with all use cases.
func NewUserHandler(
	createUC *useruc.CreateUseCase,
	getUC *useruc.GetUseCase,
	listUC *useruc.ListUseCase,
	updateUC *useruc.UpdateUseCase,
	deleteUC *useruc.DeleteUseCase,
	metrics *infratelemetry.Metrics,
) *UserHandler {
	return &UserHandler{
		createUC: createUC,
		getUC:    getUC,
		listUC:   listUC,
		updateUC: updateUC,
		deleteUC: deleteUC,
		metrics:  metrics,
	}
}

// CreateUser creates a new user.
func (h *UserHandler) CreateUser(ctx context.Context, req *userv1.CreateUserRequest) (*userv1.CreateUserResponse, error) {
	ctx, span := otel.Tracer("grpc-handler").Start(ctx, "UserHandler.CreateUser")
	defer span.End()

	input := dto.CreateInput{
		Name:  req.GetName(),
		Email: req.GetEmail(),
	}

	res, execErr := h.createUC.Execute(ctx, input)
	if execErr != nil {
		return nil, toGRPCStatus(execErr)
	}

	span.SetAttributes(attribute.String("user.id", res.ID))

	if h.metrics != nil {
		h.metrics.RecordCreate(ctx)
	}

	return &userv1.CreateUserResponse{
		Id:        res.ID,
		CreatedAt: res.CreatedAt,
	}, nil
}

// GetUser retrieves a user by ID.
func (h *UserHandler) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	ctx, span := otel.Tracer("grpc-handler").Start(ctx, "UserHandler.GetUser")
	defer span.End()

	span.SetAttributes(attribute.String("user.id", req.GetId()))

	res, execErr := h.getUC.Execute(ctx, dto.GetInput{ID: req.GetId()})
	if execErr != nil {
		return nil, toGRPCStatus(execErr)
	}

	return &userv1.GetUserResponse{
		User: toProtoUser(res),
	}, nil
}

// ListUsers retrieves a paginated list of users.
func (h *UserHandler) ListUsers(ctx context.Context, req *userv1.ListUsersRequest) (*userv1.ListUsersResponse, error) {
	ctx, span := otel.Tracer("grpc-handler").Start(ctx, "UserHandler.ListUsers")
	defer span.End()

	input := dto.ListInput{
		Page:       int(req.GetPage()),
		Limit:      int(req.GetLimit()),
		Name:       req.GetName(),
		Email:      req.GetEmail(),
		ActiveOnly: req.GetActiveOnly(),
	}

	span.SetAttributes(
		attribute.Int("filter.page", input.Page),
		attribute.Int("filter.limit", input.Limit),
	)

	res, execErr := h.listUC.Execute(ctx, input)
	if execErr != nil {
		return nil, toGRPCStatus(execErr)
	}

	span.SetAttributes(attribute.Int("result.total", res.Pagination.Total))

	protoUsers := make([]*userv1.User, len(res.Data))
	for i, u := range res.Data {
		protoUsers[i] = toProtoUser(&u)
	}

	return &userv1.ListUsersResponse{
		Users: protoUsers,
		Pagination: &commonv1.Pagination{
			Page:       toInt32(res.Pagination.Page),
			Limit:      toInt32(res.Pagination.Limit),
			Total:      toInt32(res.Pagination.Total),
			TotalPages: toInt32(res.Pagination.TotalPages),
		},
	}, nil
}

// UpdateUser updates an existing user.
func (h *UserHandler) UpdateUser(ctx context.Context, req *userv1.UpdateUserRequest) (*userv1.UpdateUserResponse, error) {
	ctx, span := otel.Tracer("grpc-handler").Start(ctx, "UserHandler.UpdateUser")
	defer span.End()

	span.SetAttributes(attribute.String("user.id", req.GetId()))

	input := dto.UpdateInput{
		ID:    req.GetId(),
		Name:  req.Name,
		Email: req.Email,
	}

	res, execErr := h.updateUC.Execute(ctx, input)
	if execErr != nil {
		return nil, toGRPCStatus(execErr)
	}

	if h.metrics != nil {
		h.metrics.RecordUpdate(ctx)
	}

	return &userv1.UpdateUserResponse{
		User: &userv1.User{
			Id:        res.ID,
			Name:      res.Name,
			Email:     res.Email,
			Active:    res.Active,
			UpdatedAt: res.UpdatedAt,
		},
	}, nil
}

// DeleteUser soft-deletes a user by ID.
func (h *UserHandler) DeleteUser(ctx context.Context, req *userv1.DeleteUserRequest) (*userv1.DeleteUserResponse, error) {
	ctx, span := otel.Tracer("grpc-handler").Start(ctx, "UserHandler.DeleteUser")
	defer span.End()

	span.SetAttributes(attribute.String("user.id", req.GetId()))

	res, execErr := h.deleteUC.Execute(ctx, dto.DeleteInput{ID: req.GetId()})
	if execErr != nil {
		return nil, toGRPCStatus(execErr)
	}

	if h.metrics != nil {
		h.metrics.RecordDelete(ctx)
	}

	return &userv1.DeleteUserResponse{
		Id: res.ID,
	}, nil
}

// toProtoUser converts a dto.GetOutput to a userv1.User proto message.
func toProtoUser(u *dto.GetOutput) *userv1.User {
	return &userv1.User{
		Id:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Active:    u.Active,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
