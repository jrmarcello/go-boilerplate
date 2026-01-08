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

func TestListUseCase_Execute_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)

	id1, id2 := vo.NewID(), vo.NewID()
	cpf1, _ := vo.NewCPF("52998224725")
	cpf2, _ := vo.NewCPF("11144477735")
	phone, _ := vo.NewPhone("11999999999")
	email1, _ := vo.NewEmail("joao@example.com")
	email2, _ := vo.NewEmail("maria@example.com")

	persons := []*person.Person{
		{
			ID: id1, Name: "João Silva", CPF: cpf1, Phone: phone, Email: email1,
			Active: true, CreatedAt: time.Now(), UpdatedAt: time.Now(),
		},
		{
			ID: id2, Name: "Maria Santos", CPF: cpf2, Phone: phone, Email: email2,
			Active: true, CreatedAt: time.Now(), UpdatedAt: time.Now(),
		},
	}

	expectedResult := &person.ListResult{
		Persons: persons,
		Total:   2,
		Page:    1,
		Limit:   20,
	}

	mockRepo.On("List", mock.Anything, mock.AnythingOfType("person.ListFilter")).Return(expectedResult, nil)

	uc := NewListUseCase(mockRepo)
	input := dto.ListInput{Page: 1, Limit: 20}

	// Act
	output, err := uc.Execute(context.Background(), input)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.Len(t, output.Data, 2)
	assert.Equal(t, "João Silva", output.Data[0].Name)
	assert.Equal(t, "Maria Santos", output.Data[1].Name)
	assert.Equal(t, 2, output.Pagination.Total)
	assert.Equal(t, 1, output.Pagination.Page)
	mockRepo.AssertExpectations(t)
}

func TestListUseCase_Execute_EmptyResult(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)

	expectedResult := &person.ListResult{
		Persons: []*person.Person{},
		Total:   0,
		Page:    1,
		Limit:   20,
	}

	mockRepo.On("List", mock.Anything, mock.AnythingOfType("person.ListFilter")).Return(expectedResult, nil)

	uc := NewListUseCase(mockRepo)
	input := dto.ListInput{Page: 1, Limit: 20}

	// Act
	output, err := uc.Execute(context.Background(), input)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.Empty(t, output.Data)
	assert.Equal(t, 0, output.Pagination.Total)
	assert.False(t, output.Pagination.HasNext)
	assert.False(t, output.Pagination.HasPrev)
	mockRepo.AssertExpectations(t)
}

func TestListUseCase_Execute_WithFilters(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)

	expectedResult := &person.ListResult{
		Persons: []*person.Person{},
		Total:   0,
		Page:    1,
		Limit:   20,
	}

	mockRepo.On("List", mock.Anything, mock.MatchedBy(func(filter person.ListFilter) bool {
		return filter.Name == "João" &&
			filter.Email == "example.com" &&
			filter.City == "São Paulo" &&
			filter.State == "SP" &&
			filter.ActiveOnly == true
	})).Return(expectedResult, nil)

	uc := NewListUseCase(mockRepo)
	input := dto.ListInput{
		Name:       "João",
		Email:      "example.com",
		City:       "São Paulo",
		State:      "SP",
		ActiveOnly: true,
		Page:       1,
		Limit:      20,
	}

	// Act
	output, err := uc.Execute(context.Background(), input)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, output)
	mockRepo.AssertExpectations(t)
}

func TestListUseCase_Execute_RepositoryError(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)
	mockRepo.On("List", mock.Anything, mock.AnythingOfType("person.ListFilter")).
		Return(nil, errors.New("database connection failed"))

	uc := NewListUseCase(mockRepo)
	input := dto.ListInput{Page: 1, Limit: 20}

	// Act
	output, err := uc.Execute(context.Background(), input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "database connection failed")
	mockRepo.AssertExpectations(t)
}

func TestListUseCase_Execute_Pagination(t *testing.T) {
	// Arrange
	mockRepo := new(MockRepository)

	expectedResult := &person.ListResult{
		Persons: []*person.Person{},
		Total:   100,
		Page:    3,
		Limit:   20,
	}

	mockRepo.On("List", mock.Anything, mock.AnythingOfType("person.ListFilter")).Return(expectedResult, nil)

	uc := NewListUseCase(mockRepo)
	input := dto.ListInput{Page: 3, Limit: 20}

	// Act
	output, err := uc.Execute(context.Background(), input)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, 5, output.Pagination.TotalPages) // 100/20 = 5
	assert.True(t, output.Pagination.HasNext)        // page 3 of 5
	assert.True(t, output.Pagination.HasPrev)        // page 3 > 1
	mockRepo.AssertExpectations(t)
}
