package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"

	"github.com/Secure-Website-Builder/Backend/internal/config"
	"github.com/Secure-Website-Builder/Backend/internal/http/handlers"
	"github.com/Secure-Website-Builder/Backend/internal/http/router"
	"github.com/Secure-Website-Builder/Backend/internal/models"
	"github.com/Secure-Website-Builder/Backend/internal/services/category"
	"github.com/Secure-Website-Builder/Backend/internal/services/product"
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

	// sqlc generated queries
	queries := models.New(db)

	// services
	categoryService := category.NewService(queries)
	productService := product.New(queries, db)

	// handlers
	categoryHandler := handlers.NewCategoryHandler(categoryService)
	productHandler := handlers.NewProductHandler(productService)
	categoryProductHandler := handlers.NewCategoryProductHandler(productService)

	// router
	r := router.SetupRouter(categoryHandler, productHandler, categoryProductHandler)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
