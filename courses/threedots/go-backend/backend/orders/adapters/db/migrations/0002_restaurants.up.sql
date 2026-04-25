BEGIN;

CREATE TABLE orders.restaurants
(
	restaurant_uuid uuid          NOT NULL,
	name            varchar(255)  NOT NULL,
	description     varchar(1024) NOT NULL,
	address         json          NOT NULL,
	currency        varchar(3)    NOT NULL,
	PRIMARY KEY (restaurant_uuid)
);

CREATE TABLE orders.restaurant_menu_items
(
	restaurant_menu_item_uuid uuid           NOT NULL,
	restaurant_uuid               uuid           NOT NULL,
	name                          varchar(255)   NOT NULL,
	gross_price                   DECIMAL(10, 2) NOT NULL,
	ordering                      FLOAT          NOT NULL,
	is_archived                   boolean        NOT NULL,
	PRIMARY KEY (restaurant_menu_item_uuid),
	FOREIGN KEY (restaurant_uuid) REFERENCES orders.restaurants (restaurant_uuid)
);

COMMIT;
