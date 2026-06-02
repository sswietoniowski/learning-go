BEGIN;

CREATE SCHEMA IF NOT EXISTS orders;

CREATE TABLE orders.customers
(
	customer_uuid uuid         NOT NULL,
	name          varchar(255) NOT NULL,
	email         varchar(255) NOT NULL,
	address       json         NOT NULL,
	phone_number  varchar(50)  NOT NULL,
	PRIMARY KEY (customer_uuid)
);

COMMIT;