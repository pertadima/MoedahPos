// cmd/seed/main.go — Demo data seeder for MoedahPOS
// Usage: go run ./cmd/seed/main.go
//        go run ./cmd/seed/main.go --reset   (drop all data, then re-seed)
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

// ── Helpers ───────────────────────────────────────────────────────────────────

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

// ── Product catalog definition ────────────────────────────────────────────────

type ProductSeed struct {
	Name, SKU, Unit, Category string
	CostPrice, SellPrice      float64
	TaxRate                   float64
	InitQty, MinQty           float64
	Barcode                   string
}

// catalogJakarta — Cafe / coffee-shop style (Toko Utama — Jakarta)
var catalogJakarta = []ProductSeed{
	// ── Minuman ────────────────────────────────────────
	{"Kopi Americano", "BEV-001", "cup", "Minuman", 8_000, 22_000, 11, 85, 5, "8991000001"},
	{"Kopi Latte", "BEV-002", "cup", "Minuman", 10_000, 28_000, 11, 72, 5, "8991000002"},
	{"Kopi Cappuccino", "BEV-003", "cup", "Minuman", 10_000, 28_000, 11, 60, 5, "8991000003"},
	{"Kopi V60", "BEV-004", "cup", "Minuman", 12_000, 32_000, 11, 40, 5, "8991000004"},
	{"Kopi Cold Brew", "BEV-005", "cup", "Minuman", 14_000, 35_000, 11, 30, 5, "8991000005"},
	{"Teh Hijau", "BEV-006", "cup", "Minuman", 4_000, 15_000, 11, 95, 10, "8991000006"},
	{"Teh Tarik", "BEV-007", "cup", "Minuman", 5_000, 16_000, 11, 80, 10, "8991000007"},
	{"Es Teh Manis", "BEV-008", "cup", "Minuman", 3_000, 12_000, 0, 120, 10, "8991000008"},
	{"Jus Jeruk Segar", "BEV-009", "cup", "Minuman", 6_000, 18_000, 11, 50, 5, "8991000009"},
	{"Jus Alpukat", "BEV-010", "cup", "Minuman", 8_000, 22_000, 11, 40, 5, "8991000010"},
	{"Susu Segar", "BEV-011", "cup", "Minuman", 7_000, 18_000, 11, 45, 5, "8991000011"},
	{"Air Mineral 600ml", "BEV-012", "botol", "Minuman", 2_500, 6_000, 0, 200, 20, "8991000012"},
	{"Air Mineral 1500ml", "BEV-013", "botol", "Minuman", 4_000, 9_000, 0, 100, 10, "8991000013"},
	{"Cokelat Panas", "BEV-014", "cup", "Minuman", 9_000, 24_000, 11, 35, 5, "8991000014"},
	{"Matcha Latte", "BEV-015", "cup", "Minuman", 12_000, 30_000, 11, 28, 5, "8991000015"},
	// ── Makanan ────────────────────────────────────────
	{"Nasi Goreng Spesial", "FOOD-001", "porsi", "Makanan", 12_000, 30_000, 11, 30, 5, "8992000001"},
	{"Nasi Goreng Seafood", "FOOD-002", "porsi", "Makanan", 15_000, 38_000, 11, 20, 5, "8992000002"},
	{"Mie Goreng", "FOOD-003", "porsi", "Makanan", 10_000, 25_000, 11, 25, 5, "8992000003"},
	{"Kwetiau Goreng", "FOOD-004", "porsi", "Makanan", 11_000, 27_000, 11, 20, 5, "8992000004"},
	{"Roti Bakar Coklat", "FOOD-005", "pcs", "Makanan", 6_000, 15_000, 0, 40, 5, "8992000005"},
	{"Roti Bakar Keju", "FOOD-006", "pcs", "Makanan", 7_000, 17_000, 0, 35, 5, "8992000006"},
	{"Sandwich Ayam", "FOOD-007", "pcs", "Makanan", 12_000, 28_000, 11, 20, 3, "8992000007"},
	{"Croissant Butter", "FOOD-008", "pcs", "Makanan", 8_000, 20_000, 0, 25, 3, "8992000008"},
	{"Pisang Goreng (5 pcs)", "FOOD-009", "porsi", "Makanan", 5_000, 13_000, 0, 25, 5, "8992000009"},
	{"Kentang Goreng", "FOOD-010", "porsi", "Makanan", 8_000, 20_000, 11, 20, 5, "8992000010"},
	// ── Snack & Cemilan ────────────────────────────────
	{"Keripik Kentang", "SNK-001", "pcs", "Snack & Cemilan", 5_000, 12_000, 0, 80, 10, "8993000001"},
	{"Coklat Wafer", "SNK-002", "pcs", "Snack & Cemilan", 4_000, 10_000, 0, 100, 10, "8993000002"},
	{"Kacang Goreng", "SNK-003", "pcs", "Snack & Cemilan", 3_500, 8_000, 0, 60, 10, "8993000003"},
	{"Donat Gula", "SNK-004", "pcs", "Snack & Cemilan", 4_500, 10_000, 0, 35, 5, "8993000004"},
	{"Choco Chip Cookie", "SNK-005", "pcs", "Snack & Cemilan", 5_000, 12_000, 0, 50, 5, "8993000005"},
	{"Banana Cake", "SNK-006", "pcs", "Snack & Cemilan", 6_000, 14_000, 0, 30, 5, "8993000006"},
	// ── Rokok & Tembakau ───────────────────────────────
	{"Rokok Surya 16", "ROK-001", "bungkus", "Rokok & Tembakau", 18_000, 24_000, 0, 3, 5, "8994000001"},
	{"Rokok Djarum Super", "ROK-002", "bungkus", "Rokok & Tembakau", 20_000, 26_000, 0, 5, 5, "8994000002"},
	{"Rokok Gudang Garam", "ROK-003", "bungkus", "Rokok & Tembakau", 19_000, 25_000, 0, 4, 5, "8994000003"},
}

