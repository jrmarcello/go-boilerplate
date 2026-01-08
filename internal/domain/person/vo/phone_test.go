package vo

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPhone(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string // valor esperado após limpeza
		wantErr bool
	}{
		// Casos válidos - celular (11 dígitos)
		{
			name:    "valid cellphone - only digits",
			input:   "11999999999",
			want:    "11999999999",
			wantErr: false,
		},
		{
			name:    "valid cellphone - formatted with parenthesis",
			input:   "(11) 99999-9999",
			want:    "11999999999",
			wantErr: false,
		},
		{
			name:    "valid cellphone - formatted with spaces",
			input:   "11 99999 9999",
			want:    "11999999999",
			wantErr: false,
		},
		{
			name:    "valid cellphone - with country code +55",
			input:   "+5511999999999",
			want:    "11999999999",
			wantErr: false,
		},
		// Casos válidos - fixo (10 dígitos)
		{
			name:    "valid landline - only digits",
			input:   "1133334444",
			want:    "1133334444",
			wantErr: false,
		},
		{
			name:    "valid landline - formatted",
			input:   "(11) 3333-4444",
			want:    "1133334444",
			wantErr: false,
		},
		// Casos inválidos
		{
			name:    "invalid phone - too short",
			input:   "119999999", // 9 dígitos
			wantErr: true,
		},
		{
			name:    "invalid phone - too long",
			input:   "119999999999", // 12 dígitos
			wantErr: true,
		},
		{
			name:    "invalid phone - empty",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid phone - letters only",
			input:   "abcdefghij",
			wantErr: true,
		},
		{
			name:    "invalid phone - mixed letters and numbers",
			input:   "11999abc99",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewPhone(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, ErrInvalidPhone), "expected ErrInvalidPhone")
				assert.Empty(t, got.String())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got.String())
			}
		})
	}
}

func TestPhone_ValueAndScan(t *testing.T) {
	phone, _ := NewPhone("11999999999")

	// Test Value - retorna string
	val, err := phone.Value()
	assert.NoError(t, err)
	assert.Equal(t, "11999999999", val)

	// Test Value com telefone vazio - retorna nil
	emptyPhone := Phone{}
	val, err = emptyPhone.Value()
	assert.NoError(t, err)
	assert.Nil(t, val)

	// Test Scan with string
	var scannedPhone Phone
	err = scannedPhone.Scan("11999999999")
	assert.NoError(t, err)
	assert.Equal(t, "11999999999", scannedPhone.String())

	// Test Scan with []byte
	var scannedPhone2 Phone
	err = scannedPhone2.Scan([]byte("11999999999"))
	assert.NoError(t, err)
	assert.Equal(t, "11999999999", scannedPhone2.String())

	// Test Scan with nil - retorna vazio (nullable)
	var scannedPhone3 Phone
	err = scannedPhone3.Scan(nil)
	assert.NoError(t, err)
	assert.Empty(t, scannedPhone3.String())

	// Test Scan with invalid type
	var scannedPhone4 Phone
	err = scannedPhone4.Scan(123)
	assert.Error(t, err)
}

func TestPhone_IsEmpty(t *testing.T) {
	// Telefone vazio
	emptyPhone := Phone{}
	assert.True(t, emptyPhone.IsEmpty())

	// Telefone preenchido
	phone, _ := NewPhone("11999999999")
	assert.False(t, phone.IsEmpty())
}

func TestParsePhone(t *testing.T) {
	// ParsePhone não valida - usado para dados do banco
	phone := ParsePhone("11999999999")
	assert.Equal(t, "11999999999", phone.String())

	// Aceita qualquer string (dados já validados)
	invalidPhone := ParsePhone("abc")
	assert.Equal(t, "abc", invalidPhone.String())
}
