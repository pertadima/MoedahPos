package service

import (
	"context"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/dto"
	repomocks "github.com/moedahpos/backend/internal/repository/mocks"
	"github.com/moedahpos/backend/internal/service/mocks"
)

//nolint:funlen
func TestTransactionService_LoyaltyRedemption(t *testing.T) {
	ctx := context.Background()
	log := zerolog.Nop()

	t.Run("PayDraft with Loyalty Points Success", func(t *testing.T) {
		tRepo := new(repomocks.TransactionRepository)
		lRepo := new(repomocks.LoyaltyRepository)
		sRepo := new(repomocks.StoreRepository)
		aSvc := new(mocks.ActivityLogServiceInterface)

		tid := "d1"
		cid := "c1"
		req := &dto.PayDraftRequest{
			CustomerID:     cid,
			PointsRedeemed: 10,
			PaymentAmount:  49990, // 50000 - 10 point discount (1pt=1IDR)
			PaymentMethod:  "cash",
		}

		tRepo.On("FindByID", ctx, tid).Return(&domain.Transaction{
			ID:      tid,
			StoreID: "s1",
			Status:  "draft",
			Total:   50000, // enough to earn points: floor(50000/1000)*1.0 = 50 pts
		}, nil)

		lRepo.On("GetBalance", ctx, cid).Return(20.0, nil)
		sRepo.On("FindByID", ctx, "s1").Return(&domain.Store{
			ID:                     "s1",
			LoyaltyRupiahPerPoint:  1,
			LoyaltyPointsPerRupiah: 1000,
		}, nil)

		tRepo.On("PayDraft", ctx, mock.MatchedBy(func(in domain.PayDraftInput) bool {
			return in.PointsRedeemed == 10 && in.PointsDiscount == 10
		}), "s1", "u1").Return(&domain.Transaction{ID: tid, Status: "completed"}, nil)

		lRepo.On("SpendPoints", ctx, cid, mock.Anything, 10.0).Return(&domain.LoyaltyLedger{ID: "lt1"}, nil)
		// New: earn points hook after PayDraft
		lRepo.On("GetCustomerTier", ctx, cid).Return(&domain.MembershipTier{Multiplier: 1.0}, nil)
		lRepo.On("EarnPoints", ctx, cid, mock.Anything, mock.AnythingOfType("float64")).Return(&domain.LoyaltyLedger{ID: "lt2"}, nil)
		aSvc.On("LogActivity", ctx, "u1", "s1", mock.Anything, mock.Anything, tid, mock.Anything).Return()

		s := NewTransactionService(tRepo, nil, nil, nil, nil, aSvc, sRepo, lRepo, log)
		resp, err := s.PayDraft(ctx, "s1", tid, "u1", req)

		assert.NoError(t, err)
		assert.Equal(t, "completed", resp.Status)
		lRepo.AssertExpectations(t)
	})

	t.Run("PayDraft Insufficient Points", func(t *testing.T) {
		tRepo := new(repomocks.TransactionRepository)
		lRepo := new(repomocks.LoyaltyRepository)
		s := NewTransactionService(tRepo, nil, nil, nil, nil, nil, nil, lRepo, log)

		tid := "d1"
		cid := "c1"
		req := &dto.PayDraftRequest{
			CustomerID:     cid,
			PointsRedeemed: 100,
		}

		tRepo.On("FindByID", ctx, tid).Return(&domain.Transaction{
			ID:      tid,
			StoreID: "s1",
			Status:  "draft",
		}, nil)

		lRepo.On("GetBalance", ctx, cid).Return(20.0, nil)

		_, err := s.PayDraft(ctx, "s1", tid, "u1", req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient loyalty points")
	})

	t.Run("PayDraft Store Not Found", func(t *testing.T) {
		tRepo := new(repomocks.TransactionRepository)
		lRepo := new(repomocks.LoyaltyRepository)
		sRepo := new(repomocks.StoreRepository)
		s := NewTransactionService(tRepo, nil, nil, nil, nil, nil, sRepo, lRepo, log)

		tid := "d1"
		cid := "c1"
		req := &dto.PayDraftRequest{
			CustomerID:     cid,
			PointsRedeemed: 10,
		}

		tRepo.On("FindByID", ctx, tid).Return(&domain.Transaction{
			ID:      tid,
			StoreID: "s1",
			Status:  "draft",
		}, nil)

		lRepo.On("GetBalance", ctx, cid).Return(20.0, nil)
		sRepo.On("FindByID", ctx, "s1").Return(nil, assert.AnError)

		_, err := s.PayDraft(ctx, "s1", tid, "u1", req)
		assert.Error(t, err)
	})
}