// catalogBandung — Minimart / convenience-store style (Cabang Bandung)
var catalogBandung = []ProductSeed{
	// ── Minuman Kemasan ────────────────────────────────
	{"Aqua 600ml", "BDG-BEV-001", "botol", "Minuman Kemasan", 3_000, 6_000, 0, 144, 20, "8995000001"},
	{"Aqua 1500ml", "BDG-BEV-002", "botol", "Minuman Kemasan", 5_000, 9_000, 0, 72, 10, "8995000002"},
	{"Teh Botol Sosro 450ml", "BDG-BEV-003", "botol", "Minuman Kemasan", 4_500, 8_000, 0, 96, 15, "8995000003"},
	{"Pocari Sweat 500ml", "BDG-BEV-004", "botol", "Minuman Kemasan", 7_000, 12_000, 0, 60, 10, "8995000004"},
	{"Coca-Cola 330ml", "BDG-BEV-005", "kaleng", "Minuman Kemasan", 6_500, 11_000, 0, 48, 10, "8995000005"},
	{"Sprite 330ml", "BDG-BEV-006", "kaleng", "Minuman Kemasan", 6_500, 11_000, 0, 48, 10, "8995000006"},
	{"Fanta Orange 330ml", "BDG-BEV-007", "kaleng", "Minuman Kemasan", 6_500, 11_000, 0, 36, 10, "8995000007"},
	{"Susu Ultra 250ml", "BDG-BEV-008", "kotak", "Minuman Kemasan", 5_000, 9_000, 0, 72, 10, "8995000008"},
	{"Good Day Coffee 250ml", "BDG-BEV-009", "botol", "Minuman Kemasan", 5_500, 10_000, 0, 60, 10, "8995000009"},
	{"Nescafe RTD 220ml", "BDG-BEV-010", "kaleng", "Minuman Kemasan", 6_000, 11_000, 0, 48, 10, "8995000010"},
	{"Milo 200ml", "BDG-BEV-011", "kotak", "Minuman Kemasan", 5_500, 10_000, 0, 60, 10, "8995000011"},
	{"Teh Pucuk 350ml", "BDG-BEV-012", "botol", "Minuman Kemasan", 4_000, 7_000, 0, 96, 15, "8995000012"},
	// ── Makanan Instan ─────────────────────────────────
	{"Indomie Goreng", "BDG-FOOD-001", "pcs", "Makanan Instan", 2_800, 5_000, 0, 200, 30, "8996000001"},
	{"Indomie Kuah Ayam", "BDG-FOOD-002", "pcs", "Makanan Instan", 2_800, 5_000, 0, 180, 30, "8996000002"},
	{"Mie Sedaap Goreng", "BDG-FOOD-003", "pcs", "Makanan Instan", 2_600, 4_500, 0, 120, 20, "8996000003"},
	{"Pop Mie Ayam 75g", "BDG-FOOD-004", "pcs", "Makanan Instan", 4_000, 7_000, 0, 96, 15, "8996000004"},
	{"Bihun Goreng", "BDG-FOOD-005", "pcs", "Makanan Instan", 2_500, 4_500, 0, 100, 15, "8996000005"},
	{"Sarimi Soto", "BDG-FOOD-006", "pcs", "Makanan Instan", 2_700, 4_800, 0, 80, 15, "8996000006"},
	// ── Snack & Cemilan ────────────────────────────────
	{"Chitato Sapi Panggang", "BDG-SNK-001", "pcs", "Snack & Cemilan", 8_000, 14_000, 0, 60, 10, "8997000001"},
	{"Lays Original", "BDG-SNK-002", "pcs", "Snack & Cemilan", 9_000, 15_000, 0, 48, 10, "8997000002"},
	{"Taro Net", "BDG-SNK-003", "pcs", "Snack & Cemilan", 4_000, 7_000, 0, 80, 10, "8997000003"},
	{"Qtela Tempe", "BDG-SNK-004", "pcs", "Snack & Cemilan", 5_000, 9_000, 0, 60, 10, "8997000004"},
	{"Cheetos Jagung", "BDG-SNK-005", "pcs", "Snack & Cemilan", 6_000, 11_000, 0, 60, 10, "8997000005"},
	{"Oreo Original", "BDG-SNK-006", "pcs", "Snack & Cemilan", 6_000, 11_000, 0, 72, 10, "8997000006"},
	{"Roma Kelapa", "BDG-SNK-007", "pcs", "Snack & Cemilan", 4_500, 8_000, 0, 80, 10, "8997000007"},
	{"Wafer Astor", "BDG-SNK-008", "pcs", "Snack & Cemilan", 5_000, 9_000, 0, 60, 10, "8997000008"},
	{"Gery Saluut", "BDG-SNK-009", "pcs", "Snack & Cemilan", 3_500, 6_500, 0, 100, 15, "8997000009"},
	{"Silverqueen Chunky 95g", "BDG-SNK-010", "pcs", "Snack & Cemilan", 16_000, 25_000, 0, 30, 5, "8997000010"},
	// ── Kebutuhan Harian ───────────────────────────────
	{"Sabun Lifebuoy 90g", "BDG-DLY-001", "pcs", "Kebutuhan Harian", 4_500, 8_000, 0, 60, 10, "8998000001"},
	{"Shampoo Pantene Sachet", "BDG-DLY-002", "sachet", "Kebutuhan Harian", 1_000, 2_000, 0, 200, 30, "8998000002"},
	{"Pasta Gigi Pepsodent 75g", "BDG-DLY-003", "pcs", "Kebutuhan Harian", 8_000, 14_000, 0, 40, 10, "8998000003"},
	{"Sikat Gigi Formula", "BDG-DLY-004", "pcs", "Kebutuhan Harian", 7_000, 12_000, 0, 30, 5, "8998000004"},
	{"Tisu Paseo 250 sheets", "BDG-DLY-005", "pcs", "Kebutuhan Harian", 8_000, 14_000, 0, 30, 5, "8998000005"},
	{"Pembalut Softex Regular", "BDG-DLY-006", "pcs", "Kebutuhan Harian", 12_000, 19_000, 0, 24, 5, "8998000006"},
	{"Kopi Kapal Api Sachet", "BDG-DLY-007", "sachet", "Kebutuhan Harian", 1_200, 2_500, 0, 300, 50, "8998000007"},
	{"Gula Pasir 250g", "BDG-DLY-008", "pcs", "Kebutuhan Harian", 5_000, 8_000, 0, 50, 10, "8998000008"},
}

