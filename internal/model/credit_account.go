package model

type CreditAccount struct {
	UserID        int64 `db:"user_id" json:"user_id"`
	Balance       int64 `db:"balance" json:"balance"`
	FrozenBalance int64 `db:"frozen_balance" json:"frozen_balance"`
	Version       int64 `db:"version" json:"version"`
}
