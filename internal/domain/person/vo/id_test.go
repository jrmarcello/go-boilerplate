package vo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewID(t *testing.T) {
	// Act
	id := NewID()

	// Assert
	assert.NotEmpty(t, id)
	assert.Len(t, id.String(), 26) // ULID has 26 characters
}

func TestNewID_Uniqueness(t *testing.T) {
	// Generate multiple IDs and ensure they are unique
	ids := make(map[string]bool)
	count := 100

	for i := 0; i < count; i++ {
		id := NewID()
		assert.False(t, ids[id.String()], "duplicate ID generated")
		ids[id.String()] = true
	}

	assert.Len(t, ids, count)
}

func TestParseID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid ULID",
			input:   "01ARZ3NDEKTSV4RRFFQ69G5FAV",
			wantErr: false,
		},
		{
			name:    "valid ULID - lowercase",
			input:   "01arz3ndektsv4rrffq69g5fav",
			wantErr: false,
		},
		{
			name:    "invalid ULID - too short",
			input:   "01ARZ3NDEK",
			wantErr: true,
		},
		{
			name:    "invalid ULID - too long",
			input:   "01ARZ3NDEKTSV4RRFFQ69G5FAVXXX",
			wantErr: true,
		},
		{
			name:    "invalid ULID - invalid characters",
			input:   "INVALID-ULID-FORMAT!!!",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := ParseID(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, id)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, id)
			}
		})
	}
}

func TestID_String(t *testing.T) {
	// Arrange
	id := NewID()

	// Act
	str := id.String()

	// Assert
	assert.Equal(t, string(id), str)
}

func TestID_ValueAndScan(t *testing.T) {
	// Arrange
	original := NewID()

	// Act - Value
	value, err := original.Value()
	require.NoError(t, err)

	// Assert - Value returns string
	strValue, ok := value.(string)
	require.True(t, ok)
	assert.Equal(t, original.String(), strValue)

	// Act - Scan from string
	var scannedFromString ID
	err = scannedFromString.Scan(strValue)
	require.NoError(t, err)
	assert.Equal(t, original, scannedFromString)

	// Act - Scan from []byte
	var scannedFromBytes ID
	err = scannedFromBytes.Scan([]byte(strValue))
	require.NoError(t, err)
	assert.Equal(t, original, scannedFromBytes)
}

func TestID_Scan_Error(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
	}{
		{
			name:  "nil value",
			value: nil,
		},
		{
			name:  "invalid type - int",
			value: 12345,
		},
		{
			name:  "invalid type - float",
			value: 123.45,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var id ID
			err := id.Scan(tt.value)
			assert.Error(t, err)
		})
	}
}

func TestParseID_RoundTrip(t *testing.T) {
	// Test that a generated ID can be parsed back
	original := NewID()

	parsed, err := ParseID(original.String())
	require.NoError(t, err)
	assert.Equal(t, original, parsed)
}
