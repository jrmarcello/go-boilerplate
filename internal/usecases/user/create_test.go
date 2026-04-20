package user

import (
	"context"
	"errors"
	"testing"

	userdomain "github.com/jrmarcello/gopherplate/internal/domain/user"
	ucshared "github.com/jrmarcello/gopherplate/internal/usecases/shared"
	"github.com/jrmarcello/gopherplate/internal/usecases/user/dto"
	"github.com/jrmarcello/gopherplate/pkg/apperror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel/codes"
)

func TestCreateUseCase_Execute_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*user.User")).Return(nil)

	uc := NewCreateUseCase(mockRepo)
	input := dto.CreateInput{
		Name:  "João Silva",
		Email: "joao@example.com",
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

// TC-UC-30: invalid email -> AppError(CodeInvalidRequest); span attribute
// app.validation_error=<msg>; status=Unset (warn path).
func TestCreateUseCase_Execute_InvalidEmail(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	uc := NewCreateUseCase(mockRepo)
	input := dto.CreateInput{
		Name:  "João Silva",
		Email: "invalid-email",
	}

	ctx, finalize := newRecordingSpanContext(t)

	// Act
	output, executeErr := uc.Execute(ctx, input)

	// Assert
	assert.Error(t, executeErr)
	assert.Nil(t, output)

	var appErr *apperror.AppError
	assert.True(t, errors.As(executeErr, &appErr), "expected *apperror.AppError")
	assert.Equal(t, apperror.CodeInvalidRequest, appErr.Code)
	assert.Equal(t, "invalid email", appErr.Message)
	mockRepo.AssertNotCalled(t, "Create")

	stub := finalize()
	assert.Equal(t, codes.Unset, stub.Status.Code,
		"validation failure must keep span status Unset")
	assert.NotEmpty(t, attrValue(stub, ucshared.AttrKeyAppValidationError),
		"expected attribute %q on span", ucshared.AttrKeyAppValidationError)
}

// TC-UC-31: duplicate email -> AppError(CodeConflict); span attribute
// app.result=duplicate_email; status=Unset (warn path).
func TestCreateUseCase_Execute_DuplicateEmail(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*user.User")).
		Return(userdomain.ErrDuplicateEmail)

	uc := NewCreateUseCase(mockRepo)
	input := dto.CreateInput{
		Name:  "João Silva",
		Email: "joao@example.com",
	}

	ctx, finalize := newRecordingSpanContext(t)

	// Act
	output, executeErr := uc.Execute(ctx, input)

	// Assert
	assert.Error(t, executeErr)
	assert.Nil(t, output)

	var appErr *apperror.AppError
	assert.True(t, errors.As(executeErr, &appErr), "expected *apperror.AppError")
	assert.Equal(t, apperror.CodeConflict, appErr.Code)
	assert.Equal(t, "email already exists", appErr.Message)
	mockRepo.AssertExpectations(t)

	stub := finalize()
	assert.Equal(t, codes.Unset, stub.Status.Code,
		"duplicate-email must classify as warn (status Unset)")
	assert.True(t, hasAttr(stub, ucshared.AttrKeyAppResult, "duplicate_email"),
		"expected attribute %s=duplicate_email on span; got attrs=%v",
		ucshared.AttrKeyAppResult, stub.Attributes)
}

// TC-UC-32: generic infra error -> FailSpan path: status=Error;
// `error.type` recorded on the exception event (TASK-1 enrichment).
// This is the only test that asserts error.type — all other tests stay
// on the warn path.
func TestCreateUseCase_Execute_RepositoryError(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*user.User")).
		Return(errors.New("database connection failed"))

	uc := NewCreateUseCase(mockRepo)
	input := dto.CreateInput{
		Name:  "João Silva",
		Email: "joao@example.com",
	}

	ctx, finalize := newRecordingSpanContext(t)

	// Act
	output, executeErr := uc.Execute(ctx, input)

	// Assert
	assert.Error(t, executeErr)
	assert.Nil(t, output)

	var appErr *apperror.AppError
	assert.True(t, errors.As(executeErr, &appErr), "expected *apperror.AppError")
	assert.Equal(t, apperror.CodeInternalError, appErr.Code)
	assert.Equal(t, "internal server error", appErr.Message)
	mockRepo.AssertExpectations(t)

	stub := finalize()
	assert.Equal(t, codes.Error, stub.Status.Code,
		"unexpected infra error must mark span Error")
	assert.True(t, hasExceptionEventAttr(stub, "error.type"),
		"expected `error.type` attribute on the exception event (TASK-1 enrichment)")
}
