// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
//go:build integration

package db_test

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"eats/backend/billing/adapters/db"
	"eats/backend/billing/adapters/db/dbtests"
	"eats/backend/billing/adapters/tax"
	"eats/backend/billing/domain"
	"eats/backend/common"
	"eats/backend/common/shared"
	"eats/backend/common/testutils"
)

func TestCreateDocument_ConcurrentDocumentNumbers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	database := testutils.NewDB(t)

	repo := db.NewPostgresRepository(database)

	seriesStr := common.NewUUIDv7().String()
	series, err := domain.NewDocumentSeries(seriesStr)
	require.NoError(t, err)

	q := dbtests.New(database)
	err = q.SaveDocumentSeries(ctx, seriesStr)
	require.NoError(t, err)

	documentData := domain.NewDocumentData{
		IssueDate: time.Now(),
		Currency:  shared.MustNewCurrency("USD"),
		Seller:    newLegalEntity(t),
		Buyer:     newLegalEntity(t),
		LineItems: []domain.NewLineItemData{
			{
				Name:         "Test Item",
				LineItemType: shared.LineItemTypeFood,
				Quantity:     1,
				UnitAmount:   shared.NewGrossAmount(decimal.RequireFromString("10.00")),
			},
		},
	}

	wg := sync.WaitGroup{}

	concurrentDocuments := 100

	f := domain.NewDocumentFactory(tax.NewStub())

	for i := 0; i < concurrentDocuments; i++ {
		wg.Go(func() {
			builder, builderErr := f.NewInvoiceBuilder(ctx, documentData)
			assert.NoError(t, builderErr)

			_, err := repo.CreateDocument(ctx, series, func(documentNumber domain.DocumentNumber) (*domain.Document, error) {
				return builder.Build(documentNumber)
			})
			assert.NoError(t, err)
		})
	}

	wg.Wait()

	docs, err := q.GetDocumentsBySeriesPrefix(ctx, series.String())
	require.NoError(t, err)

	require.Len(t, docs, concurrentDocuments)

	// Assert no sequence gaps and no duplicates
	for i := 0; i < concurrentDocuments; i++ {
		doc := docs[i]
		docNumber := strings.ReplaceAll(doc, seriesStr+"-", "")
		number, err := strconv.ParseInt(docNumber, 10, 64)
		assert.NoError(t, err)

		assert.Equal(t, i+1, int(number))
	}
}

func TestCreateDocument_ExternalReference(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	database := testutils.NewDB(t)

	repo := db.NewPostgresRepository(database)

	seriesStr := common.NewUUIDv7().String()
	series, err := domain.NewDocumentSeries(seriesStr)
	require.NoError(t, err)

	q := dbtests.New(database)
	err = q.SaveDocumentSeries(ctx, seriesStr)
	require.NoError(t, err)

	externalRef := uuid.NewString()

	documentData := domain.NewDocumentData{
		ExternalReference: &externalRef,
		IssueDate:         time.Now(),
		Currency:          shared.MustNewCurrency("USD"),
		Seller:            newLegalEntity(t),
		Buyer:             newLegalEntity(t),
		LineItems: []domain.NewLineItemData{
			{
				Name:         "Test Item",
				LineItemType: shared.LineItemTypeFood,
				Quantity:     1,
				UnitAmount:   shared.NewGrossAmount(decimal.RequireFromString("10.00")),
			},
		},
	}

	f := domain.NewDocumentFactory(tax.NewStub())

	builder, err := f.NewInvoiceBuilder(ctx, documentData)
	require.NoError(t, err)

	docUUID, err := repo.CreateDocument(ctx, series, func(documentNumber domain.DocumentNumber) (*domain.Document, error) {
		return builder.Build(documentNumber)
	})
	require.NoError(t, err)

	doc, err := repo.DocumentByUUID(ctx, docUUID)
	require.NoError(t, err)
	assert.NotNil(t, doc.ExternalReference())
	assert.Equal(t, externalRef, *doc.ExternalReference())

	// Saving the document with the same external reference should be idempotent
	builder2, err := f.NewInvoiceBuilder(ctx, documentData)
	require.NoError(t, err)

	doc2UUID, err := repo.CreateDocument(ctx, series, func(documentNumber domain.DocumentNumber) (*domain.Document, error) {
		return builder2.Build(documentNumber)
	})
	require.NoError(t, err)

	assert.Equal(t, docUUID, doc2UUID, "expected the same document UUID to be returned for idempotent create")
}

