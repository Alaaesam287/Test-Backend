package product

import (
	"context"
	"database/sql"
	"fmt"
	"mime/multipart"
	"os"
	"strings"

	"github.com/Secure-Website-Builder/Backend/internal/database"
	"github.com/Secure-Website-Builder/Backend/internal/models"
	"github.com/Secure-Website-Builder/Backend/internal/storage"
	"github.com/Secure-Website-Builder/Backend/internal/utils"
	"github.com/Secure-Website-Builder/Backend/internal/services/media"
	"github.com/google/uuid"
)

type Service struct {
	q       *models.Queries
	storage storage.ObjectStorage
	media   *media.Service
	db      *sql.DB
}

// New creates the product service; pass sqlc queries struct and raw *sql.DB
func New(q *models.Queries, db *sql.DB, storage storage.ObjectStorage, mediaService *media.Service) *Service {
	return &Service{q: q, db: db, storage: storage, media: mediaService}
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
				ImageURL:      utils.NullStringToPtr(v.PrimaryImageUrl),
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
				ImageURL:      utils.NullStringToPtr(v.PrimaryImageUrl),
				Attributes:    variantAttributes,
			}
			continue
		}

		variants = append(variants, models.VariantDTO{
			VariantID:     v.VariantID,
			SKU:           v.Sku,
			Price:         v.Price,
			StockQuantity: v.StockQuantity,
			ImageURL:      utils.NullStringToPtr(v.PrimaryImageUrl),
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
		PrimaryImage:   utils.NullStringToPtr(p.PrimaryImage),
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

// TODO: Success Response with note
// 			- Product created but image not uploaded try to upload it again
// 			- The product were already exist we increased the stock of the existing one
// 			- The product were already exist we did not change it's image if you want to change it edit the product not insert a new one

func (s *Service) CreateProduct(
	ctx context.Context,
	storeID int64,
	in models.CreateProductInput,
	image multipart.File,
	header *multipart.FileHeader,
) (*models.Product, *models.ProductVariant, error) {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback()

	qtx := s.q.WithTx(tx)

	// 2. Validate attributes in request against preset category attributes
	if err := validateVariantAttributes(ctx, qtx, in.CategoryID, in.Variant.Attributes); err != nil {
		return nil, nil, err
	}

	// 3. Find or create product
	product, productWasOutOfStock, err := findOrCreateProduct(ctx,qtx,storeID,in,)
	if err != nil {
		return nil, nil, err
	}

	// 5. Compute attribute hash
	hash := utils.HashAttributes(in.Variant.Attributes)

	// 6. Find or create variant
	finalVariant, isNewVariant, err := findOrCreateVariant(
		ctx,
		qtx,
		storeID,
		product.ProductID,
		hash,
		in.Variant,
	)
	if err != nil {
		return nil, nil, err
	}

	var uploadedKey string
	if image != nil && finalVariant.PrimaryImageUrl.Valid == false {
		key := fmt.Sprintf("stores/%d/variants/%d/%s",
			storeID,
			finalVariant.VariantID,
			uuid.NewString(),
		)

		url, _, err := s.media.UploadImage(ctx, key, image) 
		if err == nil {
		uploadedKey = key
		err = qtx.SetPrimaryVariantImage(ctx, models.SetPrimaryVariantImageParams{
			VariantID: finalVariant.VariantID,
			PrimaryImageUrl: sql.NullString{
				String: url,
				Valid:  true,
				},
			})

			if err != nil {
				// Delete uploaded image if DB insert fails
				_ = s.storage.Delete(ctx, uploadedKey)
				return nil, nil, fmt.Errorf("failed to set variant image in DB: %w", err)
			}
		}
	}

	// 7. Set default variant if product was previously not sellable
	if productWasOutOfStock && isNewVariant {
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
		if uploadedKey != "" {
			_ = s.storage.Delete(ctx, uploadedKey)
		}
		return nil, nil, err
	}

	return &product, &finalVariant, nil
}

// TODO: Success Response with note
// 			- Variant created but image not uploaded try to upload it again
// 			- The variant were already exist we increased the stock of the existing one
// 			- The variant were already exist we did not change it's image if you want to change it edit the product not insert a new one

func (s *Service) AddVariant(
	ctx context.Context,
	storeID, productID int64,
	in models.VariantInput,
	image multipart.File,
	header *multipart.FileHeader,
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

	if err := validateVariantAttributes(ctx, qtx, product.CategoryID, in.Attributes); err != nil {
		return nil, err
	}

	// 3. Compute attribute hash
	hash := utils.HashAttributes(in.Attributes)

	// 4. Find or create variant
	finalVariant, _, err := findOrCreateVariant(ctx,qtx,storeID,productID,hash,in,)
	if err != nil {
		return nil, err
	}

	var uploadedKey string
	if image != nil && finalVariant.PrimaryImageUrl.Valid == false {
		key := fmt.Sprintf("stores/%d/variants/%d/%s",
			storeID,
			finalVariant.VariantID,
			uuid.NewString(),
		)

		url, _, err := s.media.UploadImage(ctx, key, image) 
		if err == nil {
		uploadedKey = key
		err = qtx.SetPrimaryVariantImage(ctx, models.SetPrimaryVariantImageParams{
			VariantID: finalVariant.VariantID,
			PrimaryImageUrl: sql.NullString{
				String: url,
				Valid:  true,
				},
			})

			if err != nil {
				// Delete uploaded image if DB insert fails
				_ = s.storage.Delete(ctx, uploadedKey)
				return nil, fmt.Errorf("failed to set variant image in DB: %w", err)
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
		if uploadedKey != "" {
			_ = s.storage.Delete(ctx, uploadedKey)
		}
		return nil, err
	}

	return &finalVariant, nil
}
