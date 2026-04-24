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

func TestReportService(t *testing.T) {
	repo := new(repomocks.ReportRepository)
	log := zerolog.Nop()
	svc := NewReportService(repo, log)

	ctx := context.Background()

	t.Run("SalesSummary", func(t *testing.T) {
		filter := dto.ReportFilter{StoreID: "s1", DateFrom: "2024-01-01", DateTo: "2024-01-02"}
		rows := []dto.SalesSummaryRow{
			{Date: "2024-01-01", TotalSales: 1000, TotalCost: 600, GrossProfit: 400, NetProfit: 350, TransactionCount: 5},
		}
		repo.On("SalesSummary", ctx, "s1", mock.Anything, mock.Anything).Return(rows, nil).Once()

		resp, err := svc.SalesSummary(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, 1000.0, resp.TotalSales)
		assert.Equal(t, 350.0, resp.NetProfit)
		assert.Equal(t, 5, resp.TotalTransactions)
		repo.AssertExpectations(t)
	})

	t.Run("StockValuation", func(t *testing.T) {
		rows := []dto.StockValuationRow{
			{ProductID: "p1", ProductName: "P1", Quantity: 10, CostPrice: 100, TotalValue: 1000},
			{ProductID: "p2", ProductName: "P2", Quantity: 5, CostPrice: 200, TotalValue: 1000},
		}
		repo.On("StockValuation", ctx, "s1").Return(rows, nil).Once()

		resp, err := svc.StockValuation(ctx, "s1")
		assert.NoError(t, err)
		assert.Equal(t, 2000.0, resp.GrandTotal)
		assert.Len(t, resp.Rows, 2)
	})

	t.Run("CashFlow", func(t *testing.T) {
		filter := dto.ReportFilter{StoreID: "s1"}
		rows := []dto.CashFlowDayRow{
			{
				Date:   "2024-01-01",
				CashIn: 500, CashOut: 200,
				CashInByMethod: map[string]float64{"cash": 300, "bank": 200},
			},
		}
		repo.On("CashFlowSummary", ctx, "s1", mock.Anything, mock.Anything).Return(rows, nil).Once()

		resp, err := svc.CashFlow(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, 500.0, resp.TotalCashIn)
		assert.Equal(t, 300.0, resp.NetCash)
		assert.Equal(t, 300.0, resp.CashInByMethod["cash"])
	})

	t.Run("SalesByProduct", func(t *testing.T) {
		filter := dto.ReportFilter{StoreID: "s1"}
		repo.On("SalesByProduct", ctx, "s1", mock.Anything, mock.Anything).Return([]dto.SalesByProductRow{{ProductName: "P1", TotalQuantity: 10}}, nil).Once()
		res, err := svc.SalesByProduct(ctx, filter)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run("SalesByCashier", func(t *testing.T) {
		filter := dto.ReportFilter{StoreID: "s1"}
		repo.On("SalesByCashier", ctx, "s1", mock.Anything, mock.Anything).Return([]dto.SalesByCashierRow{{CashierName: "U1", TotalSales: 100}}, nil).Once()
		res, err := svc.SalesByCashier(ctx, filter)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run("ProfitSummary", func(t *testing.T) {
		filter := dto.ReportFilter{StoreID: "s1"}
		repo.On("ProfitSummary", ctx, "s1", mock.Anything, mock.Anything, "day").Return([]dto.ProfitPeriodRow{{Period: "2024-01-01", GrossProfit: 100}}, nil).Once()
		res, err := svc.ProfitSummary(ctx, filter, "day")
		assert.NoError(t, err)
		assert.Len(t, res.Rows, 1)
	})

	t.Run("CashFlowDetail", func(t *testing.T) {
		repo.On("CashFlowDetail", ctx, "s1", mock.Anything, mock.Anything).Return([]dto.CashFlowDetailEntry{{Amount: 100}}, nil).Once()
		res, err := svc.CashFlowDetail(ctx, "s1", "2024-01-01")
		assert.NoError(t, err)
		assert.Len(t, res, 1)
	})
}
