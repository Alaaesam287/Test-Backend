-- name: ListCategoriesByStore :many
SELECT category_id, store_id, name, parent_id, created_at
FROM product_category
WHERE store_id = $1
ORDER BY name;

-- name: GetProductBase :one
SELECT
  p.product_id,
  p.store_id,
  p.name,
  p.slug,
  p.description,
  p.brand,
  p.category_id,
  p.default_variant_id,
  v.price,
  p.in_stock,
  v.primary_image_url AS primary_image
FROM product p
JOIN product_variant v 
  ON v.variant_id = p.default_variant_id
WHERE p.store_id = $1 AND p.product_id = $2 AND p.deleted_at IS NULL;


-- name: GetProductAttributes :many
SELECT
  ad.attribute_id,
  ad.name,
  pav.value
FROM product_attribute_value pav
JOIN attribute_definition ad 
  ON pav.attribute_id = ad.attribute_id
WHERE pav.product_id = $1;


-- name: GetProductVariants :many
SELECT
  variant_id,
  product_id,
  sku,
  price,
  stock_quantity,
  primary_image_url
FROM product_variant
WHERE product_id = $1 AND deleted_at IS NULL;


-- name: GetVariantOptions :many
SELECT
  vo.variant_id,
  ot.name AS option_type,
  ov.value AS option_value
FROM variant_option vo
JOIN option_value ov ON vo.option_value_id = ov.option_value_id
JOIN option_type ot ON ov.option_type_id = ot.option_type_id
WHERE vo.variant_id = $1;

-- name: GetTopProductsByCategory :many
SELECT 
  p.product_id,
  p.store_id,
  p.name,
  p.slug,
  p.description,
  p.brand,
  p.category_id,
  p.default_variant_id,
  v.price,
  p.in_stock,
  v.primary_image_url AS primary_image
FROM product p
JOIN product_variant v 
  ON v.variant_id = p.default_variant_id
WHERE 
  p.store_id = $1 
  AND p.category_id = $2
  AND p.deleted_at IS NULL
ORDER BY 
  v.stock_quantity DESC
LIMIT $3;

-- name: ResolveAttributeIDByName :one
SELECT attribute_id
FROM attribute_definition
WHERE store_id = $1 AND name = $2
LIMIT 1;

-- name: ResolveCategoryIDByName :one
SELECT category_id
FROM product_category
WHERE store_id = $1 AND name = $2
LIMIT 1;
