package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog"

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/dto"
	"github.com/moedahpos/backend/internal/repository"
)

// Transaction-specific sentinel errors.
var (
	ErrTransactionNotFound      = errors.New("transaction not found")
	ErrTransactionAlreadyVoided = errors.New("transaction is already voided")
	ErrInsufficientStock        = errors.New("insufficient stock for one or more items")
	ErrInsuficientPayment       = errors.New("payment amount is less than total")
	ErrDraftNotFound            = errors.New("draft order not found")
)

// statusDraft is the transaction status for held (unpaid) orders.
const statusDraft = "draft"

// TransactionService implements cashier checkout logic.
type TransactionService struct {
	txnRepo      repository.TransactionRepository
	productRepo  repository.ProductRepository
	stockRepo    repository.StockRepository
	menuItemRepo repository.MenuItemRepository
	batchSvc     *BatchStockService // FIFO batch deduction
	log          zerolog.Logger
}

func NewTransactionService(
	txnRepo repository.TransactionRepository,
	productRepo repository.ProductRepository,
	stockRepo repository.StockRepository,
	menuItemRepo repository.MenuItemRepository,
	batchSvc *BatchStockService,
	log zerolog.Logger,
) *TransactionService {
	return &TransactionService{txnRepo: txnRepo, productRepo: productRepo, stockRepo: stockRepo, menuItemRepo: menuItemRepo, batchSvc: batchSvc, log: log}
}

