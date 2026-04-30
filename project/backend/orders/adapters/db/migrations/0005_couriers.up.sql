BEGIN;

CREATE TABLE orders.couriers
(
	courier_uuid uuid         NOT NULL,
	name         varchar(255) NOT NULL,
	phone_number varchar(50)  NOT NULL,
	city         varchar(100) NOT NULL,
	PRIMARY KEY (courier_uuid)
);

COMMIT;
