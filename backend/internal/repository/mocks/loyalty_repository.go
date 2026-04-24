package mocks

import (
	context "context"

	domain "github.com/moedahpos/backend/internal/domain"
	mock "github.com/stretchr/testify/mock"
)

type LoyaltyRepository struct {
	mock.Mock
}

func (_m *LoyaltyRepository) GetBalance(ctx context.Context, customerID string) (float64, error) {
	ret := _m.Called(ctx, customerID)
	return ret.Get(0).(float64), ret.Error(1)
}

func (_m *LoyaltyRepository) EarnPoints(ctx context.Context, customerID string, transactionID *string, points float64) (*domain.LoyaltyLedger, error) {
	ret := _m.Called(ctx, customerID, transactionID, points)
	if ret.Get(0) != nil {
		return ret.Get(0).(*domain.LoyaltyLedger), ret.Error(1)
	}
	return nil, ret.Error(1)
}

func (_m *LoyaltyRepository) SpendPoints(ctx context.Context, customerID string, transactionID *string, points float64) (*domain.LoyaltyLedger, error) {
	ret := _m.Called(ctx, customerID, transactionID, points)
	if ret.Get(0) != nil {
		return ret.Get(0).(*domain.LoyaltyLedger), ret.Error(1)
	}
	return nil, ret.Error(1)
}

func (_m *LoyaltyRepository) GetHistory(ctx context.Context, customerID string) ([]*domain.LoyaltyLedger, error) {
	ret := _m.Called(ctx, customerID)
	return ret.Get(0).([]*domain.LoyaltyLedger), ret.Error(1)
}

func (_m *LoyaltyRepository) AssignTier(ctx context.Context, customerID, tierID string) error {
	ret := _m.Called(ctx, customerID, tierID)
	return ret.Error(0)
}

func (_m *LoyaltyRepository) GetCustomerTier(ctx context.Context, customerID string) (*domain.MembershipTier, error) {
	ret := _m.Called(ctx, customerID)
	if ret.Get(0) != nil {
		return ret.Get(0).(*domain.MembershipTier), ret.Error(1)
	}
	return nil, ret.Error(1)
}
