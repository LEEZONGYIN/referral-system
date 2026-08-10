package mysql

import (
	"context"
	"database/sql"

	"referral-system/internal/txctx"
)

type dbTX interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

func connFromContext(ctx context.Context, db *sql.DB) dbTX {
	if tx, ok := txctx.FromContext(ctx); ok && tx != nil {
		return tx
	}
	return db
}
