package postgres

import (
	"context"
	"database/sql"
)

type txKey struct{}

type querier interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func withTx(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

func getTx(ctx context.Context) *sql.Tx {
	tx, _ := ctx.Value(txKey{}).(*sql.Tx)
	return tx
}

func (r *Repository) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if getTx(ctx) != nil {
		return fn(ctx)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	txCtx := withTx(ctx, tx)
	err = fn(txCtx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

type Transactor struct {
	db *sql.DB
}
