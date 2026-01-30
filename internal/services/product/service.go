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
)

type Service struct {
	storage storage.ObjectStorage
	media   *media.Service
	db      *database.DB
}

// New creates the product service; pass sqlc queries struct and raw *sql.DB
func New(db *database.DB, storage storage.ObjectStorage, mediaService *media.Service) *Service {
	return &Service{db: db, storage: storage, media: mediaService}
}

func (s *Service) GetFullProduct(ctx context.Context, storeID, productID int64) (*models.ProductFullDetailsDTO, error) {

	//  Base Product
	p, err := s.db.Queries.GetProductBase(ctx, models.GetProductBaseParams{
		StoreID:   storeID,
		ProductID: productID,
	})

	if err != nil {
		return nil, err
	}

	var defaultVariantDTO models.VariantDTO

	//  All Variants
	variantsRaw, err := s.db.Queries.GetProductVariants(ctx, productID)
	if err != nil {
		return nil, err
	}

	variants := make([]models.VariantDTO, 0)

	for i, v := range variantsRaw {
		// Get Attributes for each variant
		variantAttributesRows, err := s.db.Queries.GetProductVariantAttributes(ctx, v.VariantID)
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
	id, err := s.db.Queries.ResolveAttributeIDByName(ctx, name)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// ResolveCategoryNameToID uses sqlc generated query: ResolveCategoryIDByName
// This wraps the generated method for convenience if needed.
func (s *Service) ResolveCategoryNameToID(ctx context.Context, storeID int64, name string) (int64, error) {
	id, err := s.db.Queries.ResolveCategoryIDByName(ctx, models.ResolveCategoryIDByNameParams{StoreID: storeID, Name: name})
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

// CreateProduct creates or updates a product and its variant inside a single database transaction.
//
// The transaction guarantees atomicity for all product- and variant-related database changes,
// including attribute validation, product creation or lookup, variant creation or lookup,
// default variant assignment, and stock updates.
//
// Image upload is intentionally executed AFTER the transaction commits.
// This follows a Saga-style approach with soft failure handling:
//   - If image upload fails, the product and variant remain successfully created.
//   - If setting the primary image in the database fails after a successful upload,
//     the uploaded image is deleted to avoid orphaned media.
//   - The client may retry the image upload later via a separate endpoint.
//
// This design avoids long-running transactions and external side effects
// (such as network or storage operations) inside the database transaction.
func (s *Service) CreateProduct(
	ctx context.Context,
	storeID int64,
	in models.CreateProductInput,
	image multipart.File,
) (*models.Product, *models.ProductVariant, error) {

	var (
		product      models.Product
		finalVariant models.ProductVariant
		err          error
	)

	// Start transaction
	err = s.db.RunInTx(ctx, func(qtx *models.Queries) error {

		var productWasOutOfStock bool
		

		// Validate attributes in request against preset category attributes
		if err := validateVariantAttributes(ctx, qtx, in.CategoryID, in.Variant.Attributes); err != nil {
			return err
		}

		// Find or create product
		product, productWasOutOfStock, err = findOrCreateProduct(ctx, qtx, storeID, in)
		if err != nil {
			return err
		}

		// Compute attribute hash
		hash := utils.HashAttributes(in.Variant.Attributes)

		// Find or create variant
		finalVariant, err = findOrCreateVariant(
			ctx,
			qtx,
			storeID,
			product.ProductID,
			hash,
			in.Variant,
		)
		if err != nil {
			return err
		}

		// Set default variant if product was previously not sellable
		// either new product or existing product out of stock
		if productWasOutOfStock {
			err = qtx.SetDefaultVariant(ctx, models.SetDefaultVariantParams{
				ProductID: product.ProductID,
				DefaultVariantID: sql.NullInt64{
					Int64: finalVariant.VariantID,
					Valid: true,
				},
			})
			if err != nil {
				return err
			}
		}

		// Update product stock
		err = qtx.UpdateProductStock(ctx, models.UpdateProductStockParams{
			ProductID:     product.ProductID,
			StockQuantity: in.Variant.Stock,
		})
		if err != nil {
			return err
		}
			return nil
	})

	if err != nil {
		return nil, nil, err
	}

	// Upload image if provided and variant has no image yet
	// either new variant or existing variant without image
	if image != nil && finalVariant.PrimaryImageUrl.Valid == false {
		key := generateImageUploadKey(storeID, finalVariant.VariantID)
		url, _, err := s.media.UploadImage(ctx, key, image) 
		// if the upload image fails we do not rollback the whole transaction as the product and variant were created/updated successfully
		// we just skip setting the image and return success to the user
		// the user can try to upload the image again later
		// so we silently skip handling err != nil later
		// TODO: Log that the image was not uploaded on err != nil
		if err == nil {
			err = s.db.Queries.SetPrimaryVariantImage(ctx, models.SetPrimaryVariantImageParams{
				VariantID: finalVariant.VariantID,
				PrimaryImageUrl: sql.NullString{
					String: url,
					Valid:  true,
					},
			})
			
			// we will not return error here also, just delete the uploaded image
			if err != nil {
					// TODO: Later we can use outbox pattern to handler the failure of Deleting image 
					// right now we assume that delete always succeeds
					_ = s.storage.Delete(ctx, key)
			}else {
				finalVariant.PrimaryImageUrl = sql.NullString{
					String: url,
					Valid:  true,
				}
			}
		}
	}

	return &product, &finalVariant, nil
}


// AddVariant creates or updates a product variant inside a single database transaction.
//
// The transaction guarantees atomicity for all variant-related database changes,
// including product locking and ownership validation, attribute validation,
// variant creation or lookup, default variant assignment, and product stock updates.
//
// Image upload is intentionally executed AFTER the transaction commits.
// This follows a Saga-style approach with soft failure handling:
//   - If image upload fails, the variant remains successfully created or updated.
//   - If setting the primary image in the database fails after a successful upload,
//     the uploaded image is deleted to avoid orphaned media.
//   - The client may retry the image upload later via a separate endpoint.
//
// This design avoids long-running transactions and external side effects
// (such as network or storage operations) inside the database transaction.
func (s *Service) AddVariant(
	ctx context.Context,
	storeID, productID int64,
	in models.VariantInput,
	image multipart.File,
) (*models.ProductVariant, error) {

	var finalVariant  models.ProductVariant
	
	err := s.db.RunInTx(ctx, func(qtx *models.Queries) error {
		// Lock + fetch product
		product, err := qtx.GetProductForUpdate(ctx, productID)
		if err != nil {
			return fmt.Errorf("product not found")
		}

		if product.StoreID != storeID {
			return fmt.Errorf("product does not belong to store")
		}

		productWasOutOfStock := product.StockQuantity == 0

		if err := validateVariantAttributes(ctx, qtx, product.CategoryID, in.Attributes); err != nil {
			return err
		}

		// Compute attribute hash
		hash := utils.HashAttributes(in.Attributes)

		// Find or create variant
		finalVariant, err = findOrCreateVariant(ctx, qtx, storeID, productID, hash, in)
		if err != nil {
			return err
		}

		// Update product stock
		err = qtx.UpdateProductStock(ctx, models.UpdateProductStockParams{
			ProductID:     productID,
			StockQuantity: in.Stock,
		})
		if err != nil {
			return err
		}

		// Set default variant if product was previously out of stock
		if productWasOutOfStock {
			err = qtx.SetDefaultVariant(ctx, models.SetDefaultVariantParams{
				ProductID: productID,
				DefaultVariantID: sql.NullInt64{
					Int64: finalVariant.VariantID,
					Valid: true,
				},
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Upload image if provided and variant has no image yet
	// either new variant or existing variant without image
	if image != nil && finalVariant.PrimaryImageUrl.Valid == false {
		key := generateImageUploadKey(storeID, finalVariant.VariantID)
		url, _, err := s.media.UploadImage(ctx, key, image) 
		// if the upload image fails we do not rollback the whole transaction as the product and variant were created/updated successfully
		// we just skip setting the image and return success to the user
		// the user can try to upload the image again later
		// so we silently skip handling err != nil later
		// TODO: Log that the image was not uploaded on err != nil
		if err == nil {
			err = s.db.Queries.SetPrimaryVariantImage(ctx, models.SetPrimaryVariantImageParams{
				VariantID: finalVariant.VariantID,
				PrimaryImageUrl: sql.NullString{
					String: url,
					Valid:  true,
					},
			})
			
			// we will not return error here also, just delete the uploaded image
			if err != nil {
					// TODO: Later we can use outbox pattern to handler the failure of Deleting image 
					// right now we assume that delete always succeeds
					_ = s.storage.Delete(ctx, key)
			}else {
				finalVariant.PrimaryImageUrl = sql.NullString{
					String: url,
					Valid:  true,
				}
			}
		}
	}

	return &finalVariant, nil
}
