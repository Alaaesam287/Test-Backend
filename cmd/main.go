package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/Secure-Website-Builder/Backend/internal/config"
	"github.com/Secure-Website-Builder/Backend/internal/database"
	"github.com/Secure-Website-Builder/Backend/internal/http/handlers"
	"github.com/Secure-Website-Builder/Backend/internal/http/middleware"
	"github.com/Secure-Website-Builder/Backend/internal/http/router"
	"github.com/Secure-Website-Builder/Backend/internal/limiter"
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

	secrets, err := config.LoadSecrets()
	if err != nil {
		log.Fatalf("failed to load config secrets: %v", err)
	}

	appConfig, err := config.LoadAppConfig(config.CONFIG_FILE_PATH)
	if err != nil {
		log.Fatalf("failed to load app config: %v", err)
	}

	// Build PostgreSQL connection string
	dbURI := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		secrets.DBUser, secrets.DBPass, secrets.DBHost, secrets.DBPort, secrets.DBName,
	)

	// Open DB connection
	dbPool, err := sql.Open("postgres", dbURI)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer dbPool.Close()

	// Ensure DB is reachable
	if err := dbPool.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	// database wrapper
	db := database.NewDB(dbPool)
	// storage
	storage, err := storage.NewMinIOStorage(
		secrets.MinIOEndpoint,
		secrets.MinIOUser,
		secrets.MinIOPass,
		secrets.MinIOBucket,
		false,
	)

	if err != nil {
		log.Fatalf("failed to initialize image storage: %v", err)
	}

	// Services
	mediaService := media.New(storage)
	categoryService := category.New(db)
	productService := product.New(db, storage, mediaService)
	cartService := cart.New(db)
	storeService := store.New(db, storage)
	authService := auth.New(db, secrets.JWTSecret)

	// Middleware helpers
	storeOwnerChecker := middleware.NewStoreOwnerChecker(storeService)
	rateLimiterManager := limiter.NewManager(
		appConfig.RateLimit.RequestsPerSecond,
		appConfig.RateLimit.Burst,
		appConfig.RateLimit.CleanupInterval(),
	)

	// Rate limiter middleware
	rateLimiter := middleware.NewRateLimiter(rateLimiterManager)

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
		rateLimiter,
		storeOwnerChecker,
		secrets.JWTSecret,
	)

	port := secrets.AppPort
	if port == "" {
		port = "8080"
	}

	r.Run(":" + port)
}
