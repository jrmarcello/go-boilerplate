package vo

import (
	"crypto/rand"
	"database/sql/driver"
	"errors"

	"github.com/oklog/ulid/v2"
)

// ID é um Value Object que encapsula um ULID (Universally Unique Lexicographically Sortable Identifier).
//
// Por que ULID em vez de UUID?
//   - Ordenável: ULIDs são ordenados cronologicamente (ótimo para índices de banco)
//   - Compacto: 26 caracteres vs 36 do UUID
//   - URL-safe: não contém caracteres especiais
//
// Exemplo: 01ARZ3NDEKTSV4RRFFQ69G5FAV
type ID string

func NewID() ID {
	// Usa crypto/rand para geração segura de entropia (G404 fix)
	return ID(ulid.MustNew(ulid.Now(), rand.Reader).String())
}

// ParseID converte uma string em ID, validando o formato ULID.
// Use quando receber um ID de fonte externa (API, banco, etc).
func ParseID(s string) (ID, error) {
	if _, err := ulid.Parse(s); err != nil {
		return "", errors.New("invalid ULID") // TODO lançar erro de domínio?
	}
	return ID(s), nil
}

// String retorna a representação em string do ID.
// Implementa fmt.Stringer para uso em logs, prints, etc.
func (i ID) String() string { return string(i) }

// Value implementa driver.Valuer para o pacote database/sql.
// É chamado automaticamente quando o ID é usado em INSERT/UPDATE.
//
// Isso permite usar o ID diretamente em queries:
//
//	db.Exec("INSERT INTO users (id) VALUES ($1)", user.ID)
func (i ID) Value() (driver.Value, error) {
	return string(i), nil
}

// Scan implementa sql.Scanner para o pacote database/sql.
// É chamado automaticamente quando o ID é lido do banco (SELECT).
//
// Isso permite escanear diretamente para o tipo ID:
//
//	var id vo.ID
//	db.QueryRow("SELECT id FROM users WHERE ...").Scan(&id)
func (i *ID) Scan(value interface{}) error {
	if value == nil {
		return errors.New("ID não pode ser nulo")
	}
	// O banco pode retornar string ou []byte dependendo do driver
	switch v := value.(type) {
	case string:
		*i = ID(v)
	case []byte:
		*i = ID(string(v))
	default:
		return errors.New("tipo inválido para ID")
	}
	return nil
}