func TestCreateDocument_WithLineItemsAndTaxes(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	database := testutils.NewDB(t)

	repo := db.NewPostgresRepository(database)

	seriesStr := common.NewUUIDv7().String()
	series, err := domain.NewDocumentSeries(seriesStr)
	require.NoError(t, err)

	q := dbtests.New(database)
	err = q.SaveDocumentSeries(ctx, seriesStr)
	require.NoError(t, err)

	seller := newLegalEntity(t)
	buyer := newLegalEntity(t)

	documentData := domain.NewDocumentData{
		IssueDate: time.Now(),
		Currency:  shared.MustNewCurrency("USD"),
		Seller:    seller,
		Buyer:     buyer,
		LineItems: []domain.NewLineItemData{
			{
				Name:         "Pizza Margherita",
				LineItemType: shared.LineItemTypeFood,
				Quantity:     2,
				UnitAmount:   shared.NewGrossAmount(decimal.RequireFromString("12.30")),
			},
			{
				Name:         "Delivery Fee",
				LineItemType: shared.LineItemTypeDelivery,
				Quantity:     1,
				UnitAmount:   shared.NewGrossAmount(decimal.RequireFromString("5.00")),
			},
		},
	}

	f := domain.NewDocumentFactory(tax.NewStub())

	builder, err := f.NewInvoiceBuilder(ctx, documentData)
	require.NoError(t, err)

	docUUID, err := repo.CreateDocument(ctx, series, func(documentNumber domain.DocumentNumber) (*domain.Document, error) {
		return builder.Build(documentNumber)
	})
	require.NoError(t, err)

	doc, err := repo.DocumentByUUID(ctx, docUUID)
	require.NoError(t, err)

	assert.Equal(t, domain.DocumentTypeInvoice, doc.DocumentType())
	assert.Len(t, doc.LineItems(), 2)

	assert.Equal(t, "Pizza Margherita", doc.LineItems()[0].Name())
	assert.Equal(t, 2, doc.LineItems()[0].Quantity())
	assert.Equal(t, shared.LineItemTypeFood, doc.LineItems()[0].LineItemType())

	assert.Equal(t, "Delivery Fee", doc.LineItems()[1].Name())
	assert.Equal(t, 1, doc.LineItems()[1].Quantity())
	assert.Equal(t, shared.LineItemTypeDelivery, doc.LineItems()[1].LineItemType())

	assert.True(t, doc.Summary().GrossAmount().GreaterThan(decimal.Zero))
	assert.True(t, doc.Summary().NetAmount().GreaterThan(decimal.Zero))

	assert.Equal(t, seller.Name(), doc.Seller().Name())
	assert.Equal(t, seller.Address().Line1(), doc.Seller().Address().Line1())
	assert.Equal(t, seller.Address().City(), doc.Seller().Address().City())
	assert.Equal(t, seller.Address().PostalCode(), doc.Seller().Address().PostalCode())
	assert.Equal(t, seller.Address().CountryCode(), doc.Seller().Address().CountryCode())
	require.NotNil(t, doc.Seller().TaxID())
	assert.Equal(t, seller.TaxID().String(), doc.Seller().TaxID().String())

	assert.Equal(t, buyer.Name(), doc.Buyer().Name())
	assert.Equal(t, buyer.Address().Line1(), doc.Buyer().Address().Line1())
	assert.Equal(t, buyer.Address().City(), doc.Buyer().Address().City())
	assert.Equal(t, buyer.Address().PostalCode(), doc.Buyer().Address().PostalCode())
	assert.Equal(t, buyer.Address().CountryCode(), doc.Buyer().Address().CountryCode())
	require.NotNil(t, doc.Buyer().TaxID())
	assert.Equal(t, buyer.TaxID().String(), doc.Buyer().TaxID().String())

	taxes := doc.Summary().Taxes()
	require.Len(t, taxes, 1)
	assert.True(t, taxes[0].TaxRate().Rate().Equal(decimal.NewFromFloat(0.23)))
	assert.True(t, taxes[0].NetAmount().GreaterThan(decimal.Zero))
	assert.True(t, taxes[0].TaxAmount().GreaterThan(decimal.Zero))
}

func newLegalEntity(t *testing.T) domain.LegalEntity {
	addr := gofakeit.Address()

	address, err := shared.NewAddress(addr.Street, "", addr.Zip, addr.City, shared.MustNewCountryCode("US"))
	require.NoError(t, err)

	taxID, err := shared.NewTaxID(gofakeit.Numerify("##########"))
	require.NoError(t, err)

	entity, err := domain.NewLegalEntity(
		gofakeit.Company(),
		address,
		&taxID,
	)
	require.NoError(t, err)

	return entity
}

func newBuyer(t *testing.T) domain.LegalEntity {
	addr := gofakeit.Address()

	address, err := shared.NewAddress(addr.Street, "", addr.Zip, addr.City, shared.MustNewCountryCode("US"))
	require.NoError(t, err)

	entity, err := domain.NewLegalEntity(
		gofakeit.Company(),
		address,
		nil,
	)
	require.NoError(t, err)

	return entity
}
