-- GIN index for full-text search on menu items
-- Indexes the tsvector for efficient @@ queries
CREATE INDEX idx_menu_items_fts
    ON orders.restaurant_menu_items
    USING gin (to_tsvector('english', name));
