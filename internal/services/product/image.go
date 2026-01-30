package product

import (
	"context"
	"database/sql"
	"fmt"
	"mime/multipart"

	"github.com/Secure-Website-Builder/Backend/internal/models"
)

func (s *Service) UploadVariantImage(
	ctx context.Context,
	storeID, productID, variantID int64,
	file multipart.File,
	isPrimary bool,
) (string, error) {

	// Verify ownership
	variant, err := s.db.Queries.GetVariant(ctx, variantID)
	if err != nil {
		return "", fmt.Errorf("variant not found")
	}
	if variant.StoreID != storeID || variant.ProductID != productID {
		return "", fmt.Errorf("not your variant")
	}
	
	// Generate S3 key
	key := generateImageUploadKey(storeID, variantID)

	// Upload image using media service
	url, _, err := s.media.UploadImage(ctx, key, file)
	if err != nil {
		return "", fmt.Errorf("failed to upload image: %w", err)
	}

	err = s.db.RunInTx(ctx, func(qtx *models.Queries) error {

		// Lock the variant for update
		_, err := qtx.GetVariantForUpdate(ctx, variantID)
		if err != nil {
			return err
		}

		if isPrimary {
			err = qtx.SetPrimaryVariantImage(ctx, models.SetPrimaryVariantImageParams{
				VariantID: variantID,
				PrimaryImageUrl: sql.NullString{
					String: url,
					Valid:  true,
				},
			})
		} else {
			_, err = qtx.InsertVariantImage(ctx, models.InsertVariantImageParams{
				ProductVariantID: variantID,
				ImageUrl:         url,
			})
		}

		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		// TODO: Later we can use outbox pattern to handler the failure of Deleting image 
		// right now we assume that delete always succeeds
		s.storage.Delete(ctx, key)
		return "", err
	}

	return url, nil
}
