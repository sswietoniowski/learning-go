-- name: ListCustomerOrders :many
SELECT
    o.order_uuid,
    o.restaurant_uuid,
    r.name AS restaurant_name,
    o.courier_uuid,
    o.delivery_address,
    o.ordered_at,
    o.restaurant_confirmed_at,
    o.courier_accepted_at,
    o.restaurant_prepared_at,
    o.picked_up_at,
    o.delivered_at,
    o.items_subtotal_gross,
    o.service_fee_gross,
    o.delivery_fee_gross,
    o.total_amount_gross,
    o.total_tax,
    o.currency
FROM orders.orders o
JOIN orders.restaurants r ON o.restaurant_uuid = r.restaurant_uuid
WHERE o.customer_uuid = $1
ORDER BY o.ordered_at DESC;

-- name: ListRestaurantOrders :many
SELECT
    o.order_uuid,
    o.customer_uuid,
    o.courier_uuid,
    o.ordered_at,
    o.restaurant_confirmed_at,
    o.courier_accepted_at,
    o.restaurant_prepared_at,
    o.picked_up_at,
    o.delivered_at,
    o.items_subtotal_gross
FROM orders.orders o
WHERE o.restaurant_uuid = $1
ORDER BY o.ordered_at DESC;

-- name: ListAssignedCourierOrders :many
SELECT
    o.order_uuid,
    o.customer_uuid,
    o.courier_uuid,
    o.restaurant_uuid,
    r.name AS restaurant_name,
    o.delivery_address,
    o.ordered_at,
    o.restaurant_confirmed_at,
    o.courier_accepted_at,
    o.restaurant_prepared_at,
    o.picked_up_at,
    o.delivered_at,
    o.items_subtotal_gross
FROM orders.orders o
JOIN orders.restaurants r ON o.restaurant_uuid = r.restaurant_uuid
WHERE o.courier_uuid = $1
ORDER BY o.ordered_at DESC;

-- name: ListAvailableOrdersForCourier :many
SELECT
    o.order_uuid,
    o.customer_uuid,
    o.courier_uuid,
    o.restaurant_uuid,
    r.name AS restaurant_name,
    o.delivery_address,
    o.ordered_at,
    o.restaurant_confirmed_at,
    o.courier_accepted_at,
    o.restaurant_prepared_at,
    o.picked_up_at,
    o.delivered_at,
    o.items_subtotal_gross
FROM orders.orders o
JOIN orders.restaurants r ON o.restaurant_uuid = r.restaurant_uuid
WHERE
    o.restaurant_confirmed_at IS NOT NULL AND
    o.courier_uuid IS NULL AND
    o.delivered_at IS NULL AND
    (o.delivery_address ->> 'city') = (
        SELECT city
        FROM orders.couriers
        WHERE couriers.courier_uuid = $1
    )
ORDER BY o.ordered_at DESC;

-- name: ListMenuItems :many
-- Lists menu items with optional restaurant name filter, optional full-text search, and dynamic ordering.
-- Uses CASE WHEN to support multiple ordering options in a single query.
SELECT
    mi.restaurant_menu_item_uuid AS menu_item_uuid,
    mi.name AS menu_item_name,
    mi.gross_price,
    r.currency,
    r.restaurant_uuid,
    r.name AS restaurant_name,
    CASE WHEN sqlc.narg(search_term)::text IS NOT NULL
         THEN ts_rank(
             to_tsvector('english', mi.name),
             plainto_tsquery('english', sqlc.narg(search_term)::text)
         )
         ELSE NULL
    END AS relevance
FROM orders.restaurant_menu_items mi
JOIN orders.restaurants r ON mi.restaurant_uuid = r.restaurant_uuid
WHERE mi.is_archived = false
  AND (sqlc.narg(search_term)::text IS NULL
       OR to_tsvector('english', mi.name) @@ plainto_tsquery('english', sqlc.narg(search_term)::text))
  AND (sqlc.narg(restaurant_name_filter)::text IS NULL
       OR LOWER(r.name) LIKE LOWER('%' || sqlc.narg(restaurant_name_filter)::text || '%'))
ORDER BY
    CASE WHEN sqlc.narg(order_by)::text = 'relevance'
         THEN ts_rank(
             to_tsvector('english', mi.name),
             plainto_tsquery('english', sqlc.narg(search_term)::text)
         )
    END DESC,
    CASE WHEN (sqlc.narg(order_by)::text IS NULL OR sqlc.narg(order_by)::text = 'default')
         THEN r.name END ASC,
    CASE WHEN (sqlc.narg(order_by)::text IS NULL OR sqlc.narg(order_by)::text = 'default')
         THEN mi.ordering END ASC,
    CASE WHEN sqlc.narg(order_by)::text = 'price_asc' THEN mi.gross_price END ASC,
    CASE WHEN sqlc.narg(order_by)::text = 'price_desc' THEN mi.gross_price END DESC,
    CASE WHEN sqlc.narg(order_by)::text = 'name_asc' THEN mi.name END ASC,
    CASE WHEN sqlc.narg(order_by)::text = 'name_desc' THEN mi.name END DESC;
