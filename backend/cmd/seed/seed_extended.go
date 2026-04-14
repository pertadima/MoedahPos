// cmd/seed/seed_extended.go — Extended seeder for financial, inventory, and customer data
// Called from main.go after products and transactions have been seeded.
package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// ── Date helpers ───────────────────────────────────────────────────────────────

// tsNDaysAgo returns a timestamp N days in the past at a realistic business hour (08:00–22:00).
func tsNDaysAgo(n int) time.Time {
	now := time.Now()
	base := now.AddDate(0, 0, -n)
	hour := 8 + rand.Intn(14) //nolint:gosec // demo seeder, not cryptographic
	min := rand.Intn(60)      //nolint:gosec
	return time.Date(base.Year(), base.Month(), base.Day(), hour, min, 0, 0, base.Location())
}

// dateOnly returns a date string (YYYY-MM-DD) N days ago.
func dateOnly(n int) string {
	return time.Now().AddDate(0, 0, -n).Format("2006-01-02")
}

// ── Seed extended data for all stores ─────────────────────────────────────────

//nolint:funlen,gocognit,cyclop // Demo seeder is inherently long
func seedExtendedData(ctx context.Context, db *sqlx.DB, mainStoreID, branchStoreID, padangStoreID, adminID string) {
	stores := []struct {
		ID    string
		Label string
	}{
		{mainStoreID, "Toko Utama Jakarta"},
		{branchStoreID, "Cabang Bandung"},
		{padangStoreID, "Padang Saiyo"},
	}

	// ── 1. Income & Expense Categories ────────────────────────────────────────
	log.Println("   💰 Seeding income & expense categories...")
	incomeCatIDs := seedIncomeCategoriesExtended(ctx, db)
	expenseCatIDs := seedExpenseCategoriesExtended(ctx, db)
	log.Printf("   ✓ income cats=%d  expense cats=%d", len(incomeCatIDs), len(expenseCatIDs))

	for _, s := range stores {
		log.Printf("   🏪 [%s] seeding extended data...", s.Label)

		// Get products for this store
		var prods []struct {
			ID        string  `db:"id"`
			Name      string  `db:"name"`
			SKU       string  `db:"sku"`
			CostPrice float64 `db:"cost_price"`
			SellPrice float64 `db:"sell_price"`
		}
		must(db.SelectContext(ctx, &prods,
			`SELECT id, name, sku, cost_price, sell_price FROM products WHERE store_id=$1 AND deleted_at IS NULL ORDER BY created_at LIMIT 30`,
			s.ID))

		if len(prods) == 0 {
			log.Printf("   warn: no products found for store %s, skipping", s.Label)
			continue
		}

		// ── 2. Price History ────────────────────────────────────────────────
		nPriceHistory := seedPriceHistory(ctx, db, s.ID, prods, adminID)
		log.Printf("   ✓ [%s] price_history=%d records", s.Label, nPriceHistory)

		// ── 3. Customers ────────────────────────────────────────────────────
		nCustomers := seedCustomers(ctx, db, s.ID)
		log.Printf("   ✓ [%s] customers=%d", s.Label, nCustomers)

		// ── 4. Purchase Orders + Stock Batches ──────────────────────────────
		nPOs := seedPurchaseOrders(ctx, db, s.ID, prods, adminID)
		log.Printf("   ✓ [%s] purchase_orders=%d", s.Label, nPOs)

		// ── 5. Incomes ──────────────────────────────────────────────────────
		nIncome := seedIncomes(ctx, db, s.ID, incomeCatIDs, adminID)
		log.Printf("   ✓ [%s] incomes=%d", s.Label, nIncome)

		// ── 6. Expenses ─────────────────────────────────────────────────────
		nExpense := seedExpenses(ctx, db, s.ID, expenseCatIDs, adminID)
		log.Printf("   ✓ [%s] expenses=%d", s.Label, nExpense)
	}
}

// ── Income categories ─────────────────────────────────────────────────────────

