package logutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaskEmail(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "normal email", input: "user@example.com", want: "u***@example.com"},
		{name: "single char local", input: "u@example.com", want: "u***@example.com"},
		{name: "empty string", input: "", want: "***"},
		{name: "no at sign", input: "invalid", want: "***"},
		{name: "at sign at start", input: "@example.com", want: "***"},
		{name: "long local part", input: "joao.silva@email.com", want: "j***@email.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MaskEmail(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMaskDocument(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "CPF 11 digits", input: "12345678901", want: "***8901"},
		{name: "CNPJ 14 digits", input: "12345678000195", want: "***0195"},
		{name: "short value", input: "abc", want: "***"},
		{name: "exactly 4 chars", input: "abcd", want: "***"},
		{name: "5 chars", input: "abcde", want: "***bcde"},
		{name: "empty string", input: "", want: "***"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MaskDocument(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMaskName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "two parts", input: "Joao Silva", want: "J*** S***"},
		{name: "three parts", input: "Joao Carlos Silva", want: "J*** C*** S***"},
		{name: "single name", input: "Joao", want: "J***"},
		{name: "single char name", input: "J", want: "J"},
		{name: "empty string", input: "", want: "***"},
		{name: "whitespace only", input: "   ", want: "***"},
		{name: "leading/trailing spaces", input: "  Joao Silva  ", want: "J*** S***"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MaskName(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMaskPhone(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "with country code", input: "+5511999998888", want: "+55***8888"},
		{name: "without plus", input: "5511999998888", want: "55***8888"},
		{name: "short number", input: "1234", want: "***"},
		{name: "empty string", input: "", want: "***"},
		{name: "formatted phone", input: "+55 (11) 99999-8888", want: "+55***8888"},
		{name: "5 digits", input: "12345", want: "12***2345"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MaskPhone(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMaskSensitivePayload(t *testing.T) {
	t.Run("nil input returns nil", func(t *testing.T) {
		result := MaskSensitivePayload(nil)
		assert.Nil(t, result)
	})

	t.Run("non-map input returned unchanged", func(t *testing.T) {
		result := MaskSensitivePayload("just a string")
		assert.Equal(t, "just a string", result)
	})

	t.Run("masks email field", func(t *testing.T) {
		input := map[string]any{"email": "user@example.com", "id": 123}
		result := MaskSensitivePayload(input)

		m, ok := result.(map[string]any)
		assert.True(t, ok)
		assert.Equal(t, "u***@example.com", m["email"])
		assert.Equal(t, 123, m["id"])
	})

	t.Run("masks document field", func(t *testing.T) {
		input := map[string]any{"document": "12345678901"}
		result := MaskSensitivePayload(input)

		m, ok := result.(map[string]any)
		assert.True(t, ok)
		assert.Equal(t, "***8901", m["document"])
	})

	t.Run("masks name field", func(t *testing.T) {
		input := map[string]any{"name": "Joao Silva"}
		result := MaskSensitivePayload(input)

		m, ok := result.(map[string]any)
		assert.True(t, ok)
		assert.Equal(t, "J*** S***", m["name"])
	})

	t.Run("masks phone field", func(t *testing.T) {
		input := map[string]any{"phone": "+5511999998888"}
		result := MaskSensitivePayload(input)

		m, ok := result.(map[string]any)
		assert.True(t, ok)
		assert.Equal(t, "+55***8888", m["phone"])
	})

	t.Run("masks nested maps", func(t *testing.T) {
		input := map[string]any{
			"user": map[string]any{
				"email":    "user@example.com",
				"document": "12345678901",
			},
		}
		result := MaskSensitivePayload(input)

		m, ok := result.(map[string]any)
		assert.True(t, ok)
		nested, ok := m["user"].(map[string]any)
		assert.True(t, ok)
		assert.Equal(t, "u***@example.com", nested["email"])
		assert.Equal(t, "***8901", nested["document"])
	})

	t.Run("masks items in slices", func(t *testing.T) {
		input := map[string]any{
			"users": []any{
				map[string]any{"email": "a@b.com"},
				map[string]any{"email": "c@d.com"},
			},
		}
		result := MaskSensitivePayload(input)

		m, ok := result.(map[string]any)
		assert.True(t, ok)
		users, ok := m["users"].([]any)
		assert.True(t, ok)
		assert.Len(t, users, 2)

		first, ok := users[0].(map[string]any)
		assert.True(t, ok)
		assert.Equal(t, "a***@b.com", first["email"])
	})

	t.Run("skips empty string values", func(t *testing.T) {
		input := map[string]any{"email": "", "name": ""}
		result := MaskSensitivePayload(input)

		m, ok := result.(map[string]any)
		assert.True(t, ok)
		assert.Equal(t, "", m["email"])
		assert.Equal(t, "", m["name"])
	})
}
