BEGIN;

CREATE TABLE orders.quotes
(
	quote_uuid           uuid           NOT NULL,
	customer_uuid        uuid           NOT NULL,
	restaurant_uuid      uuid           NOT NULL,

	delivery_address     json           NOT NULL,

	created_at           TIMESTAMPTZ    NOT NULL,

	items_subtotal_gross DECIMAL(10, 2) NOT NULL,
	service_fee_gross    DECIMAL(10, 2) NOT NULL,
	delivery_fee_gross   DECIMAL(10, 2) NOT NULL,
	total_amount_gross   DECIMAL(10, 2) NOT NULL,
	total_tax            DECIMAL(10, 2) NOT NULL,

	currency             varchar(3)     NOT NULL,

	PRIMARY KEY (quote_uuid),
	FOREIGN KEY (customer_uuid) REFERENCES orders.customers (customer_uuid),
	FOREIGN KEY (restaurant_uuid) REFERENCES orders.restaurants (restaurant_uuid)
);

CREATE TABLE orders.quote_items
(
	quote_item_uuid    uuid           NOT NULL,
	quote_uuid         uuid           NOT NULL,
	menu_item_uuid uuid           NOT NULL,

	gross_price        DECIMAL(10, 2) NOT NULL,
	quantity           INT            NOT NULL,

	PRIMARY KEY (quote_item_uuid),
	FOREIGN KEY (quote_uuid) REFERENCES orders.quotes (quote_uuid),
	FOREIGN KEY (menu_item_uuid) REFERENCES orders.restaurant_menu_items (restaurant_menu_item_uuid)
);

COMMIT;