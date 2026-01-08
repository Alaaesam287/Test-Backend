package product

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/Secure-Website-Builder/Backend/internal/database"
	"github.com/Secure-Website-Builder/Backend/internal/models"
	"github.com/Secure-Website-Builder/Backend/internal/utils"
)

type Service struct {
	q  *models.Queries
	db *sql.DB
}

// New creates the product service; pass sqlc queries struct and raw *sql.DB
func New(q *models.Queries, db *sql.DB) *Service {
	return &Service{q: q, db: db}
}

func (s *Service) GetFullProduct(ctx context.Context, storeID, productID int64) (*models.ProductFullDetailsDTO, error) {

	//  Base Product
	p, err := s.q.GetProductBase(ctx, models.GetProductBaseParams{
		StoreID:   storeID,
		ProductID: productID,
	})

	if err != nil {
		return nil, err
	}

	var defaultVariantDTO models.VariantDTO

	//  All Variants
	variantsRaw, err := s.q.GetProductVariants(ctx, productID)
	if err != nil {
		return nil, err
	}

	variants := make([]models.VariantDTO, 0)

	for i, v := range variantsRaw {
		// Get Attributes for each variant
		variantAttributesRows, err := s.q.GetProductVariantAttributes(ctx, v.VariantID)
		if err != nil {
			return nil, err
		}
		variantAttributes := make([]models.AttributeDTO, 0)
		for _, a := range variantAttributesRows {
			variantAttributes = append(variantAttributes, models.AttributeDTO{
				AttributeID:    a.AttributeID,
				AttributeName:  a.AttributeName,
				AttributeValue: a.AttributeValue,
			})
		}

		// Fall over scenario: no default variant set, use first variant as default
		if i == 0 && !p.DefaultVariantID.Valid {
			defaultVariantDTO = models.VariantDTO{
				VariantID:     v.VariantID,
				SKU:           v.Sku,
				Price:         v.Price,
				StockQuantity: v.StockQuantity,
				ImageURL:      v.PrimaryImageUrl,
				Attributes:    variantAttributes,
			}
			continue
		}

		// If this is the default variant save it separately
		if p.DefaultVariantID.Valid && v.VariantID == p.DefaultVariantID.Int64 {
			defaultVariantDTO = models.VariantDTO{
				VariantID:     v.VariantID,
				SKU:           v.Sku,
				Price:         v.Price,
				StockQuantity: v.StockQuantity,
				ImageURL:      v.PrimaryImageUrl,
				Attributes:    variantAttributes,
			}
			continue
		}

		variants = append(variants, models.VariantDTO{
			VariantID:     v.VariantID,
			SKU:           v.Sku,
			Price:         v.Price,
			StockQuantity: v.StockQuantity,
			ImageURL:      v.PrimaryImageUrl,
			Attributes:    variantAttributes,
		})
	}

	return &models.ProductFullDetailsDTO{
		ProductID:      p.ProductID,
		StoreID:        p.StoreID,
		ProductName:    p.Name,
		Slug:           p.Slug,
		Description:    p.Description,
		Brand:          p.Brand,
		TotalStock:     p.TotalStock,
		CategoryID:     p.CategoryID,
		CategoryName:   p.CategoryName,
		InStock:        p.InStock,
		Price:          p.Price,
		PrimaryImage:   p.PrimaryImage,
		DefaultVariant: defaultVariantDTO,
		Variants:       variants,
	}, nil
}

// ListProductFilters input shape
type ListProductFilters struct {
	Page       int
	Limit      int
	CategoryID *int64
	MinPrice   *float64
	MaxPrice   *float64
	Brand      *string
	InStock    *bool
	Attributes []database.AttributeFilter
}

