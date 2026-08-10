package txctx

import (
	"context"
	"database/sql"
)

type key struct{}

func WithTx(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, key{}, tx)
}

func FromContext(ctx context.Context) (*sql.Tx, bool) {
	tx, ok := ctx.Value(key{}).(*sql.Tx)
	return tx, ok
}
