package repository

import (
	"context"

	"referral-system/internal/model"
)

type ReferralEventRepository interface {
	Create(ctx context.Context, event *model.ReferralEvent) error
	GetByID(ctx context.Context, id int64) (*model.ReferralEvent, error)
	GetByIdempotencyKey(ctx context.Context, key string) (*model.ReferralEvent, error)
	ListByRelationID(ctx context.Context, relationID int64, limit, offset int) ([]*model.ReferralEvent, error)
	ListByInviteeUserID(ctx context.Context, inviteeUserID int64, limit, offset int) ([]*model.ReferralEvent, error)
}