func seedIncomeCategoriesExtended(ctx context.Context, db *sqlx.DB) []string {
	extraCats := []struct{ Name, Desc string }{
		{"Penjualan Tunai", "Pendapatan dari penjualan langsung secara tunai"},
		{"Pendapatan Jasa", "Pendapatan dari layanan atau jasa tambahan"},
		{"Pendapatan Lain-lain", "Pendapatan di luar ops utama"},
		{"Injeksi Modal", "Penambahan modal dari pemilik atau investor"},
		{"Pengembalian Supplier", "Refund atau kelebihan bayar dari supplier"},
		{"Piutang Diterima", "Penerimaan pembayaran dari piutang pelanggan"},
		{"Bunga Bank", "Pendapatan bunga dari tabungan atau deposito"},
		{"Komisi Agen", "Pendapatan komisi dari agen penjualan"},
		{"Lainnya", "Penerimaan kas lainnya yang tidak terklasifikasi"},
	}

	var ids []string
	for _, c := range extraCats {
		// Upsert by name (global table, no store_id)
		var id string
		err := db.QueryRowContext(ctx,
			`SELECT id FROM income_categories WHERE name=$1`, c.Name,
		).Scan(&id)
		if err != nil {
			id = uuid.NewString()
			_, err = db.ExecContext(ctx,
				`INSERT INTO income_categories (id, name, description) VALUES ($1,$2,$3) ON CONFLICT DO NOTHING`,
				id, c.Name, c.Desc)
			must(err)
			// Re-read actual id in case of conflict
			_ = db.QueryRowContext(ctx, `SELECT id FROM income_categories WHERE name=$1`, c.Name).Scan(&id)
		}
		ids = append(ids, id)
	}
	return ids
}

// ── Expense categories ────────────────────────────────────────────────────────

func seedExpenseCategoriesExtended(ctx context.Context, db *sqlx.DB) []string {
	extraCats := []struct{ Name, Desc string }{
		{"Sewa Tempat", "Biaya sewa gedung atau ruko"},
		{"Listrik", "Tagihan listrik bulanan"},
		{"Air", "Tagihan PDAM bulanan"},
		{"Internet & Telepon", "Biaya internet dan tagihan telepon"},
		{"Gaji Karyawan", "Gaji dan tunjangan karyawan"},
		{"Bonus & Insentif", "Bonus kinerja dan insentif staf"},
		{"Pembelian Bahan Baku", "Pembelian bahan baku non-PO"},
		{"Pemeliharaan Peralatan", "Servis dan perbaikan peralatan"},
		{"Pemasaran & Promosi", "Biaya iklan, promosi, dan sosmed"},
		{"Perizinan & Pajak", "Biaya izin usaha dan pajak daerah"},
		{"Perlengkapan Kantor", "ATK dan perlengkapan ops kantor"},
		{"Transportasi", "Ongkir, pengiriman, dan biaya transportasi"},
		{"Biaya Bank", "Administrasi bank dan biaya transfer"},
		{"Asuransi", "Premi asuransi aset dan kecelakaan kerja"},
		{"Lainnya", "Pengeluaran ops lainnya"},
	}

	var ids []string
	for _, c := range extraCats {
		var id string
		err := db.QueryRowContext(ctx,
			`SELECT id FROM expense_categories WHERE name=$1`, c.Name,
		).Scan(&id)
		if err != nil {
			id = uuid.NewString()
			_, err = db.ExecContext(ctx,
				`INSERT INTO expense_categories (id, name, description) VALUES ($1,$2,$3) ON CONFLICT DO NOTHING`,
				id, c.Name, c.Desc)
			must(err)
			_ = db.QueryRowContext(ctx, `SELECT id FROM expense_categories WHERE name=$1`, c.Name).Scan(&id)
		}
		ids = append(ids, id)
	}
	return ids
}

// ── Price History ─────────────────────────────────────────────────────────────

