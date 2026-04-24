package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/moedahpos/backend/internal/config"
	"github.com/moedahpos/backend/internal/handler"
	"github.com/moedahpos/backend/internal/repository/postgres"
	"github.com/moedahpos/backend/internal/router"
	"github.com/moedahpos/backend/internal/service"
	"github.com/moedahpos/backend/internal/validator"
	"github.com/moedahpos/backend/pkg/db"
	"github.com/moedahpos/backend/pkg/jwt"
	"github.com/moedahpos/backend/pkg/logger"
	"github.com/moedahpos/backend/pkg/rbac"
)

//nolint:gocognit,funlen // bootstrap wiring is inherently long
func main() {
	// ── Config ────────────────────────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	// ── Logger ────────────────────────────────────────────────────────────────
	log := logger.New(cfg.App.Env)
	log.Info().Str("env", cfg.App.Env).Msg("starting moedah-pos api")

	// ── Database ──────────────────────────────────────────────────────────────
	sqlxDB, err := db.Connect(&cfg.DB, log)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer func() {
		if cerr := sqlxDB.Close(); cerr != nil {
			log.Error().Err(cerr).Msg("error closing database")
		}
	}()

	// ── Migrations ────────────────────────────────────────────────────────────
	if cfg.Migration.Run {
		if err := db.RunMigrations(sqlxDB, cfg.Migration.Dir, log); err != nil {
			log.Fatal().Err(err).Msg("failed to run migrations")
		}
	}

	// ── RBAC (loaded once at startup) ─────────────────────────────────────────
	roleStore, err := rbac.New(sqlxDB)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load RBAC role store")
	}
	log.Info().Msg("RBAC role store loaded")

	// ── Shared Utilities ──────────────────────────────────────────────────────
	jwtMgr := jwt.New(cfg.JWT.Secret, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL)
	validate := validator.New()

	// ── Repositories ──────────────────────────────────────────────────────────
	dbRaw := sqlxDB.DB
	userRepo := postgres.NewUserRepository(dbRaw)
	refreshTokenRepo := postgres.NewRefreshTokenRepository(dbRaw)
	storeRepo := postgres.NewStoreRepository(dbRaw)
	categoryRepo := postgres.NewCategoryRepository(dbRaw)
	productRepo := postgres.NewProductRepository(dbRaw)
	stockRepo := postgres.NewStockRepository(dbRaw)
	transactionRepo := postgres.NewTransactionRepository(dbRaw)
	poRepo := postgres.NewPurchaseOrderRepository(dbRaw)
	supplierRepo := postgres.NewSupplierRepository(dbRaw)
	reportRepo := postgres.NewReportRepository(dbRaw)
	tableRepo := postgres.NewTableRepository(dbRaw)
	menuItemRepo := postgres.NewMenuItemRepository(dbRaw)
	priceHistoryRepo := postgres.NewPriceHistoryRepository(dbRaw)
	poPaymentRepo := postgres.NewPOPaymentRepository(dbRaw)
	customerRepo := postgres.NewCustomerRepository(dbRaw)
	roleRepo := postgres.NewRoleRepository(dbRaw)
	batchRepo := postgres.NewBatchRepository(dbRaw)                 // FIFO batch inventory
	terminRepo := postgres.NewTerminRepository(dbRaw)               // PO installment schedule
	paymentRecordRepo := postgres.NewPaymentRecordRepository(dbRaw) // PO payment records
	expenseRepo := postgres.NewExpenseRepository(dbRaw)
	stockAdjustmentRepo := postgres.NewStockAdjustmentRepository(dbRaw)
	incomeRepo := postgres.NewIncomeRepository(dbRaw)
	activityLogRepo := postgres.NewActivityLogRepository(dbRaw)
	loyaltyRepo := postgres.NewLoyaltyRepository(dbRaw)
	tierRepo := postgres.NewMembershipTierRepository(dbRaw)

	// ── Services ──────────────────────────────────────────────────────────────
	batchSvc := service.NewBatchStockService(batchRepo, log) // FIFO batch inventory
	activityLogSvc := service.NewActivityLogService(activityLogRepo, log)
	authSvc := service.NewAuthService(userRepo, refreshTokenRepo, activityLogSvc, jwtMgr, cfg.Bcrypt.Cost, log)
	storeSvc := service.NewStoreService(storeRepo, userRepo, log)
	priceHistorySvc := service.NewPriceHistoryService(priceHistoryRepo, log)
	productSvc := service.NewProductService(productRepo, categoryRepo, stockRepo, priceHistorySvc, log)
	stockSvc := service.NewStockService(stockRepo, productRepo, log)
	transactionSvc := service.NewTransactionService(transactionRepo, productRepo, stockRepo, menuItemRepo, batchSvc, activityLogSvc, storeRepo, loyaltyRepo, log)
	poSvc := service.NewPurchaseOrderService(poRepo, productRepo, poPaymentRepo, terminRepo, paymentRecordRepo, priceHistorySvc, activityLogSvc, log)
	supplierSvc := service.NewSupplierService(supplierRepo, log)
	reportSvc := service.NewReportService(reportRepo, log)
	tableSvc := service.NewTableService(tableRepo, log)
	menuItemSvc := service.NewMenuItemService(menuItemRepo, log)
	customerSvc := service.NewCustomerService(customerRepo, log)
	userAdminSvc := service.NewUserAdminService(userRepo, roleRepo, cfg.Bcrypt.Cost, log)
	terminSvc := service.NewTerminService(terminRepo, paymentRecordRepo, poRepo, storeRepo, activityLogSvc, log)
	expenseSvc := service.NewExpenseService(expenseRepo, activityLogSvc, log)
	stockAdjustmentSvc := service.NewStockAdjustmentService(stockAdjustmentRepo, activityLogSvc)
	incomeSvc := service.NewIncomeService(incomeRepo, activityLogSvc, log)
	syncSvc := service.NewSyncService(categoryRepo, productRepo, stockRepo, transactionRepo, customerRepo, log)
	loyaltySvc := service.NewLoyaltyService(loyaltyRepo, tierRepo, customerRepo, storeRepo, log)

	// ── Handlers ──────────────────────────────────────────────────────────────
	authHandler := handler.NewAuthHandler(authSvc, validate, log)
	storeHandler := handler.NewStoreHandler(storeSvc, validate, log)
	productHandler := handler.NewProductHandler(productSvc, validate, log)
	stockHandler := handler.NewStockHandler(stockSvc, validate, log)
	transactionHandler := handler.NewTransactionHandler(transactionSvc, validate, log)
	poHandler := handler.NewPurchaseOrderHandler(poSvc, validate, log)
	supplierHandler := handler.NewSupplierHandler(supplierSvc, validate, log)
	reportHandler := handler.NewReportHandler(reportSvc, log)
	restaurantHandler := handler.NewRestaurantHandler(tableSvc, menuItemSvc, validate, log)
	priceHistoryHandler := handler.NewPriceHistoryHandler(priceHistorySvc, log)
	customerHandler := handler.NewCustomerHandler(customerSvc, validate, log)
	userAdminHandler := handler.NewUserAdminHandler(userAdminSvc, validate, log)
	batchStockHandler := handler.NewBatchStockHandler(batchSvc, log) // FIFO batch inventory
	terminHandler := handler.NewTerminHandler(terminSvc, validate, log)
	expenseHandler := handler.NewExpenseHandler(expenseSvc, validate, log)
	stockAdjustmentHandler := handler.NewStockAdjustmentHandler(stockAdjustmentSvc, validate, &log)
	incomeHandler := handler.NewIncomeHandler(incomeSvc, validate, log)
	activityLogHandler := handler.NewActivityLogHandler(activityLogSvc, log)
	syncHandler := handler.NewSyncHandler(syncSvc, log)
	loyaltyHandler := handler.NewLoyaltyHandler(loyaltySvc, validate, log)

	// ── Router ────────────────────────────────────────────────────────────────
	r := router.New(&router.Dependencies{
		AuthHandler:            authHandler,
		JWTManager:             jwtMgr,
		StoreHandler:           storeHandler,
		ProductHandler:         productHandler,
		StockHandler:           stockHandler,
		TransactionHandler:     transactionHandler,
		PurchaseOrderHandler:   poHandler,
		SupplierHandler:        supplierHandler,
		ReportHandler:          reportHandler,
		RestaurantHandler:      restaurantHandler,
		PriceHistoryHandler:    priceHistoryHandler,
		CustomerHandler:        customerHandler,
		UserAdminHandler:       userAdminHandler,
		BatchStockHandler:      batchStockHandler,
		TerminHandler:          terminHandler,
		ExpenseHandler:         expenseHandler,
		StockAdjustmentHandler: stockAdjustmentHandler,
		IncomeHandler:          incomeHandler,
		ActivityLogHandler:     activityLogHandler,
		SyncHandler:            syncHandler,
		LoyaltyHandler:         loyaltyHandler,
		RoleStore:              roleStore,
		DB:                     sqlxDB,
		Log:                    log,
	})

	// ── HTTP Server ───────────────────────────────────────────────────────────
	srv := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info().Str("addr", srv.Addr).Msg("server listening")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	// ── Graceful Shutdown ─────────────────────────────────────────────────────
	var quit = make(chan os.Signal, 1)

	// ── Background Workers ────────────────────────────────────────────────────
	go func() {
		ticker := time.NewTicker(5 * time.Minute) // Check every 5 minutes in production (can be tweaked)
		defer ticker.Stop()
		for {
			select {
			case <-quit:
				return
			case <-ticker.C:
				if err := expenseSvc.ProcessDueRecurringExpenses(context.Background()); err != nil {
					log.Error().Err(err).Msg("background worker: failed to process recurring expenses")
				}
			}
		}
	}()

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("forced shutdown")
	}
	log.Info().Msg("server stopped")
}
