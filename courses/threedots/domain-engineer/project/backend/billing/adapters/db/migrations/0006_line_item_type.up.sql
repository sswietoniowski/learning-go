BEGIN;

CREATE TYPE billing.line_item_type AS ENUM ('food', 'beverage', 'delivery', 'service');

ALTER TABLE billing.document_line_items
    ADD COLUMN line_item_type billing.line_item_type NOT NULL;

COMMIT;