// ── Main ──────────────────────────────────────────────────────────────────────

func main() {
	reset := flag.Bool("reset", false, "Truncate all demo tables before seeding")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	db, err := sqlx.Connect("postgres", cfg.DB.DSN())
	must(err)
	defer db.Close()

	ctx := context.Background()

	if *reset {
		log.Println("🗑  Resetting demo data...")
		resetData(ctx, db)
	}

	log.Println("🌱 Seeding MoedahPOS demo data...")

	// ── Roles ─────────────────────────────────────────────────────────────────
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
		log.Fatal("No roles found — run goose migrations first")
	}
	log.Printf("   ✓ Loaded %d roles", len(roles))

	// ── Users ─────────────────────────────────────────────────────────────────
	seedUsers := []struct{ Name, Email, Password string }{
		{"Admin Sistem", "admin@moedah.com", "Admin1234!"},
		{"Budi Manager", "manager@moedah.com", "Manager1234!"},
		{"Sari Kasir", "kasir@moedah.com", "Kasir1234!"},
		{"Andi Staff", "staff@moedah.com", "Staff1234!"},
		{"Rini Kasir BDG", "kasir.bdg@moedah.com", "Kasir1234!"},
	}
	for _, u := range seedUsers {
		_, err = db.ExecContext(ctx, `
			INSERT INTO users (id, name, email, password_hash, is_active)
			VALUES ($1, $2, $3, $4, true)
			ON CONFLICT (email) DO UPDATE SET
				name=EXCLUDED.name, password_hash=EXCLUDED.password_hash,
				is_active=true, deleted_at=NULL
		`, uuid.NewString(), u.Name, u.Email, hashpw(u.Password))
		must(err)
		log.Printf("   ✓ User: %s (%s)", u.Name, u.Email)
	}

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
	type storeRow struct{ ID, Name, Address, Phone, TaxNum string }
	storeList := []storeRow{
		{uuid.NewString(), "Toko Utama — Jakarta", "Jl. Sudirman No. 12, Jakarta Pusat", "021-5551234", "02.123.456.7-001.000"},
		{uuid.NewString(), "Cabang Bandung", "Jl. Dago No. 88, Bandung", "022-7778899", "02.123.456.7-002.000"},
	}
	for _, s := range storeList {
		_, err = db.ExecContext(ctx, `
			INSERT INTO stores (id, name, address, phone, tax_number, currency, is_active)
			VALUES ($1, $2, $3, $4, $5, 'IDR', true) ON CONFLICT DO NOTHING
		`, s.ID, s.Name, s.Address, s.Phone, s.TaxNum)
		must(err)
		log.Printf("   ✓ Store: %s", s.Name)
	}
	mainStoreID := storeList[0].ID
	branchStoreID := storeList[1].ID

	// ── User ↔ Store Memberships ───────────────────────────────────────────────
	memberships := []struct{ Email, StoreID, Role string }{
		{"admin@moedah.com", mainStoreID, "superadmin"},
		{"admin@moedah.com", branchStoreID, "superadmin"},
		{"manager@moedah.com", mainStoreID, "manager"},
		{"kasir@moedah.com", mainStoreID, "cashier"},
		{"staff@moedah.com", mainStoreID, "staff"},
		{"kasir.bdg@moedah.com", branchStoreID, "cashier"},
	}
	for _, m := range memberships {
		_, err = db.ExecContext(ctx, `
			INSERT INTO user_stores (user_id, store_id, role_id, is_active)
			VALUES ($1, $2, $3, true)
			ON CONFLICT (user_id, store_id) DO UPDATE SET role_id=EXCLUDED.role_id, is_active=true
		`, userIDs[m.Email], m.StoreID, roles[m.Role])
		must(err)
	}
	log.Printf("   ✓ Assigned %d user-store memberships", len(memberships))

	// ── Suppliers ─────────────────────────────────────────────────────────────
	type supplierRow struct{ Name, Contact, Phone, Email, Address string }
	suppliers := []supplierRow{
		{"PT Sumber Makmur", "Bapak Hendra", "021-8887766", "hendra@sumbermakmur.co.id", "Jl. Industri No. 5, Tangerang"},
		{"CV Mitra Jaya Sejahtera", "Ibu Dewi", "022-3334455", "dewi@mitrajaya.com", "Jl. Raya Cimahi No. 30, Bandung"},
		{"UD Berkah Abadi", "Bapak Rudi", "031-6667788", "rudi@berkah-abadi.com", "Jl. Pahlawan No. 9, Surabaya"},
		{"PT Indofood Distributor", "Ibu Rina", "021-7778899", "rina@indofood-dist.co.id", "Jl. Raya Bekasi No. 100, Bekasi"},
	}
	supplierIDs := map[string]string{}
	for _, s := range suppliers {
		id := uuid.NewString()
		_, err = db.ExecContext(ctx, `
			INSERT INTO suppliers (id, name, contact_name, phone, email, address, is_active)
			VALUES ($1, $2, $3, $4, $5, $6, true) ON CONFLICT DO NOTHING
		`, id, s.Name, s.Contact, s.Phone, s.Email, s.Address)
		must(err)
		supplierIDs[s.Name] = id
	}
	log.Printf("   ✓ Seeded %d suppliers", len(suppliers))

	// ── Products per Store ────────────────────────────────────────────────────
	log.Println("")
	log.Println("   📦 Seeding products...")

	// Toko Utama — Jakarta (cafe / coffee-shop)
	catMapJakarta := seedCategories(ctx, db, mainStoreID, uniqueCategories(catalogJakarta))
	n1 := seedProducts(ctx, db, mainStoreID, catalogJakarta, catMapJakarta)
	log.Printf("   ✓ Toko Utama — Jakarta : %d products / %d categories", n1, len(catMapJakarta))

	// Cabang Bandung (minimart / convenience)
	catMapBandung := seedCategories(ctx, db, branchStoreID, uniqueCategories(catalogBandung))
	n2 := seedProducts(ctx, db, branchStoreID, catalogBandung, catMapBandung)
	log.Printf("   ✓ Cabang Bandung       : %d products / %d categories", n2, len(catMapBandung))

	// ── Sample Purchase Orders ────────────────────────────────────────────────
	adminID := userIDs["admin@moedah.com"]
	seedPurchaseOrders(ctx, db, mainStoreID, branchStoreID, adminID, supplierIDs)

	// ── Sample Transactions ───────────────────────────────────────────────────
	kasirID := userIDs["kasir@moedah.com"]
	kasirBDGID := userIDs["kasir.bdg@moedah.com"]
	t1 := seedTransactions(ctx, db, mainStoreID, kasirID, catalogJakarta)
	t2 := seedTransactions(ctx, db, branchStoreID, kasirBDGID, catalogBandung)

	log.Println("")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("✅  Seed completed!")
	log.Println("")
	log.Printf("   Products   : %d (Jakarta) + %d (Bandung) = %d total", n1, n2, n1+n2)
	log.Printf("   Transactions: %d (Jakarta) + %d (Bandung)", t1, t2)
	log.Println("")
	log.Println("   Demo Credentials:")
	log.Println("   ┌──────────────────────────────────────────────────────┐")
	log.Println("   │ Role           Email                  Password       │")
	log.Println("   ├──────────────────────────────────────────────────────┤")
	log.Println("   │ superadmin     admin@moedah.com       Admin1234!     │")
	log.Println("   │ manager        manager@moedah.com     Manager1234!   │")
	log.Println("   │ cashier(JKT)   kasir@moedah.com       Kasir1234!     │")
	log.Println("   │ cashier(BDG)   kasir.bdg@moedah.com   Kasir1234!     │")
	log.Println("   │ staff          staff@moedah.com        Staff1234!     │")
	log.Println("   └──────────────────────────────────────────────────────┘")
	log.Println("   Frontend : http://localhost:3000")
	log.Println("   API      : http://localhost:8080")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

