package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"referral-system/internal/model"
)

type ReferralEventRepository struct {
	db *sql.DB
}

func NewReferralEventRepository(db *sql.DB) *ReferralEventRepository {
	return &ReferralEventRepository{db: db}
}

func (r *ReferralEventRepository) Create(ctx context.Context, event *model.ReferralEvent) error {
	if event == nil {
		return errors.New("event is nil")
	}

	const query = `
INSERT INTO referral_events (
    relation_id, inviter_user_id, invitee_user_id, event_type, idempotency_key, payload, created_at
)
VALUES (?, ?, ?, ?, ?, ?, ?)`

	res, err := connFromContext(ctx, r.db).ExecContext(ctx, query,
		event.RelationID,
		event.InviterUserID,
		event.InviteeUserID,
		event.EventType,
		event.IdempotencyKey,
		event.Payload,
		event.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert referral_event: %w", err)
	}

	id, err := res.LastInsertId()
	if err == nil {
		event.ID = id
	}
	return nil
}

func (r *ReferralEventRepository) GetByID(ctx context.Context, id int64) (*model.ReferralEvent, error) {
	const query = `
SELECT id, relation_id, inviter_user_id, invitee_user_id, event_type, idempotency_key, payload, created_at
FROM referral_events
WHERE id = ?`
	row := connFromContext(ctx, r.db).QueryRowContext(ctx, query, id)
	return scanReferralEvent(row)
}

func (r *ReferralEventRepository) GetByIdempotencyKey(ctx context.Context, key string) (*model.ReferralEvent, error) {
	const query = `
SELECT id, relation_id, inviter_user_id, invitee_user_id, event_type, idempotency_key, payload, created_at
FROM referral_events
WHERE idempotency_key = ?`
	row := connFromContext(ctx, r.db).QueryRowContext(ctx, query, key)
	return scanReferralEvent(row)
}

func (r *ReferralEventRepository) ListByRelationID(ctx context.Context, relationID int64, limit, offset int) ([]*model.ReferralEvent, error) {
	const query = `
SELECT id, relation_id, inviter_user_id, invitee_user_id, event_type, idempotency_key, payload, created_at
FROM referral_events
WHERE relation_id = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?`

	rows, err := connFromContext(ctx, r.db).QueryContext(ctx, query, relationID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list referral events by relation: %w", err)
	}
	defer rows.Close()

	return scanReferralEvents(rows)
}

func (r *ReferralEventRepository) ListByInviteeUserID(ctx context.Context, inviteeUserID int64, limit, offset int) ([]*model.ReferralEvent, error) {
	const query = `
SELECT id, relation_id, inviter_user_id, invitee_user_id, event_type, idempotency_key, payload, created_at
FROM referral_events
WHERE invitee_user_id = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?`

	rows, err := connFromContext(ctx, r.db).QueryContext(ctx, query, inviteeUserID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list referral events by invitee: %w", err)
	}
	defer rows.Close()

	return scanReferralEvents(rows)
}

func scanReferralEvent(row *sql.Row) (*model.ReferralEvent, error) {
	var item model.ReferralEvent
	var payload []byte
	if err := row.Scan(
		&item.ID,
		&item.RelationID,
		&item.InviterUserID,
		&item.InviteeUserID,
		&item.EventType,
		&item.IdempotencyKey,
		&payload,
		&item.CreatedAt,
	); err != nil {
		return nil, err
	}
	if len(payload) > 0 {
		item.Payload = json.RawMessage(payload)
	}
	return &item, nil
}

func scanReferralEvents(rows *sql.Rows) ([]*model.ReferralEvent, error) {
	result := make([]*model.ReferralEvent, 0)
	for rows.Next() {
		var item model.ReferralEvent
		var payload []byte
		if err := rows.Scan(
			&item.ID,
			&item.RelationID,
			&item.InviterUserID,
			&item.InviteeUserID,
			&item.EventType,
			&item.IdempotencyKey,
			&payload,
			&item.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan referral event: %w", err)
		}
		if len(payload) > 0 {
			item.Payload = json.RawMessage(payload)
		}
		result = append(result, &item)
	}
	return result, nil
}
