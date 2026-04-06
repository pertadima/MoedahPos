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
	"github.com/moedahpos/backend/pkg/rbac"
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

	// Shared
	RoleStore *rbac.RoleStore
	DB        *sqlx.DB
	Log       zerolog.Logger
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
				r.Post("/", withPerm(deps, "suppliers.create", deps.SupplierHandler.Create))
				r.Get("/{supplierId}", deps.SupplierHandler.Get)
				r.Put("/{supplierId}", withPerm(deps, "suppliers.update", deps.SupplierHandler.Update))
				r.Delete("/{supplierId}", withPerm(deps, "suppliers.delete", deps.SupplierHandler.Delete))
			})

			r.Route("/expense-categories", func(r chi.Router) {
				r.Get("/", deps.ExpenseHandler.ListCategories)
				r.Post("/", deps.ExpenseHandler.CreateCategory)
			})

			// ── Admin: User & Role management ─────────────────────────────────
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireAdminOrSuperAdmin(deps.DB))
				r.Route("/admin", func(r chi.Router) {
					r.Get("/roles", deps.UserAdminHandler.ListRoles)
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
							r.Get("/", withPerm(deps, "users.read", deps.StoreHandler.ListMembers))
							r.Post("/", withPerm(deps, "users.create", deps.StoreHandler.AddMember))
							r.Put("/{userId}", withPerm(deps, "users.update", deps.StoreHandler.UpdateMemberRole))
							r.Delete("/{userId}", withPerm(deps, "users.delete", deps.StoreHandler.RemoveMember))
						})

						// Categories
						r.Route("/categories", func(r chi.Router) {
							r.Get("/", withPerm(deps, "products.read", deps.ProductHandler.ListCategories))
							r.Post("/", withPerm(deps, "products.create", deps.ProductHandler.CreateCategory))
							r.Put("/{categoryId}", withPerm(deps, "products.update", deps.ProductHandler.UpdateCategory))
							r.Delete("/{categoryId}", withPerm(deps, "products.delete", deps.ProductHandler.DeleteCategory))
						})

						// Products
						r.Route("/products", func(r chi.Router) {
							r.Get("/", withPerm(deps, "products.read", deps.ProductHandler.List))
							r.Post("/", withPerm(deps, "products.create", deps.ProductHandler.Create))
							r.Get("/barcode/{barcode}", withPerm(deps, "products.read", deps.ProductHandler.GetByBarcode))
							r.Get("/{productId}", withPerm(deps, "products.read", deps.ProductHandler.Get))
							r.Put("/{productId}", withPerm(deps, "products.update", deps.ProductHandler.Update))
							r.Delete("/{productId}", withPerm(deps, "products.delete", deps.ProductHandler.Delete))
							r.Get("/{productId}/price-history", withPerm(deps, "products.read", deps.PriceHistoryHandler.ListByProduct))
						})

						// Price history (store-wide)
						r.Get("/price-history", withPerm(deps, "products.read", deps.PriceHistoryHandler.ListByStore))

						// Customers
						r.Route("/customers", func(r chi.Router) {
							r.Get("/", withPerm(deps, "products.read", deps.CustomerHandler.List))
							r.Get("/search", withPerm(deps, "products.read", deps.CustomerHandler.Search))
							r.Post("/", withPerm(deps, "products.create", deps.CustomerHandler.Create))
							r.Get("/{customerId}", withPerm(deps, "products.read", deps.CustomerHandler.Get))
							r.Put("/{customerId}", withPerm(deps, "products.update", deps.CustomerHandler.Update))
							r.Delete("/{customerId}", withPerm(deps, "products.delete", deps.CustomerHandler.Delete))
						})

						// Stock
						r.Route("/stock", func(r chi.Router) {
							r.Get("/", withPerm(deps, "stock.read", deps.StockHandler.GetLevels))
							r.Get("/low", withPerm(deps, "stock.read", deps.StockHandler.GetLowStock))
							r.Get("/movements", withPerm(deps, "stock.read", deps.StockHandler.GetMovements))
							r.Post("/adjust", withPerm(deps, "stock.adjust", deps.StockHandler.Adjust))
							r.Put("/min", withPerm(deps, "stock.adjust", deps.StockHandler.SetMinStock))
							// FIFO batch endpoints — registered before wildcard so static paths win
							r.Get("/batches", withPerm(deps, "stock.read", deps.BatchStockHandler.ListBatches))
							r.Get("/batch-summary", withPerm(deps, "stock.read", deps.BatchStockHandler.GetSummary))
							r.Get("/{productId}", withPerm(deps, "stock.read", deps.StockHandler.GetProductStock))
						})

						// ── Phase 3 ───────────────────────────────────────────

						// Transactions (Cashier / POS)
						r.Route("/transactions", func(r chi.Router) {
							r.Get("/", withPerm(deps, "transactions.read", deps.TransactionHandler.List))
							r.Post("/", withPerm(deps, "transactions.create", deps.TransactionHandler.Checkout))
							// Draft / table order endpoints (restaurant)
							r.Get("/draft", withPerm(deps, "transactions.read", deps.TransactionHandler.GetDraftByTable))
							r.Post("/draft", withPerm(deps, "transactions.create", deps.TransactionHandler.CreateDraft))
							r.Get("/{txnId}", withPerm(deps, "transactions.read", deps.TransactionHandler.Get))
							r.Put("/{txnId}/draft", withPerm(deps, "transactions.create", deps.TransactionHandler.UpdateDraft))
							r.Post("/{txnId}/pay", withPerm(deps, "transactions.create", deps.TransactionHandler.PayDraft))
							r.Post("/{txnId}/void", withPerm(deps, "transactions.void", deps.TransactionHandler.Void))
						})

						// Purchase Orders
						r.Route("/purchase-orders", func(r chi.Router) {
							r.Get("/", withPerm(deps, "purchases.read", deps.PurchaseOrderHandler.List))
							r.Post("/", withPerm(deps, "purchases.create", deps.PurchaseOrderHandler.Create))
							r.Get("/payables", withPerm(deps, "purchases.read", deps.PurchaseOrderHandler.PayableSummary))
							r.Get("/{poId}", withPerm(deps, "purchases.read", deps.PurchaseOrderHandler.Get))
							r.Put("/{poId}", withPerm(deps, "purchases.update", deps.PurchaseOrderHandler.Update))
							r.Post("/{poId}/submit", withPerm(deps, "purchases.update", deps.PurchaseOrderHandler.Submit))
							r.Post("/{poId}/receive", withPerm(deps, "purchases.receive", deps.PurchaseOrderHandler.Receive))
							r.Delete("/{poId}", withPerm(deps, "purchases.delete", deps.PurchaseOrderHandler.Cancel))
							r.Get("/{poId}/payments", withPerm(deps, "purchases.read", deps.PurchaseOrderHandler.ListPayments))
							r.Post("/{poId}/payments", withPerm(deps, "purchases.update", deps.PurchaseOrderHandler.CreatePayment))
							// Termin (installment) routes
							r.Get("/{poId}/termins", withPerm(deps, "purchases.read", deps.TerminHandler.ListTermins))
							r.Post("/{poId}/termins", withPerm(deps, "purchases.update", deps.TerminHandler.CreateSchedule))
							r.Post("/{poId}/termins/{terminId}/payments", withPerm(deps, "purchases.update", deps.TerminHandler.RecordPayment))
							r.Get("/{poId}/debt", withPerm(deps, "purchases.read", deps.TerminHandler.GetDebtSummary))
							r.Get("/{poId}/document", withPerm(deps, "purchases.read", deps.TerminHandler.GetDocument))
						})

						// Reports
						r.Route("/reports", func(r chi.Router) {
							r.Get("/sales", withPerm(deps, "reports.read", deps.ReportHandler.SalesSummary))
							r.Get("/sales/by-product", withPerm(deps, "reports.read", deps.ReportHandler.SalesByProduct))
							r.Get("/sales/by-cashier", withPerm(deps, "reports.read", deps.ReportHandler.SalesByCashier))
							r.Get("/stock-valuation", withPerm(deps, "reports.read", deps.ReportHandler.StockValuation))
							r.Get("/profit", withPerm(deps, "reports.read", deps.ReportHandler.ProfitSummary))
						})

						// Expenses
						r.Route("/expenses", func(r chi.Router) {
							r.Get("/", withPerm(deps, "reports.read", deps.ExpenseHandler.ListExpenses))
							r.Post("/", withPerm(deps, "reports.read", deps.ExpenseHandler.CreateExpense))
							r.Put("/{id}", withPerm(deps, "reports.read", deps.ExpenseHandler.UpdateExpense))
							r.Put("/{id}", withPerm(deps, "reports.read", deps.ExpenseHandler.UpdateExpense))
							r.Delete("/{id}", withPerm(deps, "reports.read", deps.ExpenseHandler.DeleteExpense))
						})

						// Stock Adjustments
						r.Route("/adjustments", func(r chi.Router) {
							r.Get("/", withPerm(deps, "stock.read", deps.StockAdjustmentHandler.GetHistory))
							r.Post("/", withPerm(deps, "stock.update", deps.StockAdjustmentHandler.Create))
						})

						// Recurring Expenses
						r.Route("/recurring-expenses", func(r chi.Router) {
							r.Get("/", withPerm(deps, "reports.read", deps.ExpenseHandler.ListRecurringExpenses))
							r.Post("/", withPerm(deps, "reports.read", deps.ExpenseHandler.CreateRecurringExpense))
							r.Put("/{id}", withPerm(deps, "reports.read", deps.ExpenseHandler.UpdateRecurringExpense))
							r.Delete("/{id}", withPerm(deps, "reports.read", deps.ExpenseHandler.DeleteRecurringExpense))
						})

						// ── Restaurant Mode ───────────────────────────────────

						// Tables
						r.Route("/tables", func(r chi.Router) {
							r.Get("/", withPerm(deps, "products.read", deps.RestaurantHandler.ListTables))
							r.Post("/", withPerm(deps, "products.create", deps.RestaurantHandler.CreateTable))
							r.Put("/{tableId}", withPerm(deps, "products.update", deps.RestaurantHandler.UpdateTable))
							r.Put("/{tableId}/status", withPerm(deps, "products.update", deps.RestaurantHandler.UpdateTableStatus))
							r.Delete("/{tableId}", withPerm(deps, "products.delete", deps.RestaurantHandler.DeleteTable))
						})

						// Menu Items
						r.Route("/menu-items", func(r chi.Router) {
							r.Get("/", withPerm(deps, "products.read", deps.RestaurantHandler.ListMenuItems))
							r.Post("/", withPerm(deps, "products.create", deps.RestaurantHandler.CreateMenuItem))
							r.Put("/{menuItemId}", withPerm(deps, "products.update", deps.RestaurantHandler.UpdateMenuItem))
							r.Delete("/{menuItemId}", withPerm(deps, "products.delete", deps.RestaurantHandler.DeleteMenuItem))
						})

						// KDS (Kitchen Display System)
						r.Route("/kds", func(r chi.Router) {
							r.Get("/tickets", withPerm(deps, "transactions.read", deps.TransactionHandler.GetKDSTickets))
							r.Put("/items/{itemId}", withPerm(deps, "transactions.create", deps.TransactionHandler.UpdateKDSItemStatus))
						})
					})
				})
			})
		})
	})

	return r
}

// withPerm wraps a HandlerFunc with the RequirePermission middleware.
func withPerm(deps *Dependencies, perm string, fn http.HandlerFunc) http.HandlerFunc {
	return middleware.RequirePermission(deps.RoleStore, perm)(fn).ServeHTTP
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
