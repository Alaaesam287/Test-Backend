package product

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/Secure-Website-Builder/Backend/internal/database"
	"github.com/Secure-Website-Builder/Backend/internal/models"
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

	//  Attributes
	attrsRaw, _ := s.q.GetProductAttributes(ctx, productID)
	attrs := make([]models.AttributeDTO, 0)

	for _, a := range attrsRaw {
		attrs = append(attrs, models.AttributeDTO{
			AttributeID: a.AttributeID,
			Name:        a.Name,
			Value:       a.Value,
		})
	}

	//  Variants
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
			ImageURL:      v.PrimaryImageUrl,
			Options:       opts,
		})
	}

	return &models.ProductFullDetailsDTO{
		ProductID:        p.ProductID,
		StoreID:          p.StoreID,
		Name:             p.Name,
		Slug:             p.Slug,
		Description:      p.Description,
		Brand:            p.Brand,
		CategoryID:       p.CategoryID,
		InStock:          p.InStock,
		Price:            p.Price,
		PrimaryImage:     p.PrimaryImage,
		Attributes:       attrs,
		Variants:         variants,
		DefaultVariantID: p.DefaultVariantID,
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
	id, err := s.q.ResolveAttributeIDByName(ctx, models.ResolveAttributeIDByNameParams{StoreID: storeID, Name: name})
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

	args := []interface{}{storeID, f.Limit, offset, f.CategoryID, f.Brand, f.MinPrice, f.MaxPrice}
	paramIndex := 8 // next placeholder index

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
