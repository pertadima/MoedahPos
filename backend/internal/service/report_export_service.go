package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"html/template"
	"strconv"
	"strings"
	"time"

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

// ── PDF (printable HTML) ─────────────────────────────────────────────────────

// pdfHTMLTemplate is a self-contained HTML document styled for browser print-to-PDF.
// Values are escaped by html/template automatically.
const pdfHTMLTemplate = `<!DOCTYPE html>
<html lang="id">
<head>
<meta charset="UTF-8"/>
<title>{{.Title}}</title>
<style>
*{box-sizing:border-box;margin:0;padding:0}
body{font-family:Arial,sans-serif;font-size:11px;color:#111;padding:20px}
h1{font-size:16px;margin-bottom:4px}
.meta{font-size:10px;color:#555;margin-bottom:16px}
table{width:100%;border-collapse:collapse;margin-bottom:16px}
th{background:#1a1a2e;color:#fff;padding:6px 8px;text-align:left;font-size:10px}
td{padding:5px 8px;border-bottom:1px solid #e5e7eb;font-size:10px}
tr:last-child td{border-bottom:none}
.total-row td{font-weight:bold;background:#f9fafb;border-top:2px solid #374151}
.footer{font-size:9px;color:#9ca3af;text-align:center;margin-top:20px}
@media print{body{padding:0}@page{margin:15mm}}
</style>
</head>
<body>
<h1>{{.Title}}</h1>
<div class="meta">Periode: {{.Period}} &nbsp;|&nbsp; Diekspor: {{.ExportedAt}}</div>
<table>
<thead><tr>{{range .Headers}}<th>{{.}}</th>{{end}}</tr></thead>
<tbody>
{{range .Rows}}<tr>{{range .}}<td>{{.}}</td>{{end}}</tr>
{{end -}}
{{- if .TotalRow}}<tr class="total-row">{{range .TotalRow}}<td>{{.}}</td>{{end}}</tr>{{end}}
</tbody>
</table>
<div class="footer">Moedah POS &mdash; laporan ini digenerate otomatis</div>
</body>
</html>`

type pdfData struct {
	Title      string
	Period     string
	ExportedAt string
	Headers    []string
	Rows       [][]string
	TotalRow   []string
}

// renderPDFHTML executes the PDF HTML template with the given data.
func renderPDFHTML(data pdfData) ([]byte, error) {
	tmpl, err := template.New("pdf").Parse(pdfHTMLTemplate)
	if err != nil {
		return nil, fmt.Errorf("parse pdf template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("render pdf template: %w", err)
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

// ExportPDF generates a printable HTML byte slice for the given report type.
// The client opens this in a browser; Ctrl+P → Save as PDF produces the document.
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
		return renderPDFHTML(pdfData{
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
		return renderPDFHTML(pdfData{
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
		return renderPDFHTML(pdfData{
			Title: "Laporan Laba Rugi", Period: period,
			ExportedAt: exportedAt, Headers: headers,
			Rows: rows, TotalRow: totalRow,
		})

	default:
		return nil, fmt.Errorf("unknown report type: %s", report)
	}
}
