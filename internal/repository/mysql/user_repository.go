package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"referral-system/internal/model"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	if user == nil {
		return errors.New("user is nil")
	}

	const query = `
INSERT INTO users (name, email, phone, referral_code, status, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?)`

	res, err := connFromContext(ctx, r.db).ExecContext(ctx, query,
		user.Name,
		user.Email,
		user.Phone,
		user.ReferralCode,
		user.Status,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}

	id, err := res.LastInsertId()
	if err == nil {
		user.ID = id
	}
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {
	const query = `
SELECT id, name, email, phone, referral_code, status, created_at, updated_at
FROM users
WHERE id = ?`

	row := connFromContext(ctx, r.db).QueryRowContext(ctx, query, id)
	return scanUser(row)
}

func (r *UserRepository) GetByReferralCode(ctx context.Context, referralCode string) (*model.User, error) {
	const query = `
SELECT id, name, email, phone, referral_code, status, created_at, updated_at
FROM users
WHERE referral_code = ?`

	row := connFromContext(ctx, r.db).QueryRowContext(ctx, query, referralCode)
	return scanUser(row)
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	const query = `
SELECT id, name, email, phone, referral_code, status, created_at, updated_at
FROM users
WHERE email = ?`

	row := connFromContext(ctx, r.db).QueryRowContext(ctx, query, email)
	return scanUser(row)
}

func (r *UserRepository) GetByPhone(ctx context.Context, phone string) (*model.User, error) {
	const query = `
SELECT id, name, email, phone, referral_code, status, created_at, updated_at
FROM users
WHERE phone = ?`

	row := connFromContext(ctx, r.db).QueryRowContext(ctx, query, phone)
	return scanUser(row)
}

func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	if user == nil {
		return errors.New("user is nil")
	}

	const query = `
UPDATE users
SET name = ?, email = ?, phone = ?, referral_code = ?, status = ?, updated_at = ?
WHERE id = ?`

	_, err := connFromContext(ctx, r.db).ExecContext(ctx, query,
		user.Name,
		user.Email,
		user.Phone,
		user.ReferralCode,
		user.Status,
		user.UpdatedAt,
		user.ID,
	)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}

func scanUser(row *sql.Row) (*model.User, error) {
	var user model.User
	if err := row.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Phone,
		&user.ReferralCode,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &user, nil
}
