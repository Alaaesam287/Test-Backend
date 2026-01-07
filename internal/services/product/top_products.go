package product

import (
	"context"

	"github.com/Secure-Website-Builder/Backend/internal/models"
)

func (s *Service) GetTopProductsByCategory(
	ctx context.Context,
	storeID, categoryID int64,
	limit int32,
) ([]models.ProductDTO, error) {

	rows, err := s.q.GetTopProductsByCategory(ctx, models.GetTopProductsByCategoryParams{
		StoreID:    storeID,
		CategoryID: categoryID,
		Limit:      limit,
	})
	if err != nil {
		return nil, err
	}

	products := make([]models.ProductDTO, 0, len(rows))

	for _, p := range rows {
		products = append(products, models.ProductDTO{
			ProductID:   p.ProductID,
			Name:        p.Name,
			Slug:        p.Slug,
			Brand:       p.Brand,
			Description: p.Description,
			CategoryID:  p.CategoryID,
			TotalStock:  p.ProductTotalStock,
			ItemStock:   p.ItemStock,
			Price:       p.Price,
			ImageURL:    p.PrimaryImage,
			InStock:     p.InStock,
		})
	}

	return products, nil
}
