-- name: LegalEntityByUUID :one
SELECT * FROM settlements.legal_entities WHERE legal_entity_uuid = $1;

-- name: SaveLegalEntity :exec
INSERT INTO settlements.legal_entities (legal_entity_uuid, legal_entity_type, business_name, tax_id, address, bank_account_number, currency)
VALUES (
           sqlc.arg(legal_entity_uuid), sqlc.arg(legal_entity_type),
           sqlc.arg(business_name), sqlc.arg(tax_id),
           sqlc.arg(address),
           sqlc.arg(bank_account_number), sqlc.arg(currency)
       )
ON CONFLICT (legal_entity_uuid) DO UPDATE SET
                          business_name = EXCLUDED.business_name,
                          tax_id = EXCLUDED.tax_id,
                          address = EXCLUDED.address,
                          bank_account_number = EXCLUDED.bank_account_number,
                          currency = EXCLUDED.currency,
                          updated_at = NOW();

-- name: SavePartnerPlatformMapping :exec
INSERT INTO settlements.partner_platform_mappings (partner_uuid, platform_entity_uuid)
VALUES ($1, $2)
ON CONFLICT (partner_uuid) DO NOTHING;

-- name: PlatformByPartnerUUID :one
SELECT platform_entity_uuid
FROM settlements.partner_platform_mappings
WHERE partner_uuid = $1;
