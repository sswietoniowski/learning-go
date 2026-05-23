-- name: SaveOrder :exec
INSERT INTO settlements.orders (order_uuid, restaurant_uuid, courier_uuid, currency, commission_net_amount, ordered_at)
VALUES (
           sqlc.arg(order_uuid),
           sqlc.arg(restaurant_uuid), sqlc.arg(courier_uuid), sqlc.arg(currency),
           sqlc.arg(commission_net_amount), sqlc.arg(ordered_at)
)
ON CONFLICT (order_uuid) DO NOTHING;

-- name: SaveOrderBreakdown :exec
INSERT INTO settlements.order_breakdowns (order_uuid, breakdown_type, net_amount, tax_amount, gross_amount)
VALUES (
           sqlc.arg(order_uuid), sqlc.arg(breakdown_type),
           sqlc.arg(net_amount), sqlc.arg(tax_amount), sqlc.arg(gross_amount)
)
ON CONFLICT (order_uuid, breakdown_type) DO NOTHING;
