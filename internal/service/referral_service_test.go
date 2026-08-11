package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"referral-system/internal/model"
)

type fakeTxMgr struct{}

func (f fakeTxMgr) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

type testStore struct {
	mu sync.Mutex

	nextUserID     int64
	nextRelationID int64
	nextEventID    int64
	nextLedgerID   int64

	users     map[int64]*model.User
	relations map[int64]*model.ReferralRelation
	events    map[int64]*model.ReferralEvent
	ledgers   map[int64]*model.CreditLedger
	accounts  map[int64]*model.CreditAccount

	activeRule *model.ReferralRule
}

func newTestStore() *testStore {
	return &testStore{
		nextUserID:     1,
		nextRelationID: 1,
		nextEventID:    1,
		nextLedgerID:   1,
		users:          make(map[int64]*model.User),
		relations:      make(map[int64]*model.ReferralRelation),
		events:         make(map[int64]*model.ReferralEvent),
		ledgers:        make(map[int64]*model.CreditLedger),
		accounts:       make(map[int64]*model.CreditAccount),
	}
}

func strPtr(v string) *string { return &v }

func cloneStringPtr(v *string) *string {
	if v == nil {
		return nil
	}
	s := *v
	return &s
}

func cloneUser(u *model.User) *model.User {
	if u == nil {
		return nil
	}
	c := *u
	c.Email = cloneStringPtr(u.Email)
	c.Phone = cloneStringPtr(u.Phone)
	return &c
}

func cloneRule(r *model.ReferralRule) *model.ReferralRule {
	if r == nil {
		return nil
	}
	c := *r
	if r.EffectiveFrom != nil {
		v := *r.EffectiveFrom
		c.EffectiveFrom = &v
	}
	if r.EffectiveTo != nil {
		v := *r.EffectiveTo
		c.EffectiveTo = &v
	}
	return &c
}

func cloneRelation(r *model.ReferralRelation) *model.ReferralRelation {
	if r == nil {
		return nil
	}
	c := *r
	if r.RuleID != nil {
		v := *r.RuleID
		c.RuleID = &v
	}
	if r.QualifiedAt != nil {
		v := *r.QualifiedAt
		c.QualifiedAt = &v
	}
	if r.RewardedAt != nil {
		v := *r.RewardedAt
		c.RewardedAt = &v
	}
	return &c
}

func cloneEvent(e *model.ReferralEvent) *model.ReferralEvent {
	if e == nil {
		return nil
	}
	c := *e
	if e.RelationID != nil {
		v := *e.RelationID
		c.RelationID = &v
	}
	if e.InviterUserID != nil {
		v := *e.InviterUserID
		c.InviterUserID = &v
	}
	if e.InviteeUserID != nil {
		v := *e.InviteeUserID
		c.InviteeUserID = &v
	}
	if e.Payload != nil {
		payload := make([]byte, len(e.Payload))
		copy(payload, e.Payload)
		c.Payload = payload
	}
	return &c
}

func cloneLedger(l *model.CreditLedger) *model.CreditLedger {
	if l == nil {
		return nil
	}
	c := *l
	return &c
}

func cloneAccount(a *model.CreditAccount) *model.CreditAccount {
	if a == nil {
		return nil
	}
	c := *a
	return &c
}

