package database

import (
	"fmt"
	"strings"

	"github.com/Secure-Website-Builder/Backend/internal/utils"
)

// AttributeFilter: attribute_id and desired value
type AttributeFilter struct {
	AttributeID int64
	Values       []string
}

// BuildAttributeFilterSQL builds multi-JOINs for each attribute filter
// startIndex is the first placeholder index for these joins (e.g. 4)
func BuildAttributeFilterSQL(filters []AttributeFilter, startIndex int) (string, []interface{}) {
	if len(filters) == 0 {
		return "", nil
	}

	var sb strings.Builder
	args := make([]interface{}, 0)
	paramIndex := startIndex

	for i, f := range filters {
		alias := fmt.Sprintf("pav%d", i)

		// attribute_id placeholder
		attrParam := paramIndex
		paramIndex++

		// value placeholders
		placeholders := make([]string, 0, len(f.Values))
		for range f.Values {
			placeholders = append(placeholders, fmt.Sprintf("$%d", paramIndex))
			paramIndex++
		}

		sb.WriteString(fmt.Sprintf(`
JOIN product_attribute_value %s
  ON %s.product_id = p.product_id
 AND %s.attribute_id = $%d
 AND %s.value IN (%s)
`, alias, alias, alias, attrParam, alias, strings.Join(placeholders, ", ")))

		args = append(args, f.AttributeID)
		args = append(args, utils.InterfaceSlice(f.Values)...)
	}

	return sb.String(), args
}

// BuildCategoryFilterSQL returns WHERE fragment and args, using next placeholder index
func BuildCategoryFilterSQL(categoryID int64, startIndex int) (string, []interface{}) {
	if categoryID <= 0 {
		return "", nil
	}
	return fmt.Sprintf(" AND p.category_id = $%d ", startIndex), []interface{}{categoryID}
}

// BuildPriceFilterSQL returns WHERE fragment and args (pv.price)
func BuildPriceFilterSQL(minPrice *float64, maxPrice *float64, startIndex int) (string, []interface{}) {
	if minPrice == nil && maxPrice == nil {
		return "", nil
	}
	parts := make([]string, 0)
	args := make([]interface{}, 0)
	idx := startIndex
	if minPrice != nil {
		parts = append(parts, fmt.Sprintf(" pv.price >= $%d ", idx))
		args = append(args, *minPrice)
		idx++
	}
	if maxPrice != nil {
		parts = append(parts, fmt.Sprintf(" pv.price <= $%d ", idx))
		args = append(args, *maxPrice)
		idx++
	}
	return " AND (" + strings.Join(parts, " AND ") + ") ", args
}

// BuildBrandFilterSQL returns WHERE fragment and args
func BuildBrandFilterSQL(brand *string, startIndex int) (string, []interface{}) {
	if brand == nil || *brand == "" {
		return "", nil
	}
	return fmt.Sprintf(" AND p.brand = $%d ", startIndex), []interface{}{*brand}
}
