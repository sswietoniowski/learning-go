BEGIN;

ALTER TABLE billing.documents
    ADD COLUMN file_url VARCHAR;

COMMIT;
