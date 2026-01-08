package vo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddress_IsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		address  Address
		expected bool
	}{
		{
			name:     "empty address",
			address:  Address{},
			expected: true,
		},
		{
			name: "only street filled",
			address: Address{
				Street: "Rua das Flores",
			},
			expected: false,
		},
		{
			name: "only number filled",
			address: Address{
				Number: "123",
			},
			expected: false,
		},
		{
			name: "only complement filled",
			address: Address{
				Complement: "Apto 101",
			},
			expected: false,
		},
		{
			name: "only neighborhood filled",
			address: Address{
				Neighborhood: "Centro",
			},
			expected: false,
		},
		{
			name: "only city filled",
			address: Address{
				City: "São Paulo",
			},
			expected: false,
		},
		{
			name: "only state filled",
			address: Address{
				State: "SP",
			},
			expected: false,
		},
		{
			name: "only zip code filled",
			address: Address{
				ZipCode: "01310-100",
			},
			expected: false,
		},
		{
			name: "fully filled address",
			address: Address{
				Street:       "Av. Paulista",
				Number:       "1000",
				Complement:   "Sala 501",
				Neighborhood: "Bela Vista",
				City:         "São Paulo",
				State:        "SP",
				ZipCode:      "01310-100",
			},
			expected: false,
		},
		{
			name: "partial address - street and city",
			address: Address{
				Street: "Rua Principal",
				City:   "Rio de Janeiro",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.address.IsEmpty()
			assert.Equal(t, tt.expected, result)
		})
	}
}