func seedPriceHistory(ctx context.Context, db *sqlx.DB, storeID string, prods []struct {
	ID        string  `db:"id"`
	Name      string  `db:"name"`
	SKU       string  `db:"sku"`
	CostPrice float64 `db:"cost_price"`
	SellPrice float64 `db:"sell_price"`
}, adminID string) int {
	count := 0
	// Simulate 3–5 price changes per product over the last 60 days
	maxProds := len(prods)
	if maxProds > 20 {
		maxProds = 20 // limit to first 20 per store to keep dataset manageable
	}

	for i := 0; i < maxProds; i++ {
		p := prods[i]
		numChanges := 3 + (i % 3) // 3, 4, or 5 changes

		currentCost := p.CostPrice * 0.8  // start 20% lower 60 days ago
		currentSell := p.SellPrice * 0.85 // start 15% lower

		for c := numChanges; c >= 1; c-- {
			daysBack := c * (60 / numChanges)
			newCost := currentCost * (1 + float64(5+rand.Intn(10))/100) //nolint:gosec
			newSell := currentSell * (1 + float64(3+rand.Intn(8))/100)  //nolint:gosec

			// Cap at actual current price on the last entry
			if c == 1 {
				newCost = p.CostPrice
				newSell = p.SellPrice
			}

			_, err := db.ExecContext(ctx, `
				INSERT INTO price_history (id, product_id, store_id, changed_by, old_cost, new_cost, old_sell, new_sell, source, notes, changed_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'manual', $9, $10)
			`, uuid.NewString(), p.ID, storeID, adminID,
				currentCost, newCost, currentSell, newSell,
				fmt.Sprintf("Penyesuaian harga rutin %s", p.SKU),
				tsNDaysAgo(daysBack))
			must(err)

			currentCost = newCost
			currentSell = newSell
			count++
		}
	}
	return count
}

// ── Customers ─────────────────────────────────────────────────────────────────

var firstNames = []string{
	"Budi", "Sari", "Andi", "Rini", "Dewi", "Ahmad", "Heni", "Rizky", "Putri", "Doni",
	"Fikar", "Maya", "Tono", "Lina", "Arif", "Nadia", "Hendra", "Siska", "Wahyu", "Yuni",
	"Fajar", "Rani", "Gilang", "Tika", "Bayu", "Ayu", "Dimas", "Wulan", "Eko", "Nisa",
}

var lastNames = []string{
	"Santoso", "Wijaya", "Pratama", "Rahayu", "Susanto", "Kurniawan", "Hidayat", "Purnama",
	"Setiawan", "Utama", "Lestari", "Wahyudi", "Nugroho", "Kusuma", "Saputra", "Wibowo",
}

