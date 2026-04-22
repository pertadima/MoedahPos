package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/moedahpos/backend/internal/repository/postgres"
)

func main() {
	db, err := sqlx.Connect("postgres", "postgresql://moedah:moedahsecret@localhost:5432/moedah_pos?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	userRepo := postgres.NewUserRepo(db)
	
	// Get admin user
	var userID string
	err = db.QueryRowx("SELECT id FROM users LIMIT 1").Scan(&userID)
	if err != nil {
		log.Fatal(err)
	}

	stores, err := userRepo.FindStoresByUserID(context.Background(), userID)
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("Stores for %s:\n", userID)
	for _, s := range stores {
		fmt.Printf(" - %s (%s): %.2f points/rupiah\n", s.StoreName, s.StoreID, s.LoyaltyPointsPerRupiah)
	}
}
