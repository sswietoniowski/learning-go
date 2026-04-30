-- name: InsertCustomer :exec
INSERT INTO
	orders.customers (
	customer_uuid,
	name,
	email,
	address,
	phone_number)
VALUES
	($1, $2, $3, $4, $5)
;


-- name: GetCustomerByUUID :one
SELECT
	*
FROM
	orders.customers
WHERE
	customer_uuid = $1
;