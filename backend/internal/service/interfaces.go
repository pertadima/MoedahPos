package service

import (
	"context"

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/dto"
)

type AuthServiceInterface interface {
	Register(ctx context.Context, req *dto.RegisterRequest) (*dto.RegisterResponse, error)
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error)
	Refresh(ctx context.Context, req *dto.RefreshRequest) (*dto.RefreshResponse, error)
	Logout(ctx context.Context, userID string) error
	Me(ctx context.Context, userID string) (*dto.MeResponse, error)
}

type ActivityLogServiceInterface interface {
	LogActivity(ctx context.Context, userID, storeID string, action domain.ActionType, module domain.ActivityModule, refID string, metadata interface{})
	GetActivityLogs(ctx context.Context, storeID string, filter dto.ActivityLogFilter) ([]dto.ActivityLogResponse, dto.PaginationMeta, error)
}

type BatchStockServiceInterface interface {
	ReceivePurchaseOrder(ctx context.Context, poID, storeID string, items []dto.POBatchItem) error
	DeductStockFIFO(ctx context.Context, productID, storeID string, qty float64) error
	GetStockSummary(ctx context.Context, storeID string) ([]*domain.BatchStockSummary, error)
	GetBatchesByProduct(ctx context.Context, productID, storeID string) ([]*domain.StockBatch, error)
	GetBatchesByStore(ctx context.Context, f dto.BatchListFilter) ([]*domain.StockBatch, error)
}

type TransactionServiceInterface interface {
	Checkout(ctx context.Context, storeID string, req *dto.CreateTransactionRequest, cashierID string) (*dto.TransactionResponse, error)
	ListTransactions(ctx context.Context, filter dto.TransactionListFilter) ([]*dto.TransactionResponse, dto.PaginationMeta, error)
	GetTransaction(ctx context.Context, id string) (*dto.TransactionResponse, error)
	VoidTransaction(ctx context.Context, id, userID string) error
	GetDraftByTable(ctx context.Context, storeID, tableID string) (*dto.TransactionResponse, error)
	CreateDraft(ctx context.Context, storeID, cashierID string, req *dto.CreateDraftRequest) (*dto.TransactionResponse, error)
	UpdateDraftItems(ctx context.Context, storeID, txnID string, req *dto.UpdateDraftRequest) (*dto.TransactionResponse, error)
	PayDraft(ctx context.Context, storeID, txnID, cashierID string, req *dto.PayDraftRequest) (*dto.TransactionResponse, error)
	GetKDSTickets(ctx context.Context, storeID string) ([]*dto.TransactionResponse, error)
	UpdateKDSItemStatus(ctx context.Context, itemID string, req *dto.UpdateKDSItemStatusRequest) error
}

type PriceHistoryServiceInterface interface {
	RecordChange(ctx context.Context, productID, storeID, changedBy string, oldCost, newCost, oldSell, newSell float64, source string, refID, notes *string) error
	ListByProduct(ctx context.Context, productID string, f dto.PriceHistoryFilter) ([]*dto.PriceHistoryRow, dto.PaginationMeta, error)
	ListByStore(ctx context.Context, storeID string, f dto.PriceHistoryFilter) ([]*dto.PriceHistoryRow, dto.PaginationMeta, error)
}

type ProductServiceInterface interface {
	ListCategories(ctx context.Context, storeID string) ([]*dto.CategoryResponse, error)
	CreateCategory(ctx context.Context, storeID string, req *dto.CreateCategoryRequest) (*dto.CategoryResponse, error)
	UpdateCategory(ctx context.Context, id string, req *dto.UpdateCategoryRequest) (*dto.CategoryResponse, error)
	DeleteCategory(ctx context.Context, id string) error

	ListProducts(ctx context.Context, filter dto.ProductListFilter) ([]*dto.ProductResponse, dto.PaginationMeta, error)
	GetProduct(ctx context.Context, id string) (*dto.ProductResponse, error)
	GetProductByBarcode(ctx context.Context, storeID, barcode string) (*dto.ProductResponse, error)
	CreateProduct(ctx context.Context, storeID string, req *dto.CreateProductRequest, createdByID string) (*dto.ProductResponse, error)
	UpdateProduct(ctx context.Context, id string, req *dto.UpdateProductRequest, changedBy string) (*dto.ProductResponse, error)
	DeleteProduct(ctx context.Context, id string) error
}

type ReportServiceInterface interface {
	SalesSummary(ctx context.Context, filter dto.ReportFilter) (*dto.SalesSummaryResponse, error)
	SalesByProduct(ctx context.Context, filter dto.ReportFilter) ([]dto.SalesByProductRow, error)
	SalesByCashier(ctx context.Context, filter dto.ReportFilter) ([]dto.SalesByCashierRow, error)
	StockValuation(ctx context.Context, storeID string) (*dto.StockValuationResponse, error)
	ProfitSummary(ctx context.Context, filter dto.ReportFilter, groupBy string) (*dto.ProfitSummaryResponse, error)
	CashFlow(ctx context.Context, filter dto.ReportFilter) (*dto.CashFlowResponse, error)
	CashFlowDetail(ctx context.Context, storeID string, dateStr string) ([]dto.CashFlowDetailEntry, error)
}
