package trxutil

import (
	"context"
	"fmt"

	"github.com/art-es/queue-service/internal/infra/log"
	"github.com/art-es/queue-service/internal/infra/ops"
	"github.com/art-es/queue-service/internal/infra/trx"
)

func Do(ctx context.Context, fn func(ctx context.Context) error) (outErr error, rbErr error) {
	ctx = trx.Begin(ctx)

	if outErr = fn(ctx); outErr != nil {
		rbErr = trx.Rollback(ctx)
		return
	}

	if outErr = trx.Commit(ctx); outErr != nil {
		outErr = fmt.Errorf("commit trx: %w", outErr)
	}

	return
}

func LogError(logger log.Logger, msg string, rbErr, opErr error) {
	if rbErr != nil {
		logger.Log(log.LevelError).
			With("message", "rollback error on "+msg).
			With("rb_error", rbErr.Error()).
			With("op_error", ops.ErrorMessage(opErr)).
			Write()
	}
}

func DoOrLogError(logger log.Logger, logMsg string, ctx context.Context, fn func(ctx context.Context) error) error {
	opErr, rbErr := Do(ctx, fn)
	LogError(logger, logMsg, rbErr, opErr)
	return opErr
}
