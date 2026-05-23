package printer

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	"sort"

	"github.com/shopspring/decimal"

	"eats/backend/billing/domain"
	"eats/backend/common"
	"eats/backend/common/shared"
)

//go:embed templates/*.html
var templateFS embed.FS

type Printer struct{}

func NewPrinter() *Printer {
	return &Printer{}
}

type PrinterData struct {
	DocumentNumber string
	DocumentType   string
	IssueDate      interface{}
	Currency       shared.Currency
	Seller         LegalEntity
	Buyer          LegalEntity
	LineItems      []LineData
	TaxSummaries   []TaxSummaryData
	NetTotal       Money
	TaxTotal       Money
	GrossTotal     Money
}

type LegalEntity struct {
	Name    string
	Address AddressData
	TaxID   *string
}

type AddressData struct {
	Line1       string
	Line2       string
	PostalCode  string
	City        string
	CountryCode string
}

type LineData struct {
	Name          string
	Quantity      int
	UnitNetAmount Money
	TaxType       string
	TaxRate       Percentage
	TaxAmount     Money
	GrossAmount   Money
	Currency      shared.Currency
}

type TaxSummaryData struct {
	TaxType   string
	Rate      Percentage
	NetAmount Money
	TaxAmount Money
}

func (p *Printer) PrintDocument(ctx context.Context, doc *domain.Document) ([]byte, error) {
	templateName, err := getTemplateName(doc.DocumentType())
	if err != nil {
		return nil, err
	}

	tmpl := template.New(templateName)

	tmpl, err = tmpl.ParseFS(templateFS, "templates/"+templateName)
	if err != nil {
		return nil, fmt.Errorf("error parsing template: %w", err)
	}

	data := p.buildPrinterData(doc)

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return nil, fmt.Errorf("error rendering template: %w", err)
	}

	return buf.Bytes(), nil
}

func getTemplateName(docType domain.DocumentType) (string, error) {
	switch docType {
	case domain.DocumentTypeReceipt:
		return "receipt.html", nil
	default:
		return "", fmt.Errorf("unknown template type: %s", docType)
	}
}

func (p *Printer) buildPrinterData(doc *domain.Document) PrinterData {
	newMoney := func(amount decimal.Decimal) Money {
		return Money{
			Amount:   amount,
			Currency: doc.Currency(),
		}
	}

	lineItems := make([]LineData, len(doc.LineItems()))
	for i, lineItem := range doc.LineItems() {
		lineItems[i] = LineData{
			Name:          lineItem.Name(),
			Quantity:      lineItem.Quantity(),
			UnitNetAmount: newMoney(lineItem.PriceBreakdown().UnitNetAmount()),
			TaxType:       lineItem.PriceBreakdown().TaxRate().TaxType().DisplayName(),
			TaxRate:       Percentage{Value: lineItem.PriceBreakdown().TaxRate().Rate()},
			TaxAmount:     newMoney(lineItem.PriceBreakdown().TaxAmount()),
			GrossAmount:   newMoney(lineItem.PriceBreakdown().GrossAmount()),
			Currency:      doc.Currency(),
		}
	}

	taxSummaries := []TaxSummaryData{}
	for _, ts := range doc.Summary().Taxes() {
		taxSummaries = append(taxSummaries, TaxSummaryData{
			TaxType:   ts.TaxRate().TaxType().DisplayName(),
			Rate:      Percentage{Value: ts.TaxRate().Rate()},
			NetAmount: newMoney(ts.NetAmount()),
			TaxAmount: newMoney(ts.TaxAmount()),
		})
	}

	sort.Slice(taxSummaries, func(i, j int) bool {
		if taxSummaries[i].TaxType == taxSummaries[j].TaxType {
			return taxSummaries[i].Rate.Value.LessThan(taxSummaries[j].Rate.Value)
		}

		return taxSummaries[i].TaxType < taxSummaries[j].TaxType
	})

	var sellerTaxID *string
	if t := doc.Seller().TaxID(); t != nil {
		sellerTaxID = common.ToPtr(t.String())
	}
	var buyerTaxID *string
	if t := doc.Buyer().TaxID(); t != nil {
		buyerTaxID = common.ToPtr(t.String())
	}

	return PrinterData{
		DocumentNumber: doc.DocumentNumber().String(),
		DocumentType:   doc.DocumentType().String(),
		IssueDate:      doc.IssueDate(),
		Currency:       doc.Currency(),
		Seller: LegalEntity{
			Name: doc.Seller().Name(),
			Address: AddressData{
				Line1:       doc.Seller().Address().Line1(),
				Line2:       doc.Seller().Address().Line2(),
				PostalCode:  doc.Seller().Address().PostalCode(),
				City:        doc.Seller().Address().City(),
				CountryCode: doc.Seller().Address().CountryCode().String(),
			},
			TaxID: sellerTaxID,
		},
		Buyer: LegalEntity{
			Name: doc.Buyer().Name(),
			Address: AddressData{
				Line1:       doc.Buyer().Address().Line1(),
				Line2:       doc.Buyer().Address().Line2(),
				PostalCode:  doc.Buyer().Address().PostalCode(),
				City:        doc.Buyer().Address().City(),
				CountryCode: doc.Buyer().Address().CountryCode().String(),
			},
			TaxID: buyerTaxID,
		},
		LineItems:    lineItems,
		NetTotal:     newMoney(doc.Summary().NetAmount()),
		TaxTotal:     newMoney(doc.Summary().TaxAmount()),
		GrossTotal:   newMoney(doc.Summary().GrossAmount()),
		TaxSummaries: taxSummaries,
	}
}

type Money struct {
	Amount   decimal.Decimal
	Currency shared.Currency
}

func (m Money) String() string {
	return m.Amount.StringFixed(int32(m.Currency.DecimalPlaces()))
}

type Percentage struct {
	Value decimal.Decimal
}

func (p Percentage) String() string {
	return p.Value.Mul(decimal.NewFromInt(100)).String() + "%"
}