func seedCustomers(ctx context.Context, db *sqlx.DB, storeID string) int {
	// Check existing count
	var existing int
	_ = db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM customers WHERE store_id=$1 AND deleted_at IS NULL`, storeID,
	).Scan(&existing)
	if existing >= 20 {
		return existing // already enough
	}

	count := 0
	target := 30 + rand.Intn(21) //nolint:gosec // 30–50 per store

	for i := existing; i < target; i++ {
		fn := firstNames[rand.Intn(len(firstNames))] //nolint:gosec
		ln := lastNames[rand.Intn(len(lastNames))]   //nolint:gosec
		name := fn + " " + ln
		phone := fmt.Sprintf("08%d%07d", 1+rand.Intn(9), rand.Intn(9999999)) //nolint:gosec
		email := fmt.Sprintf("%s.%s%d@email.com",
			lower(fn), lower(ln), rand.Intn(99)) //nolint:gosec

		areas := []string{
			"Jl. Sudirman No. %d, Jakarta",
			"Jl. Dago No. %d, Bandung",
			"Jl. Pahlawan No. %d, Bekasi",
			"Jl. Raya Bogor No. %d",
			"Gang Anggrek No. %d",
		}
		area := areas[rand.Intn(len(areas))]           //nolint:gosec
		address := fmt.Sprintf(area, 1+rand.Intn(200)) //nolint:gosec
		joinedDays := rand.Intn(60)                    //nolint:gosec
		createdAt := tsNDaysAgo(joinedDays)

		_, err := db.ExecContext(ctx, `
			INSERT INTO customers (id, store_id, name, phone, email, address, notes, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)
			ON CONFLICT DO NOTHING
		`, uuid.NewString(), storeID, name, phone, email, address,
			"Pelanggan reguler", createdAt)
		must(err)
		count++
	}
	return count
}

func lower(s string) string {
	result := make([]byte, len(s))
	for i := range s {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		result[i] = c
	}
	return string(result)
}

// ── Purchase Orders + Termins + Stock Batches ─────────────────────────────────

type poProduct struct {
	ID        string  `db:"id"`
	Name      string  `db:"name"`
	SKU       string  `db:"sku"`
	CostPrice float64 `db:"cost_price"`
	SellPrice float64 `db:"sell_price"`
}

//nolint:funlen,cyclop // PO seeder covers multiple payment scenarios
func seedPurchaseOrders(ctx context.Context, db *sqlx.DB, storeID string, rawProds []struct {
	ID        string  `db:"id"`
	Name      string  `db:"name"`
	SKU       string  `db:"sku"`
	CostPrice float64 `db:"cost_price"`
	SellPrice float64 `db:"sell_price"`
}, adminID string) int {
	// Convert to poProduct slice
	prods := make([]poProduct, len(rawProds))
	for i, p := range rawProds {
		prods[i] = poProduct{p.ID, p.Name, p.SKU, p.CostPrice, p.SellPrice}
	}

	// Load supplier IDs
	var supplierRows []struct {
		ID   string `db:"id"`
		Name string `db:"name"`
	}
	must(db.SelectContext(ctx, &supplierRows, `SELECT id, name FROM suppliers WHERE is_active=true ORDER BY name`))
	if len(supplierRows) == 0 {
		return 0
	}

	type poScenario struct {
		label       string
		status      string // received | ordered | canceled
		paymentMode string // paid | partial | unpaid
		daysBack    int
		numItems    int
		supplierIdx int
	}

	scenarios := []poScenario{
		{"PO-EXT-001", "received", "paid", 55, 6, 0},
		{"PO-EXT-002", "received", "partial", 48, 5, 1},
		{"PO-EXT-003", "received", "unpaid", 40, 7, 2},
		{"PO-EXT-004", "received", "paid", 32, 4, 0},
		{"PO-EXT-005", "ordered", "unpaid", 25, 5, 1},
		{"PO-EXT-006", "received", "partial", 18, 8, 3},
		{"PO-EXT-007", "received", "paid", 10, 6, 0},
		{"PO-EXT-008", "received", "unpaid", 5, 3, 2},
	}

	count := 0
	for idx, sc := range scenarios {
		// Make PO number unique per store by appending store-index and scenario index
		poNum := fmt.Sprintf("%s-%s-%02d", sc.label, storeID[:8], idx)

		// Check for existing PO with this number
		var existingID string
		_ = db.QueryRowContext(ctx, `SELECT id FROM purchase_orders WHERE po_number=$1`, poNum).Scan(&existingID)
		if existingID != "" {
			count++
			continue // already seeded
		}

		supplierID := supplierRows[sc.supplierIdx%len(supplierRows)].ID
		supplierName := supplierRows[sc.supplierIdx%len(supplierRows)].Name
		poID := uuid.NewString()
		orderedAt := tsNDaysAgo(sc.daysBack + 2)
		receivedAt := tsNDaysAgo(sc.daysBack)

		// Determine PO status string in DB
		dbStatus := sc.status // draft | ordered | received | canceled

		_, err := db.ExecContext(ctx, `
			INSERT INTO purchase_orders (id, store_id, supplier_id, po_number, status, total_amount, ordered_by, received_by, ordered_at, received_at, notes, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, 0, $6, $7, $8, $9, $10, $8, $8)
		`, poID, storeID, supplierID, poNum, dbStatus, adminID, adminID,
			orderedAt, receivedAt,
			fmt.Sprintf("Pembelian rutin dari %s", supplierName))
		must(err)

		// Select items for this PO
		numItems := sc.numItems
		if numItems > len(prods) {
			numItems = len(prods)
		}
		var totalAmt float64
		for j := 0; j < numItems; j++ {
			p := prods[(idx*3+j)%len(prods)]
			qty := float64(20 + rand.Intn(81))                           //nolint:gosec // 20–100 units
			unitCost := p.CostPrice * (0.9 + float64(rand.Intn(20))/100) //nolint:gosec
			subtotal := qty * unitCost
			totalAmt += subtotal

			_, err = db.ExecContext(ctx, `
				INSERT INTO purchase_order_items (id, po_id, product_id, quantity, unit_cost, received_qty, subtotal)
				VALUES ($1, $2, $3, $4, $5, $6, $7)
			`, uuid.NewString(), poID, p.ID, qty, unitCost, qty, subtotal)
			must(err)

			// Only create stock batches and update stock for received POs
			if sc.status == "received" {
				_, err = db.ExecContext(ctx, `
					INSERT INTO stock_batches (id, product_id, store_id, po_id, quantity_remaining, purchase_price, received_at)
					VALUES ($1, $2, $3, $4, $5, $6, $7)
				`, uuid.NewString(), p.ID, storeID, poID, qty, unitCost, receivedAt)
				must(err)

				_, err = db.ExecContext(ctx, `
					UPDATE stock_levels SET quantity = quantity + $1, updated_at = NOW()
					WHERE product_id = $2 AND store_id = $3
				`, qty, p.ID, storeID)
				must(err)

				_, err = db.ExecContext(ctx, `
					INSERT INTO stock_movements (id, product_id, store_id, ref_type, ref_id, quantity_delta, notes, created_by, created_at)
					VALUES ($1, $2, $3, 'purchase_order', $4, $5, $6, $7, $8)
				`, uuid.NewString(), p.ID, storeID, poID, qty,
					fmt.Sprintf("Terima PO %s dari %s", poNum, supplierName),
					adminID, receivedAt)
				must(err)
			}
		}

		// Update PO total
		_, err = db.ExecContext(ctx, `UPDATE purchase_orders SET total_amount=$1 WHERE id=$2`, totalAmt, poID)
		must(err)

		// ── Payment scenarios via termins ────────────────────────────────────
		switch sc.paymentMode {
		case "paid":
			// Single termin — fully paid
			terminID := createTermin(ctx, db, poID, 1, totalAmt, tsNDaysAgo(sc.daysBack-1), "paid")
			createPaymentRecord(ctx, db, terminID, totalAmt, dateOnly(sc.daysBack-1), "transfer", adminID)

		case "partial":
			// Two termins: first paid, second still unpaid
			half := roundToIDR(totalAmt / 2)
			terminID1 := createTermin(ctx, db, poID, 1, half, tsNDaysAgo(sc.daysBack-1), "paid")
			createPaymentRecord(ctx, db, terminID1, half, dateOnly(sc.daysBack-1), "transfer", adminID)
			_ = createTermin(ctx, db, poID, 2, totalAmt-half, tsNDaysAgo(sc.daysBack/2), "unpaid")

		case "unpaid":
			// Single termin — unpaid
			_ = createTermin(ctx, db, poID, 1, totalAmt, tsNDaysAgo(sc.daysBack-3), "unpaid")
		}

		count++
	}
	return count
}

func roundToIDR(v float64) float64 {
	return float64(int(v/1000)) * 1000
}

func createTermin(ctx context.Context, db *sqlx.DB, poID string, num int, amount float64, dueAt time.Time, status string) string {
	id := uuid.NewString()
	_, err := db.ExecContext(ctx, `
		INSERT INTO purchase_order_termins (id, po_id, termin_number, amount, due_date, status, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (po_id, termin_number) DO NOTHING
	`, id, poID, num, amount, dueAt.Format("2006-01-02"), status,
		fmt.Sprintf("Termin %d", num))
	must(err)
	// Re-read actual id
	_ = db.QueryRowContext(ctx, `SELECT id FROM purchase_order_termins WHERE po_id=$1 AND termin_number=$2`, poID, num).Scan(&id)
	return id
}

func createPaymentRecord(ctx context.Context, db *sqlx.DB, terminID string, amount float64, date, method, recordedBy string) {
	_, err := db.ExecContext(ctx, `
		INSERT INTO payment_records (id, termin_id, amount_paid, payment_date, payment_method, notes, recorded_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, uuid.NewString(), terminID, amount, date, method, "Pembayaran PO", recordedBy)
	must(err)
}

// ── Incomes ───────────────────────────────────────────────────────────────────

var incomeNotes = []string{
	"Penerimaan dari pelanggan korporat",
	"Pembayaran tagihan lama",
	"Penerimaan tunai kasir",
	"Transfer masuk dari rekening",
	"Modal tambahan dari owner",
	"Refund barang retur supplier",
	"Komisi referral pelanggan baru",
	"Pendapatan event spesial",
	"Penerimaan jasa perpanjangan kontrak",
	"Dana promosi dari distributor",
}

var paymentMethods = []string{"cash", "transfer", "qris", "card"}

func seedIncomes(ctx context.Context, db *sqlx.DB, storeID string, catIDs []string, adminID string) int {
	var existing int
	_ = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM incomes WHERE store_id=$1`, storeID).Scan(&existing)
	if existing >= 30 {
		return existing
	}

	count := 0
	target := 25 + rand.Intn(16) //nolint:gosec // 25–40 incomes per store

	// Amount ranges by category index (realistic Indonesian Rupiah)
	amountRanges := []struct{ min, max float64 }{
		{500_000, 5_000_000},    // Penjualan Tunai
		{1_000_000, 8_000_000},  // Pendapatan Jasa
		{200_000, 2_000_000},    // Pendapatan Lain-lain
		{5_000_000, 50_000_000}, // Injeksi Modal
		{100_000, 1_500_000},    // Pengembalian Supplier
		{500_000, 3_000_000},    // Piutang Diterima
		{10_000, 200_000},       // Bunga Bank
		{100_000, 500_000},      // Komisi Agen
		{50_000, 500_000},       // Lainnya
	}

	for i := 0; i < target; i++ {
		catIdx := rand.Intn(len(catIDs)) //nolint:gosec
		catID := catIDs[catIdx]
		dayBack := rand.Intn(61) //nolint:gosec // 0–60 days ago

		ar := amountRanges[catIdx%len(amountRanges)]
		rawAmount := ar.min + rand.Float64()*(ar.max-ar.min) //nolint:gosec
		amount := roundToIDR(rawAmount)
		if amount <= 0 {
			amount = ar.min
		}

		method := paymentMethods[rand.Intn(len(paymentMethods))] //nolint:gosec
		note := incomeNotes[rand.Intn(len(incomeNotes))]         //nolint:gosec
		ref := fmt.Sprintf("INC-%s-%04d", storeID[:6], i+1)

		_, err := db.ExecContext(ctx, `
			INSERT INTO incomes (id, store_id, category_id, amount, income_date, payment_method, reference, notes, created_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`, uuid.NewString(), storeID, catID, amount, dateOnly(dayBack), method, ref, note, adminID)
		must(err)
		count++
	}
	return count
}

// ── Expenses ──────────────────────────────────────────────────────────────────

var expenseNotes = []string{
	"Bayar sewa bulan ini",
	"Tagihan listrik bulanan",
	"Gaji staf periode ini",
	"Servis AC dan komputer",
	"Iklan Instagram & FB",
	"Penggantian peralatan rusak",
	"Pembelian ATK bulanan",
	"Ongkir pengiriman stok",
	"Biaya administrasi bank",
	"Perpanjangan izin usaha",
	"Biaya asuransi peralatan",
	"Pembayaran tagihan air",
	"Pembelian seragam karyawan",
	"Biaya pelatihan staf baru",
	"Pembelian perlengkapan kebersihan",
}

//nolint:funlen // expense seeder is intentionally varied
func seedExpenses(ctx context.Context, db *sqlx.DB, storeID string, catIDs []string, adminID string) int {
	var existing int
	_ = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM expenses WHERE store_id=$1`, storeID).Scan(&existing)
	if existing >= 30 {
		return existing
	}

	count := 0
	target := 30 + rand.Intn(21) //nolint:gosec // 30–50 per store

	// Realistic expense ranges per category (IDR)
	expAmountRanges := []struct{ min, max float64 }{
		{3_000_000, 15_000_000}, // Sewa Tempat
		{500_000, 3_000_000},    // Listrik
		{200_000, 500_000},      // Air
		{300_000, 800_000},      // Internet & Telepon
		{3_000_000, 20_000_000}, // Gaji Karyawan
		{500_000, 3_000_000},    // Bonus & Insentif
		{500_000, 5_000_000},    // Pembelian Bahan Baku
		{200_000, 2_000_000},    // Pemeliharaan Peralatan
		{100_000, 1_500_000},    // Pemasaran & Promosi
		{200_000, 1_000_000},    // Perizinan & Pajak
		{50_000, 500_000},       // Perlengkapan Kantor
		{100_000, 500_000},      // Transportasi
		{25_000, 200_000},       // Biaya Bank
		{300_000, 2_000_000},    // Asuransi
		{50_000, 500_000},       // Lainnya
	}

	for i := 0; i < target; i++ {
		catIdx := rand.Intn(len(catIDs)) //nolint:gosec
		catID := catIDs[catIdx]
		dayBack := rand.Intn(61) //nolint:gosec

		ar := expAmountRanges[catIdx%len(expAmountRanges)]
		rawAmount := ar.min + rand.Float64()*(ar.max-ar.min) //nolint:gosec
		amount := roundToIDR(rawAmount)
		if amount <= 0 {
			amount = ar.min
		}

		note := expenseNotes[rand.Intn(len(expenseNotes))] //nolint:gosec

		_, err := db.ExecContext(ctx, `
			INSERT INTO expenses (id, store_id, category_id, amount, expense_date, notes, created_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, uuid.NewString(), storeID, catID, amount, dateOnly(dayBack), note, adminID)
		must(err)
		count++
	}
	return count
}
