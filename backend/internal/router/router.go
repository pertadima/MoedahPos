package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"

	"github.com/moedahpos/backend/internal/handler"
	"github.com/moedahpos/backend/internal/middleware"
	"github.com/moedahpos/backend/pkg/jwt"
)

// Dependencies groups all injectable dependencies for the router.
type Dependencies struct {
	// Phase 1
	AuthHandler *handler.AuthHandler
	JWTManager  *jwt.Manager

	// Phase 2
	StoreHandler   *handler.StoreHandler
	ProductHandler *handler.ProductHandler
	StockHandler   *handler.StockHandler

	// Phase 3
	TransactionHandler   *handler.TransactionHandler
	PurchaseOrderHandler *handler.PurchaseOrderHandler
	SupplierHandler      *handler.SupplierHandler
	ReportHandler        *handler.ReportHandler

	// Restaurant mode (Phase 4)
	RestaurantHandler *handler.RestaurantHandler

	// Price History
	PriceHistoryHandler *handler.PriceHistoryHandler

	// Customers
	CustomerHandler *handler.CustomerHandler

	// User Admin
	UserAdminHandler *handler.UserAdminHandler

	BatchStockHandler *handler.BatchStockHandler

	// Termin (installment payment)
	TerminHandler *handler.TerminHandler

	ExpenseHandler         *handler.ExpenseHandler
	StockAdjustmentHandler *handler.StockAdjustmentHandler
	IncomeHandler          *handler.IncomeHandler
	ActivityLogHandler     *handler.ActivityLogHandler
	SyncHandler            *handler.SyncHandler
	LoyaltyHandler         *handler.LoyaltyHandler

	// Shared
	DB  *sqlx.DB
	Log zerolog.Logger
}

