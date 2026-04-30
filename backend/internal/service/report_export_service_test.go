package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/moedahpos/backend/internal/dto"
	repomocks "github.com/moedahpos/backend/internal/repository/mocks"
)

// ── Helpers ───────────────────────────────────────────────────────────────────

func newExportSvc(t *testing.T) (*ReportService, *repomocks.ReportRepository) {
	t.Helper()
	repo := new(repomocks.ReportRepository)
	svc := NewReportService(repo, zerolog.Nop())
	return svc, repo
}

func parseCSV(t *testing.T, data []byte) [][]string {
	t.Helper()
	r := csv.NewReader(bytes.NewReader(data))
	rows, err := r.ReadAll()
	assert.NoError(t, err)
	return rows
}

// ── writeCSV unit tests ───────────────────────────────────────────────────────

func TestWriteCSV_EmptyRows(t *testing.T) {
	data, err := writeCSV([]string{"A", "B"}, nil)
	assert.NoError(t, err)
	rows := parseCSV(t, data)
	assert.Len(t, rows, 1, "only header row expected")
	assert.Equal(t, []string{"A", "B"}, rows[0])
}

func TestWriteCSV_CommasAndQuotes(t *testing.T) {
	// Fields with commas and double-quotes must be properly escaped by encoding/csv.
	data, err := writeCSV(
		[]string{"Name", "Value"},
		[][]string{
			{`Product, with comma`, `123.45`},
			{`He said "hello"`, `67.89`},
			{`Comma,"and quote"`, `0.00`},
		},
	)
	assert.NoError(t, err)

	rows := parseCSV(t, data)
	assert.Len(t, rows, 4) // header + 3 data rows
	assert.Equal(t, "Product, with comma", rows[1][0])
	assert.Equal(t, `He said "hello"`, rows[2][0])
	assert.Equal(t, `Comma,"and quote"`, rows[3][0])
}

func TestWriteCSV_NewlinesInField(t *testing.T) {
	data, err := writeCSV(
		[]string{"Notes"},
		[][]string{{"line1\nline2"}},
	)
	assert.NoError(t, err)
	rows := parseCSV(t, data)
	assert.Len(t, rows, 2)
	assert.Equal(t, "line1\nline2", rows[1][0])
}

// ── periodLabel ───────────────────────────────────────────────────────────────

func TestPeriodLabel(t *testing.T) {
	assert.Equal(t, "30 hari terakhir", periodLabel(dto.ReportFilter{}))
	assert.Equal(t, "2024-01-01", periodLabel(dto.ReportFilter{DateFrom: "2024-01-01"}))
	assert.Equal(t, "2024-01-31", periodLabel(dto.ReportFilter{DateTo: "2024-01-31"}))
	assert.Equal(t, "2024-01-01 s/d 2024-01-31", periodLabel(dto.ReportFilter{
		DateFrom: "2024-01-01", DateTo: "2024-01-31",
	}))
}

// ── renderPDFHTML ─────────────────────────────────────────────────────────────

