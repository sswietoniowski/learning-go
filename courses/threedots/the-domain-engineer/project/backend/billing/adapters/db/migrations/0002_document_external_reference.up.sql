BEGIN;

ALTER TABLE billing.documents
    ADD COLUMN external_reference VARCHAR,
    ADD UNIQUE (external_reference);

COMMIT;
