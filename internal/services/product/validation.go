package product

import (
	"context"
	"fmt"

	"github.com/Secure-Website-Builder/Backend/internal/models"
)

func validateVariantAttributes(
	ctx context.Context,
	qtx *models.Queries,
	categoryID int64,
	attrs []models.VariantAttributeInput,
) error {
	categoryAttrs, err := qtx.ListCategoryAttributes(ctx, categoryID)
	if err != nil {
		return err
	}

	allowed := make(map[int64]bool)
	required := make(map[int64]bool)

	for _, a := range categoryAttrs {
		allowed[a.AttributeID] = true
		if a.IsRequired {
			required[a.AttributeID] = true
		}
	}

	for _, attr := range attrs {
		if !allowed[attr.AttributeID] {
			return fmt.Errorf(
				"attribute %d is not allowed for category %d",
				attr.AttributeID,
				categoryID,
			)
		}
		delete(required, attr.AttributeID)
	}

	if len(required) > 0 {
		return fmt.Errorf("missing required category attributes")
	}

	return nil
}
