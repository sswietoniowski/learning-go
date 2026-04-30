BEGIN;

CREATE TABLE orders.orders
(
	order_uuid              uuid           NOT NULL,
	quote_uuid              uuid           NOT NULL,

	customer_uuid           uuid           NOT NULL,
	restaurant_uuid         uuid           NOT NULL,
	courier_uuid            uuid,

	delivery_address        json           NOT NULL,

	ordered_at              TIMESTAMPTZ    NOT NULL,
	restaurant_confirmed_at TIMESTAMPTZ,
	courier_accepted_at     TIMESTAMPTZ,
	restaurant_prepared_at  TIMESTAMPTZ,
	picked_up_at            TIMESTAMPTZ,
	delivered_at            TIMESTAMPTZ,

	items_subtotal_gross    DECIMAL(10, 2) NOT NULL,
	service_fee_gross       DECIMAL(10, 2) NOT NULL,
	delivery_fee_gross      DECIMAL(10, 2) NOT NULL,
	total_amount_gross      DECIMAL(10, 2) NOT NULL,
	total_tax               DECIMAL(10, 2) NOT NULL,

	currency                varchar(3)     NOT NULL,

	PRIMARY KEY (order_uuid),
	FOREIGN KEY (quote_uuid) REFERENCES orders.quotes (quote_uuid),
	FOREIGN KEY (customer_uuid) REFERENCES orders.customers (customer_uuid),
	FOREIGN KEY (restaurant_uuid) REFERENCES orders.restaurants (restaurant_uuid),
	FOREIGN KEY (courier_uuid) REFERENCES orders.couriers (courier_uuid)
);

-- Index for filtering orders by delivery city (used in courier available orders query)
CREATE INDEX idx_orders_delivery_city ON orders.orders ((delivery_address->>'city'));

COMMIT;
