package cart

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Secure-Website-Builder/Backend/internal/models"
	"github.com/Secure-Website-Builder/Backend/internal/utils"
	"github.com/google/uuid"
)

type Service struct {
	q *models.Queries
	db *sql.DB
}

func New(q *models.Queries, db *sql.DB) *Service {
	return &Service{q: q, db: db}
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
			return &models.CartDTO{
				StoreID: storeID,
				Items:   []models.CartItemDTO{},
				Total:   "0",
			}, nil
		}
		return nil, err
	}

	itemsRaw, err := s.q.GetCartItems(ctx, cartRow.CartID)
	if err != nil {
		return nil, err
	}

	items := make([]models.CartItemDTO, 0, len(itemsRaw))
	for _, it := range itemsRaw {
		items = append(items, models.CartItemDTO{
			CartItemID: it.CartItemID,
			VariantID:  it.VariantID,
			ProductID:  it.ProductID,
			Product:    it.ProductName,
			SKU:        it.Sku,
			ImageURL:   utils.NullStringToPtr(it.PrimaryImageUrl),
			Price:      it.UnitPrice,
			Quantity:   it.Quantity,
			Subtotal:   it.Subtotal,
		})
	}

	total, err := s.q.GetCartTotal(ctx, cartRow.CartID)
	if err != nil {
		return nil, err
	}

	return &models.CartDTO{
		CartID:    cartRow.CartID,
		StoreID:   cartRow.StoreID,
		Items:     items,
		Total:     total,
		UpdatedAt: cartRow.UpdatedAt,
	}, nil
}


func (s *Service) AddItem(
	ctx context.Context,
	storeID int64,
	sessionID uuid.UUID,
	variantID int64,
	qty int32,
) (err error) {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	qtx := s.q.WithTx(tx)

	// 1. Validate session
	session, err := qtx.GetSession(ctx, models.GetSessionParams{
		SessionID: sessionID,
		StoreID:   storeID,
	})
	if err != nil {
		return fmt.Errorf("invalid session: %w", err)
	}

	// 2. Lock or create cart
	cart, err := qtx.GetCartForSession(ctx, models.GetCartForSessionParams{
		StoreID:   storeID,
		SessionID: sessionID,
	})

	if err == sql.ErrNoRows {
		cart, err = qtx.CreateCart(ctx, models.CreateCartParams{
			StoreID:    storeID,
			SessionID:  sessionID,
			CustomerID: session.CustomerID,
		})
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	// 3. Validate variant
	variant, err := qtx.GetVariantForCart(ctx, models.GetVariantForCartParams{
		VariantID: variantID,
		StoreID:   storeID,
	})
	if err != nil {
		return fmt.Errorf("invalid variant: %w", err)
	}

	if variant.StockQuantity < qty {
		return fmt.Errorf("insufficient stock for variant %d", variantID)
	}

	// 4. Upsert item
	if err = qtx.UpsertCartItem(ctx, models.UpsertCartItemParams{
		CartID:    cart.CartID,
		VariantID: variant.VariantID,
		Quantity:  qty,
		UnitPrice: variant.Price,
	}); err != nil {
		return err
	}

	// 5. Touch cart
	if err = qtx.TouchCart(ctx, cart.CartID); err != nil {
		return err
	}

	// 6. Commit
	return tx.Commit()
}
