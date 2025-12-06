-- name: ListCategoriesByStore :many
SELECT category_id, store_id, name, parent_id, created_at
FROM product_category
WHERE store_id = $1
ORDER BY name;
