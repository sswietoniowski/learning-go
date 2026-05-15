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
	createFunc func(documentNumber domain.DocumentNumber) (*domain.Document, error),
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

		doc, err := createFunc(docNumber)
		if err != nil {
			return fmt.Errorf("error creating document: %w", err)
		}

		docUUID = doc.UUID()
		if doc.ExternalReference() != nil {
			externalReference = *doc.ExternalReference()
		}

		sellerUUID := common.NewUUIDv7()
		buyerUUID := common.NewUUIDv7()

		err = queries.SaveLegalEntitySnapshot(ctx, dbmodels.SaveLegalEntitySnapshotParams{
			SnapshotUuid: sellerUUID,
			Name:         doc.Seller().Name(),
			Address:      doc.Seller().Address(),
			TaxID:        doc.Seller().TaxID(),
		})
		if err != nil {
			return fmt.Errorf("error saving seller legal entity: %w", err)
		}

		err = queries.SaveLegalEntitySnapshot(ctx, dbmodels.SaveLegalEntitySnapshotParams{
			SnapshotUuid: buyerUUID,
			Name:         doc.Buyer().Name(),
			Address:      doc.Buyer().Address(),
			TaxID:        doc.Buyer().TaxID(),
		})
		if err != nil {
			return fmt.Errorf("error saving buyer legal entity: %w", err)
		}

		err = queries.SaveDocument(ctx, dbmodels.SaveDocumentParams{
			DocumentUuid:      doc.UUID(),
			ExternalReference: doc.ExternalReference(),
			DocumentNumber:    doc.DocumentNumber().String(),
			SeriesPrefix:      series.String(),
			DocumentType:      doc.DocumentType(),
			IssueDate:         doc.IssueDate(),
			Currency:          doc.Currency(),
			TotalNetAmount:    doc.Summary().NetAmount(),
			TotalTaxAmount:    doc.Summary().TaxAmount(),
			TotalGrossAmount:  doc.Summary().GrossAmount(),
			SellerUuid:        sellerUUID,
			BuyerUuid:         buyerUUID,
		})
		if err != nil {
			return fmt.Errorf("error saving document: %w", err)
		}

		for _, lineItem := range doc.LineItems() {
			err := queries.SaveDocumentLineItem(ctx, dbmodels.SaveDocumentLineItemParams{
				LineItemUuid:    lineItem.UUID(),
				DocumentUuid:    doc.UUID(),
				Name:            lineItem.Name(),
				Quantity:        int32(lineItem.Quantity()),
				UnitNetAmount:   lineItem.PriceBreakdown().UnitNetAmount(),
				UnitTaxAmount:   lineItem.PriceBreakdown().UnitTaxAmount(),
				UnitGrossAmount: lineItem.PriceBreakdown().UnitGrossAmount(),
				NetAmount:       lineItem.PriceBreakdown().NetAmount(),
				TaxAmount:       lineItem.PriceBreakdown().TaxAmount(),
				GrossAmount:     lineItem.PriceBreakdown().GrossAmount(),
				TaxRate:         lineItem.PriceBreakdown().TaxRate().Rate(),
				TaxType:         lineItem.PriceBreakdown().TaxRate().TaxType(),
			})
			if err != nil {
				return fmt.Errorf("error saving document line item: %w", err)
			}
		}

		for _, tax := range doc.Summary().Taxes() {
			err := queries.SaveDocumentTax(ctx, dbmodels.SaveDocumentTaxParams{
				DocumentUuid: doc.UUID(),
				TaxType:      tax.TaxRate().TaxType(),
				TaxRate:      tax.TaxRate().Rate(),
				NetAmount:    tax.NetAmount(),
				TaxAmount:    tax.TaxAmount(),
			})
			if err != nil {
				return fmt.Errorf("error saving document tax: %w", err)
			}
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

func (r *PostgresRepository) DocumentByUUID(ctx context.Context, docUUID domain.DocumentUUID) (*domain.Document, error) {
	queries := dbmodels.New(r.db)

	dbDoc, err := queries.GetDocument(ctx, docUUID)
	if err != nil {
		return nil, fmt.Errorf("error getting document by uuid: %w", err)
	}

	dbLineItems, err := queries.GetDocumentLineItems(ctx, docUUID)
	if err != nil {
		return nil, fmt.Errorf("error getting document line items: %w", err)
	}

	var lineItems []domain.LineItem
	for _, dbLineItem := range dbLineItems {
		taxRate := domain.UnmarshalTaxRate(dbLineItem.TaxRate, dbLineItem.TaxType)
		breakdown := domain.UnmarshalPriceBreakdown(
			taxRate,
			dbLineItem.UnitNetAmount,
			dbLineItem.UnitTaxAmount,
			dbLineItem.UnitGrossAmount,
			dbLineItem.NetAmount,
			dbLineItem.TaxAmount,
			dbLineItem.GrossAmount,
		)

		lineItems = append(lineItems, domain.UnmarshalLineItem(
			dbLineItem.LineItemUuid,
			dbLineItem.Name,
			breakdown,
			int(dbLineItem.Quantity),
		))
	}

	var taxes []domain.TaxSummary

	dbTaxes, err := queries.GetDocumentTaxes(ctx, docUUID)
	if err != nil {
		return nil, fmt.Errorf("error getting document taxes: %w", err)
	}

	for _, dbTax := range dbTaxes {
		taxRate := domain.UnmarshalTaxRate(dbTax.TaxRate, dbTax.TaxType)
		taxes = append(taxes, domain.UnmarshalTaxSummary(
			taxRate,
			dbTax.NetAmount,
			dbTax.TaxAmount,
		))
	}

	summary := domain.UnmarshalPriceBreakdownSummary(
		dbDoc.BillingDocument.TotalNetAmount,
		dbDoc.BillingDocument.TotalTaxAmount,
		dbDoc.BillingDocument.TotalGrossAmount,
		taxes,
	)

	docSeries, err := domain.NewDocumentSeries(dbDoc.BillingDocument.SeriesPrefix)

	docNumber, err := domain.UnmarshalDocumentNumber(docSeries, dbDoc.BillingDocument.DocumentNumber)
	if err != nil {
		return nil, fmt.Errorf("error creating document number: %w", err)
	}

	seller, err := domain.NewLegalEntity(
		dbDoc.BillingLegalEntitySnapshot.Name,
		dbDoc.BillingLegalEntitySnapshot.Address,
		dbDoc.BillingLegalEntitySnapshot.TaxID,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating seller legal entity: %w", err)
	}

	buyer, err := domain.NewLegalEntity(
		dbDoc.BillingLegalEntitySnapshot_2.Name,
		dbDoc.BillingLegalEntitySnapshot_2.Address,
		dbDoc.BillingLegalEntitySnapshot_2.TaxID,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating buyer legal entity: %w", err)
	}

	doc := domain.UnmarshalDocument(
		dbDoc.BillingDocument.DocumentUuid,
		dbDoc.BillingDocument.ExternalReference,
		docNumber,
		dbDoc.BillingDocument.DocumentType,
		dbDoc.BillingDocument.IssueDate,
		dbDoc.BillingDocument.Currency,
		seller,
		buyer,
		lineItems,
		summary,
	)

	return doc, nil
}

func (r *PostgresRepository) UpdateFileUrl(ctx context.Context, docUUID domain.DocumentUUID, fileUrl string) error {
	queries := dbmodels.New(r.db)

	err := queries.UpdateDocumentFileUrl(ctx, dbmodels.UpdateDocumentFileUrlParams{
		DocumentUuid: docUUID,
		FileUrl:      &fileUrl,
	})
	if err != nil {
		return fmt.Errorf("error updating document file url: %w", err)
	}

	return nil
}
