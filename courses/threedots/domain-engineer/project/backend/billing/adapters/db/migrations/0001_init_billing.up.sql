BEGIN;

CREATE SCHEMA IF NOT EXISTS billing;

CREATE TABLE billing.document_series
(
    prefix VARCHAR NOT NULL,
    last_number INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (prefix)
);

INSERT INTO billing.document_series (prefix)
VALUES
    ('R');

CREATE TABLE billing.documents
(
    document_uuid uuid NOT NULL,
    document_number VARCHAR NOT NULL,
    series_prefix VARCHAR NOT NULL,
    PRIMARY KEY (document_uuid),
    UNIQUE (document_number),
    FOREIGN KEY (series_prefix) REFERENCES billing.document_series(prefix)
);

COMMIT;
