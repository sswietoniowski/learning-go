package domain

import (
	"context"
	"errors"
	"fmt"
	"time"

	"eats/backend/common"
	"eats/backend/common/shared"
)

type DocumentType struct {
	common.Enum[DocumentTypeValues]
}

type DocumentTypeValues string

func (DocumentTypeValues) Values() []string {
	return []string{"receipt", "invoice"}
}

var (
	DocumentTypeReceipt = common.MustEnum[DocumentType]("receipt")
	DocumentTypeInvoice = common.MustEnum[DocumentType]("invoice")
)

type DocumentFactory struct {
	taxRateProvider TaxRateProvider
}

func NewDocumentFactory(taxRateProvider TaxRateProvider) *DocumentFactory {
	if taxRateProvider == nil {
		panic("taxRateProvider is required")
	}

	return &DocumentFactory{
		taxRateProvider: taxRateProvider,
	}
}

type DocumentRepository interface {
	DocumentByUUID(ctx context.Context, docUUID DocumentUUID) (*Document, error)
	CreateDocument(
		ctx context.Context,
		series DocumentSeries,
		createFunc func(documentNumber DocumentNumber) (*Document, error),
	) (DocumentUUID, error)
	UpdateFileUrl(ctx context.Context, docUUID DocumentUUID, fileUrl string) error
}

// DocumentBuilder separates inter-module calls (like tax rate lookups) from document creation.
// This way, Build() can run inside a database transaction without making inter-module calls.
type DocumentBuilder struct {
	externalReference *string
	documentType      DocumentType
	issueDate         time.Time
	currency          shared.Currency
	seller            LegalEntity
	buyer             LegalEntity
	lineItems         []LineItem
	summary           PriceBreakdownSummary
}

// Build creates a Document with the given document number.
// It's safe to call inside a database transaction: no external calls are made here.
func (b *DocumentBuilder) Build(docNumber DocumentNumber) (*Document, error) {
	return &Document{
		uuid:              DocumentUUID{common.NewUUIDv7()},
		externalReference: b.externalReference,
		documentType:      b.documentType,
		issueDate:         b.issueDate,
		documentNumber:    docNumber,
		currency:          b.currency,
		seller:            b.seller,
		buyer:             b.buyer,
		lineItems:         b.lineItems,
		summary:           b.summary,
	}, nil
}

// NewReceiptBuilder resolves all external data (tax rates) upfront,
// so Build() can safely run inside a database transaction.
func (f DocumentFactory) NewReceiptBuilder(ctx context.Context, data NewDocumentData) (*DocumentBuilder, error) {
	if data.Buyer.TaxID() != nil {
		return nil, errors.New("receipts cannot be issued to buyers with a tax ID")
	}

	return f.newDocumentBuilder(ctx, DocumentTypeReceipt, data)
}

// NewInvoiceBuilder resolves all external data (tax rates) upfront,
// so Build() can safely run inside a database transaction.
func (f DocumentFactory) NewInvoiceBuilder(ctx context.Context, data NewDocumentData) (*DocumentBuilder, error) {
	return f.newDocumentBuilder(ctx, DocumentTypeInvoice, data)
}

func (f DocumentFactory) newDocumentBuilder(
	ctx context.Context,
	docType DocumentType,
	data NewDocumentData,
) (*DocumentBuilder, error) {
	if data.Buyer.IsZero() {
		return nil, errors.New("buyer can't be empty")
	}
	if data.Seller.IsZero() {
		return nil, errors.New("seller can't be empty")
	}
	if data.Currency.IsZero() {
		return nil, errors.New("currency can't be empty")
	}
	if data.IssueDate.IsZero() {
		return nil, errors.New("issue date can't be empty")
	}
	tomorrow := time.Now().Truncate(24*time.Hour).AddDate(0, 0, 1)
	if !data.IssueDate.Before(tomorrow) {
		return nil, errors.New("issue date can't be in the future")
	}
	if data.Seller.TaxID() == nil {
		return nil, errors.New("seller must have a tax ID to issue billing documents")
	}

	if len(data.LineItems) == 0 {
		return nil, errors.New("document must have at least one line item")
	}

	buyerCountryCode := data.Buyer.Address().CountryCode()
	sellerCountryCode := data.Seller.Address().CountryCode()

	lineItems := make([]LineItem, 0, len(data.LineItems))
	for _, lineItemData := range data.LineItems {
		lineItem, err := f.newLineItem(
			ctx,
			lineItemData,
			buyerCountryCode,
			data.Buyer.TaxID(),
			sellerCountryCode,
			data.Currency,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create line item for document: %w", err)
		}

		lineItems = append(lineItems, lineItem)
	}

	summary := summarizeLineItems(lineItems)

	return &DocumentBuilder{
		externalReference: data.ExternalReference,
		documentType:      docType,
		issueDate:         data.IssueDate,
		currency:          data.Currency,
		seller:            data.Seller,
		buyer:             data.Buyer,
		lineItems:         lineItems,
		summary:           summary,
	}, nil
}

type NewDocumentData struct {
	ExternalReference *string
	IssueDate         time.Time
	Currency          shared.Currency
	Seller            LegalEntity
	Buyer             LegalEntity
	LineItems         []NewLineItemData
}

type NewLineItemData struct {
	Name         string
	LineItemType shared.LineItemType
	Quantity     int
	UnitAmount   shared.LineAmount
}

