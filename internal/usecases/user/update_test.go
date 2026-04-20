package user

import (
	"context"
	"errors"
	"testing"
	"time"

	userdomain "github.com/jrmarcello/gopherplate/internal/domain/user"
	"github.com/jrmarcello/gopherplate/internal/domain/user/vo"
	ucshared "github.com/jrmarcello/gopherplate/internal/usecases/shared"
	"github.com/jrmarcello/gopherplate/internal/usecases/user/dto"
	"github.com/jrmarcello/gopherplate/pkg/apperror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel/codes"
)

// TC-UC-34: invalid ID -> AppError(CodeInvalidRequest); span attribute
// app.validation_error=<msg>; status=Unset (warn path).
func TestUpdateUseCase_Execute_InvalidID_SpanWarn(t *testing.T) {
	mockRepo := new(MockRepository)
	uc := NewUpdateUseCase(mockRepo)
	newName := "Updated Name"
	input := dto.UpdateInput{ID: "invalid-id", Name: &newName}

	ctx, finalize := newRecordingSpanContext(t)

	output, executeErr := uc.Execute(ctx, input)

	assert.Error(t, executeErr)
	assert.Nil(t, output)

	var appErr *apperror.AppError
	assert.True(t, errors.As(executeErr, &appErr), "expected *apperror.AppError")
	assert.Equal(t, apperror.CodeInvalidRequest, appErr.Code)
	mockRepo.AssertNotCalled(t, "FindByID")

	stub := finalize()
	assert.Equal(t, codes.Unset, stub.Status.Code,
		"invalid-id must classify as warn (status Unset)")
	assert.NotEmpty(t, attrValue(stub, ucshared.AttrKeyAppValidationError),
		"expected attribute %q on span", ucshared.AttrKeyAppValidationError)
}

