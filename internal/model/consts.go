package model

const (
	ReferralRelationPending   int8 = 0
	ReferralRelationConfirmed int8 = 1
	ReferralRelationRewarded  int8 = 2
	ReferralRelationRevoked   int8 = 3
)

const (
	CreditDirectionIn  int8 = 1
	CreditDirectionOut int8 = 2
)

const (
	ReferralEventInviteCreated = "INVITE_CREATED"
	ReferralEventLinkClicked    = "LINK_CLICKED"
	ReferralEventRegistered     = "REGISTERED"
	ReferralEventQualified      = "QUALIFIED"
	ReferralEventRewarded       = "REWARDED"
	ReferralEventRewardFailed   = "REWARD_FAILED"
)

const (
	CreditBizTypeReferralReward   = "REFERRAL_REWARD"
	CreditBizTypeReferralReversal = "REFERRAL_REVERSAL"
	CreditBizTypeAdjustment       = "ADJUSTMENT"
)