type DocumentUUID struct {
	common.UUID
}

type Document struct {
	uuid              DocumentUUID
	externalReference *string
	documentType      DocumentType
	issueDate         time.Time
	currency          shared.Currency
	documentNumber    DocumentNumber
	seller            LegalEntity
	buyer             LegalEntity
	lineItems         []LineItem
	summary           PriceBreakdownSummary
}

func (d *Document) UUID() DocumentUUID {
	return d.uuid
}

func (d *Document) ExternalReference() *string {
	return d.externalReference
}

func (d *Document) DocumentType() DocumentType {
	return d.documentType
}

func (d *Document) DocumentNumber() DocumentNumber {
	return d.documentNumber
}

func (d *Document) IssueDate() time.Time {
	return d.issueDate
}

func (d *Document) Currency() shared.Currency {
	return d.currency
}

func (d *Document) Seller() LegalEntity {
	return d.seller
}

func (d *Document) Buyer() LegalEntity {
	return d.buyer
}

func (d *Document) LineItems() []LineItem {
	return d.lineItems
}

func (d *Document) Summary() PriceBreakdownSummary {
	return d.summary
}

type LineItemUUID struct {
	common.UUID
}

type LineItem struct {
	uuid         LineItemUUID
	name         string
	lineItemType shared.LineItemType
	breakdown    PriceBreakdown
	quantity     int
}

func (f DocumentFactory) newLineItem(
	ctx context.Context,
	data NewLineItemData,
	buyerCountryCode shared.CountryCode,
	buyerTaxID *shared.TaxID,
	sellerCountryCode shared.CountryCode,
	currency shared.Currency,
) (LineItem, error) {
	if data.Name == "" {
		return LineItem{}, errors.New("name can't be empty")
	}

	if buyerCountryCode.IsZero() {
		return LineItem{}, errors.New("buyer country code cannot be empty")
	}

	if sellerCountryCode.IsZero() {
		return LineItem{}, errors.New("seller country code cannot be empty")
	}

	if data.LineItemType.IsZero() {
		return LineItem{}, errors.New("item type can't be zero")
	}

	if data.Quantity < 1 {
		return LineItem{}, errors.New("quantity must be positive")
	}

	if data.UnitAmount.Amount().IsNegative() {
		return LineItem{}, errors.New("unit amount can't be negative")
	}

	taxRateRequest := TaxRateRequest{
		BuyerCountryCode:  buyerCountryCode,
		BuyerTaxID:        buyerTaxID,
		SellerCountryCode: sellerCountryCode,
		LineItemType:      data.LineItemType,
		TransactionDate:   time.Now().UTC(),
	}

	taxRate, err := f.taxRateProvider.GetTaxRate(ctx, taxRateRequest)
	if err != nil {
		return LineItem{}, fmt.Errorf("could not get tax rate for item: %w", err)
	}

	var priceBreakdown PriceBreakdown

	if data.UnitAmount.IsGross() {
		priceBreakdown, err = NewPriceBreakdownFromGrossAmount(taxRate, data.UnitAmount.Amount(), currency, data.Quantity)
	} else {
		priceBreakdown, err = NewPriceBreakdownFromNetAmount(taxRate, data.UnitAmount.Amount(), currency, data.Quantity)
	}

	if err != nil {
		return LineItem{}, fmt.Errorf("failed to create price breakdown: %w", err)
	}

	return LineItem{
		uuid:         LineItemUUID{common.NewUUIDv7()},
		name:         data.Name,
		breakdown:    priceBreakdown,
		lineItemType: data.LineItemType,
		quantity:     data.Quantity,
	}, nil
}

func (l LineItem) UUID() LineItemUUID {
	return l.uuid
}

func (l LineItem) Name() string {
	return l.name
}

func (l LineItem) Quantity() int {
	return l.quantity
}

func (l LineItem) LineItemType() shared.LineItemType {
	return l.lineItemType
}

func (l LineItem) PriceBreakdown() PriceBreakdown {
	return l.breakdown
}

type TaxCalculationInput struct {
	BuyerCountryCode  shared.CountryCode
	BuyerTaxID        *shared.TaxID
	SellerCountryCode shared.CountryCode
	LineItems         []NewLineItemData
	Currency          shared.Currency
}

type TaxCalculation struct {
	lineItems []LineItem
	summary   PriceBreakdownSummary
}

func (f DocumentFactory) NewTaxCalculation(ctx context.Context, input TaxCalculationInput) (TaxCalculation, error) {
	lineItems := make([]LineItem, 0, len(input.LineItems))
	for _, li := range input.LineItems {
		lineItem, err := f.newLineItem(
			ctx,
			li,
			input.BuyerCountryCode,
			input.BuyerTaxID,
			input.SellerCountryCode,
			input.Currency,
		)
		if err != nil {
			return TaxCalculation{}, fmt.Errorf("could not create line item: %w", err)
		}

		lineItems = append(lineItems, lineItem)
	}

	summary := summarizeLineItems(lineItems)

	return TaxCalculation{
		lineItems: lineItems,
		summary:   summary,
	}, nil
}

func (t TaxCalculation) LineItems() []LineItem {
	return t.lineItems
}

func (t TaxCalculation) Summary() PriceBreakdownSummary {
	return t.summary
}
