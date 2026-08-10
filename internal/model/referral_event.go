package model

import (
	"encoding/json"
	"time"
)

type ReferralEvent struct {
	ID             int64           `db:"id" json:"id"`
	RelationID     *int64          `db:"relation_id" json:"relation_id,omitempty"`
	InviterUserID  *int64          `db:"inviter_user_id" json:"inviter_user_id,omitempty"`
	InviteeUserID  *int64          `db:"invitee_user_id" json:"invitee_user_id,omitempty"`
	EventType      string          `db:"event_type" json:"event_type"`
	IdempotencyKey string          `db:"idempotency_key" json:"idempotency_key"`
	Payload        json.RawMessage `db:"payload" json:"payload,omitempty"`
	CreatedAt      time.Time       `db:"created_at" json:"created_at"`
}
