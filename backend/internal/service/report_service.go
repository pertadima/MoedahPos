package service

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"

	"github.com/moedahpos/backend/internal/dto"
	"github.com/moedahpos/backend/internal/repository"
)

// ReportService implements analytics and reporting business logic.
type ReportService struct {
	reportRepo repository.ReportRepository
	log        zerolog.Logger
}

func NewReportService(reportRepo repository.ReportRepository, log zerolog.Logger) *ReportService {
	return &ReportService{reportRepo: reportRepo, log: log}
}

// defaultDateRange returns (30 days ago, tomorrow) when from/to are empty.
func defaultDateRange(filter dto.ReportFilter) (time.Time, time.Time) {
	layout := "2006-01-02"
	now := time.Now().UTC()
	to := now.AddDate(0, 0, 1).Truncate(24 * time.Hour) // exclusive upper bound

	if filter.DateFrom == "" && filter.DateTo == "" {
		from := now.AddDate(0, 0, -30).Truncate(24 * time.Hour)
		return from, to
	}

	from := now.AddDate(0, 0, -30).Truncate(24 * time.Hour)
	if filter.DateFrom != "" {
		if t, err := time.Parse(layout, filter.DateFrom); err == nil {
			from = t
		}
	}
	if filter.DateTo != "" {
		if t, err := time.Parse(layout, filter.DateTo); err == nil {
			to = t.AddDate(0, 0, 1) // make exclusive
		}
	}
	return from, to
}

// SalesSummary returns daily aggregated sales figures for a date range.
func (s *ReportService) SalesSummary(ctx context.Context, filter dto.ReportFilter) (*dto.SalesSummaryResponse, error) {
	from, to := defaultDateRange(filter)
	rows, err := s.reportRepo.SalesSummary(ctx, filter.StoreID, from, to)
	if err != nil {
		return nil, fmt.Errorf("sales summary: %w", err)
	}

	var totalSales, totalCost, grossProfit, totalExpense, netProfit float64
	var totalTxns int
	for _, r := range rows {
		totalSales += r.TotalSales
		totalCost += r.TotalCost
		grossProfit += r.GrossProfit
		totalExpense += r.TotalExpense
		netProfit += r.NetProfit
		totalTxns += r.TransactionCount
	}
	var margin float64
	if totalSales > 0 {
		margin = netProfit / totalSales * 100
	}
	return &dto.SalesSummaryResponse{
		Rows:              rows,
		TotalSales:        totalSales,
		TotalTransactions: totalTxns,
		TotalCost:         totalCost,
		GrossProfit:       grossProfit,
		TotalExpense:      totalExpense,
		NetProfit:         netProfit,
		ProfitMargin:      margin,
	}, nil
}

// SalesByProduct returns revenue ranked by product within a date range.
func (s *ReportService) SalesByProduct(ctx context.Context, filter dto.ReportFilter) ([]dto.SalesByProductRow, error) {
	from, to := defaultDateRange(filter)
	rows, err := s.reportRepo.SalesByProduct(ctx, filter.StoreID, from, to)
	if err != nil {
		return nil, fmt.Errorf("sales by product: %w", err)
	}
	return rows, nil
}

// SalesByCashier returns revenue ranked by cashier within a date range.
func (s *ReportService) SalesByCashier(ctx context.Context, filter dto.ReportFilter) ([]dto.SalesByCashierRow, error) {
	from, to := defaultDateRange(filter)
	rows, err := s.reportRepo.SalesByCashier(ctx, filter.StoreID, from, to)
	if err != nil {
		return nil, fmt.Errorf("sales by cashier: %w", err)
	}
	return rows, nil
}

// StockValuation returns current stock value (quantity × cost_price) per product.
func (s *ReportService) StockValuation(ctx context.Context, storeID string) (*dto.StockValuationResponse, error) {
	rows, err := s.reportRepo.StockValuation(ctx, storeID)
	if err != nil {
		return nil, fmt.Errorf("stock valuation: %w", err)
	}
	var grandTotal float64
	for _, r := range rows {
		grandTotal += r.TotalValue
	}
	return &dto.StockValuationResponse{
		Rows:       rows,
		GrandTotal: grandTotal,
	}, nil
}

// ProfitSummary returns gross profit grouped by day, week, or month.
func (s *ReportService) ProfitSummary(ctx context.Context, filter dto.ReportFilter, groupBy string) (*dto.ProfitSummaryResponse, error) {
	from, to := defaultDateRange(filter)
	rows, err := s.reportRepo.ProfitSummary(ctx, filter.StoreID, from, to, groupBy)
	if err != nil {
		return nil, fmt.Errorf("profit summary: %w", err)
	}
	var totalSales, totalCost, grossProfit, totalExpense, netProfit float64
	for _, r := range rows {
		totalSales += r.TotalSales
		totalCost += r.TotalCost
		grossProfit += r.GrossProfit
		totalExpense += r.TotalExpense
		netProfit += r.NetProfit
	}
	var margin float64
	if totalSales > 0 {
		margin = netProfit / totalSales * 100
	}
	return &dto.ProfitSummaryResponse{
		Rows:         rows,
		TotalSales:   totalSales,
		TotalCost:    totalCost,
		GrossProfit:  grossProfit,
		TotalExpense: totalExpense,
		NetProfit:    netProfit,
		ProfitMargin: margin,
	}, nil
}
