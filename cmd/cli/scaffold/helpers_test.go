package scaffold

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToPlural(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty string", input: "", want: ""},
		{name: "simple word", input: "user", want: "users"},
		{name: "word ending in y with consonant", input: "entity", want: "entities"},
		{name: "word ending in y with vowel", input: "key", want: "keys"},
		{name: "word ending in s", input: "status", want: "statuses"},
		{name: "word ending in x", input: "box", want: "boxes"},
		{name: "word ending in z", input: "quiz", want: "quizzes"},
		{name: "word ending in ch", input: "match", want: "matches"},
		{name: "word ending in sh", input: "crash", want: "crashes"},
		{name: "word ending in ss", input: "address", want: "addresses"},
		{name: "already plural - users", input: "users", want: "users"},
		{name: "already plural - entities", input: "entities", want: "entities"},
		{name: "already plural - statuses", input: "statuses", want: "statuses"},
		{name: "already plural - boxes", input: "boxes", want: "boxes"},
		{name: "already plural - matches", input: "matches", want: "matches"},
		{name: "already plural - crashes", input: "crashes", want: "crashes"},
		{name: "already plural - addresses", input: "addresses", want: "addresses"},
		{name: "simple word - order", input: "order", want: "orders"},
		{name: "simple word - role", input: "role", want: "roles"},
		{name: "word ending in y - policy", input: "policy", want: "policies"},
		{name: "word ending in y - category", input: "category", want: "categories"},
		{name: "word ending in day", input: "day", want: "days"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToPlural(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestToSingular(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty string", input: "", want: ""},
		{name: "simple plural", input: "users", want: "user"},
		{name: "ies plural", input: "entities", want: "entity"},
		{name: "es plural - status", input: "statuses", want: "status"},
		{name: "es plural - box", input: "boxes", want: "box"},
		{name: "es plural - match", input: "matches", want: "match"},
		{name: "es plural - crash", input: "crashes", want: "crash"},
		{name: "es plural - address", input: "addresses", want: "address"},
		{name: "already singular - user", input: "user", want: "user"},
		{name: "already singular - status", input: "status", want: "status"},
		{name: "simple plural - orders", input: "orders", want: "order"},
		{name: "simple plural - roles", input: "roles", want: "role"},
		{name: "ies plural - policies", input: "policies", want: "policy"},
		{name: "ies plural - categories", input: "categories", want: "category"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToSingular(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty string", input: "", want: ""},
		{name: "single word lowercase", input: "user", want: "User"},
		{name: "single word uppercase", input: "USER", want: "User"},
		{name: "snake_case", input: "order_item", want: "OrderItem"},
		{name: "kebab-case", input: "order-item", want: "OrderItem"},
		{name: "already PascalCase", input: "OrderItem", want: "OrderItem"},
		{name: "camelCase", input: "orderItem", want: "OrderItem"},
		{name: "single char", input: "a", want: "A"},
		{name: "triple snake", input: "foo_bar_baz", want: "FooBarBaz"},
		{name: "triple kebab", input: "foo-bar-baz", want: "FooBarBaz"},
		{name: "mixed separators", input: "foo_bar-baz", want: "FooBarBaz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToPascalCase(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty string", input: "", want: ""},
		{name: "single word", input: "user", want: "user"},
		{name: "snake_case", input: "order_item", want: "orderItem"},
		{name: "kebab-case", input: "order-item", want: "orderItem"},
		{name: "PascalCase", input: "OrderItem", want: "orderItem"},
		{name: "already camelCase", input: "orderItem", want: "orderItem"},
		{name: "single char", input: "A", want: "a"},
		{name: "triple snake", input: "foo_bar_baz", want: "fooBarBaz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToCamelCase(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty string", input: "", want: ""},
		{name: "single word", input: "user", want: "user"},
		{name: "PascalCase", input: "OrderItem", want: "order_item"},
		{name: "camelCase", input: "orderItem", want: "order_item"},
		{name: "kebab-case", input: "order-item", want: "order_item"},
		{name: "already snake_case", input: "order_item", want: "order_item"},
		{name: "single uppercase", input: "User", want: "user"},
		{name: "all uppercase", input: "URL", want: "url"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToSnakeCase(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestToKebabCase(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty string", input: "", want: ""},
		{name: "single word", input: "user", want: "user"},
		{name: "PascalCase", input: "OrderItem", want: "order-item"},
		{name: "camelCase", input: "orderItem", want: "order-item"},
		{name: "snake_case", input: "order_item", want: "order-item"},
		{name: "already kebab-case", input: "order-item", want: "order-item"},
		{name: "single uppercase", input: "User", want: "user"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToKebabCase(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestToLower(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty string", input: "", want: ""},
		{name: "already lowercase", input: "user", want: "user"},
		{name: "uppercase", input: "USER", want: "user"},
		{name: "mixed case", input: "OrderItem", want: "orderitem"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToLower(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTemplateFuncs(t *testing.T) {
	funcs := TemplateFuncs()

	expectedKeys := []string{"plural", "singular", "pascalCase", "camelCase", "snakeCase", "kebabCase", "lower"}
	for _, key := range expectedKeys {
		assert.NotNilf(t, funcs[key], "TemplateFuncs should contain %q", key)
	}
}
