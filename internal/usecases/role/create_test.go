package role

import (
	"context"
	"errors"
	"testing"

	roledomain "github.com/jrmarcello/gopherplate/internal/domain/role"
	"github.com/jrmarcello/gopherplate/internal/usecases/role/dto"
	ucshared "github.com/jrmarcello/gopherplate/internal/usecases/shared"
	"github.com/jrmarcello/gopherplate/pkg/apperror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func TestCreateUseCase_Execute_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	mockRepo.On("FindByName", mock.Anything, "admin").Return(nil, roledomain.ErrRoleNotFound)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*role.Role")).Return(nil)

	uc := NewCreateUseCase(mockRepo)
	input := dto.CreateInput{
		Name:        "admin",
		Description: "Administrator role",
	}

	// Act
	output, executeErr := uc.Execute(context.Background(), input)

	// Assert
	assert.NoError(t, executeErr)
	assert.NotNil(t, output)
	assert.NotEmpty(t, output.ID)
	assert.NotEmpty(t, output.CreatedAt)
	mockRepo.AssertExpectations(t)
}

func TestCreateUseCase_Execute_DuplicateName(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	existingRole := roledomain.NewRole("admin", "Existing admin role")
	mockRepo.On("FindByName", mock.Anything, "admin").Return(existingRole, nil)

	uc := NewCreateUseCase(mockRepo)
	input := dto.CreateInput{
		Name:        "admin",
		Description: "Administrator role",
	}

	// Act
	output, executeErr := uc.Execute(context.Background(), input)

	// Assert
	assert.Error(t, executeErr)
	assert.Nil(t, output)

	var appErr *apperror.AppError
	assert.True(t, errors.As(executeErr, &appErr))
	assert.Equal(t, apperror.CodeConflict, appErr.Code)

	mockRepo.AssertNotCalled(t, "Create")
}

func TestCreateUseCase_Execute_FindByNameError(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	mockRepo.On("FindByName", mock.Anything, "admin").
		Return(nil, errors.New("database connection lost"))

	uc := NewCreateUseCase(mockRepo)
	input := dto.CreateInput{
		Name:        "admin",
		Description: "Administrator role",
	}

	// Act
	output, executeErr := uc.Execute(context.Background(), input)

	// Assert
	assert.Error(t, executeErr)
	assert.Nil(t, output)

	var appErr *apperror.AppError
	assert.True(t, errors.As(executeErr, &appErr))
	assert.Equal(t, apperror.CodeInternalError, appErr.Code)

	mockRepo.AssertNotCalled(t, "Create")
}

func TestCreateUseCase_Execute_RepositoryError(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	mockRepo.On("FindByName", mock.Anything, "admin").Return(nil, roledomain.ErrRoleNotFound)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*role.Role")).
		Return(errors.New("database connection failed"))

	uc := NewCreateUseCase(mockRepo)
	input := dto.CreateInput{
		Name:        "admin",
		Description: "Administrator role",
	}

	// Act
	output, executeErr := uc.Execute(context.Background(), input)

	// Assert
	assert.Error(t, executeErr)
	assert.Nil(t, output)

	var appErr *apperror.AppError
	assert.True(t, errors.As(executeErr, &appErr))
	assert.Equal(t, apperror.CodeInternalError, appErr.Code)

	mockRepo.AssertExpectations(t)
}

// TC-UC-37: Create with duplicate name records semantic span attribute
// `app.result=duplicate_role_name` and leaves span status Unset.
func TestCreateUseCase_Execute_TC_UC_37_DuplicateName_SpanAttribute(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	existingRole := roledomain.NewRole("admin", "Existing admin role")
	mockRepo.On("FindByName", mock.Anything, "admin").Return(existingRole, nil)

	uc := NewCreateUseCase(mockRepo)
	input := dto.CreateInput{Name: "admin", Description: "Administrator role"}

	ctx, exp, endAndFlush := newRecordingContext(t)

	// Act
	output, executeErr := uc.Execute(ctx, input)
	endAndFlush()

	// Assert: error path produces AppError(Conflict)
	require.Error(t, executeErr)
	require.Nil(t, output)
	var appErr *apperror.AppError
	require.True(t, errors.As(executeErr, &appErr))
	assert.Equal(t, apperror.CodeConflict, appErr.Code)

	// Assert: semantic span attribute applied; status stays Unset.
	finished := exp.GetSpans()
	require.Len(t, finished, 1)
	stub := finished[0]
	assert.Equal(t, codes.Unset, stub.Status.Code,
		"duplicate-name is an expected outcome — span status must remain Unset")
	assert.Contains(t, stub.Attributes,
		attribute.String(ucshared.AttrKeyAppResult, "duplicate_role_name"),
		"expected attribute %s=duplicate_role_name", ucshared.AttrKeyAppResult)
}
