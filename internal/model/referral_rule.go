package model

import "time"

type ReferralRule struct {
	ID            int64      `db:"id" json:"id"`
	RuleCode      string     `db:"rule_code" json:"rule_code"`
	RewardAmount  int64      `db:"reward_amount" json:"reward_amount"`
	TriggerEvent  string     `db:"trigger_event" json:"trigger_event"`
	Status        int8       `db:"status" json:"status"`
	EffectiveFrom *time.Time `db:"effective_from" json:"effective_from,omitempty"`
	EffectiveTo   *time.Time `db:"effective_to" json:"effective_to,omitempty"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updated_at"`
}
