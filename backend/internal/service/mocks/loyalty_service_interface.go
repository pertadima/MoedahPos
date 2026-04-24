package mocks

import (
	context "context"

	dto "github.com/moedahpos/backend/internal/dto"
	mock "github.com/stretchr/testify/mock"
)

type LoyaltyServiceInterface struct {
	mock.Mock
}

func (_m *LoyaltyServiceInterface) ListTiers(ctx context.Context) ([]*dto.MembershipTierResponse, error) {
	ret := _m.Called(ctx)
	if ret.Get(0) != nil {
		return ret.Get(0).([]*dto.MembershipTierResponse), ret.Error(1)
	}
	return nil, ret.Error(1)
}

func (_m *LoyaltyServiceInterface) GetBalance(ctx context.Context, customerID string) (*dto.LoyaltyBalanceResponse, error) {
	ret := _m.Called(ctx, customerID)
	if ret.Get(0) != nil {
		return ret.Get(0).(*dto.LoyaltyBalanceResponse), ret.Error(1)
	}
	return nil, ret.Error(1)
}

func (_m *LoyaltyServiceInterface) EarnPoints(ctx context.Context, storeID, customerID string, transactionID *string, total float64) (*dto.LoyaltyLedgerResponse, error) {
	ret := _m.Called(ctx, storeID, customerID, transactionID, total)
	if ret.Get(0) != nil {
		return ret.Get(0).(*dto.LoyaltyLedgerResponse), ret.Error(1)
	}
	return nil, ret.Error(1)
}

func (_m *LoyaltyServiceInterface) RedeemPoints(ctx context.Context, customerID string, transactionID *string, points float64) (*dto.LoyaltyLedgerResponse, error) {
	ret := _m.Called(ctx, customerID, transactionID, points)
	if ret.Get(0) != nil {
		return ret.Get(0).(*dto.LoyaltyLedgerResponse), ret.Error(1)
	}
	return nil, ret.Error(1)
}

func (_m *LoyaltyServiceInterface) GetHistory(ctx context.Context, customerID string) ([]*dto.LoyaltyLedgerResponse, error) {
	ret := _m.Called(ctx, customerID)
	if ret.Get(0) != nil {
		return ret.Get(0).([]*dto.LoyaltyLedgerResponse), ret.Error(1)
	}
	return nil, ret.Error(1)
}

func (_m *LoyaltyServiceInterface) AssignTier(ctx context.Context, customerID, tierID string) error {
	ret := _m.Called(ctx, customerID, tierID)
	return ret.Error(0)
}
