package repository

import (
	"context"
	"time"

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/dto"
)

// ─── Phase 1 ──────────────────────────────────────────────────────────────────

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) (*domain.User, error)
	FindByID(ctx context.Context, id string) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	FindStoresByUserID(ctx context.Context, userID string) ([]domain.UserStore, error)
	// Admin operations
	ListAll(ctx context.Context, search string, includeInactive bool, page, perPage int) ([]*domain.User, int, error)
	Update(ctx context.Context, id, name, email string) (*domain.User, error)
	SoftDelete(ctx context.Context, id string) error
	Deactivate(ctx context.Context, id string) error
	ResetPassword(ctx context.Context, id, hash string) error
	SetStores(ctx context.Context, userID string, assignments []domain.StoreAssignment) error
}

type RoleRepository interface {
	ListRoles(ctx context.Context) ([]*domain.Role, error)
	CreateRole(ctx context.Context, role *domain.Role, permissionIDs []string) (*domain.Role, error)
	UpdateRole(ctx context.Context, role *domain.Role, permissionIDs []string) (*domain.Role, error)
	DeleteRole(ctx context.Context, id string) error
	ListPermissions(ctx context.Context) ([]*domain.Permission, error)
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *domain.RefreshToken) error
	FindByHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error)
	RevokeByID(ctx context.Context, id string) error
	RevokeAllByUserID(ctx context.Context, userID string) error
	DeleteExpired(ctx context.Context, before time.Time) error
}

// ─── Phase 2 ──────────────────────────────────────────────────────────────────

type StoreRepository interface {
	Create(ctx context.Context, store *domain.Store) (*domain.Store, error)
	FindAll(ctx context.Context, filter dto.StoreListFilter) ([]*domain.Store, int, error)
	FindByID(ctx context.Context, id string) (*domain.Store, error)
	Update(ctx context.Context, store *domain.Store) (*domain.Store, error)
	SoftDelete(ctx context.Context, id string) error
	FindMember(ctx context.Context, userID, storeID string) (*domain.UserStore, error)
	ListMembers(ctx context.Context, storeID string) ([]*domain.StoreMember, error)
	AddMember(ctx context.Context, us *domain.UserStore) error
	UpdateMemberRole(ctx context.Context, userID, storeID, roleID string) error
	DeactivateMember(ctx context.Context, userID, storeID string) error
}

type CategoryRepository interface {
	Create(ctx context.Context, cat *domain.Category) (*domain.Category, error)
	FindAllByStore(ctx context.Context, storeID string) ([]*domain.Category, error)
	FindByID(ctx context.Context, id string) (*domain.Category, error)
	Update(ctx context.Context, cat *domain.Category) (*domain.Category, error)
	SoftDelete(ctx context.Context, id string) error
	GetModifiedSince(ctx context.Context, storeID string, since time.Time) ([]*domain.Category, error)
}

type ProductRepository interface {
	Create(ctx context.Context, p *domain.Product) (*domain.Product, error)
	FindAll(ctx context.Context, filter dto.ProductListFilter) ([]*domain.Product, int, error)
	FindByID(ctx context.Context, id string) (*domain.Product, error)
	FindByBarcode(ctx context.Context, storeID, barcode string) (*domain.Product, error)
	ExistsBySKU(ctx context.Context, storeID, sku string, excludeID string) (bool, error)
	Update(ctx context.Context, p *domain.Product) (*domain.Product, error)
	SoftDelete(ctx context.Context, id string) error
	GetModifiedSince(ctx context.Context, storeID string, since time.Time) ([]*domain.Product, error)
}

