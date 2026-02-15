package trx

import (
	"context"
	"fmt"
)

func Begin(ctx context.Context) context.Context {
	return newContextWithTx(ctx, newTrx())
}

func Exists(ctx context.Context) bool {
	return txFromContext(ctx) != nil
}

func Value(ctx context.Context, key any) (any, bool) {
	if tx := txFromContext(ctx); tx != nil {
		return tx.value(key)
	}
	return nil, false
}

func SetValue(ctx context.Context, key, value any) {
	if tx := txFromContext(ctx); tx != nil {
		tx.setValue(key, value)
	}
}

func AddRollback(ctx context.Context, rollbackFunc func() error) {
	if tx := txFromContext(ctx); tx != nil {
		if rollbackFunc != nil {
			tx.addRollback(rollbackFunc)
		}
	}
}

func AddCommit(ctx context.Context, commitFunc func() error) {
	if tx := txFromContext(ctx); tx != nil {
		if commitFunc != nil {
			tx.addCommit(commitFunc)
		}
	}
}

func Rollback(ctx context.Context) error {
	if tx := txFromContext(ctx); tx != nil {
		return tx.rollback()
	}
	return nil
}

func Commit(ctx context.Context) error {
	if tx := txFromContext(ctx); tx != nil {
		return tx.commit()
	}
	return nil
}

func Do(ctx context.Context, fn func(ctx context.Context) error) (outErr error, rbErr error) {
	ctx = Begin(ctx)

	if outErr = fn(ctx); outErr != nil {
		rbErr = Rollback(ctx)
		return
	}

	if outErr = Commit(ctx); outErr != nil {
		outErr = fmt.Errorf("commit trx: %w", outErr)
	}

	return
}
