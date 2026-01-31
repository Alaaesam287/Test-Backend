 -- name: GetTotalRevenueCurrentMonth :one
-- In Overview and Analytics pages 
SELECT 
  COALESCE(SUM(total_amount), 0) AS total_revenue 
FROM customer_order 
WHERE store_id = $1 
  AND status = 'completed' 
  AND created_at >= $2 
  AND created_at < $3;

-- name: GetRevenueGrowthPercent :one 
WITH current_period AS (
  SELECT COALESCE(SUM(total_amount), 0) AS revenue
  FROM customer_order
  WHERE store_id = $1
    AND status = 'completed'
    AND created_at >= $2
    AND created_at < $3
),
previous_period AS (
  SELECT COALESCE(SUM(total_amount), 0) AS revenue
  FROM customer_order
  WHERE store_id = $1
    AND status = 'completed'
    AND created_at >= ($2 - ($3 - $2))
    AND created_at <  $2
)
SELECT
  CASE
    WHEN previous_period.revenue = 0 THEN NULL
    ELSE ROUND(
      ((current_period.revenue - previous_period.revenue)
       / previous_period.revenue) * 100, 2
    )
  END AS revenue_growth_percent
FROM current_period, previous_period;

-- name: GetNewOrdersThisMonth :one
-- Overview page
SELECT
  COUNT(*) AS new_orders
FROM customer_order
WHERE store_id = $1
  AND created_at >= $2 
  AND created_at < $3;

-- name: GetOrdersGrowthPercent :one
WITH current_period AS (
  SELECT COALESCE(COUNT(*), 0) AS orders_count
  FROM customer_order
  WHERE store_id = $1
    AND status = 'completed'
    AND created_at >= $2
    AND created_at < $3
),
previous_period AS (
  SELECT COALESCE(COUNT(*), 0) AS orders_count
  FROM customer_order
  WHERE store_id = $1
    AND status = 'completed'
    AND created_at >= ($2 - ($3 - $2))
    AND created_at <  $2
)
SELECT
  CASE
    WHEN previous_period.orders_count = 0 THEN NULL
    ELSE ROUND(
      ((current_period.orders_count - previous_period.orders_count)::numeric
       / previous_period.orders_count) * 100,
      2
    )
  END AS orders_growth_percent
FROM current_period, previous_period;

-- name: GetTotalOrders :one
-- Orders page
SELECT COUNT(*) AS total_orders
FROM customer_order
WHERE store_id = $1;

-- name: GetPendingOrders :one
SELECT COUNT(*) AS pending_orders
FROM customer_order
WHERE store_id = $1
  AND status = 'pending';

-- name: GetShippedOrders :one
SELECT COUNT(*) AS shipped_orders
FROM customer_order
WHERE store_id = $1
  AND status = 'shipped';

-- name: GetCompletedOrders :one
SELECT COUNT(*) AS completed_orders
FROM customer_order
WHERE store_id = $1
  AND status = 'completed';

-- name: GetAverageDeliveryDays :one
SELECT ROUND(AVG(EXTRACT(EPOCH FROM (delivered_at - shipped_at))/86400), 2) AS avg_delivery_days
FROM shipment s
JOIN customer_order o ON o.order_id = s.order_id
WHERE o.store_id = $1
  AND s.delivered_at IS NOT NULL;

-- name: GetTotalVisitors :one
-- Total visitors(Overview page) / Visitors(Analytics page)
SELECT
  COUNT(DISTINCT session_id) AS visitors_this_month
FROM visitor_session
WHERE store_id = $1
  AND first_seen_at >= $2
  AND first_seen_at < $3;

-- name: GetNewVisitors :one
SELECT
  COUNT(DISTINCT session_id) AS new_visitors
FROM visitor_session
WHERE store_id = $1
  AND first_seen_at >= $2
  AND first_seen_at < $3;

-- name: GetPageViews :one
SELECT 
    COUNT(pv.product_view_id) AS total_page_views
FROM product_view pv
WHERE pv.store_id = $1
  AND pv.viewed_at BETWEEN $2 AND $3;

-- name: GetRegisteredCustomers :one
-- Overview & Customers page
SELECT COUNT(*) AS total_customers
FROM customer
WHERE store_id = $1;