type StockRepository interface {
	FindLevelsByStore(ctx context.Context, storeID string, lowStockOnly bool) ([]*domain.StockLevel, error)
	FindLevelByProduct(ctx context.Context, productID, storeID string) (*domain.StockLevel, error)
	SetMinQuantity(ctx context.Context, productID, storeID string, min float64) error
	Adjust(ctx context.Context, input domain.AdjustInput) (*domain.StockLevel, error)
	FindMovements(ctx context.Context, filter dto.StockMovementFilter) ([]*domain.StockMovement, int, error)
	// DeductStock subtracts qty from stock_levels and records a stock_movement.
	DeductStock(ctx context.Context, productID, storeID string, qty float64, refID, cashierID string) error
	GetModifiedSince(ctx context.Context, storeID string, since time.Time) ([]*domain.StockLevel, error)
}

// TableRepository manages physical tables in a restaurant.
type TableRepository interface {
	FindAllByStore(ctx context.Context, storeID string) ([]*domain.RestaurantTable, error)
	FindByID(ctx context.Context, id string) (*domain.RestaurantTable, error)
	Create(ctx context.Context, t *domain.RestaurantTable) (*domain.RestaurantTable, error)
	Update(ctx context.Context, t *domain.RestaurantTable) (*domain.RestaurantTable, error)
	UpdateStatus(ctx context.Context, id string, status domain.TableStatus) error
	SoftDelete(ctx context.Context, id string) error
}

// MenuItemRepository handles menu item retrieval and management.
type MenuItemRepository interface {
	FindByID(ctx context.Context, id string) (*domain.MenuItem, error)
	FindAllByStore(ctx context.Context, storeID string) ([]*domain.MenuItem, error)
	Create(ctx context.Context, item *domain.MenuItem) (*domain.MenuItem, error)
	Update(ctx context.Context, item *domain.MenuItem) (*domain.MenuItem, error)
	ReplaceIngredients(ctx context.Context, menuItemID string, ings []domain.MenuItemIngredient) error
	SoftDelete(ctx context.Context, id string) error
}

// ─── Phase 3 ──────────────────────────────────────────────────────────────────

// TransactionRepository handles cashier transactions with atomic stock deduction.
type TransactionRepository interface {
	// Create persists a new transaction with items; status field controls draft vs completed.
	Create(ctx context.Context, input domain.CreateTransactionInput) (*domain.Transaction, error)
	// FindAll returns a paginated list of transactions.
	FindAll(ctx context.Context, filter dto.TransactionListFilter) ([]*domain.Transaction, int, error)
	// FindByID returns a transaction with its items.
	FindByID(ctx context.Context, id string) (*domain.Transaction, error)
	// Void marks a transaction as voided and reverses stock movements atomically.
	Void(ctx context.Context, txnID, userID string) error
	// GetDraftByTable returns the open draft for a given table, or nil.
	GetDraftByTable(ctx context.Context, storeID, tableID string) (*domain.Transaction, error)
	// UpdateDraftItems replaces items on a draft and recalculates totals.
	UpdateDraftItems(ctx context.Context, txnID string, items []domain.CreateTransactionItemInput,
		subtotal, discountAmt, taxAmt, total float64, customerName, notes string) (*domain.Transaction, error)
	// PayDraft finalizes a held order with payment info and deducts stock.
	PayDraft(ctx context.Context, input domain.PayDraftInput, storeID, cashierID string) (*domain.Transaction, error)
	// GetKDSTickets returns active tickets (draft restaurant transactions) for Kitchen Display.
	GetKDSTickets(ctx context.Context, storeID string) ([]*domain.Transaction, error)
	// UpdateKDSItemStatus updates the completion status of a specific KDS ticket item.
	UpdateKDSItemStatus(ctx context.Context, itemID, status string) error
	GetModifiedSince(ctx context.Context, storeID string, since time.Time) ([]*domain.Transaction, error)
}

// PurchaseOrderRepository handles PO lifecycle and stock updates on receive.
type PurchaseOrderRepository interface {
	Create(ctx context.Context, po *domain.PurchaseOrder, items []domain.POItem) (*domain.PurchaseOrder, error)
	FindAll(ctx context.Context, filter dto.POListFilter) ([]*domain.PurchaseOrder, int, error)
	FindByID(ctx context.Context, id string) (*domain.PurchaseOrder, error)
	Update(ctx context.Context, po *domain.PurchaseOrder, items []domain.POItem) (*domain.PurchaseOrder, error)
	Submit(ctx context.Context, poID, userID string) error
	// Receive marks the PO received and atomically updates stock levels.
	Receive(ctx context.Context, poID, userID string) error
	Cancel(ctx context.Context, poID string) error
}

