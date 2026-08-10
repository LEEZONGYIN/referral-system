package repository

import (
	"context"

	"referral-system/internal/model"
)

type CreditLedgerRepository interface {
	Create(ctx context.Context, ledger *model.CreditLedger) error
	GetByID(ctx context.Context, id int64) (*model.CreditLedger, error)
	GetByIdempotencyKey(ctx context.Context, key string) (*model.CreditLedger, error)
	GetByBiz(ctx context.Context, bizType, bizID string) (*model.CreditLedger, error)
	ListByUserID(ctx context.Context, userID int64, limit, offset int) ([]*model.CreditLedger, error)
}
