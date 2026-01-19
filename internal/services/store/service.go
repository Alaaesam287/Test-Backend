package store

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Secure-Website-Builder/Backend/internal/models"
	"github.com/Secure-Website-Builder/Backend/internal/storage"
)

type Service struct {
	db   *sql.DB
	q    *models.Queries
	site *storage.MinIOStorage
}

func New(db *sql.DB, q *models.Queries, site *storage.MinIOStorage) *Service {
	return &Service{
		db:   db,
		q:   q,
		site: site,
	}
}

// IsOwner checks if user is owner of the store
func (s *Service) IsOwner(ctx context.Context, userID, storeID int64) (bool, error) {
	return s.q.IsStoreOwner(ctx, models.IsStoreOwnerParams{
		StoreOwnerID: userID,
		StoreID:      storeID,
	})
}


func (s *Service) CreateStore(
	ctx context.Context,
	storeOwnerID int64,
	name string,
	domain string,
	currency string,
	timezone string,
	siteConfig json.RawMessage,
) (int64, error) {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	qtx := s.q.WithTx(tx)

	if currency == "" {
		currency = "EGP"
	}

	if timezone == "" {
		timezone = "UTC"
	}

	store, err := qtx.CreateStore(ctx, models.CreateStoreParams{
		StoreOwnerID: storeOwnerID,
		Name:         name,
		Domain:       sql.NullString{String: domain, Valid: domain != ""},
		Currency:     sql.NullString{String: currency, Valid: true},
		Timezone:     sql.NullString{String: timezone, Valid: true},
	})
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	
	key := fmt.Sprintf("stores/%d/site.json", store.StoreID)

	_, err = s.site.Upload(
		ctx,
		key,
		bytes.NewReader(siteConfig),
		int64(len(siteConfig)),
		"application/json",
	)

	if err != nil {
		_ = s.q.DeleteStore(ctx, store.StoreID)
		return 0, errors.New("failed to upload site config")
	}

	return store.StoreID, nil
}
