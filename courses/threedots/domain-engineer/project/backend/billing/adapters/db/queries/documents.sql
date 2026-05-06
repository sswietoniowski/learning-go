-- name: NextDocumentNumber :one
UPDATE billing.document_series
SET last_number = last_number + 1,
    updated_at = NOW()
WHERE prefix = $1
RETURNING last_number;

-- name: SaveDocument :exec
INSERT INTO billing.documents (
    document_uuid, document_number, series_prefix
)
VALUES (
    sqlc.arg(document_uuid), sqlc.arg(document_number), sqlc.arg(series_prefix)
);
