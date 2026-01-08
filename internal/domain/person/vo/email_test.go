package vo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEmail(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid email",
			input:   "test@example.com",
			wantErr: false,
		},
		{
			name:    "invalid email format",
			input:   "invalid-email",
			wantErr: true,
		},
		{
			name:    "empty email",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewEmail(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.input, got.String())
			}
		})
	}
}