func TestRenderPDFHTML_EmptyData(t *testing.T) {
	// Must not panic or error on empty rows / nil TotalRow.
	data, err := renderPDFHTML(pdfData{
		Title:      "Test",
		Period:     "30 hari terakhir",
		ExportedAt: "2024-01-01 00:00:00",
		Headers:    []string{"A", "B"},
		Rows:       nil,
		TotalRow:   nil,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
	html := string(data)
	assert.Contains(t, html, "Test")
	assert.Contains(t, html, "<table>")
}

func TestRenderPDFHTML_XSSEscape(t *testing.T) {
	// html/template must escape user-supplied values.
	data, err := renderPDFHTML(pdfData{
		Title:      "<script>alert(1)</script>",
		Period:     "2024-01-01",
		ExportedAt: "2024-01-01",
		Headers:    []string{"<h1>"},
		Rows:       [][]string{{"<b>value</b>"}},
		TotalRow:   []string{"<em>total</em>"},
	})
	assert.NoError(t, err)
	html := string(data)
	// Raw script tag must not appear
	assert.NotContains(t, html, "<script>alert(1)</script>")
	// Escaped form must appear
	assert.Contains(t, html, "&lt;script&gt;")
}

func TestRenderPDFHTML_TotalRowRendered(t *testing.T) {
	data, err := renderPDFHTML(pdfData{
		Title: "T", Period: "P", ExportedAt: "E",
		Headers:  []string{"X"},
		Rows:     [][]string{{"row1"}},
		TotalRow: []string{"total1"},
	})
	assert.NoError(t, err)
	html := string(data)
	assert.Contains(t, html, "total-row")
	assert.Contains(t, html, "total1")
}

// ── ExportCSV ─────────────────────────────────────────────────────────────────

func TestExportCSV_Sales(t *testing.T) {
	svc, repo := newExportSvc(t)
	ctx := context.Background()
	filter := dto.ReportFilter{StoreID: "s1", DateFrom: "2024-01-01", DateTo: "2024-01-31"}

	rows := []dto.SalesSummaryRow{
		{Date: "2024-01-01", TransactionCount: 3, TotalSales: 1500, TotalCost: 900, GrossProfit: 600, NetProfit: 500},
	}
	repo.On("SalesSummary", ctx, "s1", mock.Anything, mock.Anything).Return(rows, nil).Once()

	data, err := svc.ExportCSV(ctx, "sales", filter)
	assert.NoError(t, err)

	parsed := parseCSV(t, data)
	// header + 1 data row + 1 total row
	assert.Len(t, parsed, 3)
	assert.Equal(t, "Tanggal", parsed[0][0])
	assert.Equal(t, "2024-01-01", parsed[1][0])
	assert.Equal(t, "TOTAL", parsed[2][0])
	repo.AssertExpectations(t)
}

func TestExportCSV_Inventory(t *testing.T) {
	svc, repo := newExportSvc(t)
	ctx := context.Background()
	filter := dto.ReportFilter{StoreID: "s1"}

	valRows := []dto.StockValuationRow{
		{ProductName: "Beras", SKU: "BR001", Unit: "kg", CostPrice: 10000, Quantity: 5, TotalValue: 50000},
	}
	repo.On("StockValuation", ctx, "s1").Return(valRows, nil).Once()

	data, err := svc.ExportCSV(ctx, "inventory", filter)
	assert.NoError(t, err)

	parsed := parseCSV(t, data)
	assert.Len(t, parsed, 3) // header + 1 row + total
	assert.Equal(t, "Produk", parsed[0][0])
	assert.Equal(t, "Beras", parsed[1][0])
	repo.AssertExpectations(t)
}

func TestExportCSV_Profit(t *testing.T) {
	svc, repo := newExportSvc(t)
	ctx := context.Background()
	filter := dto.ReportFilter{StoreID: "s1"}

	profitRows := []dto.ProfitPeriodRow{
		{Period: "2024-01-01", TotalSales: 1000, GrossProfit: 400, NetProfit: 300, ProfitMargin: 30},
	}
	repo.On("ProfitSummary", ctx, "s1", mock.Anything, mock.Anything, "day").Return(profitRows, nil).Once()

	data, err := svc.ExportCSV(ctx, "profit", filter)
	assert.NoError(t, err)

	parsed := parseCSV(t, data)
	assert.Len(t, parsed, 3)
	assert.Equal(t, "Periode", parsed[0][0])
	repo.AssertExpectations(t)
}

func TestExportCSV_UnknownReport(t *testing.T) {
	svc, _ := newExportSvc(t)
	_, err := svc.ExportCSV(context.Background(), "unknown", dto.ReportFilter{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown report type")
}

func TestExportCSV_ServiceError(t *testing.T) {
	svc, repo := newExportSvc(t)
	ctx := context.Background()
	repo.On("SalesSummary", ctx, "", mock.Anything, mock.Anything).
		Return(nil, errors.New("db error")).Once()

	_, err := svc.ExportCSV(ctx, "sales", dto.ReportFilter{})
	assert.Error(t, err)
	repo.AssertExpectations(t)
}

// ── ExportPDF ─────────────────────────────────────────────────────────────────

func TestExportPDF_Sales(t *testing.T) {
	svc, repo := newExportSvc(t)
	ctx := context.Background()
	filter := dto.ReportFilter{StoreID: "s1", DateFrom: "2024-01-01", DateTo: "2024-01-31"}

	rows := []dto.SalesSummaryRow{
		{Date: "2024-01-01", TransactionCount: 2, TotalSales: 2000, GrossProfit: 800, NetProfit: 600},
	}
	repo.On("SalesSummary", ctx, "s1", mock.Anything, mock.Anything).Return(rows, nil).Once()

	data, err := svc.ExportPDF(ctx, "sales", filter)
	assert.NoError(t, err)
	html := string(data)
	assert.Contains(t, html, "Laporan Penjualan")
	assert.Contains(t, html, "2024-01-01")
	assert.Contains(t, html, "<!DOCTYPE html>")
	repo.AssertExpectations(t)
}

func TestExportPDF_Inventory(t *testing.T) {
	svc, repo := newExportSvc(t)
	ctx := context.Background()
	filter := dto.ReportFilter{StoreID: "s1"}

	valRows := []dto.StockValuationRow{
		{ProductName: "Gula", SKU: "GU001", Unit: "kg", CostPrice: 12000, Quantity: 10, TotalValue: 120000},
	}
	repo.On("StockValuation", ctx, "s1").Return(valRows, nil).Once()

	data, err := svc.ExportPDF(ctx, "inventory", filter)
	assert.NoError(t, err)
	html := string(data)
	assert.Contains(t, html, "Audit Inventaris")
	assert.Contains(t, html, "Gula")
	repo.AssertExpectations(t)
}

func TestExportPDF_Profit(t *testing.T) {
	svc, repo := newExportSvc(t)
	ctx := context.Background()
	filter := dto.ReportFilter{StoreID: "s1", DateFrom: "2024-01-01"}

	profitRows := []dto.ProfitPeriodRow{
		{Period: "2024-01-01", TotalSales: 5000, GrossProfit: 2000, NetProfit: 1500, ProfitMargin: 30},
	}
	repo.On("ProfitSummary", ctx, "s1", mock.Anything, mock.Anything, "day").Return(profitRows, nil).Once()

	data, err := svc.ExportPDF(ctx, "profit", filter)
	assert.NoError(t, err)
	html := string(data)
	assert.Contains(t, html, "Laporan Laba Rugi")
	repo.AssertExpectations(t)
}

func TestExportPDF_EmptyDataNoError(t *testing.T) {
	// Empty data (no rows) must not cause a panic or error.
	svc, repo := newExportSvc(t)
	ctx := context.Background()
	filter := dto.ReportFilter{StoreID: "s1"}

	repo.On("SalesSummary", ctx, "s1", mock.Anything, mock.Anything).
		Return([]dto.SalesSummaryRow{}, nil).Once()

	data, err := svc.ExportPDF(ctx, "sales", filter)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
	assert.Contains(t, string(data), "<table>")
	repo.AssertExpectations(t)
}

func TestExportPDF_UnknownReport(t *testing.T) {
	svc, _ := newExportSvc(t)
	_, err := svc.ExportPDF(context.Background(), "accounting", dto.ReportFilter{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown report type")
}

func TestExportPDF_ServiceError(t *testing.T) {
	svc, repo := newExportSvc(t)
	ctx := context.Background()
	repo.On("StockValuation", ctx, "").Return(nil, errors.New("db fail")).Once()

	_, err := svc.ExportPDF(ctx, "inventory", dto.ReportFilter{})
	assert.Error(t, err)
	repo.AssertExpectations(t)
}

// ── ExportReport handler tests (via report_handler_test.go pattern) ───────────
// Those live in the handler package; here we just verify the service layer.

// TestReportExportNow verifies the reportExportNow override works for tests.
func TestReportExportNow(t *testing.T) {
	fixed := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	orig := reportExportNow
	reportExportNow = func() time.Time { return fixed }
	defer func() { reportExportNow = orig }()

	assert.Equal(t, "2024-01-15 10:00:00", reportExportNow().Format("2006-01-02 15:04:05"))
}

// ── salesSummaryCSV edge cases ────────────────────────────────────────────────

func TestSalesSummaryCSV_EmptyRows(t *testing.T) {
	data, err := salesSummaryCSV(&dto.SalesSummaryResponse{})
	assert.NoError(t, err)
	parsed := parseCSV(t, data)
	// header + 1 totals row (all zeros)
	assert.Len(t, parsed, 2)
	assert.Equal(t, "TOTAL", parsed[1][0])
}

func TestInventoryCSV_EmptyRows(t *testing.T) {
	data, err := inventoryCSV(&dto.StockValuationResponse{GrandTotal: 0})
	assert.NoError(t, err)
	parsed := parseCSV(t, data)
	assert.Len(t, parsed, 2) // header + total
	assert.Equal(t, "TOTAL", parsed[1][0])
}

func TestProfitCSV_EmptyRows(t *testing.T) {
	data, err := profitCSV(&dto.ProfitSummaryResponse{})
	assert.NoError(t, err)
	parsed := parseCSV(t, data)
	assert.Len(t, parsed, 2)
	assert.Equal(t, "TOTAL", parsed[1][0])
}

// TestCSV_ProductNameWithComma verifies product names containing commas round-trip correctly.
func TestCSV_ProductNameWithComma(t *testing.T) {
	data, err := inventoryCSV(&dto.StockValuationResponse{
		Rows: []dto.StockValuationRow{
			{ProductName: "Kopi, Arabika", SKU: "KA001", Unit: "kg", CostPrice: 50000, Quantity: 2, TotalValue: 100000},
		},
		GrandTotal: 100000,
	})
	assert.NoError(t, err)
	parsed := parseCSV(t, data)
	assert.Equal(t, "Kopi, Arabika", parsed[1][0], "comma in product name must survive CSV round-trip")
}

// TestCSV_ProductNameWithQuote verifies double-quote escaping.
func TestCSV_ProductNameWithQuote(t *testing.T) {
	data, err := inventoryCSV(&dto.StockValuationResponse{
		Rows: []dto.StockValuationRow{
			{ProductName: `Gula "Premium"`, SKU: "GP001", Unit: "kg"},
		},
	})
	assert.NoError(t, err)
	// Verify it parses back correctly via standard csv reader
	r := csv.NewReader(strings.NewReader(string(data)))
	records, err := r.ReadAll()
	assert.NoError(t, err)
	assert.Equal(t, `Gula "Premium"`, records[1][0])
}
