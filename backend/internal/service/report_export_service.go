package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"

	"github.com/moedahpos/backend/internal/dto"
)

// ── CSV Helpers ───────────────────────────────────────────────────────────────

// writeCSV encodes a header row + data rows into a UTF-8 CSV byte slice.
func writeCSV(header []string, rows [][]string) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if err := w.Write(header); err != nil {
		return nil, fmt.Errorf("csv header: %w", err)
	}
	for _, r := range rows {
		if err := w.Write(r); err != nil {
			return nil, fmt.Errorf("csv row: %w", err)
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, fmt.Errorf("csv flush: %w", err)
	}
	return buf.Bytes(), nil
}

func fmtF(f float64) string { return strconv.FormatFloat(f, 'f', 2, 64) }
func fmtI(i int) string     { return strconv.Itoa(i) }

// ── Per-report CSV builders ───────────────────────────────────────────────────

func salesSummaryCSV(data *dto.SalesSummaryResponse) ([]byte, error) {
	header := []string{
		"Tanggal", "Jml Transaksi", "Total Penjualan",
		"Pajak", "Diskon", "HPP", "Laba Kotor", "Beban", "Laba Bersih",
	}
	rows := make([][]string, 0, len(data.Rows)+1)
	for _, r := range data.Rows {
		rows = append(rows, []string{
			r.Date, fmtI(r.TransactionCount),
			fmtF(r.TotalSales), fmtF(r.TotalTax),
			fmtF(r.TotalDiscount), fmtF(r.TotalCost),
			fmtF(r.GrossProfit), fmtF(r.TotalExpense),
			fmtF(r.NetProfit),
		})
	}
	rows = append(rows, []string{
		"TOTAL", fmtI(data.TotalTransactions),
		fmtF(data.TotalSales), "", "",
		fmtF(data.TotalCost), fmtF(data.GrossProfit),
		fmtF(data.TotalExpense), fmtF(data.NetProfit),
	})
	return writeCSV(header, rows)
}

func inventoryCSV(data *dto.StockValuationResponse) ([]byte, error) {
	header := []string{"Produk", "SKU", "Satuan", "Harga Pokok", "Stok", "Nilai Total"}
	rows := make([][]string, 0, len(data.Rows)+1)
	for _, r := range data.Rows {
		rows = append(rows, []string{
			r.ProductName, r.SKU, r.Unit,
			fmtF(r.CostPrice), fmtF(r.Quantity), fmtF(r.TotalValue),
		})
	}
	rows = append(rows, []string{"TOTAL", "", "", "", "", fmtF(data.GrandTotal)})
	return writeCSV(header, rows)
}

func profitCSV(data *dto.ProfitSummaryResponse) ([]byte, error) {
	header := []string{
		"Periode", "Total Penjualan", "HPP", "Laba Kotor",
		"Beban", "Laba Bersih", "Margin (%)",
	}
	rows := make([][]string, 0, len(data.Rows)+1)
	for _, r := range data.Rows {
		rows = append(rows, []string{
			r.Period, fmtF(r.TotalSales), fmtF(r.TotalCost),
			fmtF(r.GrossProfit), fmtF(r.TotalExpense),
			fmtF(r.NetProfit), fmtF(r.ProfitMargin),
		})
	}
	rows = append(rows, []string{
		"TOTAL",
		fmtF(data.TotalSales), fmtF(data.TotalCost),
		fmtF(data.GrossProfit), fmtF(data.TotalExpense),
		fmtF(data.NetProfit), fmtF(data.ProfitMargin),
	})
	return writeCSV(header, rows)
}

// ── PDF Generator ─────────────────────────────────────────────────────────────

type pdfData struct {
	Title      string
	Period     string
	ExportedAt string
	Headers    []string
	Rows       [][]string
	TotalRow   []string
}

// renderPDF generates an actual PDF using go-pdf/fpdf
func renderPDF(data pdfData) ([]byte, error) {
	// A4, portrait, mm
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	// Title
	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(0, 10, data.Title, "", 1, "L", false, 0, "")

	// Meta info
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(85, 85, 85)
	meta := fmt.Sprintf("Periode: %s   |   Diekspor: %s", data.Period, data.ExportedAt)
	pdf.CellFormat(0, 8, meta, "", 1, "L", false, 0, "")
	pdf.Ln(5)

	// Calculate column widths based on page width
	pageWidth, _ := pdf.GetPageSize()
	usableWidth := pageWidth - 30 // margins 15*2
	colCount := len(data.Headers)
	colWidth := usableWidth / float64(colCount)

	// Table Header
	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(26, 26, 46) // dark header
	pdf.SetTextColor(255, 255, 255)

	for _, h := range data.Headers {
		pdf.CellFormat(colWidth, 8, h, "1", 0, "L", true, 0, "")
	}
	pdf.Ln(-1)

	// Table Body
	pdf.SetFont("Arial", "", 8)
	pdf.SetTextColor(17, 17, 17)

	for _, row := range data.Rows {
		for _, cell := range row {
			// Limit cell text length so it doesn't overflow wildly
			if len(cell) > 20 && colCount > 5 {
				cell = cell[:17] + "..."
			}
			pdf.CellFormat(colWidth, 7, cell, "B", 0, "L", false, 0, "")
		}
		pdf.Ln(-1)
	}

	// Total Row
	if len(data.TotalRow) > 0 {
		pdf.SetFont("Arial", "B", 8)
		pdf.SetFillColor(249, 250, 251) // light gray
		for _, cell := range data.TotalRow {
			pdf.CellFormat(colWidth, 8, cell, "T", 0, "L", true, 0, "")
		}
		pdf.Ln(-1)
	}

	// Footer
	pdf.SetY(-20)
	pdf.SetFont("Arial", "I", 8)
	pdf.SetTextColor(156, 163, 175)
	pdf.CellFormat(0, 10, "Moedah POS -- laporan ini digenerate otomatis", "", 0, "C", false, 0, "")

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("pdf generation failed: %w", err)
	}

	return buf.Bytes(), nil
}

