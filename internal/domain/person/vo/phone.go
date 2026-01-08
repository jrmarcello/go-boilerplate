package vo

import (
	"database/sql/driver"
	"fmt"
	"regexp"
)

type Phone struct {
	value string
}

func NewPhone(value string) (Phone, error) {
	cleaned := cleanPhone(value)

	// Valida: 10-11 dígitos (fixo ou celular)
	if len(cleaned) < 10 || len(cleaned) > 11 {
		return Phone{}, ErrInvalidPhone
	}

	return Phone{value: cleaned}, nil
}

func ParsePhone(value string) Phone {
	return Phone{value: value}
}

func (p Phone) String() string { return p.value }

func (p Phone) IsEmpty() bool { return p.value == "" }

func (p Phone) Value() (driver.Value, error) {
	if p.value == "" {
		return nil, nil // Retorna NULL se vazio
	}
	return p.value, nil
}

func (p *Phone) Scan(value interface{}) error {
	if value == nil {
		p.value = ""
		return nil
	}
	switch v := value.(type) {
	case string:
		p.value = v
	case []byte:
		p.value = string(v)
	default:
		return fmt.Errorf("tipo inválido para Phone: %T", value) // TODO lançar erro de domínio?
	}
	return nil
}

func cleanPhone(phone string) string {
	re := regexp.MustCompile(`\D`)
	cleaned := re.ReplaceAllString(phone, "")

	// Remove código do país se presente (+55)
	if len(cleaned) == 13 && cleaned[:2] == "55" {
		cleaned = cleaned[2:]
	}

	return cleaned
}
