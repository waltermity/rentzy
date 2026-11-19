// Package repository defines the storage interfaces the services depend on.
// The postgres subpackage implements them; services never import it directly.
package repository

import "context"

// TxManager runs fn in one database transaction. Repository calls made with the
// ctx passed to fn join that transaction. Returning an error rolls it back.
// Nested calls reuse the outer transaction.
type TxManager interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}
