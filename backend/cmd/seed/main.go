// cmd/seed/main.go — Demo data seeder for MoedahPOS
// Usage: go run ./cmd/seed/main.go
//
//	go run ./cmd/seed/main.go --reset   (drop all data, then re-seed)
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
	"golang.org/x/crypto/bcrypt"

	"github.com/moedahpos/backend/internal/config"
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
	{"The Hijau", "BEV-006", "cup", "Minuman", 4_000, 15_000, 11, 95, 10, "8991000006"},
	{"The Tarik", "BEV-007", "cup", "Minuman", 5_000, 16_000, 11, 80, 10, "8991000007"},
	{"Es The Manis", "BEV-008", "cup", "Minuman", 3_000, 12_000, 0, 120, 10, "8991000008"},
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
	{"The Botol Sosro 450ml", "BDG-BEV-003", "botol", "Minuman Kemasan", 4_500, 8_000, 0, 96, 15, "8995000003"},
	{"Pocari Sweat 500ml", "BDG-BEV-004", "botol", "Minuman Kemasan", 7_000, 12_000, 0, 60, 10, "8995000004"},
	{"Coca-Cola 330ml", "BDG-BEV-005", "kaleng", "Minuman Kemasan", 6_500, 11_000, 0, 48, 10, "8995000005"},
	{"Sprite 330ml", "BDG-BEV-006", "kaleng", "Minuman Kemasan", 6_500, 11_000, 0, 48, 10, "8995000006"},
	{"Fanta Orange 330ml", "BDG-BEV-007", "kaleng", "Minuman Kemasan", 6_500, 11_000, 0, 36, 10, "8995000007"},
	{"Susu Ultra 250ml", "BDG-BEV-008", "kotak", "Minuman Kemasan", 5_000, 9_000, 0, 72, 10, "8995000008"},
	{"Good Day Coffee 250ml", "BDG-BEV-009", "botol", "Minuman Kemasan", 5_500, 10_000, 0, 60, 10, "8995000009"},
	{"Nescafe RTD 220ml", "BDG-BEV-010", "kaleng", "Minuman Kemasan", 6_000, 11_000, 0, 48, 10, "8995000010"},
	{"Milo 200ml", "BDG-BEV-011", "kotak", "Minuman Kemasan", 5_500, 10_000, 0, 60, 10, "8995000011"},
	{"The Pucuk 350ml", "BDG-BEV-012", "botol", "Minuman Kemasan", 4_000, 7_000, 0, 96, 15, "8995000012"},
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

// catalogLargeRetail — A larger, more varied set of products for high-volume testing
var catalogLargeRetail = []ProductSeed{
	{"Samsung Galaxy S24", "ELC-001", "unit", "Elektronik", 12_000_000, 16_000_000, 11, 0, 5, "880609531"},
	{"iPhone 15 Pro", "ELC-002", "unit", "Elektronik", 15_000_000, 21_000_000, 11, 0, 5, "19425370"},
	{"MacBook Air M2", "ELC-003", "unit", "Elektronik", 14_000_000, 18_500_000, 11, 0, 2, "19425312"},
	{"Sony WH-1000XM5", "ELC-004", "unit", "Elektronik", 4_500_000, 5_999_000, 11, 0, 10, "45487361"},
	{"Logitech MX Master 3S", "ELC-005", "unit", "Elektronik", 1_100_000, 1_689_000, 11, 0, 15, "09785517"},
	{"T-Shirt Basic White", "APR-001", "pcs", "Apparel", 45_000, 129_000, 0, 0, 20, "20000001"},
	{"Levi's 501 Original", "APR-002", "pcs", "Apparel", 650_000, 1_199_000, 0, 0, 10, "20000002"},
	{"Nike Air Jordan 1", "APR-003", "pasang", "Apparel", 1_800_000, 2_499_000, 11, 0, 5, "20000003"},
	{"Uniqlo Heattech", "APR-004", "pcs", "Apparel", 95_000, 199_000, 0, 0, 30, "20000004"},
	{"Minyak Goreng Bimoli 2L", "GRC-001", "pouch", "Grocery", 28_000, 36_500, 0, 0, 50, "899123401"},
	{"Beras Pandan Wangi 5kg", "GRC-002", "karung", "Grocery", 75_000, 92_000, 0, 0, 20, "899123402"},
	{"Susu Ultra Milk 1L", "GRC-003", "box", "Grocery", 16_000, 21_500, 0, 0, 48, "899123403"},
	{"Deterjen Rinso 700g", "GRC-004", "pack", "Grocery", 18_000, 24_500, 0, 0, 30, "899123404"},
	{"Sabun Mandi Dettol 100g", "GRC-005", "bar", "Grocery", 4_500, 7_800, 0, 0, 100, "899123405"},
	{"Pasta Gigi Pepsodent 190g", "GRC-006", "pcs", "Grocery", 12_000, 18_900, 0, 0, 60, "899123406"},
	{"Shampoo Pantene 170ml", "GRC-007", "botol", "Grocery", 22_000, 32_500, 0, 0, 40, "899123407"},
	{"Teh Pucuk Harum 350ml", "GRC-008", "botol", "Grocery", 2_800, 4_000, 0, 0, 240, "899123408"},
	{"Aqua Mineral 600ml", "GRC-009", "botol", "Grocery", 3_000, 5_500, 0, 0, 480, "899123409"},
	{"Indomie Goreng Original", "GRC-010", "pcs", "Grocery", 2_600, 3_500, 0, 0, 1000, "899123410"},
	{"Kopi Kapal Api 165g", "GRC-011", "pack", "Grocery", 12_500, 16_800, 0, 0, 100, "899123411"},
	{"Gula Pasir Gulaku 1kg", "GRC-012", "kg", "Grocery", 14_500, 18_000, 0, 0, 200, "899123412"},
	{"Garam Meja 250g", "GRC-013", "pcs", "Grocery", 2_000, 3_500, 0, 0, 150, "899123413"},
	{"Kecap Manis Bango 550ml", "GRC-014", "pouch", "Grocery", 22_000, 28_500, 0, 0, 80, "899123414"},
	{"Saos Sambal ABC 335ml", "GRC-015", "botol", "Grocery", 14_000, 19_800, 0, 0, 60, "899123415"},
	{"Tisu Paseo 250s", "GRC-016", "pack", "Grocery", 12_000, 16_500, 0, 0, 120, "899123416"},
	{"Popok MamyPoko S38", "GRC-017", "pack", "Grocery", 65_000, 84_500, 0, 0, 40, "899123417"},
	{"Pembalut Laurier 20s", "GRC-018", "pack", "Grocery", 16_000, 22_500, 0, 0, 80, "899123418"},
	{"Obat Nyamuk Baygon 600ml", "GRC-019", "kaleng", "Grocery", 32_000, 45_000, 0, 0, 50, "899123419"},
	{"Pewangi Molto 800ml", "GRC-020", "pouch", "Grocery", 22_000, 29_900, 0, 0, 60, "899123420"},
}

