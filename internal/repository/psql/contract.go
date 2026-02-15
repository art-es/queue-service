//go:generate mockgen -source=contract.go -destination=psqlmock/contract.go -package=psqlmock
package psql

import (
	"context"
)

type Conn interface {
	Executer
	Close() error
	beginTx(ctx context.Context) (tx, error)
}

type tx interface {
	Executer
	rollback() error
	commit() error
}

type ExecGetter interface {
	Get(ctx context.Context) (Executer, error)
}

type Executer interface {
	Exec(ctx context.Context, query string, args ...any) (Result, error)
	Query(ctx context.Context, query string, args ...any) (Rows, error)
	QueryRow(ctx context.Context, query string, args ...any) Row
}

type Result interface {
	LastInsertId() (int64, error)
	RowsAffected() (int64, error)
}

type Rows interface {
	Next() bool
	Err() error
	Close() error
	Scan(...interface{}) error
}

type Row interface {
	Scan(...any) error
	Err() error
}
