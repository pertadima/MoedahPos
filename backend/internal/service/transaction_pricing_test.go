package service

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/moedahpos/backend/internal/domain"
)

func TestComputeItemPricing(t *testing.T) {
	tests := []struct {
		name          string
		originalPrice float64
		qty           float64
		taxRate       float64
		discType      string
		discValue     float64
		wantPrice     float64
		wantSubtotal  float64
	}{
		{"Percentage", 100, 2, 10, "PERCENTAGE", 10, 90, 198},
		{"Fixed", 100, 2, 10, "FIXED", 20, 80, 176},
		{"Override", 100, 2, 10, "OVERRIDE", 50, 50, 110},
		{"Negative Price Clamped", 100, 1, 0, "FIXED", 150, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			finalPrice, _, _, _, lineSubtotal := computeItemPricing(tt.originalPrice, tt.qty, tt.taxRate, tt.discType, tt.discValue)
			assert.Equal(t, tt.wantPrice, finalPrice)
			assert.Equal(t, tt.wantSubtotal, lineSubtotal)
		})
	}
}

func TestDistributeCartDiscount_Fixed(t *testing.T) {
	items := []domain.CreateTransactionItemInput{
		{UnitPrice: 100, Quantity: 1, TaxRate: 0},
		{UnitPrice: 300, Quantity: 1, TaxRate: 0},
	}
	// Total subtotal = 400. Discount = 100 FIXED.
	// Item 1: 1/4 of total -> 25 discount.
	// Item 2: 3/4 of total -> 75 discount.

	updated, totalDisc := distributeCartDiscount(items, "FIXED", 100)
	assert.Equal(t, 100.0, totalDisc)
	assert.Equal(t, 75.0, updated[0].UnitPrice)
	assert.Equal(t, 225.0, updated[1].UnitPrice)
}

func TestDistributeCartDiscount_Percentage(t *testing.T) {
	items := []domain.CreateTransactionItemInput{
		{UnitPrice: 100, Quantity: 1, TaxRate: 0},
		{UnitPrice: 200, Quantity: 2, TaxRate: 0},
	}
	// Total subtotal = 100 + 400 = 500. Discount = 10 PERCENTAGE = 50.
	updated, totalDisc := distributeCartDiscount(items, "PERCENTAGE", 10)
	assert.Equal(t, 50.0, totalDisc)
	assert.Equal(t, 90.0, updated[0].UnitPrice)
	assert.Equal(t, 180.0, updated[1].UnitPrice)
}

func TestDistributeCartDiscount_Clamp(t *testing.T) {
	items := []domain.CreateTransactionItemInput{
		{UnitPrice: 100, Quantity: 1, TaxRate: 0},
	}
	updated, totalDisc := distributeCartDiscount(items, "FIXED", 500)
	assert.Equal(t, 100.0, totalDisc)
	assert.Equal(t, 0.0, updated[0].UnitPrice)
}