func TestUpdateUseCase_Execute_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	id := vo.NewID()
	email, _ := vo.NewEmail("joao@example.com")

	existingEntity := &userdomain.User{
		ID:        id,
		Name:      "João Silva",
		Email:     email,
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockRepo.On("FindByID", mock.Anything, id).Return(existingEntity, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*user.User")).Return(nil)

	uc := NewUpdateUseCase(mockRepo)
	newName := "João Silva Updated"
	input := dto.UpdateInput{
		ID:   id.String(),
		Name: &newName,
	}

	// Act
	output, executeErr := uc.Execute(context.Background(), input)

	// Assert
	assert.NoError(t, executeErr)
	assert.NotNil(t, output)
	assert.Equal(t, "João Silva Updated", output.Name)
	mockRepo.AssertExpectations(t)
}

func TestUpdateUseCase_Execute_NotFound(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	mockRepo.On("FindByID", mock.Anything, mock.AnythingOfType("vo.ID")).
		Return(nil, userdomain.ErrUserNotFound)

	uc := NewUpdateUseCase(mockRepo)
	newName := "Updated Name"
	input := dto.UpdateInput{
		ID:   "018e4a2c-6b4d-7000-9410-abcdef123456",
		Name: &newName,
	}

	// Act
	output, executeErr := uc.Execute(context.Background(), input)

	// Assert
	assert.Error(t, executeErr)
	assert.Nil(t, output)

	var appErr *apperror.AppError
	assert.True(t, errors.As(executeErr, &appErr), "expected *apperror.AppError")
	assert.Equal(t, apperror.CodeNotFound, appErr.Code)
	assert.Equal(t, "user not found", appErr.Message)
	mockRepo.AssertNotCalled(t, "Update")
}

func TestUpdateUseCase_Execute_InvalidEmail(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	id := vo.NewID()
	email, _ := vo.NewEmail("joao@example.com")

	existingEntity := &userdomain.User{
		ID:        id,
		Name:      "João Silva",
		Email:     email,
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockRepo.On("FindByID", mock.Anything, id).Return(existingEntity, nil)

	uc := NewUpdateUseCase(mockRepo)
	invalidEmail := "invalid-email"
	input := dto.UpdateInput{
		ID:    id.String(),
		Email: &invalidEmail,
	}

	// Act
	output, executeErr := uc.Execute(context.Background(), input)

	// Assert
	assert.Error(t, executeErr)
	assert.Nil(t, output)

	var appErr *apperror.AppError
	assert.True(t, errors.As(executeErr, &appErr), "expected *apperror.AppError")
	assert.Equal(t, apperror.CodeInvalidRequest, appErr.Code)
	assert.Equal(t, "invalid email", appErr.Message)
	mockRepo.AssertNotCalled(t, "Update")
}

func TestUpdateUseCase_Execute_InvalidID(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	uc := NewUpdateUseCase(mockRepo)
	newName := "Updated Name"
	input := dto.UpdateInput{
		ID:   "invalid-id",
		Name: &newName,
	}

	// Act
	output, executeErr := uc.Execute(context.Background(), input)

	// Assert
	assert.Error(t, executeErr)
	assert.Nil(t, output)

	var appErr *apperror.AppError
	assert.True(t, errors.As(executeErr, &appErr), "expected *apperror.AppError")
	assert.Equal(t, apperror.CodeInvalidRequest, appErr.Code)
	assert.Equal(t, "invalid ID", appErr.Message)
	mockRepo.AssertNotCalled(t, "FindByID")
}

func TestUpdateUseCase_Execute_RepositoryUpdateError(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	id := vo.NewID()
	email, _ := vo.NewEmail("joao@example.com")

	existingEntity := &userdomain.User{
		ID:        id,
		Name:      "João Silva",
		Email:     email,
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockRepo.On("FindByID", mock.Anything, id).Return(existingEntity, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*user.User")).
		Return(errors.New("database error"))

	uc := NewUpdateUseCase(mockRepo)
	newName := "João Silva Updated"
	input := dto.UpdateInput{
		ID:   id.String(),
		Name: &newName,
	}

	// Act
	output, executeErr := uc.Execute(context.Background(), input)

	// Assert
	assert.Error(t, executeErr)
	assert.Nil(t, output)

	var appErr *apperror.AppError
	assert.True(t, errors.As(executeErr, &appErr), "expected *apperror.AppError")
	assert.Equal(t, apperror.CodeInternalError, appErr.Code)
	mockRepo.AssertExpectations(t)
}

func TestUpdateUseCase_Execute_CacheDeleteError_StillSucceeds(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	mockCache := new(MockCache)
	id := vo.NewID()
	email, _ := vo.NewEmail("joao@example.com")
	cacheKey := "user:" + id.String()

	existingEntity := &userdomain.User{
		ID:        id,
		Name:      "João Silva",
		Email:     email,
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockRepo.On("FindByID", mock.Anything, id).Return(existingEntity, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*user.User")).Return(nil)
	mockCache.On("Delete", mock.Anything, cacheKey).Return(errors.New("redis connection refused"))

	uc := NewUpdateUseCase(mockRepo).WithCache(mockCache)
	newName := "João Silva Updated"
	input := dto.UpdateInput{
		ID:   id.String(),
		Name: &newName,
	}

	// Act
	output, updateErr := uc.Execute(context.Background(), input)

	// Assert — update succeeds even though cache delete failed
	assert.NoError(t, updateErr)
	assert.NotNil(t, output)
	assert.Equal(t, "João Silva Updated", output.Name)
	mockCache.AssertCalled(t, "Delete", mock.Anything, cacheKey)
}

func TestUpdateUseCase_Execute_CacheInvalidation(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	mockCache := new(MockCache)
	id := vo.NewID()
	email, _ := vo.NewEmail("joao@example.com")
	cacheKey := "user:" + id.String()

	existingEntity := &userdomain.User{
		ID:        id,
		Name:      "João Silva",
		Email:     email,
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockRepo.On("FindByID", mock.Anything, id).Return(existingEntity, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*user.User")).Return(nil)
	mockCache.On("Delete", mock.Anything, cacheKey).Return(nil)

	uc := NewUpdateUseCase(mockRepo).WithCache(mockCache)
	newName := "João Silva Updated"
	input := dto.UpdateInput{
		ID:   id.String(),
		Name: &newName,
	}

	// Act
	output, executeErr := uc.Execute(context.Background(), input)

	// Assert
	assert.NoError(t, executeErr)
	assert.NotNil(t, output)
	mockCache.AssertCalled(t, "Delete", mock.Anything, cacheKey)
	mockRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}
