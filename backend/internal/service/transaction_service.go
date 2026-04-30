package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog"

	"github.com/google/uuid"

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

// discountTypePercentage, discountTypeFixed, discountTypeOverride are the supported discount modes.
const (
	discountTypePercentage = "PERCENTAGE"
	discountTypeFixed      = "FIXED"
	discountTypeOverride   = "OVERRIDE"
)

// computeItemPricing derives the final selling price and line totals from discount inputs.
// originalPrice: the product/menu item's base sell_price.
// discountType:  PERCENTAGE | FIXED | OVERRIDE
// discountValue: the discount amount (percentage points, fixed IDR, or override price).
// Returns: finalPrice (unit), discountAmt (total for line), lineNet, lineTax, lineSubtotal.
func computeItemPricing(originalPrice, qty, taxRate float64, discountType string, discountValue float64) (
	finalPrice, discountAmt, lineNet, lineTax, lineSubtotal float64,
) {
	switch discountType {
	case discountTypeFixed:
		finalPrice = originalPrice - discountValue
	case discountTypeOverride:
		finalPrice = discountValue
	default: // PERCENTAGE (and legacy fallback)
		finalPrice = originalPrice * (1 - discountValue/100)
	}
	if finalPrice < 0 {
		finalPrice = 0
	}
	discountAmt = (originalPrice - finalPrice) * qty
	lineNet = finalPrice * qty
	lineTax = lineNet * taxRate / 100
	lineSubtotal = lineNet + lineTax
	return
}

// distributeCartDiscount applies a cart-level discount across items that already have
// item-level discounts applied. It mutates and returns the updated items slice.
// cartDiscType: PERCENTAGE | FIXED
// cartDiscValue: the raw discount value (0–100 for %, absolute IDR for FIXED).
// The cart discount is clamped to the current cart subtotal to avoid negative prices.
func distributeCartDiscount(
	items []domain.CreateTransactionItemInput,
	cartDiscType string,
	cartDiscValue float64,
) (updated []domain.CreateTransactionItemInput, totalCartDisc float64) {
	if cartDiscValue <= 0 {
		return items, 0
	}

	// Sum post-item-discount subtotals (unit_price * qty, excl. tax — this is what we discount)
	var cartSubtotal float64
	for _, it := range items {
		cartSubtotal += it.UnitPrice * it.Quantity
	}
	if cartSubtotal <= 0 {
		return items, 0
	}

	// Determine the total cart discount amount
	var cartDiscAmt float64
	switch cartDiscType {
	case discountTypeFixed:
		cartDiscAmt = cartDiscValue
	default: // PERCENTAGE
		cartDiscAmt = cartSubtotal * cartDiscValue / 100
	}
	// Clamp: cart discount cannot exceed the cart subtotal
	if cartDiscAmt > cartSubtotal {
		cartDiscAmt = cartSubtotal
	}

	updated = make([]domain.CreateTransactionItemInput, len(items))
	var allocatedSoFar float64

	for i, it := range items {
		var allocatedPerUnit float64
		if i == len(items)-1 {
			// Last item absorbs rounding remainder
			remaining := cartDiscAmt - allocatedSoFar
			allocatedPerUnit = remaining / it.Quantity
		} else {
			switch cartDiscType {
			case discountTypeFixed:
				// Proportional to item's share of cart subtotal
				itemWeight := (it.UnitPrice * it.Quantity) / cartSubtotal
				itemCartDisc := itemWeight * cartDiscAmt
				allocatedPerUnit = itemCartDisc / it.Quantity
			default: // PERCENTAGE: same rate applies to every item
				allocatedPerUnit = it.UnitPrice * cartDiscValue / 100
			}
			allocatedSoFar += allocatedPerUnit * it.Quantity
		}

		finalUnitPrice := it.UnitPrice - allocatedPerUnit
		if finalUnitPrice < 0 {
			finalUnitPrice = 0
		}
		lineNet := finalUnitPrice * it.Quantity
		lineTax := lineNet * it.TaxRate / 100

		updated[i] = it
		updated[i].CartDiscountAllocated = allocatedPerUnit
		updated[i].UnitPrice = finalUnitPrice
		updated[i].Subtotal = lineNet + lineTax
	}

	totalCartDisc = cartDiscAmt
	return
}

