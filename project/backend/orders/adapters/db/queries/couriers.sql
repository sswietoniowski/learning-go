-- name: InsertCourier :exec
INSERT INTO orders.couriers (
	courier_uuid,
	name,
	phone_number,
	city
)
VALUES
	($1, $2, $3, $4);

-- name: GetCourierCity :one
SELECT
	city
FROM
	orders.couriers
WHERE
	courier_uuid = $1
;