package main

import (
	"context"
	"fmt"
	"os"

	"github.com/moedahpos/backend/internal/config"
	"github.com/moedahpos/backend/internal/dto"
	"github.com/moedahpos/backend/internal/repository/postgres"
	"github.com/moedahpos/backend/pkg/db"
	"github.com/moedahpos/backend/pkg/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("config error: %v\n", err)
		os.Exit(1)
	}

	log := logger.New(cfg.App.Env)
	sqlxDB, err := db.Connect(&cfg.DB, log)
	if err != nil {
		fmt.Printf("db error: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		_ = sqlxDB.Close()
	}()

	if err := db.RunMigrations(sqlxDB, cfg.Migration.Dir, log); err != nil {
		fmt.Printf("migrate error: %v\n", err)
	}

	repo := postgres.NewProductRepo(sqlxDB)

	// Get first product
	products, _, err := repo.FindAll(context.Background(), dto.ProductListFilter{
		StoreID: "88e9fc92-d761-4b0f-bbaa-e4cd859b9581",
		PaginationQuery: dto.PaginationQuery{
			PerPage: 1,
			Page:    1,
		},
	})

	if err != nil {
		fmt.Printf("findall error: %v\n", err)
		return
	}

	if len(products) == 0 {
		fmt.Println("No products found for this store")
		return
	}

	p := products[0]
	fmt.Printf("Updating product %s (ID: %s)\n", p.Name, p.ID)

	p.UseGlobalTax = true
	var tp float64 = 0
	p.TaxPercentage = &tp

	updated, err := repo.Update(context.Background(), p)
	if err != nil {
		fmt.Printf("Update error: %v\n", err)
		return
	}

	fmt.Printf("Success! Updated product %s\n", updated.ID)
}
