BEGIN;

ALTER TABLE billing.documents
    ADD COLUMN seller_uuid uuid NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000',
    ADD COLUMN buyer_uuid uuid NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000';

CREATE TABLE billing.legal_entity_snapshots
(
    snapshot_uuid uuid NOT NULL,
    name VARCHAR NOT NULL,
    address JSON NOT NULL,
    tax_id VARCHAR,
    PRIMARY KEY (snapshot_uuid)
);

ALTER TABLE billing.documents
    ADD FOREIGN KEY (seller_uuid) REFERENCES billing.legal_entity_snapshots(snapshot_uuid),
    ADD FOREIGN KEY (buyer_uuid) REFERENCES billing.legal_entity_snapshots(snapshot_uuid);

CREATE TABLE billing.document_taxes
(
    document_uuid uuid NOT NULL,
    tax_rate DECIMAL(10,2) NOT NULL,
    tax_type billing.tax_type NOT NULL,
    net_amount DECIMAL(10,2) NOT NULL,
    tax_amount DECIMAL(10,2) NOT NULL,
    PRIMARY KEY (document_uuid, tax_rate, tax_type),
    FOREIGN KEY (document_uuid) REFERENCES billing.documents(document_uuid)
);

COMMIT;
