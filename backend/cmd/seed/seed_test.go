package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUniqueCategories(t *testing.T) {
	catalog := []ProductSeed{
		{Category: "A"},
		{Category: "B"},
		{Category: "A"},
		{Category: "C"},
	}
	expected := []string{"A", "B", "C"}
	assert.Equal(t, expected, uniqueCategories(catalog))
}
