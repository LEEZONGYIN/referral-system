package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"referral-system/internal/model"
)

type ReferralRelationRepository struct {
	db *sql.DB
}

func NewReferralRelationRepository(db *sql.DB) *ReferralRelationRepository {
	return &ReferralRelationRepository{db: db}
}

func (r *ReferralRelationRepository) Create(ctx context.Context, relation *model.ReferralRelation) error {
	if relation == nil {
		return errors.New("relation is nil")
	}

	const query = `
INSERT INTO referral_relations (
    inviter_user_id, invitee_user_id, referral_code, rule_id, status,
    qualified_at, rewarded_at, created_at, updated_at
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	res, err := connFromContext(ctx, r.db).ExecContext(ctx, query,
		relation.InviterUserID,
		relation.InviteeUserID,
		relation.ReferralCode,
		relation.RuleID,
		relation.Status,
		relation.QualifiedAt,
		relation.RewardedAt,
		relation.CreatedAt,
		relation.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert referral_relation: %w", err)
	}

	id, err := res.LastInsertId()
	if err == nil {
		relation.ID = id
	}
	return nil
}

func (r *ReferralRelationRepository) GetByID(ctx context.Context, id int64) (*model.ReferralRelation, error) {
	const query = `
SELECT id, inviter_user_id, invitee_user_id, referral_code, rule_id, status, qualified_at, rewarded_at, created_at, updated_at
FROM referral_relations
WHERE id = ?`
	row := connFromContext(ctx, r.db).QueryRowContext(ctx, query, id)
	return scanReferralRelation(row)
}

func (r *ReferralRelationRepository) GetByInviteeUserID(ctx context.Context, inviteeUserID int64) (*model.ReferralRelation, error) {
	const query = `
SELECT id, inviter_user_id, invitee_user_id, referral_code, rule_id, status, qualified_at, rewarded_at, created_at, updated_at
FROM referral_relations
WHERE invitee_user_id = ?`
	row := connFromContext(ctx, r.db).QueryRowContext(ctx, query, inviteeUserID)
	return scanReferralRelation(row)
}

func (r *ReferralRelationRepository) GetByInviterUserID(ctx context.Context, inviterUserID int64, limit, offset int) ([]*model.ReferralRelation, error) {
	const query = `
SELECT id, inviter_user_id, invitee_user_id, referral_code, rule_id, status, qualified_at, rewarded_at, created_at, updated_at
FROM referral_relations
WHERE inviter_user_id = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?`

	rows, err := connFromContext(ctx, r.db).QueryContext(ctx, query, inviterUserID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list referral relations: %w", err)
	}
	defer rows.Close()

	result := make([]*model.ReferralRelation, 0)
	for rows.Next() {
		item := new(model.ReferralRelation)
		if err := rows.Scan(
			&item.ID,
			&item.InviterUserID,
			&item.InviteeUserID,
			&item.ReferralCode,
			&item.RuleID,
			&item.Status,
			&item.QualifiedAt,
			&item.RewardedAt,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan referral relation: %w", err)
		}
		result = append(result, item)
	}
	return result, nil
}

func (r *ReferralRelationRepository) UpdateStatus(ctx context.Context, relationID int64, status int8, qualifiedAt, rewardedAt *time.Time) error {
	const query = `
UPDATE referral_relations
SET status = ?, qualified_at = ?, rewarded_at = ?, updated_at = ?
WHERE id = ?`

	if _, err := connFromContext(ctx, r.db).ExecContext(ctx, query, status, qualifiedAt, rewardedAt, time.Now(), relationID); err != nil {
		return fmt.Errorf("update referral relation status: %w", err)
	}
	return nil
}

func (r *ReferralRelationRepository) CountByInviterUserID(ctx context.Context, inviterUserID int64) (int64, error) {
	const query = `SELECT COUNT(1) FROM referral_relations WHERE inviter_user_id = ?`
	var count int64
	if err := connFromContext(ctx, r.db).QueryRowContext(ctx, query, inviterUserID).Scan(&count); err != nil {
		return 0, fmt.Errorf("count referral relations: %w", err)
	}
	return count, nil
}

func scanReferralRelation(row *sql.Row) (*model.ReferralRelation, error) {
	var item model.ReferralRelation
	if err := row.Scan(
		&item.ID,
		&item.InviterUserID,
		&item.InviteeUserID,
		&item.ReferralCode,
		&item.RuleID,
		&item.Status,
		&item.QualifiedAt,
		&item.RewardedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &item, nil
}
