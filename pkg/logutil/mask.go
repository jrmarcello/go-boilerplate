package logutil

import (
	"strings"
	"unicode/utf8"
)

// MaskEmail masks an email address, keeping the first character and domain visible.
// Example: "user@example.com" -> "u***@example.com"
func MaskEmail(email string) string {
	if email == "" {
		return "***"
	}
	atIdx := strings.LastIndex(email, "@")
	if atIdx <= 0 {
		return "***"
	}
	return string(email[0]) + "***" + email[atIdx:]
}

// MaskDocument masks a document number (CPF/CNPJ), keeping only the last 4 characters visible.
// Example: "12345678901" -> "***8901"
func MaskDocument(value string) string {
	if len(value) <= 4 {
		return "***"
	}
	return "***" + value[len(value)-4:]
}

// MaskName masks a full name, keeping only first initials visible.
// Example: "Joao Silva" -> "J*** S***"
func MaskName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "***"
	}

	parts := strings.Fields(name)
	masked := make([]string, len(parts))
	for i, part := range parts {
		r, _ := utf8.DecodeRuneInString(part)
		if r == utf8.RuneError || len(part) <= 1 {
			masked[i] = part
		} else {
			masked[i] = string(r) + "***"
		}
	}
	return strings.Join(masked, " ")
}

// MaskPhone masks a phone number, keeping the country code prefix and last 4 digits visible.
// Example: "+5511999998888" -> "+55***8888"
func MaskPhone(phone string) string {
	if phone == "" {
		return "***"
	}

	digits := make([]byte, 0, len(phone))
	prefix := ""

	for i, ch := range phone {
		if ch == '+' && i == 0 {
			prefix = "+"
			continue
		}
		if ch >= '0' && ch <= '9' {
			digits = append(digits, byte(ch))
		}
	}

	if len(digits) <= 4 {
		return "***"
	}

	return prefix + string(digits[:2]) + "***" + string(digits[len(digits)-4:])
}

// sensitiveKeys maps field names that should be masked to their masking function.
var sensitiveKeys = map[string]func(string) string{
	"email":        MaskEmail,
	"document":     MaskDocument,
	"cpf":          MaskDocument,
	"cnpj":         MaskDocument,
	"name":         MaskName,
	"full_name":    MaskName,
	"first_name":   MaskName,
	"last_name":    MaskName,
	"phone":        MaskPhone,
	"telefone":     MaskPhone,
	"company_name": MaskName,
	"trade_name":   MaskName,
}

// MaskSensitivePayload recursively masks known sensitive fields in a map.
// Recognized fields: email, document, cpf, cnpj, name, *_name, phone, telefone.
// Returns the input unchanged for non-map types, or a new map with masked values.
func MaskSensitivePayload(input any) any {
	if input == nil {
		return nil
	}

	switch v := input.(type) {
	case map[string]any:
		return maskMap(v)
	default:
		return input
	}
}

// maskMap creates a shallow copy of the map with sensitive string fields masked.
// Nested maps and slices are processed recursively.
func maskMap(m map[string]any) map[string]any {
	result := make(map[string]any, len(m))
	for k, v := range m {
		result[k] = maskValue(k, v)
	}
	return result
}

// maskValue applies masking to a single value based on its key.
func maskValue(key string, value any) any {
	switch v := value.(type) {
	case string:
		if maskFn, found := sensitiveKeys[key]; found && v != "" {
			return maskFn(v)
		}
		return v
	case map[string]any:
		return maskMap(v)
	case []any:
		result := make([]any, len(v))
		for i, item := range v {
			if nested, ok := item.(map[string]any); ok {
				result[i] = maskMap(nested)
			} else {
				result[i] = item
			}
		}
		return result
	default:
		return value
	}
}