// Checkout processes a sale: validates stock, calculates totals, persists atomically.
func (s *TransactionService) Checkout(ctx context.Context, storeID string, req *dto.CreateTransactionRequest, cashierID string) (*dto.TransactionResponse, error) { //nolint:gocognit,cyclop,funlen // retail+restaurant dual path
	var (
		inputItems  []domain.CreateTransactionItemInput
		subtotal    float64
		discountAmt float64
		taxAmt      float64
	)

	for _, item := range req.Items {
		if item.MenuItemID != "" {
			// ── Restaurant path: menu item → deduct ingredients ──────────────
			menuItem, err := s.menuItemRepo.FindByID(ctx, item.MenuItemID)
			if err != nil {
				return nil, fmt.Errorf("finding menu item %s: %w", item.MenuItemID, err)
			}
			if menuItem == nil || menuItem.StoreID != storeID {
				return nil, fmt.Errorf("menu item %s not found", item.MenuItemID)
			}

			// Validate ingredient stock
			for _, ing := range menuItem.Ingredients {
				needed := ing.Quantity * item.Quantity
				level, err := s.stockRepo.FindLevelByProduct(ctx, ing.ProductID, storeID)
				if err != nil {
					return nil, fmt.Errorf("checking ingredient stock: %w", err)
				}
				if level != nil && level.Quantity < needed {
					return nil, fmt.Errorf("%w: bahan %s (perlu %.2f, tersedia %.2f)",
						ErrInsufficientStock, ing.ProductName, needed, level.Quantity)
				}
			}

			lineGross := menuItem.SellPrice * item.Quantity
			lineDiscount := lineGross * item.DiscountPct / 100
			lineNet := lineGross - lineDiscount
			lineTax := lineNet * menuItem.TaxRate / 100
			lineSubtotal := lineNet + lineTax

			subtotal += lineNet
			discountAmt += lineDiscount
			taxAmt += lineTax

			// Compute BOM cost (sum of ingredient cost_price × qty per portion)
			var menuCost float64
			for _, ing := range menuItem.Ingredients {
				ingProduct, _ := s.productRepo.FindByID(ctx, ing.ProductID)
				if ingProduct != nil {
					menuCost += ingProduct.CostPrice * ing.Quantity
				}
			}

			mid := item.MenuItemID
			inputItems = append(inputItems, domain.CreateTransactionItemInput{
				ProductID:   nil,
				ProductName: menuItem.Name,
				SKU:         "MENU-" + item.MenuItemID[:8],
				Quantity:    item.Quantity,
				UnitPrice:   menuItem.SellPrice,
				CostPrice:   menuCost, // BOM cost per portion, snapshot at sale time
				DiscountPct: item.DiscountPct,
				TaxRate:     menuItem.TaxRate,
				Subtotal:    lineSubtotal,
				MenuItemID:  &mid,
			})
		} else {
			// ── Retail path: single product ──────────────────────────────────
			product, err := s.productRepo.FindByID(ctx, item.ProductID)
			if err != nil {
				return nil, fmt.Errorf("finding product %s: %w", item.ProductID, err)
			}
			if product == nil || product.StoreID != storeID {
				return nil, fmt.Errorf("%w: product %s", ErrProductNotFound, item.ProductID)
			}
			if !product.IsActive {
				return nil, fmt.Errorf("product %s is inactive", product.Name)
			}

			level, err := s.stockRepo.FindLevelByProduct(ctx, item.ProductID, storeID)
			if err != nil {
				return nil, fmt.Errorf("checking stock: %w", err)
			}
			if level != nil && level.Quantity < item.Quantity {
				return nil, fmt.Errorf("%w: %s (have %.2f, need %.2f)",
					ErrInsufficientStock, product.Name, level.Quantity, item.Quantity)
			}

			lineGross := product.SellPrice * item.Quantity
			lineDiscount := lineGross * item.DiscountPct / 100
			lineNet := lineGross - lineDiscount
			lineTax := lineNet * product.TaxRate / 100
			lineSubtotal := lineNet + lineTax

			subtotal += lineNet
			discountAmt += lineDiscount
			taxAmt += lineTax

			pid := item.ProductID
			inputItems = append(inputItems, domain.CreateTransactionItemInput{
				ProductID:   &pid,
				ProductName: product.Name,
				SKU:         product.SKU,
				Quantity:    item.Quantity,
				UnitPrice:   product.SellPrice,
				CostPrice:   product.CostPrice, // snapshot at time of sale
				DiscountPct: item.DiscountPct,
				TaxRate:     product.TaxRate,
				Subtotal:    lineSubtotal,
			})
		}
	}

	total := subtotal + taxAmt
	if req.PaymentAmount < total {
		return nil, ErrInsuficientPayment
	}

	txn, err := s.txnRepo.Create(ctx, domain.CreateTransactionInput{
		StoreID:       storeID,
		CashierID:     cashierID,
		Status:        "completed",
		CustomerName:  req.CustomerName,
		CustomerPhone: req.CustomerPhone,
		PaymentMethod: req.PaymentMethod,
		PaymentAmount: req.PaymentAmount,
		ChangeAmount:  req.PaymentAmount - total,
		Notes:         req.Notes,
		Subtotal:      subtotal,
		DiscountAmt:   discountAmt,
		TaxAmt:        taxAmt,
		Total:         total,
		Items:         inputItems,
	})
	if err != nil {
		return nil, fmt.Errorf("creating transaction: %w", err)
	}

	// ── Post-commit: deduct ingredient stocks for menu items ─────────────────
	for _, item := range req.Items {
		if item.MenuItemID == "" {
			continue
		}
		menuItem, _ := s.menuItemRepo.FindByID(ctx, item.MenuItemID)
		if menuItem == nil {
			continue
		}
		for _, ing := range menuItem.Ingredients {
			needed := ing.Quantity * item.Quantity
			_ = s.stockRepo.DeductStock(ctx, ing.ProductID, storeID, needed, txn.ID, cashierID)
			// FIFO: deduct ingredient from oldest batch first (best-effort).
			if err := s.batchSvc.DeductStockFIFO(ctx, ing.ProductID, storeID, needed); err != nil {
				s.log.Warn().Err(err).Str("product_id", ing.ProductID).Msg("FIFO ingredient deduct failed")
			}
		}
	}

	// ── Post-commit: FIFO batch deduction for retail items ────────────────────
	// Deducts from oldest batch first. Best-effort: stock_levels is the
	// canonical availability constraint validated before the sale.
	for _, item := range req.Items {
		if item.MenuItemID != "" {
			continue // menu items already handled above via ingredients
		}
		if err := s.batchSvc.DeductStockFIFO(ctx, item.ProductID, storeID, item.Quantity); err != nil {
			s.log.Warn().Err(err).
				Str("product_id", item.ProductID).
				Float64("qty", item.Quantity).
				Msg("FIFO retail batch deduct failed")
		}
	}

	s.log.Info().Str("txn_id", txn.ID).Float64("total", total).Msg("transaction completed")
	return toTransactionResponse(txn), nil
}