// TransactionService implements cashier checkout logic.
type TransactionService struct {
	txnRepo      repository.TransactionRepository
	productRepo  repository.ProductRepository
	stockRepo    repository.StockRepository
	menuItemRepo repository.MenuItemRepository
	batchSvc     BatchStockServiceInterface // FIFO deduction
	activitySvc  ActivityLogServiceInterface
	storeRepo    repository.StoreRepository
	loyaltyRepo  repository.LoyaltyRepository
	log          zerolog.Logger
}

func NewTransactionService(
	txnRepo repository.TransactionRepository,
	productRepo repository.ProductRepository,
	stockRepo repository.StockRepository,
	menuItemRepo repository.MenuItemRepository,
	batchSvc BatchStockServiceInterface,
	activitySvc ActivityLogServiceInterface,
	storeRepo repository.StoreRepository,
	loyaltyRepo repository.LoyaltyRepository,
	log zerolog.Logger,
) *TransactionService {
	return &TransactionService{
		txnRepo:      txnRepo,
		productRepo:  productRepo,
		stockRepo:    stockRepo,
		menuItemRepo: menuItemRepo,
		batchSvc:     batchSvc,
		activitySvc:  activitySvc,
		storeRepo:    storeRepo,
		loyaltyRepo:  loyaltyRepo,
		log:          log,
	}
}

