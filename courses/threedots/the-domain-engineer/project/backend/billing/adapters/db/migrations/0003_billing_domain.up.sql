BEGIN;

ALTER TABLE billing.documents
    ADD COLUMN document_type VARCHAR NOT NULL DEFAULT 'receipt',
    ADD COLUMN issue_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ADD COLUMN currency VARCHAR NOT NULL DEFAULT 'USD',
    ADD COLUMN total_net_amount DECIMAL(10,2) NOT NULL DEFAULT 0,
    ADD COLUMN total_tax_amount DECIMAL(10,2) NOT NULL DEFAULT 0,
    ADD COLUMN total_gross_amount DECIMAL(10,2) NOT NULL DEFAULT 0;

CREATE TYPE billing.tax_type AS ENUM ('vat', 'gst', 'sales-tax');

CREATE TABLE billing.document_line_items
(
    line_item_uuid uuid NOT NULL,
    document_uuid uuid NOT NULL,
    name VARCHAR NOT NULL,
    quantity INT NOT NULL,
    unit_net_amount DECIMAL(10,2) NOT NULL,
    unit_tax_amount DECIMAL(10,2) NOT NULL,
    unit_gross_amount DECIMAL(10,2) NOT NULL,
    net_amount DECIMAL(10,2) NOT NULL,
    tax_amount DECIMAL(10,2) NOT NULL,
    gross_amount DECIMAL(10,2) NOT NULL,
    tax_rate DECIMAL(10,2) NOT NULL,
    tax_type billing.tax_type NOT NULL,
    PRIMARY KEY (line_item_uuid),
    FOREIGN KEY (document_uuid) REFERENCES billing.documents(document_uuid)
);

COMMIT;