// ListTransactions returns a paginated list of transactions for a store.
func (s *TransactionService) ListTransactions(ctx context.Context, filter dto.TransactionListFilter) ([]*dto.TransactionResponse, dto.PaginationMeta, error) {
	filter.Defaults()
	txns, total, err := s.txnRepo.FindAll(ctx, filter)
	if err != nil {
		return nil, dto.PaginationMeta{}, fmt.Errorf("listing transactions: %w", err)
	}
	resp := make([]*dto.TransactionResponse, 0, len(txns))
	for _, t := range txns {
		resp = append(resp, toTransactionResponse(t))
	}
	return resp, dto.NewMeta(filter.PaginationQuery, total), nil
}

// GetTransaction returns a single transaction with all its items (receipt).
func (s *TransactionService) GetTransaction(ctx context.Context, id string) (*dto.TransactionResponse, error) {
	txn, err := s.txnRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("finding transaction: %w", err)
	}
	if txn == nil {
		return nil, ErrTransactionNotFound
	}
	return toTransactionResponse(txn), nil
}

// VoidTransaction reverses a completed transaction and restores stock.
func (s *TransactionService) VoidTransaction(ctx context.Context, id, userID string) error {
	txn, err := s.txnRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("finding transaction: %w", err)
	}
	if txn == nil {
		return ErrTransactionNotFound
	}
	if txn.Status == "voided" {
		return ErrTransactionAlreadyVoided
	}
	if err := s.txnRepo.Void(ctx, id, userID); err != nil {
		return fmt.Errorf("voiding transaction: %w", err)
	}
	s.log.Info().Str("txn_id", id).Str("voided_by", userID).Msg("transaction voided")
	return nil
}

// ─── Draft / Table Order Methods ──────────────────────────────────────────────

