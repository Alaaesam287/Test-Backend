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

-- name: GetProductForUpdate :one
SELECT product_id, store_id, category_id, stock_quantity
FROM product
WHERE product_id = $1
FOR UPDATE;


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
SELECT
  c.cart_id,
  c.store_id,
  c.updated_at
FROM cart c
WHERE c.session_id = $1
  AND c.store_id = $2
LIMIT 1;

-- name: GetCartTotal :one
SELECT
  COALESCE(SUM(ci.unit_price * ci.quantity), 0)::NUMERIC AS total
FROM cart_item ci
WHERE ci.cart_id = $1;

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

-- name: GetSession :one
SELECT session_id, customer_id
FROM visitor_session
WHERE session_id = $1 AND store_id = $2;

-- name: GetCartForSession :one
SELECT *
FROM cart
WHERE store_id = $1 AND session_id = $2
FOR UPDATE;

-- name: CreateCart :one
INSERT INTO cart (store_id, session_id, customer_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetVariantForCart :one
SELECT
  variant_id,
  price,
  stock_quantity
FROM product_variant
WHERE variant_id = $1
  AND store_id = $2
  AND deleted_at IS NULL
FOR UPDATE;

-- name: UpsertCartItem :exec
INSERT INTO cart_item (cart_id, variant_id, quantity, unit_price)
VALUES ($1, $2, $3, $4)
ON CONFLICT (cart_id, variant_id)
DO UPDATE SET
  quantity = cart_item.quantity + EXCLUDED.quantity;

-- name: TouchCart :exec
UPDATE cart SET updated_at = NOW() WHERE cart_id = $1;

-- name: IsStoreOwner :one
SELECT EXISTS (
    SELECT 1
    FROM store
    WHERE store_id = $1
      AND store_owner_id = $2
);

-- name: CreateStoreOwner :one
INSERT INTO store_owner (
  name,
  email,
  password_hash,
  phone,
  address
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING store_owner_id, name, email, created_at;

-- name: GetStoreOwnerByEmail :one
SELECT
  store_owner_id,
  name,
  email,
  password_hash,
  created_at
FROM store_owner
WHERE email = $1;

-- name: CreateCustomer :one
INSERT INTO customer (
  store_id,
  name,
  email,
  password_hash,
  phone,
  address
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING customer_id, store_id, name, email, created_at;

-- name: GetCustomerByEmail :one
SELECT
  customer_id,
  store_id,
  name,
  email,
  password_hash,
  created_at
FROM customer
WHERE email = $1
  AND store_id = $2;

-- name: ListCategoryAttributes :many
SELECT a.attribute_id, a.name, ca.is_required
FROM category_attribute ca
JOIN attribute_definition a 
ON a.attribute_id = ca.attribute_id
WHERE ca.category_id = $1;

-- name: GetAdminByEmail :one
SELECT admin_id, email, password_hash
FROM admin
WHERE email = $1;

-- name: CreateRefreshToken :exec
INSERT INTO refresh_token (token, user_id, user_role, store_id, expires_at)
VALUES ($1, $2, $3, $4, $5);

-- name: GetRefreshToken :one
SELECT *
FROM refresh_token
WHERE token = $1 AND revoked = FALSE;

-- name: RevokeRefreshToken :exec
UPDATE refresh_token
SET revoked = TRUE
WHERE token = $1;

-- name: GetProductByStoreAndName :one
SELECT *
FROM product
WHERE store_id = $1 AND name = $2 AND deleted_at IS NULL
LIMIT 1;

-- name: CreateProduct :one
INSERT INTO product (
  store_id, category_id, name, slug, description, brand
)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateProductStock :exec
UPDATE product
SET stock_quantity = stock_quantity + $2,
    in_stock = (stock_quantity + $2) > 0,
    updated_at = NOW()
WHERE product_id = $1;

-- name: SetDefaultVariant :exec
UPDATE product
SET default_variant_id = $2
WHERE product_id = $1;

-- name: GetVariantByAttributeHash :one
SELECT *
FROM product_variant
WHERE product_id = $1
  AND attribute_hash = $2
  AND deleted_at IS NULL
LIMIT 1;

-- name: GetVariant :one
SELECT *
FROM product_variant
WHERE variant_id = $1
  AND deleted_at IS NULL;

-- name: GetVariantForUpdate :one
SELECT *
FROM product_variant
WHERE variant_id = $1
  AND deleted_at IS NULL
FOR UPDATE; 

-- name: CreateVariant :one
INSERT INTO product_variant (
  product_id, store_id, attribute_hash,
  sku, price, stock_quantity
)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: IncreaseVariantStock :exec
UPDATE product_variant
SET stock_quantity = stock_quantity + $2,
    updated_at = NOW()
WHERE variant_id = $1;

-- name: InsertVariantAttribute :exec
INSERT INTO variant_attribute_value (
  variant_id, attribute_id, value
)
VALUES ($1, $2, $3);

-- name: GetProductByIdentity :one
SELECT *
FROM product
WHERE store_id = $1
  AND name = $2
  AND category_id = $3
  AND brand = $4
  AND deleted_at IS NULL
LIMIT 1;

-- name: CategoryHasAttribute :one
SELECT 1
FROM category_attribute
WHERE category_id = $1
  AND attribute_id = $2;

-- name: SetPrimaryVariantImage :exec
UPDATE product_variant
SET primary_image_url = $2
WHERE variant_id = $1;

-- name: InsertVariantImage :one
INSERT INTO product_variant_image (product_variant_id, image_url)
VALUES ($1, $2)
RETURNING *;

-- name: CreateStore :one
INSERT INTO store (
    store_owner_id,
    name,
    domain,
    currency,
    timezone
) VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateStoreDownloadStatus :exec
UPDATE store
SET download_status = $2,
    updated_at = NOW()
WHERE store_id = $1;

-- name: GetStoreByOwnerID :one
SELECT *
FROM store
WHERE store_owner_id = $1;

-- name: DeleteStore :exec
DELETE FROM store
WHERE store_id = $1;


-- name: GetStore :one
SELECT *
FROM store
WHERE store_id = $1;
