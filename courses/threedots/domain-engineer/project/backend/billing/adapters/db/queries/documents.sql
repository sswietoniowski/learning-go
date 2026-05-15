-- name: NextDocumentNumber :one
UPDATE billing.document_series
SET last_number = last_number + 1,
    updated_at = NOW()
WHERE prefix = $1
RETURNING last_number;

-- name: SaveDocument :exec
INSERT INTO billing.documents (
    document_uuid, external_reference, document_number, series_prefix, document_type, issue_date, currency, total_net_amount, total_tax_amount, total_gross_amount, seller_uuid, buyer_uuid
)
VALUES (
    sqlc.arg(document_uuid), sqlc.arg(external_reference), sqlc.arg(document_number), sqlc.arg(series_prefix),
    sqlc.arg(document_type), sqlc.arg(issue_date), sqlc.arg(currency),
    sqlc.arg(total_net_amount), sqlc.arg(total_tax_amount), sqlc.arg(total_gross_amount),
    sqlc.arg(seller_uuid), sqlc.arg(buyer_uuid)
);

-- name: SaveDocumentLineItem :exec
INSERT INTO billing.document_line_items (
    line_item_uuid, document_uuid, name, quantity, line_item_type,
    unit_net_amount, unit_tax_amount, unit_gross_amount,
    net_amount, tax_amount, gross_amount,
    tax_rate, tax_type
) VALUES (
    sqlc.arg(line_item_uuid),
    sqlc.arg(document_uuid),
    sqlc.arg(name),
    sqlc.arg(quantity),
    sqlc.arg(line_item_type),
    sqlc.arg(unit_net_amount),
    sqlc.arg(unit_tax_amount),
    sqlc.arg(unit_gross_amount),
    sqlc.arg(net_amount),
    sqlc.arg(tax_amount),
    sqlc.arg(gross_amount),
    sqlc.arg(tax_rate),
    sqlc.arg(tax_type)
);

-- name: SaveDocumentTax :exec
INSERT INTO billing.document_taxes (document_uuid, tax_type, tax_rate, net_amount, tax_amount
) VALUES (
    sqlc.arg(document_uuid),
    sqlc.arg(tax_type),
    sqlc.arg(tax_rate),
    sqlc.arg(net_amount),
    sqlc.arg(tax_amount)
 );

-- name: UpdateDocumentFileUrl :exec
UPDATE billing.documents
SET file_url = sqlc.arg(file_url)
WHERE document_uuid = sqlc.arg(document_uuid);

-- name: GetDocument :one
SELECT sqlc.embed(documents), sqlc.embed(seller), sqlc.embed(buyer)
FROM billing.documents AS documents
INNER JOIN billing.legal_entity_snapshots seller ON seller.snapshot_uuid = documents.seller_uuid
INNER JOIN billing.legal_entity_snapshots buyer ON buyer.snapshot_uuid = documents.buyer_uuid
WHERE documents.document_uuid = $1 LIMIT 1;

-- name: GetDocumentByExternalReference :one
SELECT * FROM billing.documents WHERE external_reference = $1;

-- name: GetDocumentLineItems :many
SELECT * from billing.document_line_items
WHERE document_uuid = $1;

-- name: GetDocumentTaxes :many
SELECT document_uuid, tax_rate, tax_type, net_amount, tax_amount from billing.document_taxes
WHERE document_uuid = $1;
