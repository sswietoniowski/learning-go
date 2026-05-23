-- name: SaveDocumentSeries :exec
INSERT INTO billing.document_series (prefix)
VALUES ($1)
ON CONFLICT DO NOTHING;

-- name: GetDocumentsBySeriesPrefix :many
SELECT document_number FROM billing.documents WHERE series_prefix = $1 ORDER BY document_number;
