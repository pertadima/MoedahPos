package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/moedahpos/backend/internal/repository/postgres"
	"github.com/moedahpos/backend/internal/service"
)

func main() {
	db, err := sqlx.Connect("postgres", "postgresql://moedah:moedahsecret@localhost:5432/moedah_pos?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	userRepo := postgres.NewUserRepo(db)
	
	var userID string
	err = db.QueryRowx("SELECT id FROM users LIMIT 1").Scan(&userID)
	if err != nil {
		log.Fatal(err)
	}

	logger := zerolog.Nop()
	authSvc := service.NewAuthService(userRepo, nil, nil, nil, 10, logger)
	resp, err := authSvc.Me(context.Background(), userID)
	if err != nil {
		log.Fatal(err)
	}
	
	b, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Println(string(b))
}
