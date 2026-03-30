package scaffold

import (
	"strings"
	"text/template"
	"unicode"
)

// ToPlural returns a simple English plural form of the given word.
// Handles common suffixes: "y" -> "ies", "s/x/z/ch/sh" -> "es".
func ToPlural(s string) string {
	if s == "" {
		return s
	}

	lower := strings.ToLower(s)

	// Already plural heuristic: if it ends in "ies" (from y->ies), "ses", "xes", "zes", "ches", "shes"
	if strings.HasSuffix(lower, "ies") && !strings.HasSuffix(lower, "series") {
		// Check if it's a word like "entities" that is already plural
		// by seeing if removing "ies" and adding "y" gives a valid singular
		candidate := lower[:len(lower)-3] + "y"
		if len(candidate) > 1 {
			return s // already plural
		}
	}
	if strings.HasSuffix(lower, "ses") || strings.HasSuffix(lower, "xes") ||
		strings.HasSuffix(lower, "zes") || strings.HasSuffix(lower, "ches") ||
		strings.HasSuffix(lower, "shes") {
		return s // already plural
	}
	if strings.HasSuffix(lower, "s") && !strings.HasSuffix(lower, "ss") &&
		!strings.HasSuffix(lower, "us") && !strings.HasSuffix(lower, "is") {
		return s // already plural (e.g., "users")
	}

	// Consonant + y -> ies
	if strings.HasSuffix(lower, "y") {
		if len(lower) >= 2 && !isVowel(rune(lower[len(lower)-2])) {
			return s[:len(s)-1] + "ies"
		}
		return s + "s"
	}

	// z -> double z + es (quiz -> quizzes)
	if strings.HasSuffix(lower, "z") && !strings.HasSuffix(lower, "zz") {
		return s + "zes"
	}

	// s, x, zz, ch, sh -> es
	if strings.HasSuffix(lower, "s") || strings.HasSuffix(lower, "x") ||
		strings.HasSuffix(lower, "zz") || strings.HasSuffix(lower, "ch") ||
		strings.HasSuffix(lower, "sh") {
		return s + "es"
	}

	return s + "s"
}

// ToSingular returns a simple English singular form of the given word.
// Reverses the pluralization rules from ToPlural.
func ToSingular(s string) string {
	if s == "" {
		return s
	}

	lower := strings.ToLower(s)

	// "ies" -> "y" (e.g., "entities" -> "entity")
	if strings.HasSuffix(lower, "ies") && len(lower) > 3 {
		return s[:len(s)-3] + "y"
	}

	// "sses" -> "ss" (e.g., "addresses" -> "address" needs special handling via "ses")
	if strings.HasSuffix(lower, "sses") {
		return s[:len(s)-2]
	}

	// "shes" -> "sh"
	if strings.HasSuffix(lower, "shes") {
		return s[:len(s)-2]
	}

	// "ches" -> "ch"
	if strings.HasSuffix(lower, "ches") {
		return s[:len(s)-2]
	}

	// "xes" -> "x"
	if strings.HasSuffix(lower, "xes") {
		return s[:len(s)-2]
	}

	// "zes" -> "z"
	if strings.HasSuffix(lower, "zes") {
		return s[:len(s)-2]
	}

	// "ses" -> "s" (e.g., "statuses" -> "status")
	if strings.HasSuffix(lower, "ses") {
		return s[:len(s)-2]
	}

	// Generic "s" removal (but not "ss", "us", "is")
	if strings.HasSuffix(lower, "s") && !strings.HasSuffix(lower, "ss") &&
		!strings.HasSuffix(lower, "us") && !strings.HasSuffix(lower, "is") {
		return s[:len(s)-1]
	}

	return s
}

// ToPascalCase converts a string to PascalCase.
// "order_item" -> "OrderItem", "order-item" -> "OrderItem", "order" -> "Order"
func ToPascalCase(s string) string {
	words := splitWords(s)
	var b strings.Builder
	for _, w := range words {
		if w == "" {
			continue
		}
		b.WriteString(capitalizeFirst(w))
	}
	return b.String()
}

// ToCamelCase converts a string to camelCase.
// "order_item" -> "orderItem", "OrderItem" -> "orderItem"
func ToCamelCase(s string) string {
	pascal := ToPascalCase(s)
	if pascal == "" {
		return ""
	}
	runes := []rune(pascal)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

// ToSnakeCase converts a string to snake_case.
// "OrderItem" -> "order_item", "orderItem" -> "order_item"
func ToSnakeCase(s string) string {
	words := splitWords(s)
	lowered := make([]string, 0, len(words))
	for _, w := range words {
		if w == "" {
			continue
		}
		lowered = append(lowered, strings.ToLower(w))
	}
	return strings.Join(lowered, "_")
}

// ToKebabCase converts a string to kebab-case.
// "OrderItem" -> "order-item", "order_item" -> "order-item"
func ToKebabCase(s string) string {
	words := splitWords(s)
	lowered := make([]string, 0, len(words))
	for _, w := range words {
		if w == "" {
			continue
		}
		lowered = append(lowered, strings.ToLower(w))
	}
	return strings.Join(lowered, "-")
}

// ToLower is a wrapper for strings.ToLower, exposed for template use.
func ToLower(s string) string {
	return strings.ToLower(s)
}

// TemplateFuncs returns all helper functions as a template.FuncMap.
func TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"plural":     ToPlural,
		"singular":   ToSingular,
		"pascalCase": ToPascalCase,
		"camelCase":  ToCamelCase,
		"snakeCase":  ToSnakeCase,
		"kebabCase":  ToKebabCase,
		"lower":      ToLower,
	}
}

// splitWords splits a string into words by camelCase, PascalCase, snake_case, or kebab-case boundaries.
func splitWords(s string) []string {
	// First, split on underscores and hyphens
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.ReplaceAll(s, "-", " ")

	// Then split on camelCase/PascalCase boundaries
	var words []string
	var current strings.Builder

	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == ' ' {
			if current.Len() > 0 {
				words = append(words, current.String())
				current.Reset()
			}
			continue
		}

		if unicode.IsUpper(r) && current.Len() > 0 {
			// Look at what came before
			prev := runes[i-1]
			if unicode.IsLower(prev) {
				// aB -> split: "a" "B..."
				words = append(words, current.String())
				current.Reset()
			} else if unicode.IsUpper(prev) && i+1 < len(runes) && unicode.IsLower(runes[i+1]) {
				// ABc -> split: "A" "Bc..."
				words = append(words, current.String())
				current.Reset()
			}
		}

		current.WriteRune(r)
	}
	if current.Len() > 0 {
		words = append(words, current.String())
	}

	return words
}

// capitalizeFirst capitalizes the first letter of a string and lowercases the rest.
func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	for i := 1; i < len(runes); i++ {
		runes[i] = unicode.ToLower(runes[i])
	}
	return string(runes)
}

// isVowel returns true if the rune is an English vowel.
func isVowel(r rune) bool {
	switch unicode.ToLower(r) {
	case 'a', 'e', 'i', 'o', 'u':
		return true
	}
	return false
}