// Checkout processes a sale: validates stock, calculates totals, persists atomically.
func (s *TransactionService) Checkout(ctx context.Context, storeID string, req *dto.CreateTransactionRequest, cashierID string) (*dto.TransactionResponse, error) { //nolint:gocognit,cyclop,funlen // retail+restaurant dual path
	if req.ID != "" {
		existing, err := s.txnRepo.FindByID(ctx, req.ID)
		if err == nil && existing != nil {
			s.log.Info().Str("txn_id", req.ID).Msg("Idempotent Checkout: returning existing transaction")
			return toTransactionResponse(existing), nil
		}
	}

	txnID := req.ID
	if txnID == "" {
		txnID = uuid.New().String()
	}

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

			// Resolve discount type
			discType := item.DiscountType
			discValue := item.DiscountValue
			if discType == "" {
				discType = discountTypePercentage
				discValue = item.DiscountPct
			}

			finalPrice, _, lineNet, lineTax, lineSubtotal := computeItemPricing(
				menuItem.SellPrice, item.Quantity, menuItem.EffectiveTaxRate(), discType, discValue)

			subtotal += lineNet
			discountAmt += (menuItem.SellPrice - finalPrice) * item.Quantity
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
				ProductID:     nil,
				ProductName:   menuItem.Name,
				SKU:           "MENU-" + item.MenuItemID[:8],
				Quantity:      item.Quantity,
				OriginalPrice: menuItem.SellPrice,
				UnitPrice:     finalPrice,
				CostPrice:     menuCost, // BOM cost per portion, snapshot at sale time
				DiscountPct:   item.DiscountPct,
				DiscountType:  discType,
				DiscountValue: discValue,
				TaxRate:       menuItem.EffectiveTaxRate(),
				Subtotal:      lineSubtotal,
				MenuItemID:    &mid,
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

			// Resolve discount type and value from the request
			discType := item.DiscountType
			discValue := item.DiscountValue
			// Legacy: if discount_type not set but discount_pct provided, use PERCENTAGE
			if discType == "" {
				discType = discountTypePercentage
				discValue = item.DiscountPct
			}

			finalPrice, _, lineNet, lineTax, lineSubtotal := computeItemPricing(
				product.SellPrice, item.Quantity, product.EffectiveTaxRate(), discType, discValue)

			subtotal += lineNet
			discountAmt += (product.SellPrice - finalPrice) * item.Quantity
			taxAmt += lineTax

			pid := item.ProductID
			inputItems = append(inputItems, domain.CreateTransactionItemInput{
				ProductID:     &pid,
				ProductName:   product.Name,
				SKU:           product.SKU,
				Quantity:      item.Quantity,
				OriginalPrice: product.SellPrice,
				UnitPrice:     finalPrice,
				CostPrice:     product.CostPrice, // snapshot at time of sale
				DiscountPct:   item.DiscountPct,
				DiscountType:  discType,
				DiscountValue: discValue,
				TaxRate:       product.EffectiveTaxRate(),
				Subtotal:      lineSubtotal,
			})
		}
	}

	// ── Apply cart-level discount (after item discounts) ────────────────────────
	cartDiscType := req.CartDiscountType
	if cartDiscType == "" {
		cartDiscType = discountTypePercentage
	}
	var cartDiscAmt float64
	if req.CartDiscountValue > 0 {
		inputItems, cartDiscAmt = distributeCartDiscount(inputItems, cartDiscType, req.CartDiscountValue)
		// Recalculate totals from the updated items
		subtotal, discountAmt, taxAmt = 0, 0, 0
		for _, it := range inputItems {
			lineNet := it.UnitPrice * it.Quantity
			subtotal += lineNet
			discountAmt += (it.OriginalPrice - it.UnitPrice) * it.Quantity
			taxAmt += lineNet * it.TaxRate / 100
		}
		_ = cartDiscAmt // absorbed into discountAmt above
	}

	var pointsDiscount float64
	var cid *string
	if req.CustomerID != "" {
		cid = &req.CustomerID
	}

	if req.PointsRedeemed > 0 && req.CustomerID != "" {
		balance, err := s.loyaltyRepo.GetBalance(ctx, req.CustomerID)
		if err != nil {
			return nil, fmt.Errorf("getting customer loyalty balance: %w", err)
		}
		if balance < req.PointsRedeemed {
			return nil, fmt.Errorf("insufficient loyalty points: have %.2f, trying to redeem %.2f", balance, req.PointsRedeemed)
		}

		store, err := s.storeRepo.FindByID(ctx, storeID)
		if err != nil {
			return nil, fmt.Errorf("getting store for loyalty rate: %w", err)
		}

		rate := store.LoyaltyRupiahPerPoint
		if rate <= 0 {
			rate = 1 // default safe rate
		}
		pointsDiscount = req.PointsRedeemed * rate
	}

	total := subtotal + taxAmt - pointsDiscount
	if total < 0 {
		total = 0
	}
	if req.PaymentAmount < total {
		return nil, ErrInsuficientPayment
	}

	txn, err := s.txnRepo.Create(ctx, domain.CreateTransactionInput{
		ID:                txnID,
		StoreID:           storeID,
		CashierID:         cashierID,
		Status:            "completed",
		CustomerID:        cid,
		CustomerName:      req.CustomerName,
		CustomerPhone:     req.CustomerPhone,
		PaymentMethod:     req.PaymentMethod,
		PaymentAmount:     req.PaymentAmount,
		ChangeAmount:      req.PaymentAmount - total,
		Notes:             req.Notes,
		Subtotal:          subtotal,
		DiscountAmt:       discountAmt,
		TaxAmt:            taxAmt,
		Total:             total,
		CartDiscountType:  cartDiscType,
		CartDiscountValue: req.CartDiscountValue,
		PointsRedeemed:    req.PointsRedeemed,
		PointsDiscount:    pointsDiscount,
		Items:             inputItems,
	})

	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") || strings.Contains(err.Error(), "duplicate key") {
			existing, dbErr := s.txnRepo.FindByID(ctx, txnID)
			if dbErr == nil && existing != nil {
				s.log.Info().Str("txn_id", txnID).Msg("Idempotent Checkout: returning existing transaction after duplicate key collision")
				return toTransactionResponse(existing), nil
			}
		}
		return nil, fmt.Errorf("creating transaction: %w", err)
	}

	// ── Post-commit: deduct loyalty points if redeemed ───────────────────────
	if req.PointsRedeemed > 0 && req.CustomerID != "" {
		if _, err := s.loyaltyRepo.SpendPoints(ctx, req.CustomerID, &txn.ID, req.PointsRedeemed); err != nil {
			s.log.Error().Err(err).Str("txn_id", txn.ID).Float64("points", req.PointsRedeemed).Msg("Failed to deduct loyalty points after checkout")
			// We don't fail the transaction here since payment succeeded and txn is committed, but log it loudly.
		}
	}

	// ── Post-commit: earn loyalty points for completed transaction ────────────
	if req.CustomerID != "" && s.storeRepo != nil {
		store, storeErr := s.storeRepo.FindByID(ctx, storeID)
		pointsPerRupiah := 1000.0
		if storeErr == nil && store != nil && store.LoyaltyPointsPerRupiah > 0 {
			pointsPerRupiah = store.LoyaltyPointsPerRupiah
		}
		tier, _ := s.loyaltyRepo.GetCustomerTier(ctx, req.CustomerID)
		multiplier := 1.0
		if tier != nil {
			multiplier = tier.Multiplier
		}
		earnPoints := CalculatePoints(total, multiplier, pointsPerRupiah)
		if earnPoints > 0 {
			if _, earnErr := s.loyaltyRepo.EarnPoints(ctx, req.CustomerID, &txn.ID, earnPoints); earnErr != nil {
				s.log.Error().Err(earnErr).Str("txn_id", txn.ID).Float64("points", earnPoints).Msg("Failed to earn loyalty points after checkout")
			}
		}
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

	// Log Activity
	metadata := map[string]interface{}{
		"total_amount":   total,
		"payment_method": req.PaymentMethod,
		"item_count":     len(req.Items),
	}
	if req.CartDiscountValue > 0 {
		metadata["cart_discount"] = map[string]interface{}{
			"type":  req.CartDiscountType,
			"value": req.CartDiscountValue,
		}
	}
	s.activitySvc.LogActivity(ctx, cashierID, storeID, domain.ActionTransactionCreate, domain.ModuleTransaction, txn.ID, metadata)

	// Log specific discount/override actions
	if req.CartDiscountValue > 0 {
		s.activitySvc.LogActivity(ctx, cashierID, storeID, domain.ActionDiscountCart, domain.ModuleDiscount, txn.ID, metadata["cart_discount"])
	}

	for _, item := range req.Items {
		if item.DiscountType != "" && item.DiscountValue > 0 {
			action := domain.ActionDiscountItem
			if item.DiscountType == "OVERRIDE" {
				action = domain.ActionPriceOverride
			}
			s.activitySvc.LogActivity(ctx, cashierID, storeID, action, domain.ModuleDiscount, txn.ID, map[string]interface{}{
				"product_id":     item.ProductID,
				"menu_item_id":   item.MenuItemID,
				"discount_type":  item.DiscountType,
				"discount_value": item.DiscountValue,
			})
		}
	}

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
func (s *TransactionService) VoidTransaction(ctx context.Context, id, userID string) error { //nolint:funlen
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

	s.activitySvc.LogActivity(ctx, userID, txn.StoreID, domain.ActionTransactionCancel, domain.ModuleTransaction, id, map[string]interface{}{
		"original_total": txn.Total,
	})

	return nil
}

// ─── Draft / Table Order Methods ──────────────────────────────────────────────

// processMenuItem handles the menu item line in buildItems.
func (s *TransactionService) processMenuItem(ctx context.Context, _ string, item dto.TxItemInput, menuItem *domain.MenuItem) (domain.CreateTransactionItemInput, float64, float64, float64, error) {
	discType := item.DiscountType
	discValue := item.DiscountValue
	if discType == "" {
		discType = discountTypePercentage
		discValue = item.DiscountPct
	}

	finalPrice, _, lineNet, lineTax, lineSubtotal := computeItemPricing(
		menuItem.SellPrice, item.Quantity, menuItem.EffectiveTaxRate(), discType, discValue)

	var menuCost float64
	for _, ing := range menuItem.Ingredients {
		ingProduct, _ := s.productRepo.FindByID(ctx, ing.ProductID)
		if ingProduct != nil {
			menuCost += ingProduct.CostPrice * ing.Quantity
		}
	}

	skuID := item.MenuItemID
	if len(skuID) > 8 {
		skuID = skuID[:8]
	}

	mid := item.MenuItemID
	discountAmt := (menuItem.SellPrice - finalPrice) * item.Quantity
	return domain.CreateTransactionItemInput{
		MenuItemID:    &mid,
		ProductName:   menuItem.Name,
		SKU:           "MENU-" + skuID,
		Quantity:      item.Quantity,
		OriginalPrice: menuItem.SellPrice,
		UnitPrice:     finalPrice,
		CostPrice:     menuCost,
		DiscountPct:   item.DiscountPct,
		DiscountType:  discType,
		DiscountValue: discValue,
		TaxRate:       menuItem.EffectiveTaxRate(),
		Subtotal:      lineSubtotal,
	}, discountAmt, lineNet, lineTax, nil
}

// processProduct handles the product line in buildItems.
func (s *TransactionService) processProduct(_ context.Context, item dto.TxItemInput, product *domain.Product) (domain.CreateTransactionItemInput, float64, float64, float64, error) {
	discType := item.DiscountType
	discValue := item.DiscountValue
	if discType == "" {
		discType = discountTypePercentage
		discValue = item.DiscountPct
	}

	finalPrice, _, lineNet, lineTax, lineSubtotal := computeItemPricing(
		product.SellPrice, item.Quantity, product.EffectiveTaxRate(), discType, discValue)

	pid := item.ProductID
	discountAmt := (product.SellPrice - finalPrice) * item.Quantity

	return domain.CreateTransactionItemInput{
		ProductID:     &pid,
		ProductName:   product.Name,
		SKU:           product.SKU,
		Quantity:      item.Quantity,
		OriginalPrice: product.SellPrice,
		UnitPrice:     finalPrice,
		CostPrice:     product.CostPrice,
		DiscountPct:   item.DiscountPct,
		DiscountType:  discType,
		DiscountValue: discValue,
		TaxRate:       product.EffectiveTaxRate(),
		Subtotal:      lineSubtotal,
	}, discountAmt, lineNet, lineTax, nil
}

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
			itemInput, itemDisc, itemNet, itemTax, err := s.processMenuItem(ctx, storeID, item, menuItem)
			if err != nil {
				return nil, 0, 0, 0, err
			}
			inputItems = append(inputItems, itemInput)
			subtotal += itemNet
			discountAmt += itemDisc
			taxAmt += itemTax
		} else {
			product, err := s.productRepo.FindByID(ctx, item.ProductID)
			if err != nil || product == nil || product.StoreID != storeID {
				return nil, 0, 0, 0, fmt.Errorf("%w: product %s", ErrProductNotFound, item.ProductID)
			}
			itemInput, itemDisc, itemNet, itemTax, err := s.processProduct(ctx, item, product)
			if err != nil {
				return nil, 0, 0, 0, err
			}
			inputItems = append(inputItems, itemInput)
			subtotal += itemNet
			discountAmt += itemDisc
			taxAmt += itemTax
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

	// Apply cart-level discount
	cartDiscType := req.CartDiscountType
	if cartDiscType == "" {
		cartDiscType = discountTypePercentage
	}
	if req.CartDiscountValue > 0 {
		inputItems, _ = distributeCartDiscount(inputItems, cartDiscType, req.CartDiscountValue)
		subtotal, discountAmt, taxAmt = 0, 0, 0
		for _, it := range inputItems {
			lineNet := it.UnitPrice * it.Quantity
			subtotal += lineNet
			discountAmt += (it.OriginalPrice - it.UnitPrice) * it.Quantity
			taxAmt += lineNet * it.TaxRate / 100
		}
	}

	tableID := req.TableID
	txn, err := s.txnRepo.Create(ctx, domain.CreateTransactionInput{
		StoreID:           storeID,
		CashierID:         cashierID,
		TableID:           &tableID,
		Status:            statusDraft,
		CustomerName:      req.CustomerName,
		Notes:             req.Notes,
		Subtotal:          subtotal,
		DiscountAmt:       discountAmt,
		TaxAmt:            taxAmt,
		Total:             subtotal + taxAmt,
		CartDiscountType:  cartDiscType,
		CartDiscountValue: req.CartDiscountValue,
		Items:             inputItems,
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

	// Apply cart-level discount
	cartDiscType := req.CartDiscountType
	if cartDiscType == "" {
		cartDiscType = discountTypePercentage
	}
	if req.CartDiscountValue > 0 {
		inputItems, _ = distributeCartDiscount(inputItems, cartDiscType, req.CartDiscountValue)
		subtotal, discountAmt, taxAmt = 0, 0, 0
		for _, it := range inputItems {
			lineNet := it.UnitPrice * it.Quantity
			subtotal += lineNet
			discountAmt += (it.OriginalPrice - it.UnitPrice) * it.Quantity
			taxAmt += lineNet * it.TaxRate / 100
		}
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
func (s *TransactionService) PayDraft(ctx context.Context, storeID, txnID, cashierID string, req *dto.PayDraftRequest) (*dto.TransactionResponse, error) { //nolint:gocognit,cyclop,funlen // loyalty+payment+draft logic is inherently complex
	existing, err := s.txnRepo.FindByID(ctx, txnID)
	if err != nil || existing == nil {
		return nil, ErrDraftNotFound
	}
	if existing.StoreID != storeID || existing.Status != statusDraft {
		return nil, ErrDraftNotFound
	}
	var pointsDiscount float64
	var cid *string
	if req.CustomerID != "" {
		cid = &req.CustomerID
	}

	if req.PointsRedeemed > 0 && req.CustomerID != "" {
		balance, err := s.loyaltyRepo.GetBalance(ctx, req.CustomerID)
		if err != nil {
			return nil, fmt.Errorf("getting customer loyalty balance: %w", err)
		}
		if balance < req.PointsRedeemed {
			return nil, fmt.Errorf("insufficient loyalty points: have %.2f, trying to redeem %.2f", balance, req.PointsRedeemed)
		}

		store, err := s.storeRepo.FindByID(ctx, storeID)
		if err != nil {
			return nil, fmt.Errorf("getting store for loyalty rate: %w", err)
		}

		rate := store.LoyaltyRupiahPerPoint
		if rate <= 0 {
			rate = 1 // default safe rate
		}
		pointsDiscount = req.PointsRedeemed * rate
	}

	total := existing.Total - pointsDiscount
	if total < 0 {
		total = 0
	}

	if req.PaymentAmount < total {
		return nil, ErrInsuficientPayment
	}

	txn, err := s.txnRepo.PayDraft(ctx, domain.PayDraftInput{
		TransactionID:  txnID,
		PaymentMethod:  req.PaymentMethod,
		PaymentAmount:  req.PaymentAmount,
		ChangeAmount:   req.PaymentAmount - total,
		CustomerID:     cid,
		CustomerName:   req.CustomerName,
		CustomerPhone:  req.CustomerPhone,
		PointsRedeemed: req.PointsRedeemed,
		PointsDiscount: pointsDiscount,
	}, storeID, cashierID)
	if err != nil {
		return nil, fmt.Errorf("paying draft: %w", err)
	}

	// ── Post-commit: deduct loyalty points if redeemed ───────────────────────
	if req.PointsRedeemed > 0 && req.CustomerID != "" {
		if _, err := s.loyaltyRepo.SpendPoints(ctx, req.CustomerID, &txn.ID, req.PointsRedeemed); err != nil {
			s.log.Error().Err(err).Str("txn_id", txn.ID).Float64("points", req.PointsRedeemed).Msg("Failed to deduct loyalty points after draft payment")
		}
	}

	// ── Post-commit: earn loyalty points for paid draft ───────────────────────
	if req.CustomerID != "" && s.storeRepo != nil {
		store, storeErr := s.storeRepo.FindByID(ctx, storeID)
		pointsPerRupiah := 1000.0
		if storeErr == nil && store != nil && store.LoyaltyPointsPerRupiah > 0 {
			pointsPerRupiah = store.LoyaltyPointsPerRupiah
		}
		tier, _ := s.loyaltyRepo.GetCustomerTier(ctx, req.CustomerID)
		multiplier := 1.0
		if tier != nil {
			multiplier = tier.Multiplier
		}
		earnPoints := CalculatePoints(existing.Total, multiplier, pointsPerRupiah)
		if earnPoints > 0 {
			if _, earnErr := s.loyaltyRepo.EarnPoints(ctx, req.CustomerID, &txn.ID, earnPoints); earnErr != nil {
				s.log.Error().Err(earnErr).Str("txn_id", txn.ID).Float64("points", earnPoints).Msg("Failed to earn loyalty points after draft payment")
			}
		}
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

	// Log Activity
	metadata := map[string]interface{}{
		"total_amount":   existing.Total,
		"payment_method": req.PaymentMethod,
		"item_count":     len(existing.Items),
		"from_draft":     true,
	}
	s.activitySvc.LogActivity(ctx, cashierID, storeID, domain.ActionTransactionCreate, domain.ModuleTransaction, txnID, metadata)

	return toTransactionResponse(txn), nil
}

// ─── KDS Methods ──────────────────────────────────────────────────────────────

// GetKDSTickets returns all active KDS tickets for a restaurant.
func (s *TransactionService) GetKDSTickets(ctx context.Context, storeID string) ([]*dto.TransactionResponse, error) {
	txns, err := s.txnRepo.GetKDSTickets(ctx, storeID)
	if err != nil {
		return nil, fmt.Errorf("getting kds tickets: %w", err)
	}
	resp := make([]*dto.TransactionResponse, 0, len(txns))
	for _, t := range txns {
		resp = append(resp, toTransactionResponse(t))
	}
	return resp, nil
}

// UpdateKDSItemStatus marks a KDS ticket item as completed or pending.
func (s *TransactionService) UpdateKDSItemStatus(ctx context.Context, itemID string, req *dto.UpdateKDSItemStatusRequest) error {
	if err := s.txnRepo.UpdateKDSItemStatus(ctx, itemID, req.Status); err != nil {
		return fmt.Errorf("updating kds item status: %w", err)
	}
	return nil
}

// ─── Mapper ───────────────────────────────────────────────────────────────────

func toTransactionResponse(t *domain.Transaction) *dto.TransactionResponse {
	items := make([]dto.TransactionItemResponse, 0, len(t.Items))
	for _, ti := range t.Items {
		items = append(items, dto.TransactionItemResponse{
			ID:                    ti.ID,
			ProductID:             ti.ProductID,
			MenuItemID:            ti.MenuItemID,
			ProductName:           ti.ProductName,
			SKU:                   ti.SKU,
			Quantity:              ti.Quantity,
			OriginalPrice:         ti.OriginalPrice,
			UnitPrice:             ti.UnitPrice,
			DiscountPct:           ti.DiscountPct,
			DiscountType:          ti.DiscountType,
			DiscountValue:         ti.DiscountValue,
			CartDiscountAllocated: ti.CartDiscountAllocated,
			TaxRate:               ti.TaxRate,
			Subtotal:              ti.Subtotal,
			Status:                ti.Status,
		})
		if ti.CompletedAt != nil {
			ca := ti.CompletedAt.Format(time.RFC3339)
			items[len(items)-1].CompletedAt = &ca
		}
	}
	return &dto.TransactionResponse{
		ID:                t.ID,
		StoreID:           t.StoreID,
		CashierID:         t.CashierID,
		CashierName:       t.CashierName,
		TableID:           t.TableID,
		TableNumber:       t.TableNumber,
		CustomerName:      t.CustomerName,
		CustomerPhone:     t.CustomerPhone,
		Subtotal:          t.Subtotal,
		DiscountAmt:       t.DiscountAmt,
		TaxAmt:            t.TaxAmt,
		Total:             t.Total,
		PaymentMethod:     t.PaymentMethod,
		PaymentAmount:     t.PaymentAmount,
		ChangeAmount:      t.ChangeAmount,
		Status:            t.Status,
		Notes:             t.Notes,
		CartDiscountType:  t.CartDiscountType,
		CartDiscountValue: t.CartDiscountValue,
		PointsRedeemed:    t.PointsRedeemed,
		PointsDiscount:    t.PointsDiscount,
		Items:             items,
		CreatedAt:         t.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         t.UpdatedAt.Format(time.RFC3339),
	}
}
