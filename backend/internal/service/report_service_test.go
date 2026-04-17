package service

import (
	"context"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/moedahpos/backend/internal/dto"
	repomocks "github.com/moedahpos/backend/internal/repository/mocks"
)

func TestReportService_SalesSummary(t *testing.T) {
	ctx := context.Background()
	log := zerolog.Nop()

	t.Run("Success", func(t *testing.T) {
		rRepo := new(repomocks.ReportRepository)
		s := NewReportService(rRepo, log)

		filter := dto.ReportFilter{
			StoreID:  "s1",
			DateFrom: "2023-01-01",
			DateTo:   "2023-01-01",
		}

		rRepo.On("SalesSummary", ctx, "s1", mock.Anything, mock.Anything).Return([]dto.SalesSummaryRow{
			{
				Date:             "2023-01-01",
				TotalSales:       1000,
				TotalCost:        500,
				GrossProfit:      500,
				TotalExpense:     100,
				NetProfit:        400,
				TransactionCount: 10,
			},
		}, nil)

		resp, err := s.SalesSummary(ctx, filter)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, 1000.0, resp.TotalSales)
		assert.Equal(t, 400.0, resp.NetProfit)
		assert.Equal(t, 40.0, resp.ProfitMargin) // 400/1000 * 100
		rRepo.AssertExpectations(t)
	})
}

func TestReportService_StockValuation(t *testing.T) {
	ctx := context.Background()
	log := zerolog.Nop()

	t.Run("Success", func(t *testing.T) {
		rRepo := new(repomocks.ReportRepository)
		s := NewReportService(rRepo, log)

		rRepo.On("StockValuation", ctx, "s1").Return([]dto.StockValuationRow{
			{ProductID: "p1", TotalValue: 500},
			{ProductID: "p2", TotalValue: 300},
		}, nil)

		resp, err := s.StockValuation(ctx, "s1")

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, 800.0, resp.GrandTotal)
		rRepo.AssertExpectations(t)
	})
}
