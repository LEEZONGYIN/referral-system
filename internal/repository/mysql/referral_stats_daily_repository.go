package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"referral-system/internal/model"
)

type ReferralStatsDailyRepository struct {
	db *sql.DB
}

func NewReferralStatsDailyRepository(db *sql.DB) *ReferralStatsDailyRepository {
	return &ReferralStatsDailyRepository{db: db}
}

func (r *ReferralStatsDailyRepository) Upsert(ctx context.Context, stat *model.ReferralStatsDaily) error {
	if stat == nil {
		return errors.New("stat is nil")
	}

	const query = `
INSERT INTO referral_stats_daily (
    stat_date, user_id, invite_count, qualified_count, reward_amount_total, created_at, updated_at
)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    invite_count = VALUES(invite_count),
    qualified_count = VALUES(qualified_count),
    reward_amount_total = VALUES(reward_amount_total),
    updated_at = VALUES(updated_at)`

	if _, err := connFromContext(ctx, r.db).ExecContext(ctx, query,
		stat.StatDate,
		stat.UserID,
		stat.InviteCount,
		stat.QualifiedCount,
		stat.RewardAmountTotal,
		stat.CreatedAt,
		stat.UpdatedAt,
	); err != nil {
		return fmt.Errorf("upsert referral stats daily: %w", err)
	}
	return nil
}

func (r *ReferralStatsDailyRepository) GetByUserAndDate(ctx context.Context, userID int64, statDate time.Time) (*model.ReferralStatsDaily, error) {
	const query = `
SELECT id, stat_date, user_id, invite_count, qualified_count, reward_amount_total, created_at, updated_at
FROM referral_stats_daily
WHERE user_id = ? AND stat_date = ?`
	row := connFromContext(ctx, r.db).QueryRowContext(ctx, query, userID, statDate)

	var item model.ReferralStatsDaily
	if err := row.Scan(
		&item.ID,
		&item.StatDate,
		&item.UserID,
		&item.InviteCount,
		&item.QualifiedCount,
		&item.RewardAmountTotal,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *ReferralStatsDailyRepository) ListByUser(ctx context.Context, userID int64, limit, offset int) ([]*model.ReferralStatsDaily, error) {
	const query = `
SELECT id, stat_date, user_id, invite_count, qualified_count, reward_amount_total, created_at, updated_at
FROM referral_stats_daily
WHERE user_id = ?
ORDER BY stat_date DESC
LIMIT ? OFFSET ?`

	rows, err := connFromContext(ctx, r.db).QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list referral stats daily: %w", err)
	}
	defer rows.Close()

	result := make([]*model.ReferralStatsDaily, 0)
	for rows.Next() {
		item := new(model.ReferralStatsDaily)
		if err := rows.Scan(
			&item.ID,
			&item.StatDate,
			&item.UserID,
			&item.InviteCount,
			&item.QualifiedCount,
			&item.RewardAmountTotal,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan referral stats daily: %w", err)
		}
		result = append(result, item)
	}
	return result, nil
}