// ── Main ────────────────────────────────────────────────────────────────────── d

func main() { //nolint:funlen // seeder bootstrap is inherently long
	reset := flag.Bool("reset", false, "Truncate all demo tables before seeding")
	resetRestaurant := flag.Bool("reset-restaurant", false, "Truncate only restaurant store data before seeding")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	db, err := sqlx.Connect("postgres", cfg.DB.DSN())
	must(err)
	defer func() { must(db.Close()) }()

	ctx := context.Background()

	if *reset {
		log.Println("🗑  Resetting all demo data...")
		resetData(ctx, db)
	}

	// ── Store ID discovery for scoped reset ──────────────────────────────────
	// We need these IDs for scoped operations
	var padangStoreID string
	_ = db.QueryRowContext(ctx, "SELECT id FROM stores WHERE name='Rumah Makan Padang Saiyo'").Scan(&padangStoreID)

	if *resetRestaurant && padangStoreID != "" {
		log.Println("🗑  Resetting Padang Restaurant data...")
		resetStoreData(ctx, db, padangStoreID)
	}

	// ── Roles ─────────────────────────────────────────────────────────────────
	roles := map[string]string{}
	rows, err := db.QueryxContext(ctx, `SELECT name, id FROM roles`)
	must(err)
	for rows.Next() {
		var name, id string
		must(rows.Scan(&name, &id))
		roles[name] = id
	}
	must(rows.Close())
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
	must(rows.Close())

	// ── Stores ────────────────────────────────────────────────────────────────
	type storeRow struct{ ID, Name, Address, Phone, TaxNum, StoreType string }
	storeList := []storeRow{
		{uuid.NewString(), "Toko Utama — Jakarta", "Jl. Sudirman No. 12, Jakarta Pusat", "021-5551234", "02.123.456.7-001.000", "retail"},
		{uuid.NewString(), "Cabang Bandung", "Jl. Dago No. 88, Bandung", "022-7778899", "02.123.456.7-002.000", "retail"},
		{uuid.NewString(), "Rumah Makan Padang Saiyo", "Jl. Minangkabau No. 17, Jakarta Selatan", "021-7779988", "02.456.789.1-003.000", "restaurant"},
	}
	for _, s := range storeList {
		_, err = db.ExecContext(ctx, `
			INSERT INTO stores (id, name, address, phone, tax_number, currency, store_type, is_active)
			VALUES ($1, $2, $3, $4, $5, 'IDR', $6, true)
			ON CONFLICT (name) DO UPDATE SET
				address=EXCLUDED.address, phone=EXCLUDED.phone,
				tax_number=EXCLUDED.tax_number, store_type=EXCLUDED.store_type,
				is_active=true, deleted_at=NULL
		`, s.ID, s.Name, s.Address, s.Phone, s.TaxNum, s.StoreType)
		must(err)
		log.Printf("   ✓ Store: %s (%s)", s.Name, s.StoreType)
	}
	mainStoreID := storeList[0].ID
	branchStoreID := storeList[1].ID
	padangStoreID = storeList[2].ID

	// ── User ↔ Store Memberships ───────────────────────────────────────────────
	memberships := []struct{ Email, StoreID, Role string }{
		{"admin@moedah.com", mainStoreID, "superadmin"},
		{"admin@moedah.com", branchStoreID, "superadmin"},
		{"admin@moedah.com", padangStoreID, "superadmin"},
		{"manager@moedah.com", mainStoreID, "manager"},
		{"manager@moedah.com", padangStoreID, "manager"},
		{"kasir@moedah.com", mainStoreID, "cashier"},
		{"kasir@moedah.com", padangStoreID, "cashier"},
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
			VALUES ($1, $2, $3, $4, $5, $6, true)
			ON CONFLICT (name) DO UPDATE SET
				contact_name=EXCLUDED.contact_name, phone=EXCLUDED.phone,
				email=EXCLUDED.email, address=EXCLUDED.address, is_active=true
		`, id, s.Name, s.Contact, s.Phone, s.Email, s.Address)
		must(err)
		supplierIDs[s.Name] = id
		log.Printf("   ✓ Seeded supplier: %s", s.Name)
	}
	log.Printf("   ✓ Seeded %d suppliers", len(suppliers))

	// ── Products per Store ────────────────────────────────────────────────────
	log.Println("")
	log.Println("   📦 Seeding products...")

	// Toko Utama — Jakarta (cafe / coffee-shop)
	catMapJakarta := seedCategories(ctx, db, mainStoreID, uniqueCategories(catalogJakarta))
	n1 := seedProducts(ctx, db, mainStoreID, catalogJakarta, catMapJakarta, false)
	log.Printf("   ✓ Toko Utama — Jakarta : %d products / %d categories", n1, len(catMapJakarta))

	// Cabang Bandung (minimart / convenience)
	catMapBandung := seedCategories(ctx, db, branchStoreID, uniqueCategories(catalogBandung))
	n2 := seedProducts(ctx, db, branchStoreID, catalogBandung, catMapBandung, false)
	log.Printf("   ✓ Cabang Bandung       : %d products / %d categories", n2, len(catMapBandung))

	// ── Restaurant Padang Seed ────────────────────────────────────────────────
	log.Println("")
	log.Println("   🍽️  Seeding restaurant (Padang) data...")
	seedRestaurantPadang(ctx, db, padangStoreID, userIDs["admin@moedah.com"])

	// ── High Volume Retail Expansion ──────────────────────────────────────────
	log.Println("")
	log.Println("   🚀 Preparing High Volume Retail Data (Jakarta Extra)...")
	// Add 50 more procedurally generated items to test scalability
	for i := 1; i <= 50; i++ {
		catalogLargeRetail = append(catalogLargeRetail, ProductSeed{
			Name:      fmt.Sprintf("Item Premium %d", i),
			SKU:       fmt.Sprintf("EXTRA-%03d", i),
			Unit:      "pcs",
			Category:  "Premium Goods",
			CostPrice: float64(10_000 * i),
			SellPrice: float64(15_000 * i),
			TaxRate:   11,
			InitQty:   0,
			MinQty:    5,
		})
	}
	catMapLarge := seedCategories(ctx, db, mainStoreID, uniqueCategories(catalogLargeRetail))
	seedProducts(ctx, db, mainStoreID, catalogLargeRetail, catMapLarge, false)

	// ── Sample Purchase Orders & Stock Population ─────────────────────────────
	adminID := userIDs["admin@moedah.com"]
	seedActivePurchaseOrders(ctx, db, mainStoreID, branchStoreID, adminID, supplierIDs)

	// ── Sample Transactions ───────────────────────────────────────────────────
	kasirID := userIDs["kasir@moedah.com"]
	kasirBDGID := userIDs["kasir.bdg@moedah.com"]
	t1 := seedTransactions(ctx, db, mainStoreID, kasirID, false)
	t2 := seedTransactions(ctx, db, branchStoreID, kasirBDGID, false)
	t3 := seedTransactions(ctx, db, padangStoreID, kasirID, true) // Restaurant transactions

	// ── Active KDS Tickets (Restaurant Only) ──────────────────────────────────
	log.Println("   🔥 Seeding active KDS tickets...")
	seedActiveKDSTickets(ctx, db, padangStoreID, kasirID)

	log.Println("")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("✅  Seed completed!")
	log.Println("")
	log.Printf("   Products   : %d (Jakarta) + %d (Bandung) = %d total", n1, n2, n1+n2)
	log.Printf("   Transactions: %d (Jakarta) + %d (Bandung) + %d (Padang)", t1, t2, t3)
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
func seedProducts(ctx context.Context, db *sqlx.DB, storeID string, catalog []ProductSeed, catMap map[string]string, isRestaurant bool) int {
	for _, p := range catalog {
		catID := catMap[p.Category]
		sellPrice := p.SellPrice
		if isRestaurant {
			sellPrice = 0 // Ingredients in restaurant have no direct sell price
		}
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
			p.Unit, p.CostPrice, sellPrice, p.TaxRate)
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

// seedActivePurchaseOrders creates sample POs and updates stock levels for received ones.
func seedActivePurchaseOrders(ctx context.Context, db *sqlx.DB, mainStoreID, branchStoreID, adminID string, supplierIDs map[string]string) {
	// 1. Get all products for mainStoreID
	var mainProds []struct {
		ID        string  `db:"id"`
		Name      string  `db:"name"`
		CostPrice float64 `db:"cost_price"`
	}
	must(db.SelectContext(ctx, &mainProds, "SELECT id, name, cost_price FROM products WHERE store_id=$1", mainStoreID))

	// 2. Create one big 'received' PO for PT Sumber Makmur for first 40 products
	poID := uuid.NewString()
	poNum := "PO-INIT-001"
	_, err := db.ExecContext(ctx, `
		INSERT INTO purchase_orders (id, store_id, supplier_id, po_number, status, total_amount, ordered_by, received_by, received_at)
		VALUES ($1, $2, $3, $4, 'received', 0, $5, $5, NOW())
		ON CONFLICT (po_number) DO NOTHING
	`, poID, mainStoreID, supplierIDs["PT Sumber Makmur"], poNum, adminID)
	must(err)

	// Since we might have skipped insertion, let's get the ID if it was already there (or use the new one)
	_ = db.QueryRowContext(ctx, "SELECT id FROM purchase_orders WHERE po_number=$1", poNum).Scan(&poID)

	var totalAmt float64
	for i, p := range mainProds {
		if i >= 40 {
			break
		}
		qty := 100.0
		cost := p.CostPrice
		subtotal := qty * cost
		totalAmt += subtotal

		_, err = db.ExecContext(ctx, `
			INSERT INTO purchase_order_items (id, po_id, product_id, quantity, unit_cost, received_qty, subtotal)
			VALUES ($1, $2, $3, $4, $5, $4, $6)
		`, uuid.NewString(), poID, p.ID, qty, cost, subtotal)
		must(err)

		// Update stock
		_, err = db.ExecContext(ctx, "UPDATE stock_levels SET quantity = quantity + $1 WHERE product_id = $2", qty, p.ID)
		must(err)
		_, err = db.ExecContext(ctx, `
			INSERT INTO stock_movements (id, product_id, store_id, ref_type, ref_id, quantity_delta, notes, created_by)
			VALUES ($1, $2, $3, 'purchase_order', $4, $5, 'Initial Seed Stock', $6)
		`, uuid.NewString(), p.ID, mainStoreID, poID, qty, adminID)
		must(err)
	}
	_, err = db.ExecContext(ctx, "UPDATE purchase_orders SET total_amount = $1 WHERE id = $2", totalAmt, poID)
	must(err)

	log.Printf("   ✓ Populated stock for %d products via received PO", 40)
}

// seedTransactions creates a large history of transactions.
func seedTransactions(ctx context.Context, db *sqlx.DB, storeID, cashierID string, isRestaurant bool) int {
	if isRestaurant {
		return seedRestaurantTransactions(ctx, db, storeID, cashierID)
	}

	// Get available products with stock
	var prods []struct {
		ID        string  `db:"id"`
		Name      string  `db:"name"`
		SKU       string  `db:"sku"`
		SellPrice float64 `db:"sell_price"`
		TaxRate   float64 `db:"tax_rate"`
	}
	must(db.SelectContext(ctx, &prods, "SELECT id, name, sku, sell_price, tax_rate FROM products WHERE store_id=$1", storeID))
	if len(prods) == 0 {
		return 0
	}

	customers := []string{"Budi", "Sari", "Andi", "Rini", "Guest", "General Customer", "Loyal Fan"}
	count := 0
	// Seed 250 transactions for this store
	for i := 0; i < 250; i++ {
		txID := uuid.NewString()
		cust := customers[i%len(customers)]
		method := "cash"
		if i%3 == 1 {
			method = "qris"
		} else if i%3 == 2 {
			method = "card"
		}

		// Random date in last 30 days
		daysAgo := i % 30
		hoursAgo := i % 24
		createdAt := time.Now().AddDate(0, 0, -daysAgo).Add(time.Duration(-hoursAgo) * time.Hour)

		// 1-3 items per transaction
		numItems := (i % 3) + 1
		var subtotal, taxTotal float64

		// Pre-calculate totals
		type itemLine struct {
			p struct {
				ID        string  `db:"id"`
				Name      string  `db:"name"`
				SKU       string  `db:"sku"`
				SellPrice float64 `db:"sell_price"`
				TaxRate   float64 `db:"tax_rate"`
			}
			qty float64
			sub float64
			tax float64
		}
		var lines []itemLine
		for j := 0; j < numItems; j++ {
			p := prods[(i+j)%len(prods)]
			qty := 1.0
			itemSub := p.SellPrice * qty
			itemTax := itemSub * (p.TaxRate / 100)
			subtotal += itemSub
			taxTotal += itemTax
			lines = append(lines, itemLine{p: p, qty: qty, sub: itemSub, tax: itemTax})
		}

		tx, err := db.BeginTxx(ctx, nil)
		must(err)

		total := subtotal + taxTotal
		_, err = tx.ExecContext(ctx, `
			INSERT INTO transactions (id, store_id, cashier_id, customer_name, subtotal, tax_amt, total, payment_method, payment_amount, change_amount, status, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 'completed', $11)
		`, txID, storeID, cashierID, cust, subtotal, taxTotal, total, method, total, 0, createdAt)
		must(err)

		for _, line := range lines {
			_, err = tx.ExecContext(ctx, `
				INSERT INTO transaction_items (id, transaction_id, product_id, product_name, sku, quantity, unit_price, tax_rate, subtotal)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			`, uuid.NewString(), txID, line.p.ID, line.p.Name, line.p.SKU, line.qty, line.p.SellPrice, line.p.TaxRate, line.sub)
			must(err)
		}

		must(tx.Commit())
		count++
	}
	return count
}

// resetData deletes all demo tables (preserves roles/permissions/migrations).
func resetData(ctx context.Context, db *sqlx.DB) {
	tables := []string{
		"menu_item_ingredients", "menu_items",
		"restaurant_tables",
		"transaction_items", "transactions",
		"stock_movements", "stock_batches", "stock_levels",
		"payment_records", "purchase_order_termins", "po_payments", "purchase_order_items", "purchase_orders",
		"price_history", "customers",
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

// resetStoreData deletes all data specifically for one store.
func resetStoreData(ctx context.Context, db *sqlx.DB, storeID string) {
	queries := []string{
		"DELETE FROM transaction_items WHERE transaction_id IN (SELECT id FROM transactions WHERE store_id = $1)",
		"DELETE FROM transactions WHERE store_id = $1",
		"DELETE FROM stock_movements WHERE store_id = $1",
		"DELETE FROM stock_batches WHERE store_id = $1",
		"DELETE FROM stock_levels WHERE store_id = $1",
		"DELETE FROM purchase_order_items WHERE po_id IN (SELECT id FROM purchase_orders WHERE store_id = $1)",
		"DELETE FROM purchase_orders WHERE store_id = $1",
		"DELETE FROM menu_item_ingredients WHERE menu_item_id IN (SELECT id FROM menu_items WHERE store_id = $1)",
		"DELETE FROM menu_items WHERE store_id = $1",
		"DELETE FROM restaurant_tables WHERE store_id = $1",
		"DELETE FROM products WHERE store_id = $1",
		"DELETE FROM categories WHERE store_id = $1",
	}
	for _, q := range queries {
		if _, err := db.ExecContext(ctx, q, storeID); err != nil {
			log.Printf("   warn: could not clear data for store %s: %v", storeID, err)
		}
	}
	log.Println("   ✓ Scoped store data cleared")
}

// seedRestaurantTransactions specifically uses menu_items for sales record.
func seedRestaurantTransactions(ctx context.Context, db *sqlx.DB, storeID, cashierID string) int {
	var menus []struct {
		ID        string  `db:"id"`
		Name      string  `db:"name"`
		SellPrice float64 `db:"sell_price"`
		TaxRate   float64 `db:"tax_rate"`
	}
	must(db.SelectContext(ctx, &menus, "SELECT id, name, sell_price, tax_rate FROM menu_items WHERE store_id=$1", storeID))
	if len(menus) == 0 {
		return 0
	}

	customers := []string{"Sultan", "Manto", "Ujang", "Nur", "Bunda"}
	count := 0
	for i := 0; i < 150; i++ {
		txID := uuid.NewString()
		cust := customers[i%len(customers)]
		daysAgo := i % 20
		hoursAgo := i % 12
		createdAt := time.Now().AddDate(0, 0, -daysAgo).Add(time.Duration(-hoursAgo) * time.Hour)

		numItems := (i % 4) + 1
		var subtotal, taxTotal float64

		type lineItem struct {
			ID        string
			Name      string
			SellPrice float64
			TaxRate   float64
		}
		var lines []lineItem
		for j := 0; j < numItems; j++ {
			m := menus[(i+j)%len(menus)]
			subtotal += m.SellPrice
			taxTotal += m.SellPrice * (m.TaxRate / 100)
			lines = append(lines, lineItem{m.ID, m.Name, m.SellPrice, m.TaxRate})
		}

		total := subtotal + taxTotal
		tx, err := db.BeginTxx(ctx, nil)
		must(err)

		_, err = tx.ExecContext(ctx, `
			INSERT INTO transactions (id, store_id, cashier_id, customer_name, subtotal, tax_amt, total, payment_method, payment_amount, change_amount, status, created_at, order_type)
			VALUES ($1, $2, $3, $4, $5, $6, $7, 'cash', $8, 0, 'completed', $9, 'dine_in')
		`, txID, storeID, cashierID, cust, subtotal, taxTotal, total, total, createdAt)
		must(err)

		for _, l := range lines {
			_, err = tx.ExecContext(ctx, `
				INSERT INTO transaction_items (id, transaction_id, menu_item_id, product_name, sku, quantity, unit_price, tax_rate, subtotal)
				VALUES ($1, $2, $3, $4, 'MENU-ITEM', 1, $5, $6, $5)
			`, uuid.NewString(), txID, l.ID, l.Name, l.SellPrice, l.TaxRate)
			must(err)
		}
		must(tx.Commit())
		count++
	}
	return count
}

// ── Restaurant Padang Seeder ──────────────────────────────────────────────────

// catalogPadang defines the raw ingredients / stock items used by a Padang restaurant.
// These are stored in products/stock_levels as ingredients.
var catalogPadang = []ProductSeed{
	// ── Daging & Protein ───────────────────────────────────
	{"Daging Sapi", "PDG-001", "kg", "Daging & Protein", 130_000, 0, 0, 20, 2, ""},
	{"Ayam Kampung", "PDG-002", "kg", "Daging & Protein", 65_000, 0, 0, 25, 3, ""},
	{"Ikan Kakap", "PDG-003", "kg", "Daging & Protein", 80_000, 0, 0, 10, 2, ""},
	{"Telur Ayam", "PDG-004", "butir", "Daging & Protein", 2_500, 0, 0, 300, 50, ""},
	{"Paru Sapi", "PDG-005", "kg", "Daging & Protein", 90_000, 0, 0, 8, 1, ""},
	{"Kikil Sapi", "PDG-006", "kg", "Daging & Protein", 70_000, 0, 0, 5, 1, ""},
	// ── Sayur & Nabati ────────────────────────────────────
	{"Nangka Muda", "PDG-011", "kg", "Sayur & Nabati", 8_000, 0, 0, 15, 2, ""},
	{"Kacang Panjang", "PDG-012", "kg", "Sayur & Nabati", 12_000, 0, 0, 10, 2, ""},
	{"Daun Singkong", "PDG-013", "ikat", "Sayur & Nabati", 5_000, 0, 0, 30, 5, ""},
	{"Terong Ungu", "PDG-014", "kg", "Sayur & Nabati", 15_000, 0, 0, 10, 2, ""},
	{"Pare", "PDG-015", "kg", "Sayur & Nabati", 12_000, 0, 0, 8, 1, ""},
	// ── Bumbu & Rempah ────────────────────────────────────
	{"Santan Kelapa", "PDG-021", "liter", "Bumbu & Rempah", 20_000, 0, 0, 40, 5, ""},
	{"Cabai Merah", "PDG-022", "kg", "Bumbu & Rempah", 45_000, 0, 0, 10, 2, ""},
	{"Bawang Merah", "PDG-023", "kg", "Bumbu & Rempah", 35_000, 0, 0, 15, 3, ""},
	{"Bawang Putih", "PDG-024", "kg", "Bumbu & Rempah", 40_000, 0, 0, 10, 2, ""},
	{"Lengkuas", "PDG-025", "kg", "Bumbu & Rempah", 25_000, 0, 0, 5, 1, ""},
	{"Serai", "PDG-026", "batang", "Bumbu & Rempah", 2_000, 0, 0, 50, 10, ""},
	{"Daun Jeruk", "PDG-027", "lembar", "Bumbu & Rempah", 500, 0, 0, 100, 20, ""},
	{"Kunyit", "PDG-028", "kg", "Bumbu & Rempah", 20_000, 0, 0, 5, 1, ""},
	// ── Karbohidrat ───────────────────────────────────────
	{"Beras", "PDG-031", "kg", "Karbohidrat", 13_000, 0, 0, 100, 20, ""},
	{"Lontong", "PDG-032", "porsi", "Karbohidrat", 3_000, 0, 0, 80, 10, ""},
	// ── Minuman ───────────────────────────────────────────
	{"Es Batu", "PDG-041", "kg", "Minuman", 3_000, 0, 0, 20, 5, ""},
	{"The Celup", "PDG-042", "sachet", "Minuman", 500, 0, 0, 200, 30, ""},
	{"Gula Pasir", "PDG-043", "kg", "Minuman", 14_000, 0, 0, 20, 5, ""},
	{"Jeruk Nipis", "PDG-044", "buah", "Minuman", 1_500, 0, 0, 50, 10, ""},
}

// menuPadang defines the sold menu items with their ingredient compositions.
type menuItemSeed struct {
	Name, Category, Description string
	SellPrice, TaxRate          float64
	Ingredients                 []struct {
		SKU string
		Qty float64
	}
}

var menuPadang = []menuItemSeed{
	{
		"Rendang Sapi", "Masakan Utama",
		"Rendang daging sapi empuk dimasak lama dengan santan dan rempah khas Minang",
		45_000, 11,
		[]struct {
			SKU string
			Qty float64
		}{
			{"PDG-001", 0.25}, // 250g daging sapi
			{"PDG-021", 0.2},  // 200ml santan
			{"PDG-022", 0.05}, // 50g cabai
			{"PDG-023", 0.03}, // 30g bawang merah
			{"PDG-024", 0.02}, // 20g bawang putih
			{"PDG-025", 0.02}, // 20g lengkuas
			{"PDG-028", 0.01}, // 10g kunyit
			{"PDG-031", 0.2},  // 200g beras (nasi)
		},
	},
	{
		"Gulai Ayam", "Masakan Utama",
		"Ayam kampung dimasak gulai kuning dengan santan gurih dan rempah lengkap",
		35_000, 11,
		[]struct {
			SKU string
			Qty float64
		}{
			{"PDG-002", 0.25},
			{"PDG-021", 0.25},
			{"PDG-022", 0.04},
			{"PDG-023", 0.03},
			{"PDG-028", 0.01},
			{"PDG-026", 2},
			{"PDG-031", 0.2},
		},
	},
	{
		"Gulai Paku (Gulai Pakis)", "Masakan Utama",
		"Sayur pakis dimasak gulai santan khas Minang",
		20_000, 11,
		[]struct {
			SKU string
			Qty float64
		}{
			{"PDG-013", 0.15},
			{"PDG-021", 0.15},
			{"PDG-022", 0.03},
			{"PDG-023", 0.02},
			{"PDG-031", 0.2},
		},
	},
	{
		"Sayur Nangka", "Masakan Utama",
		"Nangka muda dimasak dengan santan dan bumbu Padang",
		18_000, 11,
		[]struct {
			SKU string
			Qty float64
		}{
			{"PDG-011", 0.2},
			{"PDG-021", 0.15},
			{"PDG-023", 0.02},
			{"PDG-031", 0.2},
		},
	},
	{
		"Telur Balado", "Lauk Pelengkap",
		"Telur ayam goreng dibalut sambal balado merah pedas",
		12_000, 0,
		[]struct {
			SKU string
			Qty float64
		}{
			{"PDG-004", 2},
			{"PDG-022", 0.03},
			{"PDG-023", 0.02},
			{"PDG-024", 0.01},
		},
	},
	{
		"Gulai Ikan Kakap", "Masakan Utama",
		"Ikan kakap segar dimasak gulai kuning pedas khas Minanga",
		38_000, 11,
		[]struct {
			SKU string
			Qty float64
		}{
			{"PDG-003", 0.25},
			{"PDG-021", 0.2},
			{"PDG-022", 0.04},
			{"PDG-028", 0.01},
			{"PDG-026", 2},
			{"PDG-027", 3},
			{"PDG-031", 0.2},
		},
	},
	{
		"Dendeng Balado", "Masakan Utama",
		"Irisan daging sapi tipis digoreng kering dilumuri balado",
		42_000, 11,
		[]struct {
			SKU string
			Qty float64
		}{
			{"PDG-001", 0.2},
			{"PDG-022", 0.05},
			{"PDG-023", 0.03},
			{"PDG-031", 0.2},
		},
	},
	{
		"Gulai Tunjang (Kikil)", "Masakan Utama",
		"Kikil sapi empuk dimasak gulai santan kental",
		30_000, 11,
		[]struct {
			SKU string
			Qty float64
		}{
			{"PDG-006", 0.2},
			{"PDG-021", 0.2},
			{"PDG-022", 0.04},
			{"PDG-025", 0.02},
			{"PDG-031", 0.2},
		},
	},
	{
		"Nasi Putih", "Nasi",
		"Nasi putih pulen porsi dewasa",
		5_000, 0,
		[]struct {
			SKU string
			Qty float64
		}{
			{"PDG-031", 0.2},
		},
	},
	{
		"Es The Manis", "Minuman",
		"The manis segar dengan es batu",
		8_000, 0,
		[]struct {
			SKU string
			Qty float64
		}{
			{"PDG-042", 1},
			{"PDG-043", 0.02},
			{"PDG-041", 0.1},
		},
	},
	{
		"Es Jeruk", "Minuman",
		"Jeruk nipis peras segar dengan es batu dan gula",
		10_000, 0,
		[]struct {
			SKU string
			Qty float64
		}{
			{"PDG-044", 3},
			{"PDG-043", 0.03},
			{"PDG-041", 0.15},
		},
	},
	{
		"The Hangat", "Minuman",
		"The tawar atau manis, disajikan hangat",
		6_000, 0,
		[]struct {
			SKU string
			Qty float64
		}{
			{"PDG-042", 1},
			{"PDG-043", 0.02},
		},
	},
}

// seedRestaurantPadang seeds tables, ingredient products, and menu items.
func seedRestaurantPadang(ctx context.Context, db *sqlx.DB, storeID, _ string) {
	// ── Tables (Meja) ──────────────────────────────────────────────────────────
	tableNames := []struct {
		Number   string
		Capacity int
		Notes    string
	}{
		{"1", 4, "Dekat pintu masuk"},
		{"2", 4, ""},
		{"3", 6, "Meja keluarga"},
		{"4", 6, "Meja keluarga"},
		{"5", 2, "Pojok"},
		{"6", 2, "Pojok"},
		{"VIP-1", 8, "Ruang VIP ber-AC"},
		{"VIP-2", 8, "Ruang VIP ber-AC"},
	}
	for _, t := range tableNames {
		_, err := db.ExecContext(ctx, `
			INSERT INTO restaurant_tables (id, store_id, table_number, capacity, status, notes)
			VALUES ($1, $2, $3, $4, 'available', $5)
			ON CONFLICT DO NOTHING
		`, uuid.NewString(), storeID, t.Number, t.Capacity, t.Notes)
		must(err)
	}
	log.Printf("   ✓ Seeded %d restaurant tables", len(tableNames))

	// ── Ingredient products (raw materials) ───────────────────────────────────
	catMap := seedCategories(ctx, db, storeID, uniqueCategories(catalogPadang))
	n := seedProducts(ctx, db, storeID, catalogPadang, catMap, true) // isRestaurant = true
	log.Printf("   ✓ Seeded %d ingredient products / %d categories", n, len(catMap))

	// ── Menu items ────────────────────────────────────────────────────────────
	// Ensure menu categories exist
	menucatNames := []string{"Masakan Utama", "Lauk Pelengkap", "Nasi", "Minuman"}
	menucatMap := seedCategories(ctx, db, storeID, menucatNames)

	for _, m := range menuPadang {
		catID := menucatMap[m.Category]
		menuItemID := uuid.NewString()
		_, err := db.ExecContext(ctx, `
			INSERT INTO menu_items (id, store_id, category_id, name, description, sell_price, tax_rate)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT DO NOTHING
		`, menuItemID, storeID, catID, m.Name, m.Description, m.SellPrice, m.TaxRate)
		must(err)

		// re-read actual id (may already exist)
		var actualMenuID string
		must(db.QueryRowContext(ctx,
			`SELECT id FROM menu_items WHERE store_id=$1 AND name=$2`, storeID, m.Name,
		).Scan(&actualMenuID))

		for _, ing := range m.Ingredients {
			var prodID string
			err := db.QueryRowContext(ctx,
				`SELECT id FROM products WHERE store_id=$1 AND sku=$2`, storeID, ing.SKU,
			).Scan(&prodID)
			if err != nil {
				log.Printf("   warn: ingredient SKU %s not found for menu %s", ing.SKU, m.Name)
				continue
			}
			_, _ = db.ExecContext(ctx, `
				INSERT INTO menu_item_ingredients (id, menu_item_id, product_id, quantity)
				VALUES ($1, $2, $3, $4)
				ON CONFLICT DO NOTHING
			`, uuid.NewString(), actualMenuID, prodID, ing.Qty)
		}
	}
	log.Printf("   ✓ Seeded %d menu items with ingredients", len(menuPadang))
}

// seedActiveKDSTickets creates active 'hold' or 'draft' transactions with pending items.
func seedActiveKDSTickets(ctx context.Context, db *sqlx.DB, storeID, cashierID string) {
	// Get some tables
	var tables []struct {
		ID     string `db:"id"`
		Number string `db:"table_number"`
	}
	must(db.SelectContext(ctx, &tables, "SELECT id, table_number FROM restaurant_tables WHERE store_id=$1 LIMIT 4", storeID))

	// Get some menu items
	var menuItems []struct {
		ID        string  `db:"id"`
		Name      string  `db:"name"`
		SellPrice float64 `db:"sell_price"`
	}
	must(db.SelectContext(ctx, &menuItems, "SELECT id, name, sell_price FROM menu_items WHERE store_id=$1", storeID))

	for i, table := range tables {
		txID := uuid.NewString()
		subtotal := 0.0

		// Create draft transaction
		_, err := db.ExecContext(ctx, `
			INSERT INTO transactions (id, store_id, cashier_id, table_id, customer_name, subtotal, total, status, payment_method, created_at)
			VALUES ($1, $2, $3, $4, $5, 0, 0, 'hold', 'cash', NOW() - ($6 * interval '5 minute'))
		`, txID, storeID, cashierID, table.ID, fmt.Sprintf("Pelanggan %s", table.Number), i)
		must(err)

		// Add 2 random menu items
		for j := 0; j < 2; j++ {
			item := menuItems[(i+j)%len(menuItems)]
			qty := 1.0
			subtotal += item.SellPrice * qty

			status := "pending"
			if j == 0 && i%2 == 0 {
				status = "completed"
			}

			_, err = db.ExecContext(ctx, `
				INSERT INTO transaction_items (id, transaction_id, menu_item_id, product_name, sku, quantity, unit_price, subtotal, status)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			`, uuid.NewString(), txID, item.ID, item.Name, "MENU-PDG", qty, item.SellPrice, item.SellPrice, status)
			must(err)
		}

		_, err = db.ExecContext(ctx, "UPDATE transactions SET subtotal=$1, total=$1 WHERE id=$2", subtotal, txID)
		must(err)
	}
}
