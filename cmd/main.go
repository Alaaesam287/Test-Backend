package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq" // register the postgres driver

	"github.com/Secure-Website-Builder/Backend/internal/config"
	"github.com/Secure-Website-Builder/Backend/internal/http/handlers"
	"github.com/Secure-Website-Builder/Backend/internal/http/router"
	"github.com/Secure-Website-Builder/Backend/internal/models"
	"github.com/Secure-Website-Builder/Backend/internal/services/category"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Build PostgreSQL connection string
	dbURI := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName,
	)

	// Open DB connection
	db, err := sql.Open("postgres", dbURI)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Ensure DB is reachable
	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	// sqlc generated queries struct
	queries := models.New(db)

	categoryService := category.NewService(queries)
	categoryHandler := handlers.NewCategoryHandler(categoryService)
	

	r := router.SetupRouter(categoryHandler)
	r.Run(":8080")
}
