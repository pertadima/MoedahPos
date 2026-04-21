package dto

import (
	"github.com/moedahpos/backend/internal/domain"
)

type SyncPullOutput struct {
	Categories   []*domain.Category    `json:"categories"`
	Products     []*domain.Product     `json:"products"`
	StockLevels  []*domain.StockLevel  `json:"stock_levels"`
	Transactions []*domain.Transaction `json:"transactions"`
}
