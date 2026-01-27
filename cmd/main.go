package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/Secure-Website-Builder/Backend/internal/config"
	"github.com/Secure-Website-Builder/Backend/internal/http/handlers"
	"github.com/Secure-Website-Builder/Backend/internal/http/middleware"
	"github.com/Secure-Website-Builder/Backend/internal/http/router"
	"github.com/Secure-Website-Builder/Backend/internal/models"
	"github.com/Secure-Website-Builder/Backend/internal/services/auth"
	"github.com/Secure-Website-Builder/Backend/internal/services/cart"
	"github.com/Secure-Website-Builder/Backend/internal/services/category"
	"github.com/Secure-Website-Builder/Backend/internal/services/media"
	"github.com/Secure-Website-Builder/Backend/internal/services/product"
	"github.com/Secure-Website-Builder/Backend/internal/services/store"
	"github.com/Secure-Website-Builder/Backend/internal/storage"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment")
	}

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

	// storage
	storage, err := storage.NewMinIOStorage(
		cfg.MinIOEndpoint,
		cfg.MinIOUser,
		cfg.MinIOPass,
		cfg.MinIOBucket,
		false,
	)

	if err != nil {
		log.Fatalf("failed to initialize image storage: %v", err)
	}

	// Services
	mediaService := media.New(storage)
	categoryService := category.NewService(queries)
	productService := product.New(queries, db, storage, mediaService)
	cartService := cart.New(queries)
	storeService := store.New(db, queries, storage)
	authService := auth.New(queries, cfg.JWTSecret)

	// Middleware helpers
	storeOwnerChecker := middleware.NewStoreOwnerChecker(storeService)

	// Handlers
	categoryHandler := handlers.NewCategoryHandler(categoryService)
	productHandler := handlers.NewProductHandler(productService)
	categoryProductHandler := handlers.NewCategoryProductHandler(productService)
	cartHandler := handlers.NewCartHandler(cartService)
	authHandler := handlers.NewAuthHandler(authService)
	storeHandler := handlers.NewStoreHandler(storeService)

	// Router
	r := router.SetupRouter(
		categoryHandler,
		productHandler,
		categoryProductHandler,
		cartHandler,
		authHandler,
		storeHandler,
		storeOwnerChecker,
		cfg.JWTSecret,
	)

	port := cfg.AppPort
	if port == "" {
		port = "8080"
	}

	r.Run(":" + port)
}
