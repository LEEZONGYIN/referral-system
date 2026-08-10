package repository

import (
	"context"
	"time"

	"referral-system/internal/model"
)

type ReferralRelationRepository interface {
	Create(ctx context.Context, relation *model.ReferralRelation) error
	GetByID(ctx context.Context, id int64) (*model.ReferralRelation, error)
	GetByInviteeUserID(ctx context.Context, inviteeUserID int64) (*model.ReferralRelation, error)
	GetByInviterUserID(ctx context.Context, inviterUserID int64, limit, offset int) ([]*model.ReferralRelation, error)
	UpdateStatus(ctx context.Context, relationID int64, status int8, qualifiedAt, rewardedAt *time.Time) error
	CountByInviterUserID(ctx context.Context, inviterUserID int64) (int64, error)
}