// SupplierRepository handles CRUD for suppliers (soft delete via deleted_at).
type SupplierRepository interface {
	Create(ctx context.Context, s *domain.Supplier) (*domain.Supplier, error)
	FindAll(ctx context.Context, filter dto.SupplierListFilter) ([]*domain.Supplier, int, error)
	FindByID(ctx context.Context, id string) (*domain.Supplier, error)
	Update(ctx context.Context, s *domain.Supplier) (*domain.Supplier, error)
	SoftDelete(ctx context.Context, id string) error
}

// ReportRepository runs aggregate queries for reporting.
type ReportRepository interface {
	SalesSummary(ctx context.Context, storeID string, from, to time.Time) ([]dto.SalesSummaryRow, error)
	SalesByProduct(ctx context.Context, storeID string, from, to time.Time) ([]dto.SalesByProductRow, error)
	SalesByCashier(ctx context.Context, storeID string, from, to time.Time) ([]dto.SalesByCashierRow, error)
	StockValuation(ctx context.Context, storeID string) ([]dto.StockValuationRow, error)
	ProfitSummary(ctx context.Context, storeID string, from, to time.Time, groupBy string) ([]dto.ProfitPeriodRow, error)
	CashFlowSummary(ctx context.Context, storeID string, from, to time.Time) ([]dto.CashFlowDayRow, error)
	CashFlowDetail(ctx context.Context, storeID string, from, to time.Time) ([]dto.CashFlowDetailEntry, error)
}

// PriceHistoryRepository records and retrieves price-change audit logs.
type PriceHistoryRepository interface {
	Record(ctx context.Context, h domain.PriceHistory) error
	FindByProduct(ctx context.Context, productID string, f dto.PriceHistoryFilter) ([]*domain.PriceHistory, int, error)
	FindByStore(ctx context.Context, storeID string, f dto.PriceHistoryFilter) ([]*domain.PriceHistory, int, error)
}

// POPaymentRepository records and retrieves payments made against purchase orders.
type POPaymentRepository interface {
	Create(ctx context.Context, p domain.POPayment) (*domain.POPayment, error)
	FindByPO(ctx context.Context, poID string) ([]*domain.POPayment, error)
	AggregateByPO(ctx context.Context, poID string, totalAmount float64) (float64, string, error)
	PayableSummary(ctx context.Context, storeID string) (*dto.PayableSummary, error)
	PopulatePOPayments(ctx context.Context, pos []*domain.PurchaseOrder)
}

// CustomerRepository manages customer master data per store.
type CustomerRepository interface {
	Create(ctx context.Context, c *domain.Customer) (*domain.Customer, error)
	FindAll(ctx context.Context, f dto.CustomerListFilter) ([]*domain.Customer, int, error)
	FindByID(ctx context.Context, id string) (*domain.Customer, error)
	Update(ctx context.Context, c *domain.Customer) (*domain.Customer, error)
	SoftDelete(ctx context.Context, id string) error
	SearchByPhone(ctx context.Context, storeID, phone string) ([]*domain.Customer, error)
	GetModifiedSince(ctx context.Context, storeID string, since time.Time) ([]*domain.Customer, error)
}

type StockAdjustmentRepository interface {
	CreateAdjustment(ctx context.Context, storeID, userID string, input domain.CreateAdjustmentInput) error
	GetStockAdjustmentHistory(ctx context.Context, storeID string, productID *string) ([]*domain.StockAdjustment, error)
}

// ─── FIFO Batch Inventory ──────────────────────────────────────────────────────

