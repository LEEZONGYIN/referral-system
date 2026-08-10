package service

import (
	"context"
	"database/sql"

	"referral-system/internal/txctx"
)

type SQLTxManager struct {
	db *sql.DB
}

func NewSQLTxManager(db *sql.DB) *SQLTxManager {
	return &SQLTxManager{db: db}
}

func (m *SQLTxManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	txCtx := txctx.WithTx(ctx, tx)
	if err := fn(txCtx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
