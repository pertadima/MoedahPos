package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
	"github.com/moedahpos/backend/internal/handler"
	"github.com/moedahpos/backend/internal/middleware"
	"github.com/moedahpos/backend/pkg/jwt"
	"github.com/moedahpos/backend/pkg/rbac"
	"github.com/rs/zerolog"
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
	TransactionHandler    *handler.TransactionHandler
	PurchaseOrderHandler  *handler.PurchaseOrderHandler
	SupplierHandler       *handler.SupplierHandler
	ReportHandler         *handler.ReportHandler

	// Restaurant mode (Phase 4)
	RestaurantHandler *handler.RestaurantHandler

	// Shared
	RoleStore *rbac.RoleStore
	DB        *sqlx.DB
	Log       zerolog.Logger
}

// New configures and returns the full application router.
func New(deps *Dependencies) http.Handler {
	r := chi.NewRouter()

	// ── Global Middleware ──────────────────────────────────────────────────────
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.Logger(deps.Log))
	r.Use(corsMiddleware)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	r.Route("/api/v1", func(r chi.Router) {

		// ── Public Auth ────────────────────────────────────────────────────────
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", deps.AuthHandler.Register)
			r.Post("/login",    deps.AuthHandler.Login)
			r.Post("/refresh",  deps.AuthHandler.Refresh)
			r.Group(func(r chi.Router) {
				r.Use(middleware.Authenticate(deps.JWTManager))
				r.Post("/logout", deps.AuthHandler.Logout)
				r.Get("/me",      deps.AuthHandler.Me)
			})
		})

		// ── Authenticated-only ────────────────────────────────────────────────
		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticate(deps.JWTManager))

			// ── Suppliers (global, no store context needed) ────────────────────
			r.Route("/suppliers", func(r chi.Router) {
				r.Get("/",                 deps.SupplierHandler.List)
				r.Post("/",                withPerm(deps, "suppliers.create", deps.SupplierHandler.Create))
				r.Get("/{supplierId}",     deps.SupplierHandler.Get)
				r.Put("/{supplierId}",     withPerm(deps, "suppliers.update", deps.SupplierHandler.Update))
				r.Delete("/{supplierId}",  withPerm(deps, "suppliers.delete", deps.SupplierHandler.Delete))
			})

			// ── Store-scoped routes ────────────────────────────────────────────
			r.Route("/stores", func(r chi.Router) {
				r.Get("/",   deps.StoreHandler.List)
				r.Post("/",  deps.StoreHandler.Create) // JWT-auth only; superadmin bypasses store-scoped RBAC

				r.Route("/{storeId}", func(r chi.Router) {
					r.Get("/", deps.StoreHandler.Get)

					// Global store CRUD — JWT-auth only (no store-scoped role required)
					r.Put("/",    deps.StoreHandler.Update)
					r.Delete("/", deps.StoreHandler.Delete)

					r.Group(func(r chi.Router) {
						r.Use(middleware.StoreContext(deps.DB))

						// Members
						r.Route("/members", func(r chi.Router) {
							r.Get("/",            withPerm(deps, "users.read",   deps.StoreHandler.ListMembers))
							r.Post("/",           withPerm(deps, "users.create", deps.StoreHandler.AddMember))
							r.Put("/{userId}",    withPerm(deps, "users.update", deps.StoreHandler.UpdateMemberRole))
							r.Delete("/{userId}", withPerm(deps, "users.delete", deps.StoreHandler.RemoveMember))
						})

						// Categories
						r.Route("/categories", func(r chi.Router) {
							r.Get("/",               withPerm(deps, "products.read",   deps.ProductHandler.ListCategories))
							r.Post("/",              withPerm(deps, "products.create", deps.ProductHandler.CreateCategory))
							r.Put("/{categoryId}",   withPerm(deps, "products.update", deps.ProductHandler.UpdateCategory))
							r.Delete("/{categoryId}", withPerm(deps, "products.delete", deps.ProductHandler.DeleteCategory))
						})

						// Products
						r.Route("/products", func(r chi.Router) {
							r.Get("/",                        withPerm(deps, "products.read",   deps.ProductHandler.List))
							r.Post("/",                       withPerm(deps, "products.create", deps.ProductHandler.Create))
							r.Get("/barcode/{barcode}",       withPerm(deps, "products.read",   deps.ProductHandler.GetByBarcode))
							r.Get("/{productId}",             withPerm(deps, "products.read",   deps.ProductHandler.Get))
							r.Put("/{productId}",             withPerm(deps, "products.update", deps.ProductHandler.Update))
							r.Delete("/{productId}",          withPerm(deps, "products.delete", deps.ProductHandler.Delete))
						})

						// Stock
						r.Route("/stock", func(r chi.Router) {
							r.Get("/",             withPerm(deps, "stock.read",   deps.StockHandler.GetLevels))
							r.Get("/low",          withPerm(deps, "stock.read",   deps.StockHandler.GetLowStock))
							r.Get("/movements",    withPerm(deps, "stock.read",   deps.StockHandler.GetMovements))
							r.Post("/adjust",      withPerm(deps, "stock.adjust", deps.StockHandler.Adjust))
							r.Put("/min",          withPerm(deps, "stock.adjust", deps.StockHandler.SetMinStock))
							r.Get("/{productId}",  withPerm(deps, "stock.read",   deps.StockHandler.GetProductStock))
						})

						// ── Phase 3 ───────────────────────────────────────────

						// Transactions (Cashier / POS)
						r.Route("/transactions", func(r chi.Router) {
							r.Get("/",            withPerm(deps, "transactions.read",   deps.TransactionHandler.List))
							r.Post("/",           withPerm(deps, "transactions.create", deps.TransactionHandler.Checkout))
							r.Get("/{txnId}",     withPerm(deps, "transactions.read",   deps.TransactionHandler.Get))
							r.Post("/{txnId}/void", withPerm(deps, "transactions.void", deps.TransactionHandler.Void))
						})

						// Purchase Orders
						r.Route("/purchase-orders", func(r chi.Router) {
							r.Get("/",                    withPerm(deps, "purchases.read",   deps.PurchaseOrderHandler.List))
							r.Post("/",                   withPerm(deps, "purchases.create", deps.PurchaseOrderHandler.Create))
							r.Get("/{poId}",              withPerm(deps, "purchases.read",   deps.PurchaseOrderHandler.Get))
							r.Put("/{poId}",              withPerm(deps, "purchases.update", deps.PurchaseOrderHandler.Update))
							r.Post("/{poId}/submit",      withPerm(deps, "purchases.update", deps.PurchaseOrderHandler.Submit))
							r.Post("/{poId}/receive",     withPerm(deps, "purchases.receive", deps.PurchaseOrderHandler.Receive))
							r.Delete("/{poId}",           withPerm(deps, "purchases.delete", deps.PurchaseOrderHandler.Cancel))
						})

						// Reports
						r.Route("/reports", func(r chi.Router) {
							r.Get("/sales",             withPerm(deps, "reports.read", deps.ReportHandler.SalesSummary))
							r.Get("/sales/by-product",  withPerm(deps, "reports.read", deps.ReportHandler.SalesByProduct))
							r.Get("/sales/by-cashier",  withPerm(deps, "reports.read", deps.ReportHandler.SalesByCashier))
							r.Get("/stock-valuation",   withPerm(deps, "reports.read", deps.ReportHandler.StockValuation))
						})

						// ── Restaurant Mode ───────────────────────────────────

						// Tables
						r.Route("/tables", func(r chi.Router) {
							r.Get("/",                        withPerm(deps, "products.read",   deps.RestaurantHandler.ListTables))
							r.Post("/",                       withPerm(deps, "products.create", deps.RestaurantHandler.CreateTable))
							r.Put("/{tableId}",               withPerm(deps, "products.update", deps.RestaurantHandler.UpdateTable))
							r.Put("/{tableId}/status",        withPerm(deps, "products.update", deps.RestaurantHandler.UpdateTableStatus))
							r.Delete("/{tableId}",            withPerm(deps, "products.delete", deps.RestaurantHandler.DeleteTable))
						})

						// Menu Items
						r.Route("/menu-items", func(r chi.Router) {
							r.Get("/",                        withPerm(deps, "products.read",   deps.RestaurantHandler.ListMenuItems))
							r.Post("/",                       withPerm(deps, "products.create", deps.RestaurantHandler.CreateMenuItem))
							r.Put("/{menuItemId}",            withPerm(deps, "products.update", deps.RestaurantHandler.UpdateMenuItem))
							r.Delete("/{menuItemId}",         withPerm(deps, "products.delete", deps.RestaurantHandler.DeleteMenuItem))
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
