package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"referral-system/internal/model"
)

type CreditAccountRepository struct {
	db *sql.DB
}

func NewCreditAccountRepository(db *sql.DB) *CreditAccountRepository {
	return &CreditAccountRepository{db: db}
}

func (r *CreditAccountRepository) CreateIfNotExists(ctx context.Context, userID int64) error {
	const query = `
INSERT IGNORE INTO credit_accounts (user_id, balance, frozen_balance, version)
VALUES (?, 0, 0, 0)`
	if _, err := connFromContext(ctx, r.db).ExecContext(ctx, query, userID); err != nil {
		return fmt.Errorf("create credit account: %w", err)
	}
	return nil
}

func (r *CreditAccountRepository) GetByUserID(ctx context.Context, userID int64) (*model.CreditAccount, error) {
	const query = `
SELECT user_id, balance, frozen_balance, version
FROM credit_accounts
WHERE user_id = ?`
	row := connFromContext(ctx, r.db).QueryRowContext(ctx, query, userID)

	var account model.CreditAccount
	if err := row.Scan(&account.UserID, &account.Balance, &account.FrozenBalance, &account.Version); err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *CreditAccountRepository) UpdateBalanceWithOptimisticLock(ctx context.Context, userID int64, delta int64, expectedVersion int64) (bool, error) {
	const query = `
UPDATE credit_accounts
SET balance = balance + ?, version = version + 1
WHERE user_id = ? AND version = ?`

	res, err := connFromContext(ctx, r.db).ExecContext(ctx, query, delta, userID, expectedVersion)
	if err != nil {
		return false, fmt.Errorf("update credit account: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("rows affected: %w", err)
	}
	return affected > 0, nil
}

func (r *CreditAccountRepository) AddCredit(ctx context.Context, userID int64, amount int64) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	const query = `
UPDATE credit_accounts
SET balance = balance + ?, version = version + 1
WHERE user_id = ?`
	if _, err := connFromContext(ctx, r.db).ExecContext(ctx, query, amount, userID); err != nil {
		return fmt.Errorf("add credit: %w", err)
	}
	return nil
}