// ── Seed helpers ──────────────────────────────────────────────────────────────

// uniqueCategories extracts ordered unique category names from a catalog.
func uniqueCategories(catalog []ProductSeed) []string {
	seen := map[string]bool{}
	var out []string
	for _, p := range catalog {
		if !seen[p.Category] {
			seen[p.Category] = true
			out = append(out, p.Category)
		}
	}
	return out
}

// seedCategories upserts categories for a store and returns name→id map.
func seedCategories(ctx context.Context, db *sqlx.DB, storeID string, names []string) map[string]string {
	ids := map[string]string{}
	for _, n := range names {
		id := uuid.NewString()
		_, err := db.ExecContext(ctx, `
			INSERT INTO categories (id, store_id, name)
			VALUES ($1, $2, $3) ON CONFLICT DO NOTHING
		`, id, storeID, n)
		must(err)
		// read back actual id (may already exist)
		var actual string
		must(db.QueryRowContext(ctx,
			`SELECT id FROM categories WHERE store_id=$1 AND name=$2`, storeID, n,
		).Scan(&actual))
		ids[n] = actual
	}
	return ids
}

// seedProducts upserts all products + stock_levels for a given store.
// Returns the number of products seeded.
func seedProducts(ctx context.Context, db *sqlx.DB, storeID string, catalog []ProductSeed, catMap map[string]string) int {
	for _, p := range catalog {
		catID := catMap[p.Category]
		_, err := db.ExecContext(ctx, `
			INSERT INTO products
			  (id, store_id, category_id, sku, name, barcode, unit, cost_price, sell_price, tax_rate, is_active)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, true)
			ON CONFLICT (store_id, sku) DO UPDATE SET
			  name=EXCLUDED.name, category_id=EXCLUDED.category_id,
			  barcode=EXCLUDED.barcode, cost_price=EXCLUDED.cost_price,
			  sell_price=EXCLUDED.sell_price, tax_rate=EXCLUDED.tax_rate,
			  updated_at=NOW(), deleted_at=NULL
		`, uuid.NewString(), storeID, catID, p.SKU, p.Name, p.Barcode,
			p.Unit, p.CostPrice, p.SellPrice, p.TaxRate)
		must(err)

		// read actual product id after upsert
		var prodID string
		must(db.QueryRowContext(ctx,
			`SELECT id FROM products WHERE store_id=$1 AND sku=$2`, storeID, p.SKU,
		).Scan(&prodID))

		_, err = db.ExecContext(ctx, `
			INSERT INTO stock_levels (product_id, store_id, quantity, min_quantity)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (product_id, store_id) DO UPDATE SET
			  quantity=EXCLUDED.quantity, min_quantity=EXCLUDED.min_quantity, updated_at=NOW()
		`, prodID, storeID, p.InitQty, p.MinQty)
		must(err)
	}
	return len(catalog)
}

