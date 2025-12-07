package product

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

func (s *Service) GetFullProduct(ctx context.Context, storeID, productID int64) (*models.ProductDTO, error) {

	// 1️⃣ Base Product
	p, err := s.q.GetProductBase(ctx, models.GetProductBaseParams{
		StoreID:  storeID,
		ProductID: productID,
	})
	if err != nil {
		return nil, err
	}

	// 2️⃣ Attributes
	attrsRaw, _ := s.q.GetProductAttributes(ctx, productID)
	attrs := make([]models.AttributeDTO, 0)

	for _, a := range attrsRaw {
		var val any
		switch a.DataType {
		case "string":
			val = a.ValueText
		case "decimal", "integer":
			val = a.ValueNumber
		case "boolean":
			val = a.ValueBoolean
		}

		attrs = append(attrs, models.AttributeDTO{
			AttributeID: a.AttributeID,
			Name:        a.Name,
			DataType:    a.DataType,
			Value:       val,
		})
	}

	// 3️⃣ Variants
	variantsRaw, _ := s.q.GetProductVariants(ctx, productID)
	variants := make([]models.VariantDTO, 0)

	for _, v := range variantsRaw {

		optsRaw, _ := s.q.GetVariantOptions(ctx, v.VariantID)
		opts := make([]models.VariantOptionDTO, 0)

		for _, o := range optsRaw {
			opts = append(opts, models.VariantOptionDTO{
				Type:  o.OptionType,
				Value: o.OptionValue,
			})
		}

		variants = append(variants, models.VariantDTO{
			VariantID:     v.VariantID,
			SKU:           v.Sku,
			Price:         v.Price,
			StockQuantity: v.StockQuantity,
			ImageURL:      v.ImageUrl,
			Options:       opts,
		})
	}

	return &models.ProductDTO{
		ProductID:        p.ProductID,
		StoreID:          p.StoreID,
		Name:             p.Name,
		Slug:             p.Slug,
		Description:      p.Description,
		Brand:            p.Brand,
		CategoryID:       p.CategoryID,
		InStock:          p.InStock,
		PrimaryImage:     p.PrimaryImage,
		Attributes:       attrs,
		Variants:         variants,
		DefaultVariantID: p.DefaultVariantID,
	}, nil
}
