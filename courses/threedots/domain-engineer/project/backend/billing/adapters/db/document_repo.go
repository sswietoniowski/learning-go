package db

import (
	"context"
	"fmt"

	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"eats/backend/billing/adapters/db/dbmodels"
	"eats/backend/billing/domain"
	"eats/backend/common"
)

type DocumentRecord struct {
	UUID domain.DocumentUUID
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

		err = queries.SaveDocument(ctx, dbmodels.SaveDocumentParams{
			DocumentUuid:   record.UUID,
			DocumentNumber: docNumber.String(),
			SeriesPrefix:   series.String(),
		})
		if err != nil {
			return fmt.Errorf("error saving document: %w", err)
		}

		return nil
	})
	if err != nil {
		return domain.DocumentUUID{}, err
	}

	return docUUID, nil
}
