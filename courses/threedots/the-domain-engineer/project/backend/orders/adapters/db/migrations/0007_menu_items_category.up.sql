BEGIN;
CREATE TYPE orders.item_category AS ENUM ('food', 'beverage');
ALTER TABLE orders.restaurant_menu_items ADD COLUMN category orders.item_category NOT NULL;
COMMIT;