// seedPurchaseOrders creates sample POs for both stores.
func seedPurchaseOrders(ctx context.Context, db *sqlx.DB, mainStoreID, branchStoreID, adminID string, supplierIDs map[string]string) {
	pos := []struct {
		StoreID, SupplierKey, Status, Notes string
		Offset                              int // days ago
	}{
		{mainStoreID, "PT Sumber Makmur", "draft", "Pembelian rutin mingguan", 0},
		{mainStoreID, "CV Mitra Jaya Sejahtera", "ordered", "Restock minuman kemasan", 3},
		{branchStoreID, "PT Indofood Distributor", "draft", "Restock mie dan snack", 0},
		{branchStoreID, "CV Mitra Jaya Sejahtera", "received", "Pembelian bulan lalu", 14},
	}
	for i, po := range pos {
		poID := uuid.NewString()
		poNum := fmt.Sprintf("PO-%s-%03d", time.Now().Format("20060102"), i+1)
		_, _ = db.ExecContext(ctx, `
			INSERT INTO purchase_orders
			  (id, store_id, supplier_id, po_number, status, total_amount, ordered_by, notes, created_at)
			VALUES ($1, $2, $3, $4, $5, 0, $6, $7, NOW() - ($8 * interval '1 day'))
			ON CONFLICT DO NOTHING
		`, poID, po.StoreID, supplierIDs[po.SupplierKey], poNum, po.Status, adminID, po.Notes, po.Offset)
	}
	log.Printf("   ✓ Seeded %d sample purchase orders", len(pos))
}

