package model

import "time"

type ReferralStatsDaily struct {
	ID                int64     `db:"id" json:"id"`
	StatDate          time.Time `db:"stat_date" json:"stat_date"`
	UserID            int64     `db:"user_id" json:"user_id"`
	InviteCount       int       `db:"invite_count" json:"invite_count"`
	QualifiedCount    int       `db:"qualified_count" json:"qualified_count"`
	RewardAmountTotal int64     `db:"reward_amount_total" json:"reward_amount_total"`
	CreatedAt         time.Time `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time `db:"updated_at" json:"updated_at"`
}
