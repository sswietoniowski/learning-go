-- name: SaveLegalEntitySnapshot :exec
INSERT INTO billing.legal_entity_snapshots (
    snapshot_uuid, name, address, tax_id
)
VALUES (
    sqlc.arg(snapshot_uuid), sqlc.arg(name),
    sqlc.arg(address), sqlc.arg(tax_id)
);