// readTemplate reads the template file once
func readListProductsTemplate() (string, error) {
	// read the dedicated template file
	b, err := os.ReadFile("./internal/database/list_products_template.sql")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ResolveAttributeNameToID uses sqlc generated query: ResolveAttributeIDByName
// This wraps the generated method for convenience if needed.
func (s *Service) ResolveAttributeNameToID(ctx context.Context, storeID int64, name string) (int64, error) {
	// The sqlc function generated from queries.sql is called ResolveAttributeIDByName
	// (ensure names match your sqlc config; adjust name if sqlc generated a different function).
	id, err := s.q.ResolveAttributeIDByName(ctx, name)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// ResolveCategoryNameToID uses sqlc generated query: ResolveCategoryIDByName
// This wraps the generated method for convenience if needed.
func (s *Service) ResolveCategoryNameToID(ctx context.Context, storeID int64, name string) (int64, error) {
	id, err := s.q.ResolveCategoryIDByName(ctx, models.ResolveCategoryIDByNameParams{StoreID: storeID, Name: name})
	if err != nil {
		return 0, err
	}
	return id, nil
}

// ListProducts builds SQL from template + dynamic joins and executes it.
func (s *Service) ListProducts(ctx context.Context, storeID int64, f ListProductFilters) ([]models.ProductDTO, error) {
	// Load template
	tpl, err := readListProductsTemplate()
	if err != nil {
		return nil, fmt.Errorf("read template: %w", err)
	}

	// sane defaults
	if f.Page <= 0 {
		f.Page = 1
	}
	if f.Limit <= 0 || f.Limit > 200 {
		f.Limit = 20
	}
	offset := (f.Page - 1) * f.Limit

	args := []interface{}{storeID, f.Limit, offset, f.CategoryID, f.Brand, f.MinPrice, f.MaxPrice, f.InStock}
	paramIndex := len(args) + 1 // next placeholder index

	// attribute joins
	joinSQL, joinArgs := database.BuildAttributeFilterSQL(f.Attributes, paramIndex)
	if len(joinArgs) > 0 {
		args = append(args, joinArgs...)
		paramIndex += len(joinArgs)
	}

	// assemble SQL
	sqlFinal := strings.Replace(tpl, "/*{{DYNAMIC_JOINS}}*/", joinSQL, 1)

	// Debugging
	// fmt.Println("SQL:", sqlFinal)
	// fmt.Println("ARGS:", args)

	// Execute
	rows, err := s.db.QueryContext(ctx, sqlFinal, args...)
	if err != nil {
		return nil, fmt.Errorf("query exec: %w", err)
	}
	defer rows.Close()

	res := make([]models.ProductDTO, 0)
	for rows.Next() {
		var dto models.ProductDTO
		if err := rows.Scan(
			&dto.ProductID,
			&dto.Name,
			&dto.Slug,
			&dto.Brand,
			&dto.Description,
			&dto.CategoryID,
			&dto.TotalStock,
			&dto.ItemStock,
			&dto.Price,
			&dto.ImageURL,
			&dto.InStock,
		); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		res = append(res, dto)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}

	return res, nil
}

func (s *Service) CreateProduct(
	ctx context.Context,
	storeID int64,
	in models.CreateProductInput,
) (*models.Product, *models.ProductVariant, error) {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback()

	qtx := s.q.WithTx(tx)

	// 1. Fetch category attributes once
	categoryAttrs, err := qtx.ListCategoryAttributes(ctx, in.CategoryID)
	if err != nil {
		return nil, nil, err
	}

	allowed := make(map[int64]bool)
	required := make(map[int64]bool)

	for _, a := range categoryAttrs {
		allowed[a.AttributeID] = true
		if a.IsRequired {
			required[a.AttributeID] = true
		}
	}

	// 2. Validate input attributes
	for _, attr := range in.Variant.Attributes {
		if !allowed[attr.AttributeID] {
			return nil, nil, fmt.Errorf(
				"attribute %d is not allowed for category %d",
				attr.AttributeID, in.CategoryID,
			)
		}
		delete(required, attr.AttributeID)
	}

	if len(required) > 0 {
		return nil, nil, fmt.Errorf("missing required category attributes")
	}

	// 3. Try to find existing product
	product, err := qtx.GetProductByIdentity(ctx, models.GetProductByIdentityParams{
		StoreID:    storeID,
		Name:       in.Name,
		CategoryID: in.CategoryID,
		Brand:      sql.NullString{String: in.Brand, Valid: in.Brand != ""},
	})

	productExists := err == nil
	productWasOutOfStock := !productExists || product.StockQuantity == 0

	// 4. Create product if it does NOT exist
	if !productExists {
		product, err = qtx.CreateProduct(ctx, models.CreateProductParams{
			StoreID:     storeID,
			CategoryID:  in.CategoryID,
			Name:        in.Name,
			Slug:        sql.NullString{String: in.Slug, Valid: in.Slug != ""},
			Description: sql.NullString{String: in.Description, Valid: in.Description != ""},
			Brand:       sql.NullString{String: in.Brand, Valid: in.Brand != ""},
		})
		if err != nil {
			return nil, nil, err
		}
	}

	// 5. Compute attribute hash
	hash := utils.HashAttributes(in.Variant.Attributes)

	var finalVariant models.ProductVariant
	var createdNewVariant bool

	// 6. Variant handling
	if productExists {
		existingVariant, err := qtx.GetVariantByAttributeHash(ctx, models.GetVariantByAttributeHashParams{
			ProductID:     product.ProductID,
			AttributeHash: hash,
		})

		if err == nil {
			// Variant exists → increase stock
			err = qtx.IncreaseVariantStock(ctx, models.IncreaseVariantStockParams{
				VariantID:     existingVariant.VariantID,
				StockQuantity: in.Variant.Stock,
			})
			if err != nil {
				return nil, nil, err
			}

			finalVariant = existingVariant
		} else {
			// Create new variant
			newVariant, err := qtx.CreateVariant(ctx, models.CreateVariantParams{
				ProductID:       product.ProductID,
				StoreID:         storeID,
				AttributeHash:   hash,
				Sku:             in.Variant.SKU,
				Price:           fmt.Sprintf("%f", in.Variant.Price),
				StockQuantity:   in.Variant.Stock,
				PrimaryImageUrl: in.Variant.ImageURL,
			})
			if err != nil {
				return nil, nil, err
			}

			for _, a := range in.Variant.Attributes {
				err = qtx.InsertVariantAttribute(ctx, models.InsertVariantAttributeParams{
					VariantID:   newVariant.VariantID,
					AttributeID: a.AttributeID,
					Value:       a.Value,
				})
				if err != nil {
					return nil, nil, err
				}
			}

			finalVariant = newVariant
			createdNewVariant = true
		}
	} else {
		// New product → always create variant
		newVariant, err := qtx.CreateVariant(ctx, models.CreateVariantParams{
			ProductID:       product.ProductID,
			StoreID:         storeID,
			AttributeHash:   hash,
			Sku:             in.Variant.SKU,
			Price:           fmt.Sprintf("%f", in.Variant.Price),
			StockQuantity:   in.Variant.Stock,
			PrimaryImageUrl: in.Variant.ImageURL,
		})
		if err != nil {
			return nil, nil, err
		}

		for _, a := range in.Variant.Attributes {
			err = qtx.InsertVariantAttribute(ctx, models.InsertVariantAttributeParams{
				VariantID:   newVariant.VariantID,
				AttributeID: a.AttributeID,
				Value:       a.Value,
			})
			if err != nil {
				return nil, nil, err
			}
		}

		finalVariant = newVariant
		createdNewVariant = true
	}

	// 7. Set default variant if product was previously not sellable
	if productWasOutOfStock && createdNewVariant {
		err = qtx.SetDefaultVariant(ctx, models.SetDefaultVariantParams{
			ProductID: product.ProductID,
			DefaultVariantID: sql.NullInt64{
				Int64: finalVariant.VariantID,
				Valid: true,
			},
		})
		if err != nil {
			return nil, nil, err
		}
	}

	// 8. Update product stock
	err = qtx.UpdateProductStock(ctx, models.UpdateProductStockParams{
		ProductID:     product.ProductID,
		StockQuantity: in.Variant.Stock,
	})
	if err != nil {
		return nil, nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, err
	}

	return &product, &finalVariant, nil
}



func (s *Service) AddVariant(
	ctx context.Context,
	storeID, productID int64,
	in models.VariantInput,
) (*models.ProductVariant, error) {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	qtx := s.q.WithTx(tx)

	// 1. Lock + fetch product
	product, err := qtx.GetProductForUpdate(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("product not found")
	}

	if product.StoreID != storeID {
		return nil, fmt.Errorf("product does not belong to store")
	}

	productWasOutOfStock := product.StockQuantity == 0

	// 2. Validate attributes against category
	categoryAttrs, err := qtx.ListCategoryAttributes(ctx, product.CategoryID)
	if err != nil {
		return nil, err
	}

	allowed := make(map[int64]bool)
	required := make(map[int64]bool)

	for _, a := range categoryAttrs {
		allowed[a.AttributeID] = true
		if a.IsRequired {
			required[a.AttributeID] = true
		}
	}

	for _, attr := range in.Attributes {
		if !allowed[attr.AttributeID] {
			return nil, fmt.Errorf(
				"attribute %d is not allowed for category %d",
				attr.AttributeID, product.CategoryID,
			)
		}
		delete(required, attr.AttributeID)
	}

	if len(required) > 0 {
		return nil, fmt.Errorf("missing required category attributes")
	}

	// 3. Compute attribute hash
	hash := utils.HashAttributes(in.Attributes)

	var finalVariant models.ProductVariant

	// 4. Check for existing variant
	existing, err := qtx.GetVariantByAttributeHash(ctx, models.GetVariantByAttributeHashParams{
		ProductID:     productID,
		AttributeHash: hash,
	})

	if err == nil {
		// Variant exists → increase stock
		err = qtx.IncreaseVariantStock(ctx, models.IncreaseVariantStockParams{
			VariantID:     existing.VariantID,
			StockQuantity: in.Stock,
		})
		if err != nil {
			return nil, err
		}

		finalVariant = existing
	} else {
		// Create new variant
		finalVariant, err = qtx.CreateVariant(ctx, models.CreateVariantParams{
			ProductID:       productID,
			StoreID:         storeID,
			AttributeHash:   hash,
			Sku:             in.SKU,
			Price:           fmt.Sprintf("%f", in.Price),
			StockQuantity:   in.Stock,
			PrimaryImageUrl: in.ImageURL,
		})
		if err != nil {
			return nil, err
		}

		for _, a := range in.Attributes {
			err = qtx.InsertVariantAttribute(ctx, models.InsertVariantAttributeParams{
				VariantID:   finalVariant.VariantID,
				AttributeID: a.AttributeID,
				Value:       a.Value,
			})
			if err != nil {
				return nil, err
			}
		}
	}

	// 5. Update product stock
	err = qtx.UpdateProductStock(ctx, models.UpdateProductStockParams{
		ProductID:     productID,
		StockQuantity: in.Stock,
	})
	if err != nil {
		return nil, err
	}

	// 6. Set default variant if product was previously out of stock
	if productWasOutOfStock {
		err = qtx.SetDefaultVariant(ctx, models.SetDefaultVariantParams{
			ProductID: productID,
			DefaultVariantID: sql.NullInt64{
				Int64: finalVariant.VariantID,
				Valid: true,
			},
		})
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &finalVariant, nil
}
