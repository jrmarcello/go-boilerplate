package person

import (
	"context"
	"errors"
	"testing"
	"time"

	"bitbucket.org/appmax-space/ms-boilerplate-go/internal/domain/person"
	"bitbucket.org/appmax-space/ms-boilerplate-go/internal/domain/person/vo"
	"bitbucket.org/appmax-space/ms-boilerplate-go/internal/usecases/person/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateUseCase_Execute_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	id := vo.NewID()
	cpf, _ := vo.NewCPF("52998224725")
	phone, _ := vo.NewPhone("11999999999")
	email, _ := vo.NewEmail("joao@example.com")

	existingPerson := &person.Person{
		ID:        id,
		Name:      "João Silva",
		CPF:       cpf,
		Phone:     phone,
		Email:     email,
		Active:    true,
		CreatedAt: time.Now().Add(-24 * time.Hour),
		UpdatedAt: time.Now().Add(-24 * time.Hour),
	}

	mockRepo.On("FindByID", mock.Anything, id).Return(existingPerson, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*person.Person")).Return(nil)

	uc := NewUpdateUseCase(mockRepo, nil)
	newName := "João Santos"
	input := dto.UpdateInput{
		ID:   id.String(),
		Name: &newName,
	}

	// Act
	output, err := uc.Execute(context.Background(), input)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, id.String(), output.ID)
	assert.Equal(t, "João Santos", output.Name)
	mockRepo.AssertExpectations(t)
}

