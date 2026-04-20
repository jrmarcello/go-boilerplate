package role

import (
	"context"
	"errors"
	"testing"

	roledomain "github.com/jrmarcello/gopherplate/internal/domain/role"
	"github.com/jrmarcello/gopherplate/internal/domain/user/vo"
	"github.com/jrmarcello/gopherplate/internal/usecases/role/dto"
	ucshared "github.com/jrmarcello/gopherplate/internal/usecases/shared"
	"github.com/jrmarcello/gopherplate/pkg/apperror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func TestDeleteUseCase_Execute_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	id := vo.NewID()

	mockRepo.On("Delete", mock.Anything, id).Return(nil)

	uc := NewDeleteUseCase(mockRepo)
	input := dto.DeleteInput{ID: id.String()}

	// Act
	output, deleteErr := uc.Execute(context.Background(), input)

	// Assert
	assert.NoError(t, deleteErr)
	assert.NotNil(t, output)
	assert.Equal(t, id.String(), output.ID)
	mockRepo.AssertExpectations(t)
}

func TestDeleteUseCase_Execute_NotFound(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	mockRepo.On("Delete", mock.Anything, mock.AnythingOfType("vo.ID")).
		Return(roledomain.ErrRoleNotFound)

	uc := NewDeleteUseCase(mockRepo)
	input := dto.DeleteInput{ID: "018e4a2c-6b4d-7000-9410-abcdef123456"}

	// Act
	output, deleteErr := uc.Execute(context.Background(), input)

	// Assert
	assert.Error(t, deleteErr)
	assert.Nil(t, output)

	var appErr *apperror.AppError
	assert.True(t, errors.As(deleteErr, &appErr))
	assert.Equal(t, apperror.CodeNotFound, appErr.Code)

	mockRepo.AssertExpectations(t)
}

func TestDeleteUseCase_Execute_InvalidID(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	uc := NewDeleteUseCase(mockRepo)
	input := dto.DeleteInput{ID: "invalid-id"}

	// Act
	output, deleteErr := uc.Execute(context.Background(), input)

	// Assert
	assert.Error(t, deleteErr)
	assert.Nil(t, output)

	var appErr *apperror.AppError
	assert.True(t, errors.As(deleteErr, &appErr))
	assert.Equal(t, apperror.CodeInvalidRequest, appErr.Code)

	mockRepo.AssertNotCalled(t, "Delete")
}

func TestDeleteUseCase_Execute_RepositoryError(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	mockRepo.On("Delete", mock.Anything, mock.AnythingOfType("vo.ID")).
		Return(errors.New("database error"))

	uc := NewDeleteUseCase(mockRepo)
	input := dto.DeleteInput{ID: "018e4a2c-6b4d-7000-9410-abcdef123456"}

	// Act
	output, deleteErr := uc.Execute(context.Background(), input)

	// Assert
	assert.Error(t, deleteErr)
	assert.Nil(t, output)

	var appErr *apperror.AppError
	assert.True(t, errors.As(deleteErr, &appErr))
	assert.Equal(t, apperror.CodeInternalError, appErr.Code)

	mockRepo.AssertExpectations(t)
}

// TC-UC-38: Delete with role-not-found records semantic span attribute
// `app.result=not_found` and leaves span status Unset.
func TestDeleteUseCase_Execute_TC_UC_38_NotFound_SpanAttribute(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	mockRepo.On("Delete", mock.Anything, mock.AnythingOfType("vo.ID")).
		Return(roledomain.ErrRoleNotFound)

	uc := NewDeleteUseCase(mockRepo)
	input := dto.DeleteInput{ID: "018e4a2c-6b4d-7000-9410-abcdef123456"}

	ctx, exp, endAndFlush := newRecordingContext(t)

	// Act
	output, deleteErr := uc.Execute(ctx, input)
	endAndFlush()

	// Assert: AppError(NotFound)
	require.Error(t, deleteErr)
	require.Nil(t, output)
	var appErr *apperror.AppError
	require.True(t, errors.As(deleteErr, &appErr))
	assert.Equal(t, apperror.CodeNotFound, appErr.Code)

	// Assert: semantic span attribute applied; status stays Unset.
	finished := exp.GetSpans()
	require.Len(t, finished, 1)
	stub := finished[0]
	assert.Equal(t, codes.Unset, stub.Status.Code,
		"role not-found is an expected outcome — span status must remain Unset")
	assert.Contains(t, stub.Attributes,
		attribute.String(ucshared.AttrKeyAppResult, "not_found"),
		"expected attribute %s=not_found", ucshared.AttrKeyAppResult)
}

// TC-UC-39: Delete with invalid ID records semantic span attribute
// `app.validation_error=<msg>` and leaves span status Unset.
func TestDeleteUseCase_Execute_TC_UC_39_InvalidID_SpanAttribute(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	uc := NewDeleteUseCase(mockRepo)
	input := dto.DeleteInput{ID: "invalid-id"}

	ctx, exp, endAndFlush := newRecordingContext(t)

	// Act
	output, deleteErr := uc.Execute(ctx, input)
	endAndFlush()

	// Assert: AppError(InvalidRequest)
	require.Error(t, deleteErr)
	require.Nil(t, output)
	var appErr *apperror.AppError
	require.True(t, errors.As(deleteErr, &appErr))
	assert.Equal(t, apperror.CodeInvalidRequest, appErr.Code)

	// Assert: semantic span attribute uses validation key with the underlying
	// error message as the value (AttrValue empty in mapping).
	finished := exp.GetSpans()
	require.Len(t, finished, 1)
	stub := finished[0]
	assert.Equal(t, codes.Unset, stub.Status.Code,
		"invalid-ID is an expected outcome — span status must remain Unset")
	assert.Contains(t, stub.Attributes,
		attribute.String(ucshared.AttrKeyAppValidationError, vo.ErrInvalidID.Error()),
		"expected attribute %s=<vo.ErrInvalidID message>", ucshared.AttrKeyAppValidationError)

	mockRepo.AssertNotCalled(t, "Delete")
}