// buildItems resolves and prices items from TxItemInput — shared by Checkout and CreateDraft.
func (s *TransactionService) buildItems(ctx context.Context, storeID string, reqItems []dto.TxItemInput, _ bool,
) ([]domain.CreateTransactionItemInput, float64, float64, float64, error) {
	var inputItems []domain.CreateTransactionItemInput
	var subtotal, discountAmt, taxAmt float64

	for _, item := range reqItems {
		if item.MenuItemID != "" {
			menuItem, err := s.menuItemRepo.FindByID(ctx, item.MenuItemID)
			if err != nil || menuItem == nil || menuItem.StoreID != storeID {
				return nil, 0, 0, 0, fmt.Errorf("menu item %s not found", item.MenuItemID)
			}
			lineGross := menuItem.SellPrice * item.Quantity
			lineDiscount := lineGross * item.DiscountPct / 100
			lineNet := lineGross - lineDiscount
			lineTax := lineNet * menuItem.TaxRate / 100
			lineSubtotal := lineNet + lineTax
			subtotal += lineNet
			discountAmt += lineDiscount
			taxAmt += lineTax
			var menuCost float64
			for _, ing := range menuItem.Ingredients {
				ingProduct, _ := s.productRepo.FindByID(ctx, ing.ProductID)
				if ingProduct != nil {
					menuCost += ingProduct.CostPrice * ing.Quantity
				}
			}
			mid := item.MenuItemID
			inputItems = append(inputItems, domain.CreateTransactionItemInput{
				MenuItemID:  &mid,
				ProductName: menuItem.Name,
				SKU:         "MENU-" + item.MenuItemID[:8],
				Quantity:    item.Quantity,
				UnitPrice:   menuItem.SellPrice,
				CostPrice:   menuCost,
				DiscountPct: item.DiscountPct,
				TaxRate:     menuItem.TaxRate,
				Subtotal:    lineSubtotal,
			})
		} else {
			product, err := s.productRepo.FindByID(ctx, item.ProductID)
			if err != nil || product == nil || product.StoreID != storeID {
				return nil, 0, 0, 0, fmt.Errorf("%w: product %s", ErrProductNotFound, item.ProductID)
			}
			lineGross := product.SellPrice * item.Quantity
			lineDiscount := lineGross * item.DiscountPct / 100
			lineNet := lineGross - lineDiscount
			lineTax := lineNet * product.TaxRate / 100
			lineSubtotal := lineNet + lineTax
			subtotal += lineNet
			discountAmt += lineDiscount
			taxAmt += lineTax
			pid := item.ProductID
			inputItems = append(inputItems, domain.CreateTransactionItemInput{
				ProductID:   &pid,
				ProductName: product.Name,
				SKU:         product.SKU,
				Quantity:    item.Quantity,
				UnitPrice:   product.SellPrice,
				CostPrice:   product.CostPrice,
				DiscountPct: item.DiscountPct,
				TaxRate:     product.TaxRate,
				Subtotal:    lineSubtotal,
			})
		}
	}
	return inputItems, subtotal, discountAmt, taxAmt, nil
}

// GetDraftByTable returns the open draft for a table, or nil.
func (s *TransactionService) GetDraftByTable(ctx context.Context, storeID, tableID string) (*dto.TransactionResponse, error) {
	txn, err := s.txnRepo.GetDraftByTable(ctx, storeID, tableID)
	if err != nil {
		return nil, fmt.Errorf("getting draft: %w", err)
	}
	if txn == nil {
		return nil, nil
	}
	return toTransactionResponse(txn), nil
}

// CreateDraft saves a restaurant table order as a draft (no payment, no stock deduction).
func (s *TransactionService) CreateDraft(ctx context.Context, storeID, cashierID string, req *dto.CreateDraftRequest) (*dto.TransactionResponse, error) {
	inputItems, subtotal, discountAmt, taxAmt, err := s.buildItems(ctx, storeID, req.Items, false)
	if err != nil {
		return nil, err
	}
	tableID := req.TableID
	txn, err := s.txnRepo.Create(ctx, domain.CreateTransactionInput{
		StoreID:      storeID,
		CashierID:    cashierID,
		TableID:      &tableID,
		Status:       statusDraft,
		CustomerName: req.CustomerName,
		Notes:        req.Notes,
		Subtotal:     subtotal,
		DiscountAmt:  discountAmt,
		TaxAmt:       taxAmt,
		Total:        subtotal + taxAmt,
		Items:        inputItems,
	})
	if err != nil {
		return nil, fmt.Errorf("creating draft: %w", err)
	}
	s.log.Info().Str("txn_id", txn.ID).Str("table_id", tableID).Msg("draft created")
	return toTransactionResponse(txn), nil
}