-- name: GetCustomersGrowthPercent :one
-- Customers page
WITH current_period AS (
    SELECT COUNT(*) AS cnt
    FROM customer
    WHERE store_id = $1
      AND created_at >= $2
      AND created_at < $3
),
previous_period AS (
    SELECT COUNT(*) AS cnt
    FROM customer
    WHERE store_id = $1
      AND created_at >= ($2 - ($3 - $2))
      AND created_at < $2
)
SELECT
    CASE
        WHEN previous_period.cnt = 0 THEN NULL
        ELSE ROUND(
            ((current_period.cnt - previous_period.cnt)::numeric
             / previous_period.cnt) * 100, 2
        )
    END AS growth_percentage
FROM current_period, previous_period;

-- name: GetPurchasingCustomers  :one
SELECT COUNT(DISTINCT customer_id) AS purchasing_customers
FROM customer_order
WHERE store_id = $1
  AND status = 'completed'
  AND customer_id IS NOT NULL;

-- name: GetNewRegisteredCustomers  :one
-- Overview page
SELECT
  COUNT(*) AS new_customers
FROM customer
WHERE store_id = $1
  AND created_at >= $2
  AND created_at < $3;

-- name: GetCountOfNewPurchasingCustomers  :one
-- Customers page
WITH first_orders AS (
    SELECT customer_id, MIN(created_at) AS first_order_at
    FROM customer_order
    WHERE store_id = $1
      AND status = 'completed'
    GROUP BY customer_id
)
SELECT COUNT(*) AS new_purchasing_customers
FROM first_orders
WHERE first_order_at >= $2    
  AND first_order_at < $3;  

-- name: GetGrowthPercentOfPurchasingCustomers  :one
-- Customers page
WITH first_orders AS (
    SELECT customer_id, MIN(created_at) AS first_order_at
    FROM customer_order
    WHERE store_id = $1
      AND status = 'completed'
    GROUP BY customer_id
),
current_period AS (
    SELECT COUNT(*) AS cnt
    FROM first_orders
    WHERE first_order_at >= $2
      AND first_order_at < $3
),
previous_period AS (
    SELECT COUNT(*) AS cnt
    FROM first_orders
    WHERE first_order_at >= ($2 - ($3 - $2))
      AND first_order_at < $2
)
SELECT 
    CASE
        WHEN previous_period.cnt = 0 THEN NULL
        ELSE ROUND(
            ((current_period.cnt - previous_period.cnt)::numeric
             / previous_period.cnt) * 100, 2
        )
    END AS growth_percentage
FROM current_period, previous_period;

-- name: ListLowStockProducts :many
SELECT
  p.name,
  p.stock_quantity,
  CASE
    WHEN p.stock_quantity = 0 THEN 'OUT_OF_STOCK'
    WHEN p.stock_quantity BETWEEN 1 AND 10 THEN 'LOW_STOCK'
  END AS stock_status,
  p.updated_at
FROM product p
WHERE p.store_id = $1
  AND p.deleted_at IS NULL
  AND p.stock_quantity <= 10
ORDER BY
  p.stock_quantity ASC,
  p.updated_at DESC
LIMIT $2;                            -- The number of showing products 

-- name: GetTotalProducts :one
SELECT
  COUNT(*) AS total_products
FROM product
WHERE store_id = $1
  AND deleted_at IS NULL;

-- name: GetCountOfLowAndOutOfStockProducts :one
SELECT
  COUNT(*) AS low_out_stock_products
FROM product
WHERE store_id = $1
  AND deleted_at IS NULL
  AND stock_quantity <= 10;

-- name: ListProductTable :many
SELECT
    p.name AS product_name,

    -- Total units sold (completed + shipped)
    COALESCE(SUM(oi.quantity), 0) AS units_sold,
	
    -- Total views across all variants
    COALESCE(SUM(pvw.views_count), 0) AS total_views,

	-- Views-to-purchase ratio
    CASE
        WHEN COALESCE(SUM(oi.quantity), 0) = 0 THEN NULL
        ELSE ROUND(CAST(COALESCE(SUM(pvw.views_count), 0) AS numeric) 
                   / SUM(oi.quantity), 2)
    END AS views_to_purchase_ratio,
	
    -- Stock status based on sum of all variant stock
    CASE
        WHEN COALESCE(SUM(pv.stock_quantity), 0) = 0 THEN 'out of stock'
        WHEN COALESCE(SUM(pv.stock_quantity), 0) <= 10 THEN 'low stock'
        ELSE 'in stock'
    END AS stock_status

