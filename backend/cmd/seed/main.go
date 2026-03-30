// cmd/seed/main.go — Demo data seeder for MoedahPOS
// Usage: go run ./cmd/seed/main.go
//        go run ./cmd/seed/main.go --reset   (drops and re-seeds everything)
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/moedahpos/backend/internal/config"
	"golang.org/x/crypto/bcrypt"
)

// ── Helpers ──────────────────────────────────────────────────────────────────

func must(err error) {
	if err != nil {
		log.Fatalf("seed error: %v", err)
	}
}

func hashpw(plain string) string {
	h, err := bcrypt.GenerateFromPassword([]byte(plain), 10)
	must(err)
	return string(h)
}

func ptr[T any](v T) *T { return &v }

// ── Main ──────────────────────────────────────────────────────────────────────

func main() {
	reset := flag.Bool("reset", false, "Truncate all demo tables before seeding")
	flag.Parse()

	// Load env + config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	dsn := cfg.DB.DSN()
	db, err := sqlx.Connect("postgres", dsn)
	must(err)
	defer db.Close()

	ctx := context.Background()

	if *reset {
		log.Println("🗑  Resetting demo data...")
		resetData(ctx, db)
	}

	log.Println("🌱 Seeding MoedahPOS demo data...")

	// ── Roles (read from DB — already seeded by goose migrations) ────────────
	roles := map[string]string{}
	rows, err := db.QueryxContext(ctx, `SELECT name, id FROM roles`)
	must(err)
	for rows.Next() {
		var name, id string
		must(rows.Scan(&name, &id))
		roles[name] = id
	}
	rows.Close()
	if len(roles) == 0 {
		log.Fatal("No roles found — make sure goose migrations have run first")
	}
	log.Printf("   ✓ Loaded %d roles", len(roles))

	// ── Users ─────────────────────────────────────────────────────────────────
	users := []struct {
		ID       string
		Name     string
		Email    string
		Password string
	}{
		{uuid.NewString(), "Admin Sistem", "admin@moedah.com", "Admin1234!"},
		{uuid.NewString(), "Budi Manager", "manager@moedah.com", "Manager1234!"},
		{uuid.NewString(), "Sari Kasir", "kasir@moedah.com", "Kasir1234!"},
		{uuid.NewString(), "Andi Staff", "staff@moedah.com", "Staff1234!"},
	}

	for _, u := range users {
		_, err = db.ExecContext(ctx, `
			INSERT INTO users (id, name, email, password_hash, is_active)
			VALUES ($1, $2, $3, $4, true)
			ON CONFLICT (email) DO UPDATE SET
				name          = EXCLUDED.name,
				password_hash = EXCLUDED.password_hash,
				is_active     = true,
				deleted_at    = NULL
		`, u.ID, u.Name, u.Email, hashpw(u.Password))
		must(err)
		log.Printf("   ✓ User: %s (%s)", u.Name, u.Email)
	}

	// Re-read actual IDs (upsert may keep old IDs)
	userIDs := map[string]string{}
	rows, err = db.QueryxContext(ctx, `SELECT email, id FROM users WHERE deleted_at IS NULL`)
	must(err)
	for rows.Next() {
		var email, id string
		must(rows.Scan(&email, &id))
		userIDs[email] = id
	}
	rows.Close()

	// ── Stores ────────────────────────────────────────────────────────────────
	type Store struct {
		ID      string
		Name    string
		Address string
		Phone   string
		TaxNum  string
	}
	stores := []Store{
		{uuid.NewString(), "Toko Utama — Jakarta", "Jl. Sudirman No. 12, Jakarta Pusat", "021-5551234", "02.123.456.7-001.000"},
		{uuid.NewString(), "Cabang Bandung", "Jl. Dago No. 88, Bandung", "022-7778899", "02.123.456.7-002.000"},
	}

	for _, s := range stores {
		_, err = db.ExecContext(ctx, `
			INSERT INTO stores (id, name, address, phone, tax_number, currency, is_active)
			VALUES ($1, $2, $3, $4, $5, 'IDR', true)
			ON CONFLICT DO NOTHING
		`, s.ID, s.Name, s.Address, s.Phone, s.TaxNum)
		must(err)
		log.Printf("   ✓ Store: %s", s.Name)
	}

	mainStoreID := stores[0].ID
	branchStoreID := stores[1].ID

	// ── User ↔ Store Memberships ───────────────────────────────────────────────
	memberships := []struct{ UserEmail, StoreID, RoleName string }{
		{"admin@moedah.com", mainStoreID, "superadmin"},
		{"admin@moedah.com", branchStoreID, "superadmin"},
		{"manager@moedah.com", mainStoreID, "manager"},
		{"kasir@moedah.com", mainStoreID, "cashier"},
		{"staff@moedah.com", mainStoreID, "staff"},
	}
	for _, m := range memberships {
		_, err = db.ExecContext(ctx, `
			INSERT INTO user_stores (user_id, store_id, role_id, is_active)
			VALUES ($1, $2, $3, true)
			ON CONFLICT (user_id, store_id) DO UPDATE SET role_id = EXCLUDED.role_id, is_active = true
		`, userIDs[m.UserEmail], m.StoreID, roles[m.RoleName])
		must(err)
	}
	log.Printf("   ✓ Assigned %d user-store memberships", len(memberships))

	// ── Suppliers ─────────────────────────────────────────────────────────────
	supplierIDs := map[string]string{}
	suppliersData := []struct {
		Name, Contact, Phone, Email, Address string
	}{
		{"PT Sumber Makmur", "Bapak Hendra", "021-8887766", "hendra@sumbermakmur.co.id", "Jl. Industri No. 5, Tangerang"},
		{"CV Mitra Jaya Sejahtera", "Ibu Dewi", "022-3334455", "dewi@mitrajaya.com", "Jl. Raya Cimahi No. 30, Bandung"},
		{"UD Berkah Abadi", "Bapak Rudi", "031-6667788", "rudi@berkah-abadi.com", "Jl. Pahlawan No. 9, Surabaya"},
	}
	for _, s := range suppliersData {
		id := uuid.NewString()
		_, err = db.ExecContext(ctx, `
			INSERT INTO suppliers (id, name, contact_name, phone, email, address, is_active)
			VALUES ($1, $2, $3, $4, $5, $6, true)
			ON CONFLICT DO NOTHING
		`, id, s.Name, s.Contact, s.Phone, s.Email, s.Address)
		must(err)
		supplierIDs[s.Name] = id
	}
	log.Printf("   ✓ Seeded %d suppliers", len(suppliersData))

	// ── Categories ────────────────────────────────────────────────────────────
	type Category struct{ ID, Name string }
	seedCategories := func(storeID string, names []string) map[string]string {
		catIDs := map[string]string{}
		for _, n := range names {
			id := uuid.NewString()
			_, err = db.ExecContext(ctx, `
				INSERT INTO categories (id, store_id, name)
				VALUES ($1, $2, $3) ON CONFLICT DO NOTHING
			`, id, storeID, n)
			must(err)
			catIDs[n] = id
		}
		return catIDs
	}

	mainCats := seedCategories(mainStoreID, []string{"Minuman", "Makanan", "Snack & Cemilan", "Rokok & Tembakau"})
	seedCategories(branchStoreID, []string{"Minuman", "Makanan", "Snack & Cemilan"})
	log.Println("   ✓ Seeded categories for 2 stores")

	// ── Products (Toko Utama) ─────────────────────────────────────────────────
	type Product struct {
		Name, SKU, Unit, CategoryName string
		CostPrice, SellPrice, TaxRate float64
		InitQty                       float64
	}

	products := []Product{
		// Minuman
		{"Kopi Americano", "BEV-001", "cup", "Minuman", 8000, 22000, 11, 85},
		{"Kopi Latte", "BEV-002", "cup", "Minuman", 10000, 28000, 11, 72},
		{"Kopi Cappuccino", "BEV-003", "cup", "Minuman", 10000, 28000, 11, 60},
		{"Teh Hijau", "BEV-004", "cup", "Minuman", 4000, 15000, 11, 95},
		{"Jus Jeruk", "BEV-005", "cup", "Minuman", 6000, 18000, 11, 50},
		{"Air Mineral 600ml", "BEV-006", "botol", "Minuman", 2500, 6000, 0, 200},
		{"Susu Segar", "BEV-007", "cup", "Minuman", 7000, 18000, 11, 45},
		{"Es Teh Manis", "BEV-008", "cup", "Minuman", 3000, 12000, 0, 120},
		// Makanan
		{"Nasi Goreng Spesial", "FOOD-001", "porsi", "Makanan", 12000, 30000, 11, 30},
		{"Mie Goreng", "FOOD-002", "porsi", "Makanan", 10000, 25000, 11, 30},
		{"Roti Bakar Coklat", "FOOD-003", "pcs", "Makanan", 6000, 15000, 0, 40},
		{"Sandwich Ayam", "FOOD-004", "pcs", "Makanan", 12000, 28000, 11, 20},
		{"Pisang Goreng (5 pcs)", "FOOD-005", "pcs", "Makanan", 5000, 13000, 0, 25},
		// Snack
		{"Keripik Kentang", "SNK-001", "pcs", "Snack & Cemilan", 5000, 12000, 0, 80},
		{"Coklat Wafer", "SNK-002", "pcs", "Snack & Cemilan", 4000, 10000, 0, 100},
		{"Kacang Goreng", "SNK-003", "pcs", "Snack & Cemilan", 3500, 8000, 0, 60},
		{"Donat Gula", "SNK-004", "pcs", "Snack & Cemilan", 4500, 10000, 0, 35},
		// Rokok
		{"Rokok Surya 16", "ROK-001", "bungkus", "Rokok & Tembakau", 18000, 24000, 0, 3},
		{"Rokok Djarum Super", "ROK-002", "bungkus", "Rokok & Tembakau", 20000, 26000, 0, 5},
	}

	for _, p := range products {
		prodID := uuid.NewString()
		catID := mainCats[p.CategoryName]

		_, err = db.ExecContext(ctx, `
			INSERT INTO products (id, store_id, category_id, sku, name, unit, cost_price, sell_price, tax_rate, is_active)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, true)
			ON CONFLICT (store_id, sku) DO UPDATE SET
				name       = EXCLUDED.name,
				cost_price = EXCLUDED.cost_price,
				sell_price = EXCLUDED.sell_price,
				deleted_at = NULL
		`, prodID, mainStoreID, catID, p.SKU, p.Name, p.Unit, p.CostPrice, p.SellPrice, p.TaxRate)
		must(err)

		// Get actual product ID (may have existed)
		var actualID string
		must(db.QueryRowContext(ctx, `SELECT id FROM products WHERE store_id=$1 AND sku=$2`, mainStoreID, p.SKU).Scan(&actualID))

		_, err = db.ExecContext(ctx, `
			INSERT INTO stock_levels (product_id, store_id, quantity, min_quantity)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (product_id, store_id) DO UPDATE SET quantity = EXCLUDED.quantity
		`, actualID, mainStoreID, p.InitQty, 5)
		must(err)
	}
	log.Printf("   ✓ Seeded %d products with stock for Toko Utama", len(products))

	// ── Purchase Orders (sample, draft) ──────────────────────────────────────
	adminID := userIDs["admin@moedah.com"]
	poID := uuid.NewString()
	poNum := fmt.Sprintf("PO-%s", time.Now().Format("20060102-0001"))

	_, err = db.ExecContext(ctx, `
		INSERT INTO purchase_orders (id, store_id, supplier_id, po_number, status, total_amount, ordered_by, notes)
		VALUES ($1, $2, $3, $4, 'draft', 0, $5, 'Pembelian rutin mingguan')
		ON CONFLICT DO NOTHING
	`, poID, mainStoreID, supplierIDs["PT Sumber Makmur"], poNum, adminID)
	must(err)
	log.Printf("   ✓ Sample purchase order: %s", poNum)

	// ── Sample Transactions ───────────────────────────────────────────────────
	kasirID := userIDs["kasir@moedah.com"]
	seedTransactions(ctx, db, mainStoreID, kasirID)

	log.Println("")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("✅  Seed completed successfully!")
	log.Println("")
	log.Println("   Demo Credentials:")
	log.Println("   ┌─────────────────────────────────────────┐")
	log.Println("   │ Role     │ Email                │ Pass   │")
	log.Println("   ├─────────────────────────────────────────┤")
	log.Println("   │ Admin    │ admin@moedah.com     │ Admin1234!   │")
	log.Println("   │ Manager  │ manager@moedah.com   │ Manager1234! │")
	log.Println("   │ Kasir    │ kasir@moedah.com     │ Kasir1234!   │")
	log.Println("   │ Staff    │ staff@moedah.com     │ Staff1234!   │")
	log.Println("   └─────────────────────────────────────────┘")
	log.Println("   Frontend: http://localhost:3000")
	log.Println("   Backend:  http://localhost:8080")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

// seedTransactions inserts 5 sample completed transactions
func seedTransactions(ctx context.Context, db *sqlx.DB, storeID, cashierID string) {
	type txData struct {
		CustomerName  string
		ProductSKU    string
		Qty           float64
		PaymentMethod string
	}
	samples := []txData{
		{"Budi Santoso", "BEV-001", 2, "cash"},
		{"Pelanggan Umum", "FOOD-001", 1, "qris"},
		{"Ani Rahayu", "BEV-002", 1, "cash"},
		{"Pelanggan Umum", "SNK-001", 3, "cash"},
		{"Dewi Lestari", "BEV-008", 2, "card"},
	}

	for i, s := range samples {
		// Fetch product details
		var prodID string
		var sellPrice, taxRate float64
		var prodName, sku string
		err := db.QueryRowContext(ctx, `
			SELECT id, name, sku, sell_price, tax_rate FROM products WHERE store_id=$1 AND sku=$2
		`, storeID, s.ProductSKU).Scan(&prodID, &prodName, &sku, &sellPrice, &taxRate)
		if err != nil {
			continue
		}

		lineSubtotal := sellPrice * s.Qty
		lineTax := lineSubtotal * (taxRate / 100)
		total := lineSubtotal + lineTax
		payAmt := total + float64((i+1)*5000) // add some change for cash txns

		txID := uuid.NewString()

		_, err = db.ExecContext(ctx, `
			INSERT INTO transactions
			  (id, store_id, cashier_id, customer_name, subtotal, discount_amt, tax_amt, total,
			   payment_method, payment_amount, change_amount, status, created_at)
			VALUES ($1, $2, $3, $4, $5, 0, $6, $7, $8, $9, $10, 'completed',
			        NOW() - ($11 * interval '1 hour'))
			ON CONFLICT DO NOTHING
		`, txID, storeID, cashierID, s.CustomerName,
			lineSubtotal, lineTax, total,
			s.PaymentMethod, payAmt, payAmt-total,
			float64(i*3))
		if err != nil {
			continue
		}

		_, err = db.ExecContext(ctx, `
			INSERT INTO transaction_items
			  (id, transaction_id, product_id, product_name, sku, quantity, unit_price, discount_pct, tax_rate, subtotal)
			VALUES ($1, $2, $3, $4, $5, $6, $7, 0, $8, $9)
		`, uuid.NewString(), txID, prodID, prodName, sku, s.Qty, sellPrice, taxRate, lineSubtotal)
		if err != nil {
			continue
		}

		// Deduct stock
		_, _ = db.ExecContext(ctx, `
			UPDATE stock_levels SET quantity = GREATEST(0, quantity - $1), updated_at = NOW()
			WHERE product_id = $2 AND store_id = $3
		`, s.Qty, prodID, storeID)

		_, _ = db.ExecContext(ctx, `
			INSERT INTO stock_movements (id, product_id, store_id, ref_type, ref_id, quantity_delta, notes, created_by)
			VALUES ($1, $2, $3, 'sale', $4, $5, 'Penjualan kasir', $6)
		`, uuid.NewString(), prodID, storeID, txID, -s.Qty, cashierID)
	}
	log.Println("   ✓ Seeded 5 sample transactions with stock movements")
}

// resetData truncates all data tables (keeps roles/permissions/migrations)
func resetData(ctx context.Context, db *sqlx.DB) {
	tables := []string{
		"transaction_items", "transactions",
		"stock_movements", "stock_levels",
		"purchase_order_items", "purchase_orders",
		"products", "categories",
		"user_stores", "suppliers",
		"refresh_tokens", "users", "stores",
	}
	for _, t := range tables {
		_, err := db.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s", t))
		if err != nil {
			log.Printf("   warn: could not clear %s: %v", t, err)
		}
	}
	log.Println("   ✓ All demo tables cleared")
}
