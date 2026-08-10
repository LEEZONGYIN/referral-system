package model

import "time"

type ReferralRelation struct {
	ID            int64      `db:"id" json:"id"`
	InviterUserID int64      `db:"inviter_user_id" json:"inviter_user_id"`
	InviteeUserID int64      `db:"invitee_user_id" json:"invitee_user_id"`
	ReferralCode  string     `db:"referral_code" json:"referral_code"`
	RuleID        *int64     `db:"rule_id" json:"rule_id,omitempty"`
	Status        int8       `db:"status" json:"status"`
	QualifiedAt   *time.Time `db:"qualified_at" json:"qualified_at,omitempty"`
	RewardedAt    *time.Time `db:"rewarded_at" json:"rewarded_at,omitempty"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updated_at"`
}
