package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"referral-system/internal/model"
)

type ReferralRuleRepository struct {
	db *sql.DB
}

func NewReferralRuleRepository(db *sql.DB) *ReferralRuleRepository {
	return &ReferralRuleRepository{db: db}
}

func (r *ReferralRuleRepository) Create(ctx context.Context, rule *model.ReferralRule) error {
	if rule == nil {
		return errors.New("rule is nil")
	}

	const query = `
INSERT INTO referral_rules (rule_code, reward_amount, trigger_event, status, effective_from, effective_to, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	res, err := connFromContext(ctx, r.db).ExecContext(ctx, query,
		rule.RuleCode,
		rule.RewardAmount,
		rule.TriggerEvent,
		rule.Status,
		rule.EffectiveFrom,
		rule.EffectiveTo,
		rule.CreatedAt,
		rule.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert referral_rule: %w", err)
	}

	id, err := res.LastInsertId()
	if err == nil {
		rule.ID = id
	}
	return nil
}

func (r *ReferralRuleRepository) GetByID(ctx context.Context, id int64) (*model.ReferralRule, error) {
	const query = `
SELECT id, rule_code, reward_amount, trigger_event, status, effective_from, effective_to, created_at, updated_at
FROM referral_rules
WHERE id = ?`
	row := connFromContext(ctx, r.db).QueryRowContext(ctx, query, id)
	return scanReferralRule(row)
}

func (r *ReferralRuleRepository) GetByRuleCode(ctx context.Context, ruleCode string) (*model.ReferralRule, error) {
	const query = `
SELECT id, rule_code, reward_amount, trigger_event, status, effective_from, effective_to, created_at, updated_at
FROM referral_rules
WHERE rule_code = ?`
	row := connFromContext(ctx, r.db).QueryRowContext(ctx, query, ruleCode)
	return scanReferralRule(row)
}

func (r *ReferralRuleRepository) ListActiveRules(ctx context.Context, now time.Time) ([]*model.ReferralRule, error) {
	const query = `
SELECT id, rule_code, reward_amount, trigger_event, status, effective_from, effective_to, created_at, updated_at
FROM referral_rules
WHERE status = 1
  AND (effective_from IS NULL OR effective_from <= ?)
  AND (effective_to IS NULL OR effective_to >= ?)
ORDER BY id DESC`

	rows, err := connFromContext(ctx, r.db).QueryContext(ctx, query, now, now)
	if err != nil {
		return nil, fmt.Errorf("list active rules: %w", err)
	}
	defer rows.Close()

	result := make([]*model.ReferralRule, 0)
	for rows.Next() {
		item := new(model.ReferralRule)
		if err := rows.Scan(
			&item.ID,
			&item.RuleCode,
			&item.RewardAmount,
			&item.TriggerEvent,
			&item.Status,
			&item.EffectiveFrom,
			&item.EffectiveTo,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan active rule: %w", err)
		}
		result = append(result, item)
	}
	return result, nil
}

func scanReferralRule(row *sql.Row) (*model.ReferralRule, error) {
	var item model.ReferralRule
	if err := row.Scan(
		&item.ID,
		&item.RuleCode,
		&item.RewardAmount,
		&item.TriggerEvent,
		&item.Status,
		&item.EffectiveFrom,
		&item.EffectiveTo,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &item, nil
}