// seedTransactions creates sample completed transactions for a store.
// Returns number of transactions created.
func seedTransactions(ctx context.Context, db *sqlx.DB, storeID, cashierID string, catalog []ProductSeed) int {
	// pick first 8 products from this store's catalog
	picks := catalog
	if len(picks) > 8 {
		picks = picks[:8]
	}
	customers := []string{
		"Budi Santoso", "Pelanggan Umum", "Ani Rahayu",
		"Dewi Lestari", "Rudi Hartono", "Maya Indah",
	}
	methods := []string{"cash", "qris", "cash", "card", "cash", "qris"}
	count := 0

	for i, p := range picks {
		var prodID string
		var sellPrice, taxRate float64
		var prodName, sku string
		err := db.QueryRowContext(ctx, `
			SELECT id, name, sku, sell_price, tax_rate FROM products WHERE store_id=$1 AND sku=$2
		`, storeID, p.SKU).Scan(&prodID, &prodName, &sku, &sellPrice, &taxRate)
		if err != nil {
			continue
		}

		qty := float64(i%3 + 1)
		subtotal := sellPrice * qty
		tax := subtotal * (taxRate / 100)
		total := subtotal + tax
		payAmt := total + float64((i+1)*2000)
		customer := customers[i%len(customers)]
		method := methods[i%len(methods)]
		hoursAgo := float64(i * 2)

		txID := uuid.NewString()
		_, err = db.ExecContext(ctx, `
			INSERT INTO transactions
			  (id, store_id, cashier_id, customer_name, subtotal, discount_amt, tax_amt, total,
			   payment_method, payment_amount, change_amount, status, created_at)
			VALUES ($1,$2,$3,$4,$5,0,$6,$7,$8,$9,$10,'completed', NOW()-($11 * interval '1 hour'))
			ON CONFLICT DO NOTHING
		`, txID, storeID, cashierID, customer,
			subtotal, tax, total, method, payAmt, payAmt-total, hoursAgo)
		if err != nil {
			continue
		}
		_, err = db.ExecContext(ctx, `
			INSERT INTO transaction_items
			  (id, transaction_id, product_id, product_name, sku, quantity, unit_price, discount_pct, tax_rate, subtotal)
			VALUES ($1,$2,$3,$4,$5,$6,$7,0,$8,$9)
		`, uuid.NewString(), txID, prodID, prodName, sku, qty, sellPrice, taxRate, subtotal)
		if err != nil {
			continue
		}
		// deduct stock
		_, _ = db.ExecContext(ctx, `
			UPDATE stock_levels SET quantity=GREATEST(0,quantity-$1), updated_at=NOW()
			WHERE product_id=$2 AND store_id=$3
		`, qty, prodID, storeID)
		_, _ = db.ExecContext(ctx, `
			INSERT INTO stock_movements (id,product_id,store_id,ref_type,ref_id,quantity_delta,notes,created_by)
			VALUES ($1,$2,$3,'sale',$4,$5,'Penjualan kasir',$6)
		`, uuid.NewString(), prodID, storeID, txID, -qty, cashierID)
		count++
	}
	return count
}

// resetData deletes all demo tables (preserves roles/permissions/migrations).
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
		if _, err := db.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s", t)); err != nil {
			log.Printf("   warn: could not clear %s: %v", t, err)
		}
	}
	log.Println("   ✓ All demo tables cleared")
}
