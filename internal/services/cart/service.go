package cart

import (
	"context"
	"database/sql"

	"github.com/Secure-Website-Builder/Backend/internal/database"
	"github.com/Secure-Website-Builder/Backend/internal/errorx"
	"github.com/Secure-Website-Builder/Backend/internal/models"
	"github.com/Secure-Website-Builder/Backend/internal/utils"
	"github.com/google/uuid"
)

type Service struct {
	db *database.DB
}

func New(db *database.DB) *Service {
	return &Service{db: db}
}

func (s *Service) GetCart(
	ctx context.Context,
	storeID int64,
	sessionID uuid.UUID,
) (*models.CartDTO, error) {

	cartRow, err := s.db.Queries.GetCartBySession(ctx, models.GetCartBySessionParams{
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

	itemsRaw, err := s.db.Queries.GetCartItems(ctx, cartRow.CartID)
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

	total, err := s.db.Queries.GetCartTotal(ctx, cartRow.CartID)
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


	return s.db.RunInTx(ctx, func(qtx *models.Queries) error {

		if qty <= 0 {
			return errorx.ErrInvalidQuantity
		}

		// Validate session
		session, err := qtx.GetSession(ctx, models.GetSessionParams{
			SessionID: sessionID,
			StoreID:   storeID,
		})
		
		if err != nil || !session.CustomerID.Valid{
			return errorx.ErrInvalidSession
		}

		// Lock or create cart
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

		// Validate variant
		variant, err := qtx.GetVariantForCart(ctx, models.GetVariantForCartParams{
			VariantID: variantID,
			StoreID:   storeID,
		})
		if err != nil {
			return errorx.ErrInvalidVariant
		}

		if variant.StockQuantity < qty {
			return errorx.ErrInsufficientStock
		}

		// Upsert item
		if err = qtx.UpsertCartItem(ctx, models.UpsertCartItemParams{
			CartID:    cart.CartID,
			VariantID: variant.VariantID,
			Quantity:  qty,
			UnitPrice: variant.Price,
		}); err != nil {
			return err
		}

		// Touch cart
		if err = qtx.TouchCart(ctx, cart.CartID); err != nil {
			return err
		}
	return nil
	})
}

func (s *Service) Checkout(
	ctx context.Context,
	storeID int64,
	sessionID uuid.UUID,
	paymentMethod string,
) error {

	return s.db.RunInTx(ctx, func(qtx *models.Queries) error {

		// Validate session
		session, err := qtx.GetSession(ctx, models.GetSessionParams{
			SessionID: sessionID,
			StoreID:   storeID,
		})
		if err != nil {
			return err
		}

		// Lock cart
		cart, err := qtx.GetCartForSession(ctx, models.GetCartForSessionParams{
			StoreID:   storeID,
			SessionID: sessionID,
		})
		if err == sql.ErrNoRows {
			return errorx.ErrCartNotFound
		}
		if err != nil {
			return err
		}

		// Lock cart items + variants
		items, err := qtx.GetCartItemsForUpdate(ctx, cart.CartID)
		if err != nil {
			return err
		}
		if len(items) == 0 {
			return errorx.ErrCartEmpty
		}

		// Validate stock
		for _, item := range items {
			if item.AvailableStock < item.CartQuantity {
				return errorx.ErrOutOfStock
			}
		}

		// Calculate total IN SQL
		total, err := qtx.GetCartTotal(ctx, cart.CartID)
		if err != nil {
			return err
		}

		// Create order with status 'pending'
		order, err := qtx.CreateOrder(ctx, models.CreateOrderParams{
			StoreID:    storeID,
			CustomerID: session.CustomerID,
			SessionID:  sessionID,
			TotalAmount: total,
		})
		if err != nil {
			return err
		}

		// Create order items
		for _, item := range items {

			if err := qtx.CreateOrderItem(ctx, models.CreateOrderItemParams{
				OrderID:   order.OrderID,
				VariantID: item.VariantID,
				Quantity:  item.CartQuantity,
				UnitPrice: item.UnitPrice,
				Subtotal:  item.Subtotal,
			}); err != nil {
				return err
			}
		}

		// Create payment (simulated success)
		// TODO: Integrate with real payment gateway
		if err := qtx.CreatePayment(ctx, models.CreatePaymentParams{
			OrderID: order.OrderID,
			Method:  paymentMethod,
			Amount:  total,
			Status:  "completed",
			TransactionRef: sql.NullString{
				String: uuid.NewString(), // Simulated transaction reference for third-party payment gateway
				Valid:  true,
			},
		}); err != nil {
			return err
		}

		// Deduct stock
		for _, item := range items {
			if err := qtx.DecreaseVariantStock(ctx, models.DecreaseVariantStockParams{
				VariantID:  item.VariantID,
				CartQuantity: item.CartQuantity,
			}); err != nil {
				return err
			}
		}

		// Mark order as completed
		if err := qtx.UpdateOrderStatus(ctx, models.UpdateOrderStatusParams{
			OrderID: order.OrderID,
			Status:  sql.NullString{String: "completed", Valid: true},
		}); err != nil {
			return err
		}

		// Clear cart
		if err := qtx.ClearCartItems(ctx, cart.CartID); err != nil {
			return err
		}

		return nil
	})
}
