BEGIN;

CREATE TYPE settlements.partner_type AS ENUM ('restaurant', 'courier');

CREATE TABLE settlements.billing_cycles
(
    billing_cycle_uuid uuid NOT NULL,
    partner_uuid uuid NOT NULL,
    partner_type settlements.partner_type NOT NULL,
    billing_cycle_number INT NOT NULL,
    closed BOOLEAN NOT NULL DEFAULT FALSE,
    settled BOOLEAN NOT NULL DEFAULT FALSE,
    start_date TIMESTAMPTZ NOT NULL,
    end_date TIMESTAMPTZ,
    PRIMARY KEY (billing_cycle_uuid),
    UNIQUE (partner_uuid, billing_cycle_number),
    FOREIGN KEY (partner_uuid) REFERENCES settlements.legal_entities(legal_entity_uuid)
);

CREATE UNIQUE INDEX billing_cycles_one_open_per_partner
    ON settlements.billing_cycles (partner_uuid)
    WHERE closed = false;

CREATE TABLE settlements.billing_cycle_orders (
    billing_cycle_uuid uuid NOT NULL,
    order_uuid uuid NOT NULL,
    PRIMARY KEY (billing_cycle_uuid, order_uuid),
    FOREIGN KEY (billing_cycle_uuid) REFERENCES settlements.billing_cycles(billing_cycle_uuid),
    FOREIGN KEY (order_uuid) REFERENCES settlements.orders(order_uuid)
);

COMMIT;