func TestUpdateUseCase_Execute_UpdateEmail(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	id := vo.NewID()
	cpf, _ := vo.NewCPF("52998224725")
	phone, _ := vo.NewPhone("11999999999")
	email, _ := vo.NewEmail("joao@example.com")

	existingPerson := &person.Person{
		ID:        id,
		Name:      "João Silva",
		CPF:       cpf,
		Phone:     phone,
		Email:     email,
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockRepo.On("FindByID", mock.Anything, id).Return(existingPerson, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*person.Person")).Return(nil)

	uc := NewUpdateUseCase(mockRepo, nil)
	newEmail := "joao.santos@example.com"
	input := dto.UpdateInput{
		ID:    id.String(),
		Email: &newEmail,
	}

	// Act
	output, err := uc.Execute(context.Background(), input)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "joao.santos@example.com", output.Email)
	mockRepo.AssertExpectations(t)
}

func TestUpdateUseCase_Execute_UpdatePhone(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	id := vo.NewID()
	cpf, _ := vo.NewCPF("52998224725")
	phone, _ := vo.NewPhone("11999999999")
	email, _ := vo.NewEmail("joao@example.com")

	existingPerson := &person.Person{
		ID:        id,
		Name:      "João Silva",
		CPF:       cpf,
		Phone:     phone,
		Email:     email,
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockRepo.On("FindByID", mock.Anything, id).Return(existingPerson, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*person.Person")).Return(nil)

	uc := NewUpdateUseCase(mockRepo, nil)
	newPhone := "11888888888"
	input := dto.UpdateInput{
		ID:    id.String(),
		Phone: &newPhone,
	}

	// Act
	output, err := uc.Execute(context.Background(), input)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "11888888888", output.Phone)
	mockRepo.AssertExpectations(t)
}

func TestUpdateUseCase_Execute_NotFound(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	mockRepo.On("FindByID", mock.Anything, mock.AnythingOfType("vo.ID")).
		Return(nil, person.ErrPersonNotFound)

	uc := NewUpdateUseCase(mockRepo, nil)
	newName := "João Santos"
	input := dto.UpdateInput{
		ID:   "01ARZ3NDEKTSV4RRFFQ69G5FAV",
		Name: &newName,
	}

	// Act
	output, err := uc.Execute(context.Background(), input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.True(t, errors.Is(err, person.ErrPersonNotFound))
	mockRepo.AssertNotCalled(t, "Update")
}

func TestUpdateUseCase_Execute_InvalidID(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	uc := NewUpdateUseCase(mockRepo, nil)
	newName := "João Santos"
	input := dto.UpdateInput{
		ID:   "invalid-id",
		Name: &newName,
	}

	// Act
	output, err := uc.Execute(context.Background(), input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, output)
	mockRepo.AssertNotCalled(t, "FindByID")
}

func TestUpdateUseCase_Execute_InvalidEmail(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	id := vo.NewID()
	cpf, _ := vo.NewCPF("52998224725")
	phone, _ := vo.NewPhone("11999999999")
	email, _ := vo.NewEmail("joao@example.com")

	existingPerson := &person.Person{
		ID:    id,
		Name:  "João Silva",
		CPF:   cpf,
		Phone: phone,
		Email: email,
	}

	mockRepo.On("FindByID", mock.Anything, id).Return(existingPerson, nil)

	uc := NewUpdateUseCase(mockRepo, nil)
	invalidEmail := "invalid-email"
	input := dto.UpdateInput{
		ID:    id.String(),
		Email: &invalidEmail,
	}

	// Act
	output, err := uc.Execute(context.Background(), input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.True(t, errors.Is(err, vo.ErrInvalidEmail))
	mockRepo.AssertNotCalled(t, "Update")
}

func TestUpdateUseCase_Execute_RepositoryError(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	id := vo.NewID()
	cpf, _ := vo.NewCPF("52998224725")
	phone, _ := vo.NewPhone("11999999999")
	email, _ := vo.NewEmail("joao@example.com")

	existingPerson := &person.Person{
		ID:    id,
		Name:  "João Silva",
		CPF:   cpf,
		Phone: phone,
		Email: email,
	}

	mockRepo.On("FindByID", mock.Anything, id).Return(existingPerson, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*person.Person")).
		Return(errors.New("database error"))

	uc := NewUpdateUseCase(mockRepo, nil)
	newName := "João Santos"
	input := dto.UpdateInput{
		ID:   id.String(),
		Name: &newName,
	}

	// Act
	output, err := uc.Execute(context.Background(), input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "database error")
	mockRepo.AssertExpectations(t)
}

func TestUpdateUseCase_Execute_UpdateAddress(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	id := vo.NewID()
	cpf, _ := vo.NewCPF("52998224725")
	phone, _ := vo.NewPhone("11999999999")
	email, _ := vo.NewEmail("joao@example.com")

	existingPerson := &person.Person{
		ID:        id,
		Name:      "João Silva",
		CPF:       cpf,
		Phone:     phone,
		Email:     email,
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockRepo.On("FindByID", mock.Anything, id).Return(existingPerson, nil)
	mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(c *person.Person) bool {
		return c.Address.Street == "Av. Paulista" && c.Address.City == "São Paulo"
	})).Return(nil)

	uc := NewUpdateUseCase(mockRepo, nil)
	input := dto.UpdateInput{
		ID: id.String(),
		Address: &dto.UpdateAddressDTO{
			Street: "Av. Paulista",
			Number: "1000",
			City:   "São Paulo",
			State:  "SP",
		},
	}

	// Act
	output, err := uc.Execute(context.Background(), input)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, output)
	mockRepo.AssertExpectations(t)
}

func TestUpdateUseCase_Execute_InvalidPhone(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	id := vo.NewID()
	cpf, _ := vo.NewCPF("52998224725")
	phone, _ := vo.NewPhone("11999999999")
	email, _ := vo.NewEmail("joao@example.com")

	existingPerson := &person.Person{
		ID:    id,
		Name:  "João Silva",
		CPF:   cpf,
		Phone: phone,
		Email: email,
	}

	mockRepo.On("FindByID", mock.Anything, id).Return(existingPerson, nil)

	uc := NewUpdateUseCase(mockRepo, nil)
	invalidPhone := "123" // Telefone muito curto
	input := dto.UpdateInput{
		ID:    id.String(),
		Phone: &invalidPhone,
	}

	// Act
	output, err := uc.Execute(context.Background(), input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.True(t, errors.Is(err, vo.ErrInvalidPhone))
	mockRepo.AssertNotCalled(t, "Update")
}
