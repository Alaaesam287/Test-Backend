-- name: ListCategoriesByStore :many
SELECT c.category_id, c.name, pc.name as parent_name
FROM store_category s
JOIN category_definition c
  ON s.category_id = c.category_id
JOIN category_definition pc
  ON c.parent_id = pc.category_id
WHERE s.store_id = $1
ORDER BY c.name;

-- name: GetProductBase :one
SELECT
  p.product_id,
  p.store_id,
  p.name,
  p.slug,
  p.description,
  p.brand,
  p.stock_quantity as total_stock,
  p.category_id,
  c.name as category_name,
  p.default_variant_id,
  v.price,
  p.in_stock,
  v.primary_image_url AS primary_image
FROM product p
JOIN product_variant v 
  ON v.variant_id = p.default_variant_id
JOIN category_definition c
  ON p.category_id = c.category_id
WHERE p.store_id = $1 AND p.product_id = $2 AND p.deleted_at IS NULL;


-- name: GetProductVariantAttributes :many
SELECT
  ad.attribute_id,
  ad.name as attribute_name,
  vav.value as attribute_value
FROM variant_attribute_value vav
JOIN attribute_definition ad 
  ON vav.attribute_id = ad.attribute_id
WHERE vav.variant_id = $1;


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
  p.stock_quantity as product_total_stock,
  v.stock_quantity as item_stock,
  v.price,
  p.in_stock,
  v.primary_image_url AS primary_image
FROM product p
JOIN product_variant v 
  ON v.variant_id = p.default_variant_id
WHERE 
  p.store_id = $1 
  AND p.category_id = $2
  AND p.in_stock = TRUE
  AND p.deleted_at IS NULL
ORDER BY 
  v.stock_quantity DESC
LIMIT $3;

-- name: ResolveAttributeIDByName :one
SELECT attribute_id
FROM attribute_definition
WHERE name = $1
LIMIT 1;

-- name: ResolveCategoryIDByName :one
SELECT c.category_id
FROM category_definition c
JOIN store_category s
  ON s.category_id = c.category_id
WHERE s.store_id = $1 AND c.name = $2
LIMIT 1;

-- name: GetCartBySession :one
SELECT c.cart_id, c.store_id, c.updated_at
FROM cart c
JOIN visitor_session s ON s.customer_id = c.customer_id
WHERE s.session_id = $1
  AND c.store_id = $2
LIMIT 1;

-- name: GetCartItems :many
SELECT
	ci.cart_item_id,
	ci.variant_id,
	p.product_id,
	p.name AS product_name,
	v.sku,
	v.primary_image_url,
	ci.unit_price,
	ci.quantity,
	(ci.unit_price * ci.quantity)::NUMERIC AS subtotal
FROM cart_item ci
JOIN product_variant v ON v.variant_id = ci.variant_id
JOIN product p ON p.product_id = v.product_id
WHERE ci.cart_id = $1
ORDER BY ci.created_at;