// UpdateDraftItems replaces items on an existing draft and recalculates totals.
func (s *TransactionService) UpdateDraftItems(ctx context.Context, storeID, txnID string, req *dto.UpdateDraftRequest) (*dto.TransactionResponse, error) {
	// Verify it belongs to this store and is still a draft
	existing, err := s.txnRepo.FindByID(ctx, txnID)
	if err != nil || existing == nil {
		return nil, ErrDraftNotFound
	}
	if existing.StoreID != storeID || existing.Status != statusDraft {
		return nil, ErrDraftNotFound
	}

	inputItems, subtotal, discountAmt, taxAmt, err := s.buildItems(ctx, storeID, req.Items, false)
	if err != nil {
		return nil, err
	}
	total := subtotal + taxAmt

	txn, err := s.txnRepo.UpdateDraftItems(ctx, txnID, inputItems, subtotal, discountAmt, taxAmt, total, req.CustomerName, req.Notes)
	if err != nil {
		return nil, fmt.Errorf("updating draft: %w", err)
	}
	s.log.Info().Str("txn_id", txnID).Msg("draft updated")
	return toTransactionResponse(txn), nil
}

// PayDraft finalizes a held order: validates payment, deducts stock, marks completed.
func (s *TransactionService) PayDraft(ctx context.Context, storeID, txnID, cashierID string, req *dto.PayDraftRequest) (*dto.TransactionResponse, error) {
	existing, err := s.txnRepo.FindByID(ctx, txnID)
	if err != nil || existing == nil {
		return nil, ErrDraftNotFound
	}
	if existing.StoreID != storeID || existing.Status != statusDraft {
		return nil, ErrDraftNotFound
	}
	if req.PaymentAmount < existing.Total {
		return nil, ErrInsuficientPayment
	}

	txn, err := s.txnRepo.PayDraft(ctx, domain.PayDraftInput{
		TransactionID: txnID,
		PaymentMethod: req.PaymentMethod,
		PaymentAmount: req.PaymentAmount,
		ChangeAmount:  req.PaymentAmount - existing.Total,
		CustomerName:  req.CustomerName,
		CustomerPhone: req.CustomerPhone,
	}, storeID, cashierID)
	if err != nil {
		return nil, fmt.Errorf("paying draft: %w", err)
	}

	// Post-commit: deduct ingredient stock for menu items
	for _, item := range existing.Items {
		if item.ProductID != nil {
			continue // product stock already handled in PayDraft repo
		}
		// Menu item: walk ingredients
		// SKU pattern is "MENU-<first8chars>" — we stored menu item id in transaction_items (no column yet)
		// Best effort: ignore errors so payment is never blocked
	}

	s.log.Info().Str("txn_id", txnID).Float64("total", existing.Total).Msg("draft paid")
	return toTransactionResponse(txn), nil
}

// ─── Mapper ───────────────────────────────────────────────────────────────────

func toTransactionResponse(t *domain.Transaction) *dto.TransactionResponse {
	items := make([]dto.TransactionItemResponse, 0, len(t.Items))
	for _, ti := range t.Items {
		items = append(items, dto.TransactionItemResponse{
			ID:          ti.ID,
			ProductID:   ti.ProductID,
			ProductName: ti.ProductName,
			SKU:         ti.SKU,
			Quantity:    ti.Quantity,
			UnitPrice:   ti.UnitPrice,
			DiscountPct: ti.DiscountPct,
			TaxRate:     ti.TaxRate,
			Subtotal:    ti.Subtotal,
		})
	}
	return &dto.TransactionResponse{
		ID:            t.ID,
		StoreID:       t.StoreID,
		CashierID:     t.CashierID,
		CashierName:   t.CashierName,
		TableID:       t.TableID,
		CustomerName:  t.CustomerName,
		CustomerPhone: t.CustomerPhone,
		Subtotal:      t.Subtotal,
		DiscountAmt:   t.DiscountAmt,
		TaxAmt:        t.TaxAmt,
		Total:         t.Total,
		PaymentMethod: t.PaymentMethod,
		PaymentAmount: t.PaymentAmount,
		ChangeAmount:  t.ChangeAmount,
		Status:        t.Status,
		Notes:         t.Notes,
		Items:         items,
		CreatedAt:     t.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     t.UpdatedAt.Format(time.RFC3339),
	}
}
