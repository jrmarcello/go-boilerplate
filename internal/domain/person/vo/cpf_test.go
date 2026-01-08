package vo

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCPF(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid cpf - 529.982.247-25",
			input:   "52998224725",
			wantErr: false,
		},
		{
			name:    "valid cpf formatted",
			input:   "529.982.247-25",
			wantErr: false,
		},
		{
			name:    "valid cpf - 111.444.777-35",
			input:   "11144477735",
			wantErr: false,
		},
		{
			name:    "invalid cpf - wrong check digits",
			input:   "12345678901",
			wantErr: true,
		},
		{
			name:    "invalid cpf - all same digits",
			input:   "11111111111",
			wantErr: true,
		},
		{
			name:    "invalid cpf - all zeros",
			input:   "00000000000",
			wantErr: true,
		},
		{
			name:    "invalid cpf length - too short",
			input:   "123",
			wantErr: true,
		},
		{
			name:    "empty cpf",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewCPF(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, ErrInvalidCPF), "expected ErrInvalidCPF")
				assert.Empty(t, got.String())
			} else {
				assert.NoError(t, err)
				// CPF é armazenado sem formatação
				assert.Len(t, got.String(), 11)
			}
		})
	}
}

func TestCPF_ValueAndScan(t *testing.T) {
	cpf, _ := NewCPF("52998224725")

	// Test Value
	val, err := cpf.Value()
	assert.NoError(t, err)
	assert.Equal(t, "52998224725", val)

	// Test Scan with string
	var scannedCPF CPF
	err = scannedCPF.Scan("52998224725")
	assert.NoError(t, err)
	assert.Equal(t, "52998224725", scannedCPF.String())

	// Test Scan with []byte
	var scannedCPF2 CPF
	err = scannedCPF2.Scan([]byte("52998224725"))
	assert.NoError(t, err)
	assert.Equal(t, "52998224725", scannedCPF2.String())

	// Test Scan with nil
	var scannedCPF3 CPF
	err = scannedCPF3.Scan(nil)
	assert.Error(t, err)
}
