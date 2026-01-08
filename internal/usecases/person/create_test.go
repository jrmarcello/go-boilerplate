package person

import (
	"context"
	"errors"
	"testing"

	"bitbucket.org/appmax-space/ms-boilerplate-go/internal/domain/person/vo"
	"bitbucket.org/appmax-space/ms-boilerplate-go/internal/usecases/person/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateUseCase_Execute_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*person.Person")).Return(nil)

	uc := NewCreateUseCase(mockRepo)
	input := dto.CreateInput{
		Name:     "João Silva",
		Document: "52998224725", // CPF válido
		Phone:    "11999999999",
		Email:    "joao@example.com",
	}

	// Act
	output, err := uc.Execute(context.Background(), input)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.NotEmpty(t, output.ID)
	assert.NotEmpty(t, output.CreatedAt)
	mockRepo.AssertExpectations(t)
}

func TestCreateUseCase_Execute_InvalidCPF(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	uc := NewCreateUseCase(mockRepo)
	input := dto.CreateInput{
		Name:     "João Silva",
		Document: "12345678901", // CPF com dígitos verificadores inválidos
		Phone:    "11999999999",
		Email:    "joao@example.com",
	}

	// Act
	output, err := uc.Execute(context.Background(), input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.True(t, errors.Is(err, vo.ErrInvalidCPF), "expected ErrInvalidCPF")
	mockRepo.AssertNotCalled(t, "Create")
}

func TestCreateUseCase_Execute_InvalidEmail(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	uc := NewCreateUseCase(mockRepo)
	input := dto.CreateInput{
		Name:     "João Silva",
		Document: "52998224725",
		Phone:    "11999999999",
		Email:    "invalid-email",
	}

	// Act
	output, err := uc.Execute(context.Background(), input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.True(t, errors.Is(err, vo.ErrInvalidEmail), "expected ErrInvalidEmail")
	mockRepo.AssertNotCalled(t, "Create")
}

func TestCreateUseCase_Execute_RepositoryError(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*person.Person")).
		Return(errors.New("database connection failed"))

	uc := NewCreateUseCase(mockRepo)
	input := dto.CreateInput{
		Name:     "João Silva",
		Document: "52998224725",
		Phone:    "11999999999",
		Email:    "joao@example.com",
	}

	// Act
	output, err := uc.Execute(context.Background(), input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "database connection failed")
	mockRepo.AssertExpectations(t)
}

func TestCreateUseCase_Execute_InvalidPhone(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	uc := NewCreateUseCase(mockRepo)
	input := dto.CreateInput{
		Name:     "João Silva",
		Document: "52998224725",
		Phone:    "123", // Telefone inválido (muito curto)
		Email:    "joao@example.com",
	}

	// Act
	output, err := uc.Execute(context.Background(), input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.True(t, errors.Is(err, vo.ErrInvalidPhone), "expected ErrInvalidPhone")
	mockRepo.AssertNotCalled(t, "Create")
}