// periodLabel converts a ReportFilter into a human-readable date range string.
func periodLabel(filter dto.ReportFilter) string {
	if filter.DateFrom == "" && filter.DateTo == "" {
		return "30 hari terakhir"
	}
	parts := []string{}
	if filter.DateFrom != "" {
		parts = append(parts, filter.DateFrom)
	}
	if filter.DateTo != "" {
		parts = append(parts, filter.DateTo)
	}
	return strings.Join(parts, " s/d ")
}

// reportExportNow allows tests to override the current timestamp.
var reportExportNow = time.Now

// ── ExportCSV ────────────────────────────────────────────────────────────────

// ExportCSV generates a CSV byte slice for the given report type.
// report must be one of: "sales", "inventory", "profit".
func (s *ReportService) ExportCSV(ctx context.Context, report string, filter dto.ReportFilter) ([]byte, error) {
	switch report {
	case "sales":
		data, err := s.SalesSummary(ctx, filter)
		if err != nil {
			return nil, err
		}
		return salesSummaryCSV(data)
	case "inventory":
		data, err := s.StockValuation(ctx, filter.StoreID)
		if err != nil {
			return nil, err
		}
		return inventoryCSV(data)
	case "profit":
		data, err := s.ProfitSummary(ctx, filter, "day")
		if err != nil {
			return nil, err
		}
		return profitCSV(data)
	default:
		return nil, fmt.Errorf("unknown report type: %s", report)
	}
}

// ── ExportPDF ────────────────────────────────────────────────────────────────

// ExportPDF generates a true PDF byte slice for the given report type.
// report must be one of: "sales", "inventory", "profit".
//
//nolint:cyclop,funlen // switch over report types is inherently multi-branch and long
func (s *ReportService) ExportPDF(ctx context.Context, report string, filter dto.ReportFilter) ([]byte, error) {
	now := reportExportNow()
	exportedAt := now.Format("2006-01-02 15:04:05")
	period := periodLabel(filter)

	switch report {
	case "sales":
		data, err := s.SalesSummary(ctx, filter)
		if err != nil {
			return nil, err
		}
		headers := []string{
			"Tanggal", "Jml Txn", "Penjualan", "Pajak",
			"Diskon", "HPP", "Laba Kotor", "Beban", "Laba Bersih",
		}
		rows := make([][]string, 0, len(data.Rows))
		for _, r := range data.Rows {
			rows = append(rows, []string{
				r.Date, fmtI(r.TransactionCount),
				fmtF(r.TotalSales), fmtF(r.TotalTax),
				fmtF(r.TotalDiscount), fmtF(r.TotalCost),
				fmtF(r.GrossProfit), fmtF(r.TotalExpense), fmtF(r.NetProfit),
			})
		}
		totalRow := []string{
			"TOTAL", fmtI(data.TotalTransactions),
			fmtF(data.TotalSales), "", "",
			fmtF(data.TotalCost), fmtF(data.GrossProfit),
			fmtF(data.TotalExpense), fmtF(data.NetProfit),
		}
		return renderPDF(pdfData{
			Title: "Laporan Penjualan", Period: period,
			ExportedAt: exportedAt, Headers: headers,
			Rows: rows, TotalRow: totalRow,
		})

	case "inventory":
		data, err := s.StockValuation(ctx, filter.StoreID)
		if err != nil {
			return nil, err
		}
		headers := []string{"Produk", "SKU", "Satuan", "Harga Pokok", "Stok", "Nilai Total"}
		rows := make([][]string, 0, len(data.Rows))
		for _, r := range data.Rows {
			rows = append(rows, []string{
				r.ProductName, r.SKU, r.Unit,
				fmtF(r.CostPrice), fmtF(r.Quantity), fmtF(r.TotalValue),
			})
		}
		return renderPDF(pdfData{
			Title: "Audit Inventaris", Period: period,
			ExportedAt: exportedAt, Headers: headers,
			Rows:     rows,
			TotalRow: []string{"TOTAL", "", "", "", "", fmtF(data.GrandTotal)},
		})

	case "profit":
		data, err := s.ProfitSummary(ctx, filter, "day")
		if err != nil {
			return nil, err
		}
		headers := []string{
			"Periode", "Penjualan", "HPP", "Laba Kotor",
			"Beban", "Laba Bersih", "Margin (%)",
		}
		rows := make([][]string, 0, len(data.Rows))
		for _, r := range data.Rows {
			rows = append(rows, []string{
				r.Period, fmtF(r.TotalSales), fmtF(r.TotalCost),
				fmtF(r.GrossProfit), fmtF(r.TotalExpense),
				fmtF(r.NetProfit), fmtF(r.ProfitMargin),
			})
		}
		totalRow := []string{
			"TOTAL",
			fmtF(data.TotalSales), fmtF(data.TotalCost),
			fmtF(data.GrossProfit), fmtF(data.TotalExpense),
			fmtF(data.NetProfit), fmtF(data.ProfitMargin),
		}
		return renderPDF(pdfData{
			Title: "Laporan Laba Rugi", Period: period,
			ExportedAt: exportedAt, Headers: headers,
			Rows: rows, TotalRow: totalRow,
		})

	default:
		return nil, fmt.Errorf("unknown report type: %s", report)
	}
}
