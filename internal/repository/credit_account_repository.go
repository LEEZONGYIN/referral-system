package repository

import (
	"context"

	"referral-system/internal/model"
)

type CreditAccountRepository interface {
	CreateIfNotExists(ctx context.Context, userID int64) error
	GetByUserID(ctx context.Context, userID int64) (*model.CreditAccount, error)
	UpdateBalanceWithOptimisticLock(ctx context.Context, userID int64, delta int64, expectedVersion int64) (bool, error)
	AddCredit(ctx context.Context, userID int64, amount int64) error
}