FROM product p
LEFT JOIN product_variant pv
    ON pv.product_id = p.product_id
LEFT JOIN order_item oi
    ON oi.variant_id = pv.variant_id
LEFT JOIN customer_order o
    ON o.order_id = oi.order_id AND o.status IN ('completed', 'shipped')
LEFT JOIN (
    SELECT product_id, COUNT(*) AS views_count
    FROM product_view
    GROUP BY product_id
) pvw
    ON pvw.product_id = p.product_id

WHERE p.store_id = $1
  AND p.deleted_at IS NULL

GROUP BY p.product_id, p.name
ORDER BY p.name ASC;

-- name: GetConversionRate :one
SELECT
    COUNT(DISTINCT co.order_id)::FLOAT / NULLIF(COUNT(DISTINCT vs.session_id),0) AS conversion_rate
FROM visitor_session vs
LEFT JOIN customer_order co ON vs.customer_id = co.customer_id
    AND co.store_id = $1
    AND co.status = 'completed'
    AND co.created_at BETWEEN $2 AND $3
WHERE vs.store_id = $1
  AND vs.first_seen_at BETWEEN $2 AND $3;

-- name: GetRevenueOverTime :many
SELECT 
    DATE(co.created_at) AS order_date,
    SUM(co.total_amount) AS revenue,        
    COUNT(DISTINCT co.order_id) AS total_orders     -- appears on the tooltip with the revenue
FROM customer_order co
WHERE co.store_id = $1
  AND co.created_at BETWEEN $2 AND $3
GROUP BY DATE(co.created_at)
ORDER BY order_date;

-- name: GetTopSellingProducts :many
SELECT 
    p.name AS product_name,
    SUM(oi.quantity) AS units_sold,                 
    SUM(oi.subtotal) AS revenue
FROM order_item oi
JOIN product_variant pv ON oi.variant_id = pv.variant_id
JOIN product p ON pv.product_id = p.product_id
JOIN customer_order co ON oi.order_id = co.order_id
WHERE co.store_id = $1
  AND co.created_at BETWEEN $2 AND $3
GROUP BY p.product_id, p.name
ORDER BY revenue DESC
LIMIT 5;

-- name: GetLoyalCustomers :many
SELECT 
    c.name AS customer_name,
    SUM(co.total_amount) AS total_spent,
    COUNT(co.order_id) AS orders_count
FROM customer c
JOIN customer_order co ON c.customer_id = co.customer_id
WHERE co.store_id = $1
  AND co.created_at BETWEEN $2 AND $3
GROUP BY c.customer_id, c.name
ORDER BY orders_count DESC
LIMIT 5;

-- name: GetProductsNeedingAttention :many
WITH product_sales AS (
    SELECT 
        pv.product_id,
        COALESCE(SUM(oi.quantity), 0) AS units_sold,
        COALESCE(SUM(oi.subtotal), 0) AS revenue
    FROM product_variant pv
    LEFT JOIN order_item oi ON pv.variant_id = oi.variant_id
    LEFT JOIN customer_order co ON oi.order_id = co.order_id
        AND co.store_id = $1
        AND co.created_at BETWEEN $2 AND $3
    WHERE pv.store_id = $1
    GROUP BY pv.product_id
)
SELECT 
    p.name AS product_name,
    s.units_sold,
    s.revenue
FROM product p
LEFT JOIN product_sales s ON p.product_id = s.product_id
WHERE p.store_id = $1
ORDER BY s.units_sold ASC, s.revenue ASC
LIMIT 5;

