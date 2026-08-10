package repository

import (
	"context"
	"time"

	"referral-system/internal/model"
)

type ReferralStatsDailyRepository interface {
	Upsert(ctx context.Context, stat *model.ReferralStatsDaily) error
	GetByUserAndDate(ctx context.Context, userID int64, statDate time.Time) (*model.ReferralStatsDaily, error)
	ListByUser(ctx context.Context, userID int64, limit, offset int) ([]*model.ReferralStatsDaily, error)
}
