BEGIN;

CREATE SCHEMA IF NOT EXISTS settlements;

CREATE TYPE settlements.legal_entity_type AS ENUM ('partner', 'platform');

CREATE TABLE settlements.legal_entities
(
    legal_entity_uuid uuid NOT NULL,
    legal_entity_type settlements.legal_entity_type NOT NULL,
    business_name VARCHAR NOT NULL,
    address JSON NOT NULL,
    tax_id VARCHAR NOT NULL,
    bank_account_number VARCHAR NOT NULL,
    currency VARCHAR NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (legal_entity_uuid)
);

CREATE TABLE settlements.partner_platform_mappings
(
    partner_uuid uuid NOT NULL,
    platform_entity_uuid uuid NOT NULL,
    PRIMARY KEY (partner_uuid),
    FOREIGN KEY (partner_uuid) REFERENCES settlements.legal_entities(legal_entity_uuid),
    FOREIGN KEY (platform_entity_uuid) REFERENCES settlements.legal_entities(legal_entity_uuid)
);

COMMIT;
