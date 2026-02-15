package trx

import "context"

type contextKeyTx struct{}

func txFromContext(ctx context.Context) *trx {
	if v, ok := ctx.Value(contextKeyTx{}).(*trx); ok {
		return v
	}
	return nil
}

func newContextWithTx(ctx context.Context, v *trx) context.Context {
	return context.WithValue(ctx, contextKeyTx{}, v)
}
