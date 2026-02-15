package psql

import (
	"context"
	"fmt"

	"github.com/art-es/queue-service/internal/infra/trx"
)

type contextKeyTx struct{}

type execGetter struct {
	conn Conn
}

func NewExecGetter(conn Conn) ExecGetter {
	return &execGetter{conn: conn}
}

func (g *execGetter) Get(ctx context.Context) (Executer, error) {
	if trx.Exists(ctx) {
		tx, err := g.getTx(ctx)
		if err != nil {
			return nil, err
		}

		return tx, nil
	}

	return g.conn, nil
}

func (g *execGetter) getTx(ctx context.Context) (tx, error) {
	if val, ok := trx.Value(ctx, contextKeyTx{}); ok {
		if tx, ok := val.(tx); ok {
			return tx, nil
		}
	}

	tx, err := g.conn.beginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("psql tx begin: %w", err)
	}

	trx.SetValue(ctx, contextKeyTx{}, tx)
	trx.AddRollback(ctx, tx.rollback)
	trx.AddCommit(ctx, tx.commit)

	return tx, nil
}
