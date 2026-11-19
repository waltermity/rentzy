// Package postgres implements the repository interfaces on PostgreSQL (pgx).
package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"rentzy/backend/internal/domain"
)

// querier is what both *pgxpool.Pool and pgx.Tx provide.
type querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type txKey struct{}

// DB is shared by all repositories. It also implements repository.TxManager:
// a transaction started by WithinTx travels in the context, and every
// repository call made with that context runs inside it.
type DB struct {
	pool *pgxpool.Pool
}

func NewDB(pool *pgxpool.Pool) *DB { return &DB{pool: pool} }

func (d *DB) q(ctx context.Context) querier {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}
	return d.pool
}

func (d *DB) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if _, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return fn(ctx) // already in a transaction
	}
	tx, err := d.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("postgres: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }() // no-op after commit

	if err := fn(context.WithValue(ctx, txKey{}, tx)); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("postgres: commit tx: %w", err)
	}
	return nil
}

// mapErr turns pgx/PostgreSQL errors into domain errors.
func mapErr(err error, what string) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%s: %w", what, domain.ErrNotFound)
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			return fmt.Errorf("%s: %w (%s)", what, domain.ErrConflict, pgErr.ConstraintName)
		case "23503": // foreign_key_violation
			return fmt.Errorf("%s: %w: referenced record does not exist (%s)", what, domain.ErrInvalidInput, pgErr.ConstraintName)
		case "23514", "23502": // check_violation, not_null_violation
			return fmt.Errorf("%s: %w: constraint %s violated", what, domain.ErrInvalidInput, pgErr.ConstraintName)
		}
	}
	return fmt.Errorf("%s: %w", what, err)
}

// expectOne returns ErrNotFound when an UPDATE matched no row.
func expectOne(tag pgconn.CommandTag, what string) error {
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w", what, domain.ErrNotFound)
	}
	return nil
}

// nullIfEmpty stores "" as NULL for optional text columns.
func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// collect scans every row with scan.
func collect[T any](rows pgx.Rows, scan func(pgx.Rows) (T, error)) ([]T, error) {
	defer rows.Close()
	out := []T{}
	for rows.Next() {
		v, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
