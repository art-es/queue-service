package trx

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTxMethods(t *testing.T) {
	t.Run("tx does not exist", func(t *testing.T) {
		ctx := context.Background()
		assert.False(t, Exists(ctx))
	})

	t.Run("tx exists", func(t *testing.T) {
		ctx := context.Background()
		ctx = newContextWithTx(ctx, newTrx())
		assert.True(t, Exists(ctx))
	})

	t.Run("tx begin", func(t *testing.T) {
		ctx := context.Background()
		ctx = Begin(ctx)
		assert.True(t, Exists(ctx))
	})

	t.Run("tx rollback without errors", func(t *testing.T) {
		ctx := context.Background()
		ctx = Begin(ctx)

		var rollbackCount int
		AddRollback(ctx, func() error {
			rollbackCount++
			return nil
		})
		AddRollback(ctx, func() error {
			rollbackCount++
			return nil
		})

		err := Rollback(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 2, rollbackCount)
	})

	t.Run("tx rollback with errors", func(t *testing.T) {
		ctx := context.Background()
		ctx = Begin(ctx)

		var rollbackCount int
		AddRollback(ctx, func() error {
			rollbackCount++
			return errors.New("error from 1st rollback")
		})
		AddRollback(ctx, func() error {
			rollbackCount++
			return nil
		})
		AddRollback(ctx, func() error {
			rollbackCount++
			return errors.New("error from 3rd rollback")
		})

		err := Rollback(ctx)
		assert.EqualError(t, err, "error from 1st rollback\nerror from 3rd rollback")
		assert.Equal(t, 3, rollbackCount)
	})

	t.Run("tx commit without errors", func(t *testing.T) {
		ctx := context.Background()
		ctx = Begin(ctx)

		var commitCount int
		AddCommit(ctx, func() error {
			commitCount++
			return nil
		})
		AddCommit(ctx, func() error {
			commitCount++
			return nil
		})

		err := Commit(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 2, commitCount)
	})

	t.Run("tx commit with errors", func(t *testing.T) {
		ctx := context.Background()
		ctx = Begin(ctx)

		var commitCount int
		AddCommit(ctx, func() error {
			commitCount++
			return errors.New("error from 1st commit")
		})
		AddCommit(ctx, func() error {
			commitCount++
			return nil
		})
		AddCommit(ctx, func() error {
			commitCount++
			return errors.New("error from 3rd commit")
		})

		err := Commit(ctx)
		assert.EqualError(t, err, "error from 1st commit\nerror from 3rd commit")
		assert.Equal(t, 3, commitCount)
	})
}
