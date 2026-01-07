package category

import (
	"context"

	"github.com/Secure-Website-Builder/Backend/internal/models"
)

type Service struct {
	db *models.Queries
}

func NewService(db *models.Queries) *Service {
	return &Service{db: db}
}

func (s *Service) ListCategoriesByStore(ctx context.Context, storeID int64) ([]models.ListCategoriesByStoreRow, error) {
	return s.db.ListCategoriesByStore(ctx, storeID)
}

func (s *Service) ListAttributesByCategory(ctx context.Context, categoryID int64) ([]models.ListCategoryAttributesRow, error) {
	return s.db.ListCategoryAttributes(ctx, categoryID)
}
