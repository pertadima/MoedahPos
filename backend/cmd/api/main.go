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
	defer sqlxDB.Close()

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
	jwtMgr   := jwt.New(cfg.JWT.Secret, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL)
	validate := validator.New()

	// ── Repositories ──────────────────────────────────────────────────────────
	userRepo         := postgres.NewUserRepo(sqlxDB)
	refreshTokenRepo := postgres.NewRefreshTokenRepo(sqlxDB)
	storeRepo        := postgres.NewStoreRepo(sqlxDB)
	categoryRepo     := postgres.NewCategoryRepo(sqlxDB)
	productRepo      := postgres.NewProductRepo(sqlxDB)
	stockRepo        := postgres.NewStockRepo(sqlxDB)
	transactionRepo  := postgres.NewTransactionRepo(sqlxDB)
	poRepo           := postgres.NewPORepo(sqlxDB)
	supplierRepo     := postgres.NewSupplierRepo(sqlxDB)
	reportRepo       := postgres.NewReportRepo(sqlxDB)
	tableRepo        := postgres.NewTableRepo(sqlxDB)
	menuItemRepo     := postgres.NewMenuItemRepo(sqlxDB)

	// ── Services ──────────────────────────────────────────────────────────────
	authSvc        := service.NewAuthService(userRepo, refreshTokenRepo, jwtMgr, cfg.Bcrypt.Cost, log)
	storeSvc       := service.NewStoreService(storeRepo, userRepo, log)
	productSvc     := service.NewProductService(productRepo, categoryRepo, stockRepo, log)
	stockSvc       := service.NewStockService(stockRepo, productRepo, log)
	transactionSvc := service.NewTransactionService(transactionRepo, productRepo, stockRepo, menuItemRepo, log)
	poSvc          := service.NewPurchaseOrderService(poRepo, productRepo, log)
	supplierSvc    := service.NewSupplierService(supplierRepo, log)
	reportSvc      := service.NewReportService(reportRepo, log)
	tableSvc       := service.NewTableService(tableRepo, log)
	menuItemSvc    := service.NewMenuItemService(menuItemRepo, log)

	// ── Handlers ──────────────────────────────────────────────────────────────
	authHandler        := handler.NewAuthHandler(authSvc, validate, log)
	storeHandler       := handler.NewStoreHandler(storeSvc, validate, log)
	productHandler     := handler.NewProductHandler(productSvc, validate, log)
	stockHandler       := handler.NewStockHandler(stockSvc, validate, log)
	transactionHandler := handler.NewTransactionHandler(transactionSvc, validate, log)
	poHandler          := handler.NewPurchaseOrderHandler(poSvc, validate, log)
	supplierHandler    := handler.NewSupplierHandler(supplierSvc, validate, log)
	reportHandler      := handler.NewReportHandler(reportSvc, log)
	restaurantHandler  := handler.NewRestaurantHandler(tableSvc, menuItemSvc, validate, log)

	// ── Router ────────────────────────────────────────────────────────────────
	r := router.New(&router.Dependencies{
		AuthHandler:           authHandler,
		JWTManager:            jwtMgr,
		StoreHandler:          storeHandler,
		ProductHandler:        productHandler,
		StockHandler:          stockHandler,
		TransactionHandler:    transactionHandler,
		PurchaseOrderHandler:  poHandler,
		SupplierHandler:       supplierHandler,
		ReportHandler:         reportHandler,
		RestaurantHandler:     restaurantHandler,
		RoleStore:             roleStore,
		DB:                    sqlxDB,
		Log:                   log,
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
	quit := make(chan os.Signal, 1)
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
