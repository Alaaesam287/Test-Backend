package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq" // register the postgres driver

	"github.com/Secure-Website-Builder/Backend/internal/config"
	"github.com/Secure-Website-Builder/Backend/internal/models"
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

	http.HandleFunc("/owners", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "only POST allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx := r.Context()

		owner, err := queries.CreateStoreOwner(ctx, models.CreateStoreOwnerParams{
			Name:         "Aboomar Store",
			Email:        "aboomar@test.com",
			PasswordHash: "hashed-password",
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Created store owner: %+v\n", owner)
	})

	log.Printf("Server running on port %s", cfg.AppPort)
	log.Fatal(http.ListenAndServe(":"+cfg.AppPort, nil))
}
