package domain

import "time"

// TableStatus represents the occupancy state of a restaurant table.
type TableStatus string

const (
	TableAvailable TableStatus = "available"
	TableOccupied  TableStatus = "occupied"
	TableReserved  TableStatus = "reserved"
)

// RestaurantTable is a physical table in a restaurant store.
type RestaurantTable struct {
	ID          string      `db:"id"`
	StoreID     string      `db:"store_id"`
	TableNumber string      `db:"table_number"`
	Capacity    int         `db:"capacity"`
	Status      TableStatus `db:"status"`
	Notes       *string     `db:"notes"`
	IsActive    bool        `db:"is_active"`
	CreatedAt   time.Time   `db:"created_at"`
	UpdatedAt   time.Time   `db:"updated_at"`
	DeletedAt   *time.Time  `db:"deleted_at"`
}

// MenuItem is a composed dish (recipe) sold as one unit in a restaurant.
// Its stock impact is defined by MenuItemIngredients.
type MenuItem struct {
	ID           string     `db:"id"`
	StoreID      string     `db:"store_id"`
	CategoryID   *string    `db:"category_id"`
	Name         string     `db:"name"`
	Description  *string    `db:"description"`
	SellPrice    float64    `db:"sell_price"`
	TaxRate      float64    `db:"tax_rate"`
	ImageURL     *string    `db:"image_url"`
	IsActive     bool       `db:"is_active"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`
	DeletedAt    *time.Time `db:"deleted_at"`

	// Populated via JOIN
	CategoryName *string              `db:"category_name"`
	Ingredients  []MenuItemIngredient `db:"-"`
}

// MenuItemIngredient is one ingredient in a menu item's recipe.
type MenuItemIngredient struct {
	ID          string  `db:"id"`
	MenuItemID  string  `db:"menu_item_id"`
	ProductID   string  `db:"product_id"`
	Quantity    float64 `db:"quantity"`

	// Populated via JOIN
	ProductName string  `db:"product_name"`
	ProductSKU  string  `db:"product_sku"`
	Unit        string  `db:"unit"`
}
