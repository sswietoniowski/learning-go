BEGIN;

CREATE TABLE settlements.orders
(
    -- Note: We don't add a foreign key to the orders schema here to avoid tight coupling between modules.
    -- Both modules are independent and run separate transactions. They should not rely on each other's database schemas.
    order_uuid uuid NOT NULL,
    restaurant_uuid uuid NOT NULL,
    courier_uuid uuid NOT NULL,
    currency VARCHAR NOT NULL,
    commission_net_amount DECIMAL(10,2) NOT NULL,
    ordered_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (order_uuid),
    FOREIGN KEY (restaurant_uuid) REFERENCES settlements.legal_entities(legal_entity_uuid),
    FOREIGN KEY (courier_uuid) REFERENCES settlements.legal_entities(legal_entity_uuid)
);

CREATE TYPE settlements.breakdown_type AS ENUM ('items', 'delivery', 'total');

CREATE TABLE settlements.order_breakdowns
(
    order_uuid uuid NOT NULL,
    breakdown_type settlements.breakdown_type NOT NULL,
    net_amount DECIMAL(10,2) NOT NULL,
    tax_amount DECIMAL(10,2) NOT NULL,
    gross_amount DECIMAL(10,2) NOT NULL,
    PRIMARY KEY (order_uuid, breakdown_type),
    FOREIGN KEY (order_uuid) REFERENCES settlements.orders(order_uuid)
);

COMMIT;
