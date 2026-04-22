package service

import (
	"context"
	"time"

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

type StockServiceInterface interface {
	GetStockLevels(ctx context.Context, storeID string, lowStockOnly bool) ([]*dto.StockLevelResponse, error)
	GetProductStock(ctx context.Context, productID, storeID string) (*dto.StockLevelResponse, error)
	AdjustStock(ctx context.Context, storeID string, req *dto.AdjustStockRequest, userID string) (*dto.StockLevelResponse, error)
	SetMinStock(ctx context.Context, storeID string, req *dto.SetMinStockRequest) error
	GetMovements(ctx context.Context, filter dto.StockMovementFilter) ([]*dto.StockMovementResponse, dto.PaginationMeta, error)
}

type CustomerServiceInterface interface {
	List(ctx context.Context, f dto.CustomerListFilter) ([]*dto.CustomerResponse, dto.PaginationMeta, error)
	Get(ctx context.Context, id string) (*dto.CustomerResponse, error)
	Create(ctx context.Context, storeID string, req dto.CreateCustomerRequest) (*dto.CustomerResponse, error)
	Update(ctx context.Context, id string, req dto.UpdateCustomerRequest) (*dto.CustomerResponse, error)
	Delete(ctx context.Context, id string) error
	Search(ctx context.Context, f dto.CustomerListFilter) ([]*dto.CustomerResponse, error)
}

type StockAdjustmentServiceInterface interface {
	CreateAdjustment(ctx context.Context, storeID, userID string, input domain.CreateAdjustmentInput) error
	GetAdjustmentHistory(ctx context.Context, storeID string, productID *string) ([]*domain.StockAdjustment, error)
}

type SupplierServiceInterface interface {
	ListSuppliers(ctx context.Context, filter dto.SupplierListFilter) ([]*dto.SupplierResponse, dto.PaginationMeta, error)
	GetSupplier(ctx context.Context, id string) (*dto.SupplierResponse, error)
	CreateSupplier(ctx context.Context, req *dto.CreateSupplierRequest) (*dto.SupplierResponse, error)
	UpdateSupplier(ctx context.Context, id string, req *dto.UpdateSupplierRequest) (*dto.SupplierResponse, error)
	DeleteSupplier(ctx context.Context, id string) error
}

type ExpenseServiceInterface interface {
	ListCategories(ctx context.Context, includeDeleted bool) ([]dto.ExpenseCategoryResponse, error)
	CreateCategory(ctx context.Context, req *dto.CreateExpenseCategoryRequest) (*dto.ExpenseCategoryResponse, error)
	UpdateCategory(ctx context.Context, id string, req *dto.UpdateExpenseCategoryRequest) (*dto.ExpenseCategoryResponse, error)
	SoftDeleteCategory(ctx context.Context, id string) error

	CreateExpense(ctx context.Context, storeID, userID string, req *dto.CreateExpenseRequest) (*dto.ExpenseResponse, error)
	ListExpenses(ctx context.Context, filter dto.ExpenseListFilter) ([]*dto.ExpenseResponse, dto.PaginationMeta, error)
	UpdateExpense(ctx context.Context, id, storeID string, req *dto.UpdateExpenseRequest) (*dto.ExpenseResponse, error)
	DeleteExpense(ctx context.Context, id, storeID string) error
	UpdatePaymentStatus(ctx context.Context, id, storeID string, req *dto.UpdateExpenseStatusRequest) (*dto.ExpenseResponse, error)

	CreateRecurringExpense(ctx context.Context, storeID, userID string, req *dto.CreateRecurringExpenseRequest) (*dto.RecurringExpenseResponse, error)
	ListRecurringExpenses(ctx context.Context, filter dto.ExpenseListFilter) ([]*dto.RecurringExpenseResponse, dto.PaginationMeta, error)
	UpdateRecurringExpense(ctx context.Context, id, storeID string, req *dto.UpdateRecurringExpenseRequest) (*dto.RecurringExpenseResponse, error)
	DeleteRecurringExpense(ctx context.Context, id, storeID string) error
	ProcessDueRecurringExpenses(ctx context.Context) error
}

type IncomeServiceInterface interface {
	ListCategories(ctx context.Context, includeDeleted bool) ([]*dto.IncomeCategoryResponse, error)
	CreateCategory(ctx context.Context, req *dto.CreateIncomeCategoryRequest) (*dto.IncomeCategoryResponse, error)
	UpdateCategory(ctx context.Context, id string, req *dto.UpdateIncomeCategoryRequest) (*dto.IncomeCategoryResponse, error)
	SoftDeleteCategory(ctx context.Context, id string) error

	CreateIncome(ctx context.Context, storeID, userID string, req *dto.CreateIncomeRequest) (*dto.IncomeResponse, error)
	ListIncomes(ctx context.Context, f dto.IncomeListFilter) ([]*dto.IncomeResponse, dto.PaginationMeta, error)
	UpdateIncome(ctx context.Context, id, storeID string, req *dto.UpdateIncomeRequest) (*dto.IncomeResponse, error)
	DeleteIncome(ctx context.Context, id, storeID string) error
}

type TerminServiceInterface interface {
	CreateTerminSchedule(ctx context.Context, poID string, req dto.CreateTerminScheduleRequest) ([]dto.TerminResponse, error)
	GetTerminSchedule(ctx context.Context, poID string) ([]dto.TerminResponse, error)
	RecordPayment(ctx context.Context, terminID, userID string, req dto.RecordPaymentRequest) (*dto.PaymentRecordResponse, error)
	CalculatePODebt(ctx context.Context, poID string) (*dto.PODebtSummaryResponse, error)
	GenerateDocumentData(ctx context.Context, poID, docType string) (*dto.PODocumentData, error)
}

type PurchaseOrderServiceInterface interface {
	ListPOs(ctx context.Context, filter dto.POListFilter) ([]*dto.POResponse, dto.PaginationMeta, error)
	GetPO(ctx context.Context, id string) (*dto.POResponse, error)
	CreatePO(ctx context.Context, storeID string, req *dto.CreatePORequest, userID string) (*dto.POResponse, error)
	UpdatePO(ctx context.Context, id string, req *dto.UpdatePORequest, storeID string) (*dto.POResponse, error)
	SubmitPO(ctx context.Context, id, userID string) error
	ReceivePO(ctx context.Context, id, userID string) error
	CancelPO(ctx context.Context, id string) error
	CreatePayment(ctx context.Context, poID, storeID, userID string, req dto.POPaymentRequest) (*dto.POPaymentResponse, error)
	ListPayments(ctx context.Context, poID string) ([]*dto.POPaymentResponse, error)
	PayableSummary(ctx context.Context, storeID string) (*dto.PayableSummary, error)
}

type UserAdminServiceInterface interface {
	ListUsers(ctx context.Context, search string, includeInactive bool, page, perPage int) ([]dto.UserResponse, int, error)
	GetUser(ctx context.Context, id string) (*dto.UserResponse, error)
	CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error)
	UpdateUser(ctx context.Context, id string, req *dto.UpdateUserRequest) (*dto.UserResponse, error)
	DeactivateUser(ctx context.Context, id string) error
	ResetPassword(ctx context.Context, id string, req *dto.ResetPasswordRequest) error
	SetUserStores(ctx context.Context, userID string, req *dto.SetUserStoresRequest) (*dto.UserResponse, error)
	ListRoles(ctx context.Context) ([]*domain.Role, error)
}

