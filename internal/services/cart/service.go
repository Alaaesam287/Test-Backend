package cart

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/Secure-Website-Builder/Backend/internal/models"
	"github.com/google/uuid"
)

type Service struct {
	q *models.Queries
}

func New(q *models.Queries) *Service {
	return &Service{q: q}
}

func (s *Service) GetCart(
	ctx context.Context,
	storeID int64,
	sessionID uuid.UUID,
) (*models.CartDTO, error) {

	cartRow, err := s.q.GetCartBySession(ctx, models.GetCartBySessionParams{
		SessionID: sessionID,
		StoreID:   storeID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			// Visitor has no cart yet
			return &models.CartDTO{
				StoreID: storeID,
				Items:   []models.CartItemDTO{},
				Total:   0,
			}, nil
		}
		// Real failure
		return nil, err
	}

	itemsRaw, err := s.q.GetCartItems(ctx, cartRow.CartID)
	if err != nil {
		return nil, err
	}

	items := make([]models.CartItemDTO, 0, len(itemsRaw))
	var total float64

	for _, it := range itemsRaw {
		subtotal, err := strconv.ParseFloat(it.Subtotal, 64)
		if err != nil {
			return nil, err
		}

		items = append(items, models.CartItemDTO{
			CartItemID: it.CartItemID,
			VariantID:  it.VariantID,
			ProductID:  it.ProductID,
			Product:    it.ProductName,
			SKU:        it.Sku,
			ImageURL:   it.PrimaryImageUrl,
			Price:      it.UnitPrice,
			Quantity:   it.Quantity,
			Subtotal:   it.Subtotal,
		})

		total += subtotal
	}

	return &models.CartDTO{
		CartID:    cartRow.CartID,
		StoreID:   cartRow.StoreID,
		Items:     items,
		Total:     total,
		UpdatedAt: cartRow.UpdatedAt,
	}, nil
}
