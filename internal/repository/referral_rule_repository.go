package repository

import (
	"context"
	"time"

	"referral-system/internal/model"
)

type ReferralRuleRepository interface {
	Create(ctx context.Context, rule *model.ReferralRule) error
	GetByID(ctx context.Context, id int64) (*model.ReferralRule, error)
	GetByRuleCode(ctx context.Context, ruleCode string) (*model.ReferralRule, error)
	ListActiveRules(ctx context.Context, now time.Time) ([]*model.ReferralRule, error)
}
