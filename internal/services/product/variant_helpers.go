package product

import (
	"context"
	"fmt"

	"github.com/Secure-Website-Builder/Backend/internal/models"
)

func insertVariantAttributes(
	ctx context.Context,
	qtx *models.Queries,
	variantID int64,
	attrs []models.VariantAttributeInput,
) error {
	for _, a := range attrs {
		if err := qtx.InsertVariantAttribute(ctx, models.InsertVariantAttributeParams{
			VariantID:   variantID,
			AttributeID: a.AttributeID,
			Value:       a.Value,
		}); err != nil {
			return err
		}
	}
	return nil
}

func findOrCreateVariant(
	ctx context.Context,
	qtx *models.Queries,
	storeID int64,
	productID int64,
	hash string,
	inputVariant models.VariantInput,
) (variant models.ProductVariant, isNew bool, err error) {

	// Try to find existing variant
	existingVariant, err := qtx.GetVariantByAttributeHash(ctx, models.GetVariantByAttributeHashParams{
		ProductID:     productID,
		AttributeHash: hash,
	})

	if err == nil {
		// Variant exists - increase stock
		if err := qtx.IncreaseVariantStock(ctx, models.IncreaseVariantStockParams{
			VariantID:     existingVariant.VariantID,
			StockQuantity: inputVariant.Stock,
		}); err != nil {
			return models.ProductVariant{}, false, err
		}

		return existingVariant, false, nil
	}

	// Variant does not exist - create new one
	newVariant, err := qtx.CreateVariant(ctx, models.CreateVariantParams{
		ProductID:     productID,
		StoreID:       storeID,
		AttributeHash: hash,
		Sku:           inputVariant.SKU,
		Price:         fmt.Sprintf("%f", inputVariant.Price),
		StockQuantity: inputVariant.Stock,
	})
	if err != nil {
		return models.ProductVariant{}, false, err
	}

	if err := insertVariantAttributes(ctx, qtx, newVariant.VariantID, inputVariant.Attributes); err != nil {
		return models.ProductVariant{}, false, err
	}

	return newVariant, true, nil
}