type StoreServiceInterface interface {
	GetStore(ctx context.Context, id string) (*dto.StoreResponse, error)
	CreateStore(ctx context.Context, req *dto.CreateStoreRequest) (*dto.StoreResponse, error)
	UpdateStore(ctx context.Context, id string, req *dto.UpdateStoreRequest) (*dto.StoreResponse, error)
	DeleteStore(ctx context.Context, id string) error
	ListStores(ctx context.Context, f dto.StoreListFilter) ([]*dto.StoreResponse, dto.PaginationMeta, error)

	ListMembers(ctx context.Context, storeID string) ([]*dto.MemberResponse, error)
	AddMember(ctx context.Context, storeID string, req *dto.AddMemberRequest) error
	UpdateMemberRole(ctx context.Context, storeID, userID string, req *dto.UpdateMemberRoleRequest) error
	RemoveMember(ctx context.Context, storeID, userID string) error
}

type SyncServiceInterface interface {
	Pull(ctx context.Context, storeID string, since time.Time) (*dto.SyncPullOutput, error)
}

type LoyaltyServiceInterface interface {
	ListTiers(ctx context.Context) ([]*dto.MembershipTierResponse, error)
	GetBalance(ctx context.Context, customerID string) (*dto.LoyaltyBalanceResponse, error)
	EarnPoints(ctx context.Context, storeID, customerID string, transactionID *string, total float64) (*dto.LoyaltyLedgerResponse, error)
	RedeemPoints(ctx context.Context, customerID string, transactionID *string, points float64) (*dto.LoyaltyLedgerResponse, error)
	GetHistory(ctx context.Context, customerID string) ([]*dto.LoyaltyLedgerResponse, error)
	AssignTier(ctx context.Context, customerID, tierID string) error
}
