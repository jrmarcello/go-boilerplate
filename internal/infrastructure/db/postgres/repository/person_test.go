package repository

import (
	"database/sql"
	"testing"
	"time"

	"bitbucket.org/appmax-space/ms-boilerplate-go/internal/domain/person"
	"bitbucket.org/appmax-space/ms-boilerplate-go/internal/domain/person/vo"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Unit Tests for internal conversions (não precisam de banco)
// =============================================================================

func TestPeopleDB_ToEntity(t *testing.T) {
	// Arrange
	dbModel := peopleDB{
		ID:           "01ARZ3NDEKTSV4RRFFQ69G5FAV",
		Name:         "João Silva",
		Document:     "52998224725",
		Phone:        sql.NullString{String: "11999999999", Valid: true},
		Email:        "joao@example.com",
		Street:       sql.NullString{String: "Av. Paulista", Valid: true},
		Number:       sql.NullString{String: "1000", Valid: true},
		Complement:   sql.NullString{String: "Apto 101", Valid: true},
		Neighborhood: sql.NullString{String: "Bela Vista", Valid: true},
		City:         sql.NullString{String: "São Paulo", Valid: true},
		State:        sql.NullString{String: "SP", Valid: true},
		ZipCode:      sql.NullString{String: "01310-100", Valid: true},
		Active:       true,
		CreatedAt:    time.Now().Add(-24 * time.Hour),
		UpdatedAt:    time.Now(),
	}

	// Act
	entity, err := dbModel.toEntity()

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, entity)
	assert.Equal(t, "01ARZ3NDEKTSV4RRFFQ69G5FAV", entity.ID.String())
	assert.Equal(t, "João Silva", entity.Name)
	assert.Equal(t, "52998224725", entity.CPF.String())
	assert.Equal(t, "11999999999", entity.Phone.String())
	assert.Equal(t, "joao@example.com", entity.Email.String())
	assert.Equal(t, "Av. Paulista", entity.Address.Street)
	assert.Equal(t, "São Paulo", entity.Address.City)
	assert.Equal(t, "SP", entity.Address.State)
	assert.True(t, entity.Active)
}

func TestPeopleDB_ToEntity_InvalidID(t *testing.T) {
	// Arrange
	dbModel := peopleDB{
		ID:       "invalid-id",
		Name:     "Test",
		Document: "52998224725",
		Email:    "test@example.com",
	}

	// Act
	entity, err := dbModel.toEntity()

	// Assert
	assert.Error(t, err)
	assert.Nil(t, entity)
	assert.Contains(t, err.Error(), "erro ao parsear ID")
}

func TestPeopleDB_ToEntity_InvalidCPF(t *testing.T) {
	// Arrange
	dbModel := peopleDB{
		ID:       "01ARZ3NDEKTSV4RRFFQ69G5FAV",
		Name:     "Test",
		Document: "12345678901", // CPF inválido
		Email:    "test@example.com",
	}

	// Act
	entity, err := dbModel.toEntity()

	// Assert
	assert.Error(t, err)
	assert.Nil(t, entity)
	assert.Contains(t, err.Error(), "erro ao parsear CPF")
}

func TestPeopleDB_ToEntity_InvalidEmail(t *testing.T) {
	// Arrange
	dbModel := peopleDB{
		ID:       "01ARZ3NDEKTSV4RRFFQ69G5FAV",
		Name:     "Test",
		Document: "52998224725",
		Email:    "invalid-email",
	}

	// Act
	entity, err := dbModel.toEntity()

	// Assert
	assert.Error(t, err)
	assert.Nil(t, entity)
	assert.Contains(t, err.Error(), "erro ao parsear email")
}

func TestFromEntity(t *testing.T) {
	// Arrange
	cpf, _ := vo.NewCPF("52998224725")
	phone, _ := vo.NewPhone("11999999999")
	email, _ := vo.NewEmail("joao@example.com")

	entity := &person.Person{
		ID:    vo.NewID(),
		Name:  "João Silva",
		CPF:   cpf,
		Phone: phone,
		Email: email,
		Address: vo.Address{
			Street:       "Av. Paulista",
			Number:       "1000",
			Complement:   "Apto 101",
			Neighborhood: "Bela Vista",
			City:         "São Paulo",
			State:        "SP",
			ZipCode:      "01310-100",
		},
		Active:    true,
		CreatedAt: time.Now().Add(-24 * time.Hour),
		UpdatedAt: time.Now(),
	}

	// Act
	dbModel := fromEntity(entity)

	// Assert
	assert.Equal(t, entity.ID.String(), dbModel.ID)
	assert.Equal(t, entity.Name, dbModel.Name)
	assert.Equal(t, entity.CPF.String(), dbModel.Document)
	assert.Equal(t, entity.Phone.String(), dbModel.Phone.String)
	assert.True(t, dbModel.Phone.Valid)
	assert.Equal(t, entity.Email.String(), dbModel.Email)
	assert.Equal(t, entity.Address.Street, dbModel.Street.String)
	assert.True(t, dbModel.Street.Valid)
	assert.Equal(t, entity.Address.City, dbModel.City.String)
	assert.True(t, dbModel.City.Valid)
	assert.Equal(t, entity.Active, dbModel.Active)
}