-- name: GetPopularButNotSellingProducts :many
WITH product_views AS (
    SELECT
        pv.product_id,
        COUNT(*) AS views
    FROM product_view pv
    WHERE pv.store_id = $1
      AND pv.viewed_at BETWEEN $2 AND $3
    GROUP BY pv.product_id
),
product_sales AS (
    SELECT
        pv.product_id,
        COUNT(DISTINCT co.order_id) AS sales_count
    FROM product_variant pv
    JOIN order_item oi ON oi.variant_id = pv.variant_id
    JOIN customer_order co ON co.order_id = oi.order_id
    WHERE pv.store_id = $1
      AND co.status = 'completed'
      AND co.created_at BETWEEN $2 AND $3
    GROUP BY pv.product_id
)
SELECT
    p.name AS product_name,
    COALESCE(v.views, 0) AS views,
    COALESCE(s.sales_count, 0) AS sales
FROM product p
LEFT JOIN product_views v ON v.product_id = p.product_id
LEFT JOIN product_sales s ON s.product_id = p.product_id
WHERE p.store_id = $1
  AND p.deleted_at IS NULL
ORDER BY
    v.views DESC,
    s.sales ASC
LIMIT 5;

-- name: GetFunnelMetrics :one
WITH
visits AS (
    SELECT session_id
    FROM visitor_session
    WHERE store_id = $1
      AND first_seen_at BETWEEN $2 AND $3
),
product_views AS (
    SELECT DISTINCT pv.session_id
    FROM product_view pv
    JOIN visits v ON v.session_id = pv.session_id
),
added_to_cart AS (
    SELECT DISTINCT ce.session_id
    FROM cart_event ce
    JOIN product_views pv ON pv.session_id = ce.session_id
    WHERE ce.event_type = 'add'
),
checkout_started AS (
    SELECT DISTINCT co.session_id
    FROM customer_order co
    JOIN added_to_cart ac ON ac.session_id = co.session_id
    WHERE co.created_at BETWEEN $2 AND $3
),
purchase_complete AS (
    SELECT DISTINCT co.session_id
    FROM customer_order co
    JOIN checkout_started cs ON cs.session_id = co.session_id
    WHERE co.status = 'completed'
      AND co.created_at BETWEEN $2 AND $3
)
SELECT
    (SELECT COUNT(*) FROM visits) AS site_visits,
    (SELECT COUNT(*) FROM product_views) AS product_views,
    (SELECT COUNT(*) FROM added_to_cart) AS added_to_cart,
    (SELECT COUNT(*) FROM checkout_started) AS checkout_started,
    (SELECT COUNT(*) FROM purchase_complete) AS purchase_complete,

    -- Drop percentages
    ROUND(((SELECT COUNT(*) FROM visits) - (SELECT COUNT(*) FROM product_views))::numeric
          / NULLIF((SELECT COUNT(*) FROM visits),0) * 100, 2) AS drop_site_to_view_pct,

    ROUND(((SELECT COUNT(*) FROM product_views) - (SELECT COUNT(*) FROM added_to_cart))::numeric
          / NULLIF((SELECT COUNT(*) FROM product_views),0) * 100, 2) AS drop_view_to_cart_pct,

    ROUND(((SELECT COUNT(*) FROM added_to_cart) - (SELECT COUNT(*) FROM checkout_started))::numeric
          / NULLIF((SELECT COUNT(*) FROM added_to_cart),0) * 100, 2) AS drop_cart_to_checkout_pct,

    ROUND(((SELECT COUNT(*) FROM checkout_started) - (SELECT COUNT(*) FROM purchase_complete))::numeric
          / NULLIF((SELECT COUNT(*) FROM checkout_started),0) * 100, 2) AS drop_checkout_to_purchase_pct;

-- name: GetConversionOverTime :many
SELECT
    DATE(vs.first_seen_at) AS day,
    ROUND(
        COUNT(DISTINCT co.session_id)::numeric
        / NULLIF(COUNT(DISTINCT vs.session_id), 0)
        * 100,
        2
    ) AS conversion_rate_percent
FROM visitor_session vs
LEFT JOIN customer_order co
    ON vs.session_id = co.session_id
   AND co.created_at >= $2
   AND co.created_at < $3
   AND co.status = 'completed'
WHERE vs.store_id = $1
  AND vs.first_seen_at >= $2
  AND vs.first_seen_at < $3
GROUP BY day
ORDER BY day;
