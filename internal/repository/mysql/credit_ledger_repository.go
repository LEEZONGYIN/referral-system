package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"referral-system/internal/model"
)

type CreditLedgerRepository struct {
	db *sql.DB
}

func NewCreditLedgerRepository(db *sql.DB) *CreditLedgerRepository {
	return &CreditLedgerRepository{db: db}
}

func (r *CreditLedgerRepository) Create(ctx context.Context, ledger *model.CreditLedger) error {
	if ledger == nil {
		return errors.New("ledger is nil")
	}

	const query = `
INSERT INTO credit_ledger (
    user_id, biz_type, biz_id, direction, amount, before_balance, after_balance, idempotency_key, created_at
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	res, err := connFromContext(ctx, r.db).ExecContext(ctx, query,
		ledger.UserID,
		ledger.BizType,
		ledger.BizID,
		ledger.Direction,
		ledger.Amount,
		ledger.BeforeBalance,
		ledger.AfterBalance,
		ledger.IdempotencyKey,
		ledger.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert credit ledger: %w", err)
	}

	id, err := res.LastInsertId()
	if err == nil {
		ledger.ID = id
	}
	return nil
}

func (r *CreditLedgerRepository) GetByID(ctx context.Context, id int64) (*model.CreditLedger, error) {
	const query = `
SELECT id, user_id, biz_type, biz_id, direction, amount, before_balance, after_balance, idempotency_key, created_at
FROM credit_ledger
WHERE id = ?`
	row := connFromContext(ctx, r.db).QueryRowContext(ctx, query, id)
	return scanCreditLedger(row)
}

func (r *CreditLedgerRepository) GetByIdempotencyKey(ctx context.Context, key string) (*model.CreditLedger, error) {
	const query = `
SELECT id, user_id, biz_type, biz_id, direction, amount, before_balance, after_balance, idempotency_key, created_at
FROM credit_ledger
WHERE idempotency_key = ?`
	row := connFromContext(ctx, r.db).QueryRowContext(ctx, query, key)
	return scanCreditLedger(row)
}

func (r *CreditLedgerRepository) GetByBiz(ctx context.Context, bizType, bizID string) (*model.CreditLedger, error) {
	const query = `
SELECT id, user_id, biz_type, biz_id, direction, amount, before_balance, after_balance, idempotency_key, created_at
FROM credit_ledger
WHERE biz_type = ? AND biz_id = ?`
	row := connFromContext(ctx, r.db).QueryRowContext(ctx, query, bizType, bizID)
	return scanCreditLedger(row)
}

func (r *CreditLedgerRepository) ListByUserID(ctx context.Context, userID int64, limit, offset int) ([]*model.CreditLedger, error) {
	const query = `
SELECT id, user_id, biz_type, biz_id, direction, amount, before_balance, after_balance, idempotency_key, created_at
FROM credit_ledger
WHERE user_id = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?`
	rows, err := connFromContext(ctx, r.db).QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list credit ledger: %w", err)
	}
	defer rows.Close()

	result := make([]*model.CreditLedger, 0)
	for rows.Next() {
		item := new(model.CreditLedger)
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.BizType,
			&item.BizID,
			&item.Direction,
			&item.Amount,
			&item.BeforeBalance,
			&item.AfterBalance,
			&item.IdempotencyKey,
			&item.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan credit ledger: %w", err)
		}
		result = append(result, item)
	}
	return result, nil
}

func scanCreditLedger(row *sql.Row) (*model.CreditLedger, error) {
	var item model.CreditLedger
	if err := row.Scan(
		&item.ID,
		&item.UserID,
		&item.BizType,
		&item.BizID,
		&item.Direction,
		&item.Amount,
		&item.BeforeBalance,
		&item.AfterBalance,
		&item.IdempotencyKey,
		&item.CreatedAt,
	); err != nil {
		return nil, err
	}
	return &item, nil
}