func TestFromEntity_EmptyOptionalFields(t *testing.T) {
	// Arrange
	cpf, _ := vo.NewCPF("52998224725")
	phone, _ := vo.NewPhone("11999999999")
	email, _ := vo.NewEmail("joao@example.com")

	entity := &person.Person{
		ID:        vo.NewID(),
		Name:      "João Silva",
		CPF:       cpf,
		Phone:     phone,
		Email:     email,
		Address:   vo.Address{}, // Endereço vazio
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Act
	dbModel := fromEntity(entity)

	// Assert
	assert.False(t, dbModel.Street.Valid, "Street should be NULL")
	assert.False(t, dbModel.Number.Valid, "Number should be NULL")
	assert.False(t, dbModel.Complement.Valid, "Complement should be NULL")
	assert.False(t, dbModel.Neighborhood.Valid, "Neighborhood should be NULL")
	assert.False(t, dbModel.City.Valid, "City should be NULL")
	assert.False(t, dbModel.State.Valid, "State should be NULL")
	assert.False(t, dbModel.ZipCode.Valid, "ZipCode should be NULL")
}

func TestToNullString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected sql.NullString
	}{
		{
			name:  "non-empty string",
			input: "test value",
			expected: sql.NullString{
				String: "test value",
				Valid:  true,
			},
		},
		{
			name:  "empty string",
			input: "",
			expected: sql.NullString{
				String: "",
				Valid:  false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toNullString(tt.input)
			assert.Equal(t, tt.expected.String, result.String)
			assert.Equal(t, tt.expected.Valid, result.Valid)
		})
	}
}

func TestPeopleDB_ToEntity_EmptyAddress(t *testing.T) {
	// Arrange
	dbModel := peopleDB{
		ID:           "01ARZ3NDEKTSV4RRFFQ69G5FAV",
		Name:         "João Silva",
		Document:     "52998224725",
		Phone:        sql.NullString{String: "11999999999", Valid: true},
		Email:        "joao@example.com",
		Street:       sql.NullString{Valid: false},
		Number:       sql.NullString{Valid: false},
		Complement:   sql.NullString{Valid: false},
		Neighborhood: sql.NullString{Valid: false},
		City:         sql.NullString{Valid: false},
		State:        sql.NullString{Valid: false},
		ZipCode:      sql.NullString{Valid: false},
		Active:       true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Act
	entity, err := dbModel.toEntity()

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, entity)
	assert.True(t, entity.Address.IsEmpty())
}

func TestFromEntity_RoundTrip(t *testing.T) {
	// Test that we can convert entity -> dbModel -> entity without data loss
	cpf, _ := vo.NewCPF("52998224725")
	phone, _ := vo.NewPhone("11999999999")
	email, _ := vo.NewEmail("joao@example.com")

	original := &person.Person{
		ID:    vo.NewID(),
		Name:  "João Silva",
		CPF:   cpf,
		Phone: phone,
		Email: email,
		Address: vo.Address{
			Street:       "Av. Paulista",
			Number:       "1000",
			Complement:   "Apto 101",
			Neighborhood: "Bela Vista",
			City:         "São Paulo",
			State:        "SP",
			ZipCode:      "01310100",
		},
		Active:    true,
		CreatedAt: time.Now().Truncate(time.Microsecond),
		UpdatedAt: time.Now().Truncate(time.Microsecond),
	}

	// Convert to DB model
	dbModel := fromEntity(original)

	// Convert back to entity
	restored, err := dbModel.toEntity()

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, original.ID, restored.ID)
	assert.Equal(t, original.Name, restored.Name)
	assert.Equal(t, original.CPF.String(), restored.CPF.String())
	assert.Equal(t, original.Phone.String(), restored.Phone.String())
	assert.Equal(t, original.Email.String(), restored.Email.String())
	assert.Equal(t, original.Address.Street, restored.Address.Street)
	assert.Equal(t, original.Address.City, restored.Address.City)
	assert.Equal(t, original.Active, restored.Active)
}
