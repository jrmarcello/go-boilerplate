package interfaces

import (
	"context"
)

type Cache interface {
	// Get recupera um valor do cache e deserializa em dest.
	// Retorna ErrCacheMiss se a chave não existir.
	Get(ctx context.Context, key string, dest interface{}) error

	// Set armazena um valor no cache com TTL configurado.
	Set(ctx context.Context, key string, value interface{}) error

	// Delete remove uma chave do cache.
	Delete(ctx context.Context, key string) error
}