func seedTestService(t *testing.T) (*ReferralService, *testStore, *model.User) {
	t.Helper()

	store := newTestStore()
	now := time.Now()
	inviter := &model.User{
		ID:           store.nextUserID,
		Name:         "Alice",
		Email:        strPtr("alice@example.com"),
		Phone:        strPtr("13800000001"),
		ReferralCode: "ref-93fcc618",
		Status:       1,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	store.users[inviter.ID] = cloneUser(inviter)
	store.nextUserID++

	store.activeRule = &model.ReferralRule{
		ID:           1,
		RuleCode:     "DEFAULT_REGISTER_REWARD",
		RewardAmount: 100,
		TriggerEvent: model.ReferralEventRegistered,
		Status:       1,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	svc := NewReferralService(ReferralServiceDeps{
		TxMgr:                fakeTxMgr{},
		UserRepo:             &fakeUserRepo{store: store},
		ReferralRuleRepo:     &fakeRuleRepo{store: store},
		ReferralRelationRepo: &fakeRelationRepo{store: store},
		ReferralEventRepo:    &fakeEventRepo{store: store},
		CreditAccountRepo:    &fakeCreditAccountRepo{store: store},
		CreditLedgerRepo:     &fakeCreditLedgerRepo{store: store},
		StatsRepo:            nil,
	})

	return svc, store, inviter
}

type fakeUserRepo struct {
	store *testStore
}

func (r *fakeUserRepo) Create(ctx context.Context, user *model.User) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	if user == nil {
		return errors.New("user is nil")
	}

	for _, existing := range r.store.users {
		if user.ReferralCode != "" && existing.ReferralCode == user.ReferralCode {
			return errors.New("duplicate referral code")
		}
		if user.Email != nil && existing.Email != nil && *user.Email == *existing.Email {
			return errors.New("duplicate email")
		}
		if user.Phone != nil && existing.Phone != nil && *user.Phone == *existing.Phone {
			return errors.New("duplicate phone")
		}
	}

	user.ID = r.store.nextUserID
	r.store.nextUserID++
	r.store.users[user.ID] = cloneUser(user)
	return nil
}

func (r *fakeUserRepo) GetByID(ctx context.Context, id int64) (*model.User, error) {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	if user, ok := r.store.users[id]; ok {
		return cloneUser(user), nil
	}
	return nil, sql.ErrNoRows
}

func (r *fakeUserRepo) GetByReferralCode(ctx context.Context, referralCode string) (*model.User, error) {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	for _, user := range r.store.users {
		if user.ReferralCode == referralCode {
			return cloneUser(user), nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *fakeUserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	for _, user := range r.store.users {
		if user.Email != nil && *user.Email == email {
			return cloneUser(user), nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *fakeUserRepo) GetByPhone(ctx context.Context, phone string) (*model.User, error) {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	for _, user := range r.store.users {
		if user.Phone != nil && *user.Phone == phone {
			return cloneUser(user), nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *fakeUserRepo) Update(ctx context.Context, user *model.User) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	if user == nil {
		return errors.New("user is nil")
	}
	if _, ok := r.store.users[user.ID]; !ok {
		return sql.ErrNoRows
	}
	r.store.users[user.ID] = cloneUser(user)
	return nil
}

func (r *fakeUserRepo) List(ctx context.Context, limit, offset int) ([]*model.User, error) {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	items := make([]*model.User, 0, len(r.store.users))
	for _, user := range r.store.users {
		items = append(items, cloneUser(user))
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID > items[j].ID })

	if offset >= len(items) {
		return []*model.User{}, nil
	}
	end := offset + limit
	if end > len(items) {
		end = len(items)
	}
	return items[offset:end], nil
}

type fakeRuleRepo struct {
	store *testStore
}

func (r *fakeRuleRepo) Create(ctx context.Context, rule *model.ReferralRule) error {
	return errors.New("not implemented")
}

func (r *fakeRuleRepo) GetByID(ctx context.Context, id int64) (*model.ReferralRule, error) {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	if r.store.activeRule != nil && r.store.activeRule.ID == id {
		return cloneRule(r.store.activeRule), nil
	}
	return nil, sql.ErrNoRows
}

func (r *fakeRuleRepo) GetByRuleCode(ctx context.Context, ruleCode string) (*model.ReferralRule, error) {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	if r.store.activeRule != nil && r.store.activeRule.RuleCode == ruleCode {
		return cloneRule(r.store.activeRule), nil
	}
	return nil, sql.ErrNoRows
}

func (r *fakeRuleRepo) ListActiveRules(ctx context.Context, now time.Time) ([]*model.ReferralRule, error) {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	if r.store.activeRule == nil {
		return []*model.ReferralRule{}, nil
	}
	return []*model.ReferralRule{cloneRule(r.store.activeRule)}, nil
}

type fakeRelationRepo struct {
	store *testStore
}

func (r *fakeRelationRepo) Create(ctx context.Context, relation *model.ReferralRelation) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	if relation == nil {
		return errors.New("relation is nil")
	}

	for _, existing := range r.store.relations {
		if existing.InviteeUserID == relation.InviteeUserID {
			return errors.New("duplicate invitee relation")
		}
		if existing.InviterUserID == relation.InviterUserID && existing.InviteeUserID == relation.InviteeUserID {
			return errors.New("duplicate inviter/invitee relation")
		}
	}

	relation.ID = r.store.nextRelationID
	r.store.nextRelationID++
	r.store.relations[relation.ID] = cloneRelation(relation)
	return nil
}

func (r *fakeRelationRepo) GetByID(ctx context.Context, id int64) (*model.ReferralRelation, error) {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	if relation, ok := r.store.relations[id]; ok {
		return cloneRelation(relation), nil
	}
	return nil, sql.ErrNoRows
}

func (r *fakeRelationRepo) GetByInviteeUserID(ctx context.Context, inviteeUserID int64) (*model.ReferralRelation, error) {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	for _, relation := range r.store.relations {
		if relation.InviteeUserID == inviteeUserID {
			return cloneRelation(relation), nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *fakeRelationRepo) GetByInviterUserID(ctx context.Context, inviterUserID int64, limit, offset int) ([]*model.ReferralRelation, error) {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	items := make([]*model.ReferralRelation, 0)
	for _, relation := range r.store.relations {
		if relation.InviterUserID == inviterUserID {
			items = append(items, cloneRelation(relation))
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })

	if offset >= len(items) {
		return []*model.ReferralRelation{}, nil
	}
	end := offset + limit
	if end > len(items) {
		end = len(items)
	}
	return items[offset:end], nil
}

func (r *fakeRelationRepo) UpdateStatus(ctx context.Context, relationID int64, status int8, qualifiedAt, rewardedAt *time.Time) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	relation, ok := r.store.relations[relationID]
	if !ok {
		return sql.ErrNoRows
	}
	relation.Status = status
	relation.QualifiedAt = qualifiedAt
	relation.RewardedAt = rewardedAt
	relation.UpdatedAt = time.Now()
	r.store.relations[relationID] = relation
	return nil
}

func (r *fakeRelationRepo) CountByInviterUserID(ctx context.Context, inviterUserID int64) (int64, error) {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	var count int64
	for _, relation := range r.store.relations {
		if relation.InviterUserID == inviterUserID {
			count++
		}
	}
	return count, nil
}

type fakeEventRepo struct {
	store *testStore
}

func (r *fakeEventRepo) Create(ctx context.Context, event *model.ReferralEvent) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	if event == nil {
		return errors.New("event is nil")
	}

	for _, existing := range r.store.events {
		if existing.IdempotencyKey == event.IdempotencyKey {
			return errors.New("duplicate event idempotency key")
		}
	}

	event.ID = r.store.nextEventID
	r.store.nextEventID++
	r.store.events[event.ID] = cloneEvent(event)
	return nil
}

func (r *fakeEventRepo) GetByID(ctx context.Context, id int64) (*model.ReferralEvent, error) {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	if event, ok := r.store.events[id]; ok {
		return cloneEvent(event), nil
	}
	return nil, sql.ErrNoRows
}

func (r *fakeEventRepo) GetByIdempotencyKey(ctx context.Context, key string) (*model.ReferralEvent, error) {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	for _, event := range r.store.events {
		if event.IdempotencyKey == key {
			return cloneEvent(event), nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *fakeEventRepo) ListByRelationID(ctx context.Context, relationID int64, limit, offset int) ([]*model.ReferralEvent, error) {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	items := make([]*model.ReferralEvent, 0)
	for _, event := range r.store.events {
		if event.RelationID != nil && *event.RelationID == relationID {
			items = append(items, cloneEvent(event))
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
	if offset >= len(items) {
		return []*model.ReferralEvent{}, nil
	}
	end := offset + limit
	if end > len(items) {
		end = len(items)
	}
	return items[offset:end], nil
}

func (r *fakeEventRepo) ListByInviteeUserID(ctx context.Context, inviteeUserID int64, limit, offset int) ([]*model.ReferralEvent, error) {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	items := make([]*model.ReferralEvent, 0)
	for _, event := range r.store.events {
		if event.InviteeUserID != nil && *event.InviteeUserID == inviteeUserID {
			items = append(items, cloneEvent(event))
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
	if offset >= len(items) {
		return []*model.ReferralEvent{}, nil
	}
	end := offset + limit
	if end > len(items) {
		end = len(items)
	}
	return items[offset:end], nil
}

type fakeCreditAccountRepo struct {
	store *testStore
}

func (r *fakeCreditAccountRepo) CreateIfNotExists(ctx context.Context, userID int64) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	if _, ok := r.store.accounts[userID]; ok {
		return nil
	}
	r.store.accounts[userID] = &model.CreditAccount{UserID: userID, Balance: 0, FrozenBalance: 0, Version: 0}
	return nil
}

func (r *fakeCreditAccountRepo) GetByUserID(ctx context.Context, userID int64) (*model.CreditAccount, error) {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	if account, ok := r.store.accounts[userID]; ok {
		return cloneAccount(account), nil
	}
	return nil, sql.ErrNoRows
}

func (r *fakeCreditAccountRepo) UpdateBalanceWithOptimisticLock(ctx context.Context, userID int64, delta int64, expectedVersion int64) (bool, error) {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	account, ok := r.store.accounts[userID]
	if !ok {
		return false, sql.ErrNoRows
	}
	if account.Version != expectedVersion {
		return false, nil
	}
	account.Balance += delta
	account.Version++
	r.store.accounts[userID] = account
	return true, nil
}

func (r *fakeCreditAccountRepo) AddCredit(ctx context.Context, userID int64, amount int64) error {
	return errors.New("not implemented")
}

type fakeCreditLedgerRepo struct {
	store *testStore
}

func (r *fakeCreditLedgerRepo) Create(ctx context.Context, ledger *model.CreditLedger) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	if ledger == nil {
		return errors.New("ledger is nil")
	}

	for _, existing := range r.store.ledgers {
		if existing.IdempotencyKey == ledger.IdempotencyKey {
			return errors.New("duplicate ledger idempotency key")
		}
		if existing.BizType == ledger.BizType && existing.BizID == ledger.BizID {
			return errors.New("duplicate biz")
		}
	}

	ledger.ID = r.store.nextLedgerID
	r.store.nextLedgerID++
	r.store.ledgers[ledger.ID] = cloneLedger(ledger)
	return nil
}

func (r *fakeCreditLedgerRepo) GetByID(ctx context.Context, id int64) (*model.CreditLedger, error) {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	if ledger, ok := r.store.ledgers[id]; ok {
		return cloneLedger(ledger), nil
	}
	return nil, sql.ErrNoRows
}

func (r *fakeCreditLedgerRepo) GetByIdempotencyKey(ctx context.Context, key string) (*model.CreditLedger, error) {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	for _, ledger := range r.store.ledgers {
		if ledger.IdempotencyKey == key {
			return cloneLedger(ledger), nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *fakeCreditLedgerRepo) GetByBiz(ctx context.Context, bizType, bizID string) (*model.CreditLedger, error) {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	for _, ledger := range r.store.ledgers {
		if ledger.BizType == bizType && ledger.BizID == bizID {
			return cloneLedger(ledger), nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *fakeCreditLedgerRepo) ListByUserID(ctx context.Context, userID int64, limit, offset int) ([]*model.CreditLedger, error) {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	items := make([]*model.CreditLedger, 0)
	for _, ledger := range r.store.ledgers {
		if ledger.UserID == userID {
			items = append(items, cloneLedger(ledger))
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })

	if offset >= len(items) {
		return []*model.CreditLedger{}, nil
	}
	end := offset + limit
	if end > len(items) {
		end = len(items)
	}
	return items[offset:end], nil
}

func TestReferralService_NormalFlow(t *testing.T) {
	svc, store, inviter := seedTestService(t)

	invitee := &model.User{Name: "Bob", Email: strPtr("bob@example.com"), Phone: strPtr("13800000002")}
	relation, err := svc.RegisterWithReferral(context.Background(), invitee, inviter.ReferralCode, "register-bob-001")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	if relation == nil {
		t.Fatal("expected relation")
	}
	if relation.InviterUserID != inviter.ID {
		t.Fatalf("unexpected inviter: %d", relation.InviterUserID)
	}
	if relation.InviteeUserID == 0 {
		t.Fatal("expected invitee user id")
	}
	if relation.RuleID == nil {
		t.Fatal("expected rule id to be set")
	}

	store.mu.Lock()
	if len(store.ledgers) != 1 {
		store.mu.Unlock()
		t.Fatalf("expected 1 ledger after auto reward, got %d", len(store.ledgers))
	}
	if len(store.events) != 2 {
		store.mu.Unlock()
		t.Fatalf("expected 2 events after auto reward, got %d", len(store.events))
	}
	account := store.accounts[inviter.ID]
	store.mu.Unlock()
	if account == nil || account.Balance != 100 {
		t.Fatalf("expected inviter balance 100, got %+v", account)
	}
}

func TestReferralService_RepeatInviteRejected(t *testing.T) {
	svc, _, inviter := seedTestService(t)

	first := &model.User{Name: "Bob", Email: strPtr("bob@example.com"), Phone: strPtr("13800000002")}
	if _, err := svc.RegisterWithReferral(context.Background(), first, inviter.ReferralCode, "register-bob-001"); err != nil {
		t.Fatalf("first register failed: %v", err)
	}

	second := &model.User{Name: "Bob Again", Email: strPtr("bob@example.com"), Phone: strPtr("13800000002")}
	_, err := svc.RegisterWithReferral(context.Background(), second, inviter.ReferralCode, "register-bob-002")
	if err == nil {
		t.Fatal("expected duplicate invite to fail")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "duplicate") {
		t.Fatalf("expected duplicate error, got: %v", err)
	}
}

func TestReferralService_SelfInviteRejected(t *testing.T) {
	svc, _, inviter := seedTestService(t)

	self := &model.User{Name: "Alice", Email: strPtr("alice@example.com"), Phone: strPtr("13800000001")}
	_, err := svc.RegisterWithReferral(context.Background(), self, inviter.ReferralCode, "register-self-001")
	if err == nil {
		t.Fatal("expected self-invite to fail")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "duplicate") {
		t.Fatalf("expected duplicate identity error, got: %v", err)
	}
}

func TestReferralService_ReferralCodeNotFound(t *testing.T) {
	svc, _, _ := seedTestService(t)

	invitee := &model.User{Name: "Bob", Email: strPtr("bob@example.com"), Phone: strPtr("13800000002")}
	_, err := svc.RegisterWithReferral(context.Background(), invitee, "not-exist-code", "register-bob-404")
	if err == nil {
		t.Fatal("expected not found error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "find inviter") {
		t.Fatalf("expected inviter lookup error, got: %v", err)
	}
}

func TestReferralService_ConcurrentRegister(t *testing.T) {
	svc, store, inviter := seedTestService(t)

	const n = 20
	errCh := make(chan error, n)
	var wg sync.WaitGroup

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			email := fmt.Sprintf("user-%02d@example.com", i)
			phone := fmt.Sprintf("1390000%04d", i)
			invitee := &model.User{Name: fmt.Sprintf("User-%02d", i), Email: &email, Phone: &phone}

			_, err := svc.RegisterWithReferral(context.Background(), invitee, inviter.ReferralCode, fmt.Sprintf("register-%02d", i))
			errCh <- err
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			t.Fatalf("concurrent register failed: %v", err)
		}
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	if len(store.relations) != n {
		t.Fatalf("expected %d relations, got %d", n, len(store.relations))
	}
	if len(store.users) != n+1 {
		t.Fatalf("expected %d users, got %d", n+1, len(store.users))
	}
}

func TestReferralService_RewardDuplicateNotAllowed(t *testing.T) {
	svc, store, inviter := seedTestService(t)

	invitee := &model.User{Name: "Bob", Email: strPtr("bob@example.com"), Phone: strPtr("13800000002")}
	relation, err := svc.RegisterWithReferral(context.Background(), invitee, inviter.ReferralCode, "register-bob-001")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	store.mu.Lock()
	if len(store.ledgers) != 1 {
		store.mu.Unlock()
		t.Fatalf("expected 1 auto ledger, got %d", len(store.ledgers))
	}
	store.mu.Unlock()

	bizID := fmt.Sprintf("reward-%d-%d", relation.ID, 100)
	err = svc.RewardReferral(context.Background(), relation.ID, bizID, 100, "reward-1")
	if err == nil {
		t.Fatal("expected duplicate reward to fail")
	}
	if !errors.Is(err, ErrRewardAlreadyProcessed) {
		t.Fatalf("expected ErrRewardAlreadyProcessed, got: %v", err)
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	account := store.accounts[inviter.ID]
	if account == nil {
		t.Fatal("expected credit account")
	}
	if account.Balance != 100 {
		t.Fatalf("expected balance 100, got %d", account.Balance)
	}
	if len(store.ledgers) != 1 {
		t.Fatalf("expected 1 ledger, got %d", len(store.ledgers))
	}
}
