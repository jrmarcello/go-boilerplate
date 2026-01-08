package vo

import (
	"database/sql/driver"
	"fmt"
	"regexp"
	"strconv"
)

type CPF struct {
	value string
}

func NewCPF(value string) (CPF, error) {
	cleaned := cleanCPF(value)

	if len(cleaned) != 11 {
		return CPF{}, ErrInvalidCPF
	}

	if allDigitsEqual(cleaned) {
		return CPF{}, ErrInvalidCPF
	}

	if !validateCPFDigits(cleaned) {
		return CPF{}, ErrInvalidCPF
	}

	return CPF{value: cleaned}, nil
}

func ParseCPF(value string) CPF {
	return CPF{value: value}
}

func (c CPF) String() string { return c.value }

func (c CPF) Value() (driver.Value, error) {
	return c.value, nil
}

func (c *CPF) Scan(value interface{}) error {
	if value == nil {
		return fmt.Errorf("CPF não pode ser nulo")
	}
	switch v := value.(type) {
	case string:
		c.value = v
	case []byte:
		c.value = string(v)
	default:
		return fmt.Errorf("tipo inválido para CPF: %T", value)
	}
	return nil
}

func cleanCPF(cpf string) string {
	re := regexp.MustCompile(`\D`)
	return re.ReplaceAllString(cpf, "")
}

func allDigitsEqual(cpf string) bool {
	first := cpf[0]
	for i := 1; i < len(cpf); i++ {
		if cpf[i] != first {
			return false
		}
	}
	return true
}

// validateCPFDigits valida os dígitos verificadores usando o algoritmo MOD 11.
//
// ALGORITMO DO CPF (simplificado):
//
// Para o CPF 529.982.247-25:
//
//  1. Pega os 9 primeiros dígitos: 529982247
//
//  2. Multiplica cada um por pesos decrescentes (10 a 2):
//     5×10 + 2×9 + 9×8 + 9×7 + 8×6 + 2×5 + 2×4 + 4×3 + 7×2 = 295
//
//  3. Calcula: 11 - (295 % 11) = 11 - 9 = 2 ← primeiro dígito verificador
//
//  4. Repete incluindo o primeiro dígito (pesos 11 a 2) para obter o segundo
//
// Se os dígitos calculados conferem com os informados, o CPF é válido.
func validateCPFDigits(cpf string) bool {
	// --- Calcula o PRIMEIRO dígito verificador ---
	sum := 0
	for i := 0; i < 9; i++ {
		digit, _ := strconv.Atoi(string(cpf[i]))
		sum += digit * (10 - i) // Pesos: 10, 9, 8, 7, 6, 5, 4, 3, 2
	}
	remainder := sum % 11
	firstDigit := 0
	if remainder >= 2 {
		firstDigit = 11 - remainder
	}

	// Verifica se o primeiro dígito calculado confere
	actualFirst, _ := strconv.Atoi(string(cpf[9]))
	if actualFirst != firstDigit {
		return false
	}

	// --- Calcula o SEGUNDO dígito verificador ---
	sum = 0
	for i := 0; i < 10; i++ {
		digit, _ := strconv.Atoi(string(cpf[i]))
		sum += digit * (11 - i) // Pesos: 11, 10, 9, 8, 7, 6, 5, 4, 3, 2
	}
	remainder = sum % 11
	secondDigit := 0
	if remainder >= 2 {
		secondDigit = 11 - remainder
	}

	// Verifica se o segundo dígito calculado confere
	actualSecond, _ := strconv.Atoi(string(cpf[10]))
	return actualSecond == secondDigit
}
