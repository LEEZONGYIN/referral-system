package service

import (
	"context"
	"testing"
)

type fakeTxMgr struct{}

func (f fakeTxMgr) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func TestNewReferralService(t *testing.T) {
	_ = NewReferralService(ReferralServiceDeps{})
}

func TestReferralServiceNilGuard(t *testing.T) {
	svc := NewReferralService(ReferralServiceDeps{TxMgr: fakeTxMgr{}})
	if _, err := svc.RegisterWithReferral(context.Background(), nil, "code", "idem"); err == nil {
		t.Fatal("expected error")
	}
}
