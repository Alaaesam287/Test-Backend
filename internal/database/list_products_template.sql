-- ListProductsBase template: the service will replace the placeholders.
SELECT
  p.product_id,
  p.name,
  p.slug,
  p.brand,
  pv.price,
  pv.primary_image_url,
  p.in_stock
FROM product p
JOIN product_variant pv
  ON pv.variant_id = p.default_variant_id

/*{{DYNAMIC_JOINS}}*/

WHERE p.store_id = $1
  AND p.deleted_at IS NULL
  AND ($4::BIGINT IS NULL OR p.category_id = $4)
  AND ($5::TEXT IS NULL OR p.brand = $5)
  AND ($6::DECIMAL IS NULL OR v.price >= $6)
  AND ($7::DECIMAL IS NULL OR v.price <= $7)
ORDER BY p.created_at DESC
LIMIT $2 OFFSET $3;
