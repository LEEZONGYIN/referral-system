package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"referral-system/internal/model"
	"referral-system/internal/repository"
)

var (
	ErrRewardAlreadyProcessed   = errors.New("reward already processed")
	ErrRegisterAlreadyProcessed = errors.New("register already processed")
)

type TransactionManager interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type ReferralService struct {
	txMgr                TransactionManager
	userRepo             repository.UserRepository
	referralRuleRepo     repository.ReferralRuleRepository
	referralRelationRepo repository.ReferralRelationRepository
	referralEventRepo    repository.ReferralEventRepository
	creditAccountRepo    repository.CreditAccountRepository
	creditLedgerRepo     repository.CreditLedgerRepository
	statsRepo            repository.ReferralStatsDailyRepository
}

type ReferralServiceDeps struct {
	TxMgr                TransactionManager
	UserRepo             repository.UserRepository
	ReferralRuleRepo     repository.ReferralRuleRepository
	ReferralRelationRepo repository.ReferralRelationRepository
	ReferralEventRepo    repository.ReferralEventRepository
	CreditAccountRepo    repository.CreditAccountRepository
	CreditLedgerRepo     repository.CreditLedgerRepository
	StatsRepo            repository.ReferralStatsDailyRepository
}

func NewReferralService(deps ReferralServiceDeps) *ReferralService {
	return &ReferralService{
		txMgr:                deps.TxMgr,
		userRepo:             deps.UserRepo,
		referralRuleRepo:     deps.ReferralRuleRepo,
		referralRelationRepo:  deps.ReferralRelationRepo,
		referralEventRepo:     deps.ReferralEventRepo,
		creditAccountRepo:     deps.CreditAccountRepo,
		creditLedgerRepo:      deps.CreditLedgerRepo,
		statsRepo:             deps.StatsRepo,
	}
}

