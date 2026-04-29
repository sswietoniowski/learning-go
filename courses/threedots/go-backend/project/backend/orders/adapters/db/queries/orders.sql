-- name: GetQuoteItems :many
SELECT *
FROM orders.quote_items
WHERE quote_uuid = $1;

-- name: GetQuote :one
SELECT
	*
FROM
	orders.quotes AS quotes
WHERE
	quote_uuid = $1
LIMIT 1;

-- name: AddQuote :exec
INSERT INTO orders.quotes (
	quote_uuid,
	customer_uuid,
	restaurant_uuid,
	delivery_address,
	items_subtotal_gross,
	service_fee_gross,
	delivery_fee_gross,
	total_amount_gross,
	total_tax,
	created_at,
	currency
)
VALUES
	($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
;

-- name: AddQuoteItems :copyfrom
INSERT INTO orders.quote_items (
	quote_item_uuid,
	quote_uuid,
	menu_item_uuid,
	gross_price,
	quantity
)
VALUES
	($1, $2, $3, $4, $5);

-- name: PlaceOrder :exec
INSERT INTO orders.orders (
	order_uuid,
	quote_uuid,
	customer_uuid,
	restaurant_uuid,
	delivery_address,
	ordered_at,
	items_subtotal_gross,
	service_fee_gross,
	delivery_fee_gross,
	total_amount_gross,
	total_tax,
	currency
)
VALUES
	($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
;

-- name: GetOrder :one
SELECT *
FROM orders.orders
WHERE order_uuid = $1
LIMIT 1;

-- name: UpdateOrder :exec
UPDATE orders.orders
SET
    courier_uuid            = COALESCE(sqlc.narg('courier_uuid'), courier_uuid),
    ordered_at              = COALESCE(sqlc.narg('ordered_at'), ordered_at),
    restaurant_confirmed_at = COALESCE(sqlc.narg('restaurant_confirmed_at'), restaurant_confirmed_at),
    courier_accepted_at     = COALESCE(sqlc.narg('courier_accepted_at'), courier_accepted_at),
    restaurant_prepared_at  = COALESCE(sqlc.narg('restaurant_prepared_at'), restaurant_prepared_at),
    picked_up_at            = COALESCE(sqlc.narg('picked_up_at'), picked_up_at),
    delivered_at            = COALESCE(sqlc.narg('delivered_at'), delivered_at)
WHERE order_uuid = $1;