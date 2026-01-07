package store

import (
	"context"

	"github.com/Secure-Website-Builder/Backend/internal/models"
)

type Service struct {
	q *models.Queries
}

func New(q *models.Queries) *Service {
	return &Service{q: q}
}

// IsOwner checks if user is owner of the store
func (s *Service) IsOwner(ctx context.Context, userID, storeID int64) (bool, error) {
	return s.q.IsStoreOwner(ctx, models.IsStoreOwnerParams{
		StoreOwnerID: userID,
		StoreID:      storeID,
	})
}