func (s *ReferralService) RegisterWithReferral(ctx context.Context, invitee *model.User, referralCode string, idempotencyKey string) (*model.ReferralRelation, error) {
	if s.txMgr == nil {
		return nil, errors.New("transaction manager is required")
	}
	if invitee == nil {
		return nil, errors.New("invitee is nil")
	}
	if referralCode == "" {
		return nil, errors.New("referral code is required")
	}
	if idempotencyKey == "" {
		return nil, errors.New("idempotency key is required")
	}

	var result *model.ReferralRelation
	err := s.txMgr.WithinTransaction(ctx, func(txCtx context.Context) error {
		existingEvent, err := s.referralEventRepo.GetByIdempotencyKey(txCtx, idempotencyKey)
		if err == nil && existingEvent != nil && existingEvent.RelationID != nil {
			relation, err := s.referralRelationRepo.GetByID(txCtx, *existingEvent.RelationID)
			if err != nil {
				return fmt.Errorf("load existing relation: %w", err)
			}
			result = relation
			return ErrRegisterAlreadyProcessed
		}

		inviter, err := s.userRepo.GetByReferralCode(txCtx, referralCode)
		if err != nil {
			return fmt.Errorf("find inviter: %w", err)
		}

		if invitee.ReferralCode == "" {
			invitee.ReferralCode = generateReferralCode()
		}

		now := time.Now()
		invitee.CreatedAt = now
		invitee.UpdatedAt = now
		if err := s.userRepo.Create(txCtx, invitee); err != nil {
			return fmt.Errorf("create invitee: %w", err)
		}

		rule, err := s.findActiveRule(txCtx, now)
		if err != nil {
			return fmt.Errorf("find referral rule: %w", err)
		}

		relation := &model.ReferralRelation{
			InviterUserID: inviter.ID,
			InviteeUserID: invitee.ID,
			ReferralCode:  referralCode,
			Status:        model.ReferralRelationConfirmed,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		if rule != nil {
			relation.RuleID = &rule.ID
		}
		if err := s.referralRelationRepo.Create(txCtx, relation); err != nil {
			return fmt.Errorf("create referral relation: %w", err)
		}

		event := &model.ReferralEvent{
			RelationID:     &relation.ID,
			InviterUserID:  &inviter.ID,
			InviteeUserID:  &invitee.ID,
			EventType:      model.ReferralEventRegistered,
			IdempotencyKey: idempotencyKey,
			CreatedAt:      now,
		}
		if err := s.referralEventRepo.Create(txCtx, event); err != nil {
			return fmt.Errorf("create referral event: %w", err)
		}

		result = relation
		return nil
	})
	if err != nil && !errors.Is(err, ErrRegisterAlreadyProcessed) {
		return nil, err
	}
	return result, nil
}

func (s *ReferralService) RewardReferral(ctx context.Context, relationID int64, bizID string, amount int64, idempotencyKey string) error {
	if s.txMgr == nil {
		return errors.New("transaction manager is required")
	}
	if amount <= 0 {
		return errors.New("amount must be positive")
	}
	if bizID == "" {
		return errors.New("biz id is required")
	}
	if idempotencyKey == "" {
		return errors.New("idempotency key is required")
	}

	return s.txMgr.WithinTransaction(ctx, func(txCtx context.Context) error {
		existingLedger, err := s.creditLedgerRepo.GetByIdempotencyKey(txCtx, idempotencyKey)
		if err == nil && existingLedger != nil {
			return ErrRewardAlreadyProcessed
		}

		relation, err := s.referralRelationRepo.GetByID(txCtx, relationID)
		if err != nil {
			return fmt.Errorf("get relation: %w", err)
		}

		if relation.Status == model.ReferralRelationRewarded {
			return ErrRewardAlreadyProcessed
		}

		if err := s.creditAccountRepo.CreateIfNotExists(txCtx, relation.InviterUserID); err != nil {
			return fmt.Errorf("ensure credit account: %w", err)
		}

		account, err := s.creditAccountRepo.GetByUserID(txCtx, relation.InviterUserID)
		if err != nil {
			return fmt.Errorf("get credit account: %w", err)
		}

		ok, err := s.creditAccountRepo.UpdateBalanceWithOptimisticLock(txCtx, relation.InviterUserID, amount, account.Version)
		if err != nil {
			return fmt.Errorf("update balance: %w", err)
		}
		if !ok {
			return sql.ErrTxDone
		}

		now := time.Now()
		ledger := &model.CreditLedger{
			UserID:         relation.InviterUserID,
			BizType:        model.CreditBizTypeReferralReward,
			BizID:          bizID,
			Direction:      model.CreditDirectionIn,
			Amount:         amount,
			BeforeBalance:  account.Balance,
			AfterBalance:   account.Balance + amount,
			IdempotencyKey: idempotencyKey,
			CreatedAt:      now,
		}
		if err := s.creditLedgerRepo.Create(txCtx, ledger); err != nil {
			return fmt.Errorf("create credit ledger: %w", err)
		}

		relation.Status = model.ReferralRelationRewarded
		relation.RewardedAt = &now
		if err := s.referralRelationRepo.UpdateStatus(txCtx, relation.ID, relation.Status, relation.QualifiedAt, relation.RewardedAt); err != nil {
			return fmt.Errorf("update relation status: %w", err)
		}

		event := &model.ReferralEvent{
			RelationID:     &relation.ID,
			InviterUserID:  &relation.InviterUserID,
			InviteeUserID:  &relation.InviteeUserID,
			EventType:      model.ReferralEventRewarded,
			IdempotencyKey: idempotencyKey,
			CreatedAt:      now,
		}
		if err := s.referralEventRepo.Create(txCtx, event); err != nil {
			return fmt.Errorf("create reward event: %w", err)
		}

		return nil
	})
}

func (s *ReferralService) findActiveRule(ctx context.Context, now time.Time) (*model.ReferralRule, error) {
	if s.referralRuleRepo == nil {
		return nil, nil
	}
	rules, err := s.referralRuleRepo.ListActiveRules(ctx, now)
	if err != nil {
		return nil, err
	}
	if len(rules) == 0 {
		return nil, nil
	}
	return rules[0], nil
}

func (s *ReferralService) GetReferralHistory(ctx context.Context, userID int64, limit, offset int) ([]*model.ReferralRelation, error) {
	if s.referralRelationRepo == nil {
		return nil, errors.New("referral relation repository is required")
	}
	return s.referralRelationRepo.GetByInviterUserID(ctx, userID, limit, offset)
}

func (s *ReferralService) GetDashboard(ctx context.Context, userID int64) (map[string]int64, error) {
	if s.referralRelationRepo == nil {
		return nil, errors.New("referral relation repository is required")
	}
	inviteCount, err := s.referralRelationRepo.CountByInviterUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var qualifiedCount int64
	var rewardAmountTotal int64
	if s.statsRepo != nil {
		stats, err := s.statsRepo.ListByUser(ctx, userID, 1000, 0)
		if err != nil {
			return nil, err
		}
		for _, stat := range stats {
			qualifiedCount += int64(stat.QualifiedCount)
			rewardAmountTotal += stat.RewardAmountTotal
		}
	}

	return map[string]int64{
		"invite_count":        inviteCount,
		"qualified_count":     qualifiedCount,
		"reward_amount_total": rewardAmountTotal,
	}, nil
}

func (s *ReferralService) GetCreditBalance(ctx context.Context, userID int64) (*model.CreditAccount, error) {
	if s.creditAccountRepo == nil {
		return nil, errors.New("credit account repository is required")
	}
	return s.creditAccountRepo.GetByUserID(ctx, userID)
}

func (s *ReferralService) GetCreditLedger(ctx context.Context, userID int64, limit, offset int) ([]*model.CreditLedger, error) {
	if s.creditLedgerRepo == nil {
		return nil, errors.New("credit ledger repository is required")
	}
	return s.creditLedgerRepo.ListByUserID(ctx, userID, limit, offset)
}

func (s *ReferralService) ListUsers(ctx context.Context, limit, offset int) ([]*model.User, error) {
	if s.userRepo == nil {
		return nil, errors.New("user repository is required")
	}
	return s.userRepo.List(ctx, limit, offset)
}

func generateReferralCode() string {
	buf := make([]byte, 4)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("ref-%d", time.Now().UnixNano())
	}
	return "ref-" + hex.EncodeToString(buf)
}
