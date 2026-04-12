package domain

import (
	"encoding/json"
	"time"
)

type ActionType string

const (
	ActionAuthLogin         ActionType = "AUTH_LOGIN"
	ActionAuthLogout        ActionType = "AUTH_LOGOUT"
	ActionTransactionCreate ActionType = "TRANSACTION_CREATE"
	ActionTransactionCancel ActionType = "TRANSACTION_CANCEL"
	ActionDiscountItem      ActionType = "DISCOUNT_ITEM"
	ActionDiscountCart      ActionType = "DISCOUNT_CART"
	ActionPriceOverride     ActionType = "PRICE_OVERRIDE"
	ActionStockAdjustment   ActionType = "STOCK_ADJUSTMENT"

	// Financial Activities
	ActionPurchaseOrderCreate  ActionType = "PURCHASE_ORDER_CREATE"
	ActionPurchaseOrderUpdate  ActionType = "PURCHASE_ORDER_UPDATE"
	ActionPurchaseOrderPayment ActionType = "PURCHASE_ORDER_PAYMENT"

	ActionIncomeCreate ActionType = "INCOME_CREATE"
	ActionIncomeUpdate ActionType = "INCOME_UPDATE"
	ActionIncomeDelete ActionType = "INCOME_DELETE"

	ActionExpenseCreate ActionType = "EXPENSE_CREATE"
	ActionExpenseUpdate ActionType = "EXPENSE_UPDATE"
	ActionExpenseDelete ActionType = "EXPENSE_DELETE"
)

type ActivityModule string

const (
	ModuleAuth        ActivityModule = "AUTH"
	ModuleTransaction ActivityModule = "TRANSACTION"
	ModuleDiscount    ActivityModule = "DISCOUNT"
	ModuleInventory   ActivityModule = "INVENTORY"
	ModulePurchase    ActivityModule = "PURCHASE"
	ModuleIncome      ActivityModule = "INCOME"
	ModuleExpense     ActivityModule = "EXPENSE"
)

type ActivityLog struct {
	ID          string          `db:"id" json:"id"`
	UserID      string          `db:"user_id" json:"user_id"`
	UserName    string          `db:"user_name" json:"user_name,omitempty"`
	StoreID     *string         `db:"store_id" json:"store_id"`
	ActionType  ActionType      `db:"action_type" json:"action_type"`
	Module      ActivityModule  `db:"module" json:"module"`
	ReferenceID *string         `db:"reference_id" json:"reference_id"`
	Metadata    json.RawMessage `db:"metadata" json:"metadata"`
	CreatedAt   time.Time       `db:"created_at" json:"created_at"`
}
