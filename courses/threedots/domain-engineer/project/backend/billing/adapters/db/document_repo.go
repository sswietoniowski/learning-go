package db

import (
	"context"
	"fmt"

	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"eats/backend/billing/adapters/db/dbmodels"
	"eats/backend/billing/domain"
	"eats/backend/common"
	"eats/backend/common/log"
)

type DocumentRecord struct {
	UUID              domain.DocumentUUID
	ExternalReference *string
}

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}

func (r *PostgresRepository) CreateDocument(
	ctx context.Context,
	series domain.DocumentSeries,
	createFunc func(documentNumber domain.DocumentNumber) (DocumentRecord, error),
) (domain.DocumentUUID, error) {
	var docUUID domain.DocumentUUID
	var externalReference string

	// ReadCommitted is safe here: NextDocumentNumber uses "last_number = last_number + 1",
	// so PostgreSQL re-evaluates the expression on the current row value after acquiring
	// the row lock. There is no lost update risk because no value is read into Go memory
	// and written back. RepeatableRead would cause serialization errors under contention.
	err := common.UpdateInReadCommittedTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		queries := dbmodels.New(tx)

		nextNumber, err := queries.NextDocumentNumber(ctx, series.String())
		if err != nil {
			return fmt.Errorf("error getting next document number: %w", err)
		}

		docNumber, err := domain.NewDocumentNumber(series, int(nextNumber))
		if err != nil {
			return fmt.Errorf("error creating document number: %w", err)
		}

		record, err := createFunc(docNumber)
		if err != nil {
			return fmt.Errorf("error creating document: %w", err)
		}

		docUUID = record.UUID
		if record.ExternalReference != nil {
			externalReference = *record.ExternalReference
		}

		err = queries.SaveDocument(ctx, dbmodels.SaveDocumentParams{
			DocumentUuid:      record.UUID,
			ExternalReference: record.ExternalReference,
			DocumentNumber:    docNumber.String(),
			SeriesPrefix:      series.String(),
		})
		if err != nil {
			return fmt.Errorf("error saving document: %w", err)
		}

		return nil
	})

	// We can't handle this with ON CONFLICT - we have to cancel the transaction not to generate a new document number
	if common.IsUniqueViolationError(err, "documents_external_reference_key") {
		logger := log.FromContext(ctx)
		logger.With("external_reference", externalReference).Info("Skipping document creation due to existing external reference")

		// This is outside of transaction, but it's okay - this is a read operation
		dbDoc, err := r.getDocumentByExternalReference(ctx, externalReference)
		if err != nil {
			return domain.DocumentUUID{}, fmt.Errorf("error retrieving existing document by external reference: %w", err)
		}

		return dbDoc.DocumentUuid, nil
	}
	if err != nil {
		return domain.DocumentUUID{}, err
	}

	return docUUID, nil
}

func (r *PostgresRepository) getDocumentByExternalReference(ctx context.Context, externalRef string) (dbmodels.BillingDocument, error) {
	queries := dbmodels.New(r.db)

	dbDoc, err := queries.GetDocumentByExternalReference(ctx, &externalRef)
	if err != nil {
		return dbmodels.BillingDocument{}, fmt.Errorf("error getting document by external reference: %w", err)
	}

	return dbDoc, nil
}
