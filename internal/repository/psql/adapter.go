package psql

import (
	"context"
	"database/sql"
)

var (
	_ Conn = (*connAdapter)(nil)
	_ tx   = (*txAdapter)(nil)
)

type connAdapter struct {
	db *sql.DB
}

type txAdapter struct {
	tx *sql.Tx
}

func newConnAdapter(db *sql.DB) *connAdapter {
	return &connAdapter{db: db}
}

func newTxAdapter(tx *sql.Tx) *txAdapter {
	return &txAdapter{tx: tx}
}

func (c *connAdapter) Exec(ctx context.Context, query string, args ...any) (Result, error) {
	return c.db.ExecContext(ctx, query, args...)
}

func (c *connAdapter) Query(ctx context.Context, query string, args ...any) (Rows, error) {
	return c.db.QueryContext(ctx, query, args...)
}

func (c *connAdapter) QueryRow(ctx context.Context, query string, args ...any) Row {
	return c.db.QueryRowContext(ctx, query, args...)
}

func (c *connAdapter) Close() error {
	return c.db.Close()
}

func (c *connAdapter) beginTx(ctx context.Context) (tx, error) {
	txObj, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	return newTxAdapter(txObj), nil
}

func (a *txAdapter) Exec(ctx context.Context, query string, args ...any) (Result, error) {
	return a.tx.ExecContext(ctx, query, args...)
}

func (a *txAdapter) Query(ctx context.Context, query string, args ...any) (Rows, error) {
	return a.tx.QueryContext(ctx, query, args...)
}

func (a *txAdapter) QueryRow(ctx context.Context, query string, args ...any) Row {
	return a.tx.QueryRowContext(ctx, query, args...)
}

func (a *txAdapter) rollback() error {
	return a.tx.Rollback()
}

func (a *txAdapter) commit() error {
	return a.tx.Commit()
}