// BatchRepository manages stock batch records for FIFO inventory tracking.
// Each batch corresponds to one product received in a purchase order.
type BatchRepository interface {
	// CreateBatch inserts a new stock batch when a PO item is received.
	CreateBatch(ctx context.Context, batch *domain.StockBatch) error

	// GetBatchesByProduct returns all non-empty batches for a product, oldest first (FIFO order).
	GetBatchesByProduct(ctx context.Context, productID, storeID string) ([]*domain.StockBatch, error)

	// GetBatchesByStore returns non-empty batches for a store; optionally filtered by product_id.
	GetBatchesByStore(ctx context.Context, filter dto.BatchListFilter) ([]*domain.StockBatch, error)

	// GetStockSummary returns aggregated batch totals (total qty, batch count, avg cost) per product.
	GetStockSummary(ctx context.Context, storeID string) ([]*domain.BatchStockSummary, error)

	// DeductFIFO subtracts qty from the oldest available batches within a single DB transaction.
	// Rows are locked with FOR UPDATE to prevent concurrent over-deductions.
	// Returns an error immediately if total available stock is insufficient.
	DeductFIFO(ctx context.Context, productID, storeID string, qty float64) error
}

// ─── Termin (Installment) Repositories ───────────────────────────────────────

// TerminRepository manages installment schedules for purchase orders.
type TerminRepository interface {
	// CreateSchedule atomically deletes existing termins for a PO and inserts the new schedule.
	CreateSchedule(ctx context.Context, poID string, termins []domain.POTermin) error

	// FindByPO returns all termins for a PO with their payment aggregation, ordered by termin_number.
	FindByPO(ctx context.Context, poID string) ([]*domain.POTermin, error)

	// FindByID returns a single termin with its payment aggregation.
	FindByID(ctx context.Context, terminID string) (*domain.POTermin, error)

	// UpdateStatus recalculates and persists the termin status (unpaid/partial/paid/overdue).
	UpdateStatus(ctx context.Context, terminID string) error

	// DebtSummary aggregates payment totals across all termins for a PO.
	DebtSummary(ctx context.Context, poID string, totalAmount float64) (*domain.PODebtSummary, error)
}

// PaymentRecordRepository manages individual payment transactions against termins.
type PaymentRecordRepository interface {
	// Create inserts one payment record and returns the persisted row.
	Create(ctx context.Context, record domain.PaymentRecord) (*domain.PaymentRecord, error)

	// FindByTermin returns all payment records for a termin, newest first.
	FindByTermin(ctx context.Context, terminID string) ([]*domain.PaymentRecord, error)
}

type ExpenseRepository interface {
	ListCategories(ctx context.Context, includeDeleted bool) ([]*domain.ExpenseCategory, error)
	CreateCategory(ctx context.Context, c *domain.ExpenseCategory) (*domain.ExpenseCategory, error)
	GetCategoryByID(ctx context.Context, id string) (*domain.ExpenseCategory, error)
	UpdateCategory(ctx context.Context, id string, name, desc string, isActive bool) (*domain.ExpenseCategory, error)
	SoftDeleteCategory(ctx context.Context, id string) error

	CreateExpense(ctx context.Context, e *domain.Expense) (*domain.Expense, error)
	FindAll(ctx context.Context, f dto.ExpenseListFilter) ([]*domain.Expense, int, error)
	GetByID(ctx context.Context, id, storeID string) (*domain.Expense, error)
	Update(ctx context.Context, e *domain.Expense) (*domain.Expense, error)
	Delete(ctx context.Context, id, storeID string) error
	UpdatePaymentStatus(ctx context.Context, id, storeID, status string) (*domain.Expense, error)

	CreateRecurringExpense(ctx context.Context, e *domain.RecurringExpense) (*domain.RecurringExpense, error)
	FindAllRecurring(ctx context.Context, f dto.ExpenseListFilter) ([]*domain.RecurringExpense, int, error)
	GetRecurringByID(ctx context.Context, id, storeID string) (*domain.RecurringExpense, error)
	UpdateRecurring(ctx context.Context, e *domain.RecurringExpense) (*domain.RecurringExpense, error)
	DeleteRecurring(ctx context.Context, id, storeID string) error
	GetDueRecurringExpenses(ctx context.Context) ([]*domain.RecurringExpense, error)
	BumpRecurringNextRun(ctx context.Context, id string, nextRun string) error
}