// New configures and returns the full application router.
func New(deps *Dependencies) http.Handler { //nolint:funlen // route wiring is inherently long
	r := chi.NewRouter()

	// ── Global Middleware ──────────────────────────────────────────────────────
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.Logger(deps.Log))
	r.Use(corsMiddleware)

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	r.Route("/api/v1", func(r chi.Router) {

		// ── Public Auth ────────────────────────────────────────────────────────
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", deps.AuthHandler.Register)
			r.Post("/login", deps.AuthHandler.Login)
			r.Post("/refresh", deps.AuthHandler.Refresh)
			r.Group(func(r chi.Router) {
				r.Use(middleware.Authenticate(deps.JWTManager))
				r.Post("/logout", deps.AuthHandler.Logout)
				r.Get("/me", deps.AuthHandler.Me)
			})
		})

		// ── Authenticated-only ────────────────────────────────────────────────
		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticate(deps.JWTManager))

			// ── Suppliers (global, no store context needed) ────────────────────
			r.Route("/suppliers", func(r chi.Router) {
				r.Get("/", deps.SupplierHandler.List)
				r.Post("/", withPerm(deps, "pembelian:write", deps.SupplierHandler.Create))
				r.Get("/{supplierId}", deps.SupplierHandler.Get)
				r.Put("/{supplierId}", withPerm(deps, "pembelian:update", deps.SupplierHandler.Update))
				r.Delete("/{supplierId}", withPerm(deps, "pembelian:delete", deps.SupplierHandler.Delete))
			})

			// ── Categories ────────────────────────────────────────────────────────
			// GET accessible to all authenticated, others require SuperAdmin
			r.Route("/expense-categories", func(r chi.Router) {
				r.Get("/", deps.ExpenseHandler.ListCategories)
				r.Group(func(r chi.Router) {
					r.Use(middleware.RequireSuperAdmin(deps.DB))
					r.Post("/", deps.ExpenseHandler.CreateCategory)
					r.Put("/{id}", deps.ExpenseHandler.UpdateCategory)
					r.Delete("/{id}", deps.ExpenseHandler.DeleteCategory)
				})
			})

			r.Route("/income-categories", func(r chi.Router) {
				r.Get("/", deps.IncomeHandler.ListCategories)
				r.Group(func(r chi.Router) {
					r.Use(middleware.RequireSuperAdmin(deps.DB))
					r.Post("/", deps.IncomeHandler.CreateCategory)
					r.Put("/{id}", deps.IncomeHandler.UpdateCategory)
					r.Delete("/{id}", deps.IncomeHandler.DeleteCategory)
				})
			})

			// ── Admin: User & Role management ─────────────────────────────────
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireAdminOrSuperAdmin(deps.DB))
				r.Route("/admin", func(r chi.Router) {
					r.Route("/roles", func(r chi.Router) {
						r.Get("/", deps.UserAdminHandler.ListRoles)
						r.Post("/", deps.UserAdminHandler.CreateRole)
						r.Put("/{roleId}", deps.UserAdminHandler.UpdateRole)
						r.Delete("/{roleId}", deps.UserAdminHandler.DeleteRole)
					})
					r.Get("/permissions", deps.UserAdminHandler.ListPermissions)

					r.Route("/users", func(r chi.Router) {
						r.Get("/", deps.UserAdminHandler.List)
						r.Post("/", deps.UserAdminHandler.Create)
						r.Get("/{userId}", deps.UserAdminHandler.Get)
						r.Put("/{userId}", deps.UserAdminHandler.Update)
						r.Post("/{userId}/deactivate", deps.UserAdminHandler.Deactivate)
						r.Post("/{userId}/reset-password", deps.UserAdminHandler.ResetPassword)
						r.Put("/{userId}/stores", deps.UserAdminHandler.SetStores)
					})
				})
			})

			// ── Store-scoped routes ────────────────────────────────────────────
			r.Route("/stores", func(r chi.Router) {
				r.Get("/", deps.StoreHandler.List)
				r.Post("/", deps.StoreHandler.Create) // JWT-auth only; superadmin bypasses store-scoped RBAC

				r.Route("/{storeId}", func(r chi.Router) {
					r.Get("/", deps.StoreHandler.Get)

					// Global store CRUD — JWT-auth only (no store-scoped role required)
					r.Put("/", deps.StoreHandler.Update)
					r.Delete("/", deps.StoreHandler.Delete)

					r.Group(func(r chi.Router) {
						r.Use(middleware.StoreContext(deps.DB))

						// Members
						r.Route("/members", func(r chi.Router) {
							r.Get("/", withPerm(deps, "settings:read", deps.StoreHandler.ListMembers))
							r.Post("/", withPerm(deps, "settings:write", deps.StoreHandler.AddMember))
							r.Put("/{userId}", withPerm(deps, "settings:update", deps.StoreHandler.UpdateMemberRole))
							r.Delete("/{userId}", withPerm(deps, "settings:delete", deps.StoreHandler.RemoveMember))
						})

						// Categories
						r.Route("/categories", func(r chi.Router) {
							r.Get("/", withPerm(deps, "inventory:read", deps.ProductHandler.ListCategories))
							r.Post("/", withPerm(deps, "inventory:write", deps.ProductHandler.CreateCategory))
							r.Put("/{categoryId}", withPerm(deps, "inventory:update", deps.ProductHandler.UpdateCategory))
							r.Delete("/{categoryId}", withPerm(deps, "inventory:delete", deps.ProductHandler.DeleteCategory))
						})

						// Products
						r.Route("/products", func(r chi.Router) {
							r.Get("/", withPerm(deps, "inventory:read", deps.ProductHandler.List))
							r.Post("/", withPerm(deps, "inventory:write", deps.ProductHandler.Create))
							r.Get("/barcode/{barcode}", withPerm(deps, "kasir:read", deps.ProductHandler.GetByBarcode))
							r.Get("/{productId}", withPerm(deps, "inventory:read", deps.ProductHandler.Get))
							r.Put("/{productId}", withPerm(deps, "inventory:update", deps.ProductHandler.Update))
							r.Delete("/{productId}", withPerm(deps, "inventory:delete", deps.ProductHandler.Delete))
							r.Get("/{productId}/price-history", withPerm(deps, "inventory:read", deps.PriceHistoryHandler.ListByProduct))
						})

						// Price history (store-wide)
						r.Get("/price-history", withPerm(deps, "inventory:read", deps.PriceHistoryHandler.ListByStore))

						// Customers
						r.Route("/customers", func(r chi.Router) {
							r.Get("/", withPerm(deps, "kasir:read", deps.CustomerHandler.List))
							r.Get("/search", withPerm(deps, "kasir:read", deps.CustomerHandler.Search))
							r.Post("/", withPerm(deps, "kasir:write", deps.CustomerHandler.Create))
							r.Get("/{customerId}", withPerm(deps, "kasir:read", deps.CustomerHandler.Get))
							r.Put("/{customerId}", withPerm(deps, "kasir:update", deps.CustomerHandler.Update))
							r.Delete("/{customerId}", withPerm(deps, "kasir:delete", deps.CustomerHandler.Delete))
							// Loyalty sub-routes per customer
							r.Get("/{customerId}/loyalty", withPerm(deps, "kasir:read", deps.LoyaltyHandler.GetBalance))
							r.Post("/{customerId}/loyalty/earn", withPerm(deps, "kasir:write", deps.LoyaltyHandler.EarnPoints))
							r.Post("/{customerId}/loyalty/redeem", withPerm(deps, "kasir:write", deps.LoyaltyHandler.RedeemPoints))
							r.Get("/{customerId}/loyalty/history", withPerm(deps, "kasir:read", deps.LoyaltyHandler.GetHistory))
							r.Get("/{customerId}/loyalty/history/paged", withPerm(deps, "kasir:read", deps.LoyaltyHandler.GetHistoryPaginated))
							r.Post("/{customerId}/loyalty/void", withPerm(deps, "kasir:write", deps.LoyaltyHandler.VoidTransactionPoints))
							r.Post("/{customerId}/loyalty/adjust", withPerm(deps, "admin:write", deps.LoyaltyHandler.AdjustPoints))
							r.Put("/{customerId}/loyalty/tier", withPerm(deps, "kasir:update", deps.LoyaltyHandler.AssignTier))
						})

						// Loyalty tiers (store-scoped read)
						r.Route("/loyalty", func(r chi.Router) {
							r.Get("/tiers", withPerm(deps, "kasir:read", deps.LoyaltyHandler.ListTiers))
							r.Get("/summary", withPerm(deps, "kasir:read", deps.LoyaltyHandler.GetLoyaltySummary))
						})

						// Stock
						r.Route("/stock", func(r chi.Router) {
							r.Get("/", withPerm(deps, "inventory:read", deps.StockHandler.GetLevels))
							r.Get("/low", withPerm(deps, "inventory:read", deps.StockHandler.GetLowStock))
							r.Get("/movements", withPerm(deps, "inventory:read", deps.StockHandler.GetMovements))
							r.Post("/adjust", withPerm(deps, "inventory:write", deps.StockHandler.Adjust))
							r.Put("/min", withPerm(deps, "inventory:write", deps.StockHandler.SetMinStock))
							// FIFO batch endpoints — registered before wildcard so static paths win
							r.Get("/batches", withPerm(deps, "inventory:read", deps.BatchStockHandler.ListBatches))
							r.Get("/batch-summary", withPerm(deps, "inventory:read", deps.BatchStockHandler.GetSummary))
							r.Get("/{productId}", withPerm(deps, "inventory:read", deps.StockHandler.GetProductStock))
						})

						// Sync
						r.Route("/sync", func(r chi.Router) {
							r.Get("/pull", withPerm(deps, "kasir:read", deps.SyncHandler.Pull))
						})

						// ── Phase 3 ───────────────────────────────────────────

						// Transactions (Cashier / POS)
						r.Route("/transactions", func(r chi.Router) {
							r.Get("/", withPerm(deps, "penjualan:read", deps.TransactionHandler.List))
							r.Post("/", withPerm(deps, "kasir:write", deps.TransactionHandler.Checkout))
							// Draft / table order endpoints (restaurant)
							r.Get("/draft", withPerm(deps, "kasir:read", deps.TransactionHandler.GetDraftByTable))
							r.Post("/draft", withPerm(deps, "kasir:write", deps.TransactionHandler.CreateDraft))
							r.Get("/{txnId}", withPerm(deps, "penjualan:read", deps.TransactionHandler.Get))
							r.Put("/{txnId}/draft", withPerm(deps, "kasir:update", deps.TransactionHandler.UpdateDraft))
							r.Post("/{txnId}/pay", withPerm(deps, "kasir:write", deps.TransactionHandler.PayDraft))
							r.Post("/{txnId}/void", withPerm(deps, "penjualan:delete", deps.TransactionHandler.Void))
						})

						// Purchase Orders
						r.Route("/purchase-orders", func(r chi.Router) {
							r.Get("/", withPerm(deps, "pembelian:read", deps.PurchaseOrderHandler.List))
							r.Post("/", withPerm(deps, "pembelian:write", deps.PurchaseOrderHandler.Create))
							r.Get("/payables", withPerm(deps, "pembelian:read", deps.PurchaseOrderHandler.PayableSummary))
							r.Get("/{poId}", withPerm(deps, "pembelian:read", deps.PurchaseOrderHandler.Get))
							r.Put("/{poId}", withPerm(deps, "pembelian:update", deps.PurchaseOrderHandler.Update))
							r.Post("/{poId}/submit", withPerm(deps, "pembelian:update", deps.PurchaseOrderHandler.Submit))
							r.Post("/{poId}/receive", withPerm(deps, "pembelian:update", deps.PurchaseOrderHandler.Receive))
							r.Delete("/{poId}", withPerm(deps, "pembelian:delete", deps.PurchaseOrderHandler.Cancel))
							r.Get("/{poId}/payments", withPerm(deps, "pembelian:read", deps.PurchaseOrderHandler.ListPayments))
							r.Post("/{poId}/payments", withPerm(deps, "pembelian:write", deps.PurchaseOrderHandler.CreatePayment))
							// Termin (installment) routes
							r.Get("/{poId}/termins", withPerm(deps, "pembelian:read", deps.TerminHandler.ListTermins))
							r.Post("/{poId}/termins", withPerm(deps, "pembelian:write", deps.TerminHandler.CreateSchedule))
							r.Post("/{poId}/termins/{terminId}/payments", withPerm(deps, "pembelian:write", deps.TerminHandler.RecordPayment))
							r.Get("/{poId}/debt", withPerm(deps, "pembelian:read", deps.TerminHandler.GetDebtSummary))
							r.Get("/{poId}/document", withPerm(deps, "pembelian:read", deps.TerminHandler.GetDocument))
						})

						// Reports
						r.Route("/reports", func(r chi.Router) {
							r.Get("/sales", withPerm(deps, "penjualan:read", deps.ReportHandler.SalesSummary))
							r.Get("/sales/by-product", withPerm(deps, "penjualan:read", deps.ReportHandler.SalesByProduct))
							r.Get("/sales/by-cashier", withPerm(deps, "penjualan:read", deps.ReportHandler.SalesByCashier))
							r.Get("/stock-valuation", withPerm(deps, "inventory:read", deps.ReportHandler.StockValuation))
							r.Get("/profit", withPerm(deps, "keuangan:read", deps.ReportHandler.ProfitSummary))
							r.Get("/cash-flow", withPerm(deps, "keuangan:read", deps.ReportHandler.CashFlow))
							r.Get("/cash-flow/detail", withPerm(deps, "keuangan:read", deps.ReportHandler.CashFlowDetail))
							// Export: GET /reports/export?type=csv|pdf&report=sales|inventory|profit
							// Minimum permission: penjualan:read (profit export requires keuangan:read — enforced per-report in handler)
							r.Get("/export", withPerm(deps, "penjualan:read", deps.ReportHandler.ExportReport))
						})

						// Expenses
						r.Route("/expenses", func(r chi.Router) {
							r.Get("/", withPerm(deps, "keuangan:read", deps.ExpenseHandler.ListExpenses))
							r.Post("/", withPerm(deps, "keuangan:write", deps.ExpenseHandler.CreateExpense))
							r.Put("/{id}", withPerm(deps, "keuangan:update", deps.ExpenseHandler.UpdateExpense))
							r.Delete("/{id}", withPerm(deps, "keuangan:delete", deps.ExpenseHandler.DeleteExpense))
							r.Patch("/{id}/status", withPerm(deps, "keuangan:update", deps.ExpenseHandler.UpdateExpenseStatus))
						})

						// Stock Adjustments
						r.Route("/adjustments", func(r chi.Router) {
							r.Get("/", withPerm(deps, "inventory:read", deps.StockAdjustmentHandler.GetHistory))
							r.Post("/", withPerm(deps, "inventory:write", deps.StockAdjustmentHandler.Create))
						})

						// Activity Logs
						r.Route("/activity-logs", func(r chi.Router) {
							r.Get("/", withPerm(deps, "settings:read", deps.ActivityLogHandler.List))
						})

						// Recurring Expenses
						r.Route("/recurring-expenses", func(r chi.Router) {
							r.Get("/", withPerm(deps, "keuangan:read", deps.ExpenseHandler.ListRecurringExpenses))
							r.Post("/", withPerm(deps, "keuangan:write", deps.ExpenseHandler.CreateRecurringExpense))
							r.Put("/{id}", withPerm(deps, "keuangan:update", deps.ExpenseHandler.UpdateRecurringExpense))
							r.Delete("/{id}", withPerm(deps, "keuangan:delete", deps.ExpenseHandler.DeleteRecurringExpense))
						})

						// Incomes
						r.Route("/incomes", func(r chi.Router) {
							r.Get("/", withPerm(deps, "keuangan:read", deps.IncomeHandler.ListIncomes))
							r.Post("/", withPerm(deps, "keuangan:write", deps.IncomeHandler.CreateIncome))
							r.Put("/{id}", withPerm(deps, "keuangan:update", deps.IncomeHandler.UpdateIncome))
							r.Delete("/{id}", withPerm(deps, "keuangan:delete", deps.IncomeHandler.DeleteIncome))
						})

						// ── Restaurant Mode ───────────────────────────────────

						// Tables
						r.Route("/tables", func(r chi.Router) {
							r.Get("/", withPerm(deps, "kasir:read", deps.RestaurantHandler.ListTables))
							r.Post("/", withPerm(deps, "kasir:write", deps.RestaurantHandler.CreateTable))
							r.Put("/{tableId}", withPerm(deps, "kasir:update", deps.RestaurantHandler.UpdateTable))
							r.Put("/{tableId}/status", withPerm(deps, "kasir:update", deps.RestaurantHandler.UpdateTableStatus))
							r.Delete("/{tableId}", withPerm(deps, "kasir:delete", deps.RestaurantHandler.DeleteTable))
						})

						// Menu Items
						r.Route("/menu-items", func(r chi.Router) {
							r.Get("/", withPerm(deps, "inventory:read", deps.RestaurantHandler.ListMenuItems))
							r.Post("/", withPerm(deps, "inventory:write", deps.RestaurantHandler.CreateMenuItem))
							r.Put("/{menuItemId}", withPerm(deps, "inventory:update", deps.RestaurantHandler.UpdateMenuItem))
							r.Delete("/{menuItemId}", withPerm(deps, "inventory:delete", deps.RestaurantHandler.DeleteMenuItem))
						})

						// KDS (Kitchen Display System)
						r.Route("/kds", func(r chi.Router) {
							r.Get("/tickets", withPerm(deps, "kasir:read", deps.TransactionHandler.GetKDSTickets))
							r.Put("/items/{itemId}", withPerm(deps, "kasir:update", deps.TransactionHandler.UpdateKDSItemStatus))
						})
					})
				})
			})
		})
	})

	return r
}

// withPerm wraps a HandlerFunc with the RequirePermission middleware.
//
//nolint:revive // deps is kept for API consistency
func withPerm(deps *Dependencies, perm string, fn http.HandlerFunc) http.HandlerFunc {
	return middleware.RequirePermission(perm)(fn).ServeHTTP
}

// corsMiddleware adds CORS headers and handles OPTIONS preflight requests.
// This allows the Next.js frontend (localhost:3000) to call the API (localhost:8080).
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-Request-Id")
		w.Header().Set("Access-Control-Max-Age", "86400")

		// Respond immediately to preflight
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
