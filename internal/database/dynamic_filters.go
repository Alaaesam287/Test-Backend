package database

import (
	"fmt"
	"strings"

	"github.com/Secure-Website-Builder/Backend/internal/utils"
)

// AttributeFilter: attribute_id and desired value
type AttributeFilter struct {
	AttributeID int64
	Values      []string
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
		alias := fmt.Sprintf("vav%d", i)

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
JOIN variant_attribute_value %s
  ON %s.variant_id = p.default_variant_id
 AND %s.attribute_id = $%d
 AND %s.value IN (%s)
`, alias, alias, alias, attrParam, alias, strings.Join(placeholders, ", ")))

		args = append(args, f.AttributeID)
		args = append(args, utils.InterfaceSlice(f.Values)...)
	}

	return sb.String(), args
}
