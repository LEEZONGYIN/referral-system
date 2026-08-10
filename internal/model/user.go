package model

import "time"

type User struct {
	ID           int64      `db:"id" json:"id"`
	Name         string     `db:"name" json:"name"`
	Email        *string    `db:"email" json:"email,omitempty"`
	Phone        *string    `db:"phone" json:"phone,omitempty"`
	ReferralCode string     `db:"referral_code" json:"referral_code"`
	Status       int8       `db:"status" json:"status"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
}
