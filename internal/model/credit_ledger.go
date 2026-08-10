package model

import "time"

type CreditLedger struct {
	ID             int64     `db:"id" json:"id"`
	UserID         int64     `db:"user_id" json:"user_id"`
	BizType        string    `db:"biz_type" json:"biz_type"`
	BizID          string    `db:"biz_id" json:"biz_id"`
	Direction      int8      `db:"direction" json:"direction"`
	Amount         int64     `db:"amount" json:"amount"`
	BeforeBalance  int64     `db:"before_balance" json:"before_balance"`
	AfterBalance   int64     `db:"after_balance" json:"after_balance"`
	IdempotencyKey string    `db:"idempotency_key" json:"idempotency_key"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
}
