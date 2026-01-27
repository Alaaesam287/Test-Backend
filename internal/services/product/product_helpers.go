package product

import (
	"context"
	"database/sql"

	"github.com/Secure-Website-Builder/Backend/internal/models"
)

func findOrCreateProduct(
	ctx context.Context,
	qtx *models.Queries,
	storeID int64,
	in models.CreateProductInput,
) (product models.Product, wasOutOfStock bool, err error) {

	product, err = qtx.GetProductByIdentity(ctx, models.GetProductByIdentityParams{
		StoreID:    storeID,
		Name:       in.Name,
		CategoryID: in.CategoryID,
		Brand:      sql.NullString{String: in.Brand, Valid: in.Brand != ""},
	})

	if err == nil {
		// Product exists
		wasOutOfStock = product.StockQuantity == 0
		return
	}

	// Product does not exist - create it
	product, err = qtx.CreateProduct(ctx, models.CreateProductParams{
		StoreID:     storeID,
		CategoryID:  in.CategoryID,
		Name:        in.Name,
		Slug:        sql.NullString{String: in.Slug, Valid: in.Slug != ""},
		Description: sql.NullString{String: in.Description, Valid: in.Description != ""},
		Brand:       sql.NullString{String: in.Brand, Valid: in.Brand != ""},
	})
	if err != nil {
		return
	}

	wasOutOfStock = true 
	return
}
