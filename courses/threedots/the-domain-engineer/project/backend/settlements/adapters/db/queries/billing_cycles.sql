-- name: CurrentBillingCycle :one
SELECT * FROM settlements.billing_cycles
WHERE partner_uuid = $1 AND closed = false
LIMIT 1;

-- name: SaveBillingCycle :exec
INSERT INTO settlements.billing_cycles (billing_cycle_uuid, partner_uuid, partner_type, billing_cycle_number, closed, settled, start_date, end_date)
VALUES (
        sqlc.arg(billing_cycle_uuid),
        sqlc.arg(partner_uuid),
        sqlc.arg(partner_type),
        sqlc.arg(billing_cycle_number),
        sqlc.arg(closed),
        sqlc.arg(settled),
        sqlc.arg(start_date),
        sqlc.arg(end_date)
       )
ON CONFLICT (billing_cycle_uuid) DO UPDATE SET
   -- Other fields are immutable
    closed = EXCLUDED.closed,
    settled = EXCLUDED.settled,
    end_date = EXCLUDED.end_date;

-- name: OrdersByBillingCycleUUID :many
SELECT * FROM settlements.orders
INNER JOIN settlements.billing_cycle_orders USING (order_uuid)
WHERE billing_cycle_uuid = $1;

-- name: OrderBreakdownsByBillingCycleUUID :many
SELECT ob.*
FROM settlements.order_breakdowns ob
INNER JOIN settlements.billing_cycle_orders bco USING (order_uuid)
WHERE bco.billing_cycle_uuid = $1;

-- name: AddOrderToBillingCycle :exec
INSERT INTO settlements.billing_cycle_orders (billing_cycle_uuid, order_uuid)
VALUES (
           sqlc.arg(billing_cycle_uuid), sqlc.arg(order_uuid)
)
ON CONFLICT (billing_cycle_uuid, order_uuid) DO NOTHING;

-- name: OrderInPartnerBillingCycleExists :one
SELECT EXISTS(
    SELECT 1 FROM settlements.billing_cycle_orders bco
    INNER JOIN settlements.billing_cycles bc ON bco.billing_cycle_uuid = bc.billing_cycle_uuid
    WHERE bco.order_uuid = sqlc.arg(order_uuid)
      AND bc.partner_uuid = sqlc.arg(partner_uuid)
);

-- name: BillingCyclesByPartnerUUID :many
SELECT *
FROM settlements.billing_cycles
WHERE partner_uuid = $1
ORDER BY billing_cycle_number DESC;

-- name: DeliveryInvoicesByBillingCycleUUID :many
SELECT
    bc.partner_uuid AS seller_uuid,
    o.restaurant_uuid AS buyer_uuid,
    COUNT(o.order_uuid) AS quantity,
    SUM(ob.net_amount)::DECIMAL AS net_amount
FROM settlements.orders o
JOIN settlements.order_breakdowns ob ON o.order_uuid = ob.order_uuid AND ob.breakdown_type = 'delivery'
JOIN settlements.billing_cycle_orders bco ON o.order_uuid = bco.order_uuid
JOIN settlements.billing_cycles bc USING (billing_cycle_uuid)
WHERE bc.billing_cycle_uuid = $1
    AND bc.partner_type = 'courier'
GROUP BY (bc.partner_uuid, o.restaurant_uuid)
ORDER BY (bc.partner_uuid, o.restaurant_uuid);

-- name: CommissionInvoiceByBillingCycleUUID :one
SELECT
    bc.partner_uuid AS buyer_uuid,
    COUNT(o.order_uuid) AS quantity,
    SUM(o.commission_net_amount)::DECIMAL AS net_amount
FROM settlements.orders o
JOIN settlements.billing_cycle_orders bco ON o.order_uuid = bco.order_uuid
JOIN settlements.billing_cycles bc USING (billing_cycle_uuid)
WHERE bc.billing_cycle_uuid = $1
    AND bc.partner_type = 'restaurant'
GROUP BY bc.partner_uuid;

-- name: UnsettledClosedCycles :many
SELECT * FROM settlements.billing_cycles
WHERE partner_uuid = $1 AND closed = true AND settled = false
ORDER BY billing_cycle_number ASC;

-- name: GetBillingCycleByUUID :one
SELECT * FROM settlements.billing_cycles
WHERE billing_cycle_uuid = $1;