type IncomeRepository interface {
	ListCategories(ctx context.Context, includeDeleted bool) ([]*domain.IncomeCategory, error)
	GetCategoryByID(ctx context.Context, id string) (*domain.IncomeCategory, error)
	CreateCategory(ctx context.Context, cat *domain.IncomeCategory) (*domain.IncomeCategory, error)
	UpdateCategory(ctx context.Context, id string, name, desc string, isActive bool) (*domain.IncomeCategory, error)
	SoftDeleteCategory(ctx context.Context, id string) error
	Create(ctx context.Context, inc *domain.Income) (*domain.Income, error)
	FindAll(ctx context.Context, f dto.IncomeListFilter) ([]*domain.Income, int, error)
	FindByID(ctx context.Context, id string) (*domain.Income, error)
	Update(ctx context.Context, inc *domain.Income) (*domain.Income, error)
	Delete(ctx context.Context, id, storeID string) error
	SumByDateRange(ctx context.Context, storeID string, from, to time.Time) (float64, error)
}

// ─── Loyalty System ───────────────────────────────────────────────────────────

// MembershipTierRepository manages loyalty tier configuration records.
type MembershipTierRepository interface {
	// FindAll returns all configured membership tiers.
	FindAll(ctx context.Context) ([]*domain.MembershipTier, error)
	// FindByID returns a single tier by its UUID.
	FindByID(ctx context.Context, id string) (*domain.MembershipTier, error)
}

// LoyaltyRepository manages the immutable loyalty ledger for customers.
type LoyaltyRepository interface {
	// GetBalance returns the sum of all point deltas for a customer (current balance).
	GetBalance(ctx context.Context, customerID string) (float64, error)
	// EarnPoints appends a positive EARN entry to the ledger.
	EarnPoints(ctx context.Context, customerID string, transactionID *string, points float64) (*domain.LoyaltyLedger, error)
	// SpendPoints appends a negative SPEND entry to the ledger.
	SpendPoints(ctx context.Context, customerID string, transactionID *string, points float64) (*domain.LoyaltyLedger, error)
	// VoidPoints revokes points earned by a specific transaction (inserts a VOID entry).
	VoidPoints(ctx context.Context, customerID string, transactionID *string, points float64) (*domain.LoyaltyLedger, error)
	// AdjustPoints inserts a manual ADJUST entry (positive or negative delta).
	AdjustPoints(ctx context.Context, customerID string, delta float64, note string) (*domain.LoyaltyLedger, error)
	// GetHistory returns all ledger entries for a customer, newest first.
	GetHistory(ctx context.Context, customerID string) ([]*domain.LoyaltyLedger, error)
	// GetHistoryPaginated returns a page of ledger entries for a customer.
	GetHistoryPaginated(ctx context.Context, customerID string, page, perPage int) ([]*domain.LoyaltyLedger, int, error)
	// AssignTier updates the customer's loyalty_tier_id.
	AssignTier(ctx context.Context, customerID, tierID string) error
	// GetCustomerTier returns the tier for a customer (nil if unassigned).
	GetCustomerTier(ctx context.Context, customerID string) (*domain.MembershipTier, error)
	// GetTopCustomersByBalance returns top N customers by total point balance for a store.
	GetTopCustomersByBalance(ctx context.Context, storeID string, limit int) ([]dto.TopCustomerLoyalty, error)
	// GetPointsSummary returns total points earned and used within [from, to) for a store.
	GetPointsSummary(ctx context.Context, storeID string, from, to time.Time) (earned, used float64, err error)
}
