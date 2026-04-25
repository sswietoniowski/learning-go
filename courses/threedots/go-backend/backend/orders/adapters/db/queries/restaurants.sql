-- name: UpsertRestaurant :one
INSERT INTO orders.restaurants (restaurant_uuid, name, description, address, currency)
VALUES
	($1, $2, $3, $4, $5)
ON CONFLICT (restaurant_uuid) DO UPDATE SET
	name = EXCLUDED.name,
	description = EXCLUDED.description,
	address = EXCLUDED.address
RETURNING *;

-- name: GetRestaurant :one
SELECT
	*
FROM
	orders.restaurants
WHERE
	restaurant_uuid = $1
;

-- name: GetRestaurantMenu :many
SELECT
	sqlc.embed(restaurant_menu_items)
FROM
	orders.restaurant_menu_items AS restaurant_menu_items
WHERE
	restaurant_uuid = $1 AND
	is_archived = FALSE
ORDER BY
	ordering ASC
;

-- name: UpsertRestaurantMenuItem :exec
INSERT INTO orders.restaurant_menu_items (
	restaurant_menu_item_uuid,
	restaurant_uuid,
	name,
	gross_price,
	ordering,
	is_archived
)
VALUES
	($1, $2, $3, $4, $5, $6)
ON CONFLICT (restaurant_menu_item_uuid) DO UPDATE SET
	restaurant_uuid = EXCLUDED.restaurant_uuid,
	name = EXCLUDED.name,
	gross_price = EXCLUDED.gross_price,
	ordering = EXCLUDED.ordering,
	is_archived = EXCLUDED.is_archived
;

-- name: ArchiveMenuItems :exec
UPDATE
	orders.restaurant_menu_items
SET
	is_archived = TRUE
WHERE
	restaurant_menu_item_uuid = ANY ($1::UUID[])
;

-- name: GetMenuItemsByUUIDs :many
SELECT
	restaurant_menu_items.*
FROM
	orders.restaurant_menu_items AS restaurant_menu_items
WHERE
	restaurant_uuid = $1 AND
	restaurant_menu_item_uuid = ANY ($2::UUID[])
;
