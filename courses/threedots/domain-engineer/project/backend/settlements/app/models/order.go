package models

import (
	"context"
	"errors"
	"time"

	"github.com/shopspring/decimal"

	"eats/backend/billing/api/module/client"
	"eats/backend/common"
	"eats/backend/common/shared"
	"eats/backend/settlements/domain"
)

// CommissionRate is the platform's cut of items.Net (food + beverage).
var CommissionRate = decimal.New(20, -2) // 20%

type OrderUUID struct {
	common.UUID
}

type OrderRepository interface {
	SaveOrder(ctx context.Context, order Order) error
}

// Order is the settlements-side aggregate that captures everything we learn
// when StartSettlement runs: the parties involved, the receipt totals, and
// the platform commission.
type Order struct {
	orderUUID      OrderUUID
	restaurantUUID domain.LegalEntityUUID
	courierUUID    domain.LegalEntityUUID
	currency       shared.Currency

	itemsBreakdown    AmountBreakdown
	deliveryBreakdown AmountBreakdown
	totalBreakdown    AmountBreakdown

	commissionNetAmount decimal.Decimal

	orderedAt time.Time
}

// AmountBreakdown holds a triple of (net, tax, gross) amounts.
type AmountBreakdown struct {
	Net   decimal.Decimal
	Tax   decimal.Decimal
	Gross decimal.Decimal
}

func NewAmountBreakdown() AmountBreakdown {
	return AmountBreakdown{
		Net:   decimal.Zero,
		Tax:   decimal.Zero,
		Gross: decimal.Zero,
	}
}

func (o AmountBreakdown) Add(net, tax, gross decimal.Decimal) AmountBreakdown {
	return AmountBreakdown{
		Net:   o.Net.Add(net),
		Tax:   o.Tax.Add(tax),
		Gross: o.Gross.Add(gross),
	}
}

func NewOrder(
	orderUUID OrderUUID,
	restaurantUUID domain.LegalEntityUUID,
	courierUUID domain.LegalEntityUUID,
	currency shared.Currency,
	orderedAt time.Time,
	receipt client.DocumentReadModel,
) (Order, error) {
	if orderUUID.IsZero() {
		return Order{}, errors.New("order uuid is zero")
	}
	if restaurantUUID.IsZero() {
		return Order{}, errors.New("restaurant uuid is zero")
	}
	if courierUUID.IsZero() {
		return Order{}, errors.New("courier uuid is zero")
	}
	if currency.IsZero() {
		return Order{}, errors.New("currency is zero")
	}
	if orderedAt.IsZero() {
		return Order{}, errors.New("ordered at is zero")
	}

	deliveryBreakdown := NewAmountBreakdown()
	itemsBreakdown := NewAmountBreakdown()

	for _, lineItem := range receipt.LineItems {
		switch lineItem.Type {
		case shared.LineItemTypeDelivery:
			deliveryBreakdown = deliveryBreakdown.Add(lineItem.NetAmount, lineItem.TaxAmount, lineItem.GrossAmount)
		case shared.LineItemTypeFood:
			fallthrough
		case shared.LineItemTypeBeverage:
			itemsBreakdown = itemsBreakdown.Add(lineItem.NetAmount, lineItem.TaxAmount, lineItem.GrossAmount)
		}
	}

	commissionNetAmount := itemsBreakdown.Net.Mul(CommissionRate).Round(2)

	totalBreakdown := NewAmountBreakdown()
	totalBreakdown = totalBreakdown.Add(
		receipt.NetTotal,
		receipt.TaxTotal,
		receipt.GrossTotal,
	)

	return Order{
		orderUUID:           orderUUID,
		restaurantUUID:      restaurantUUID,
		courierUUID:         courierUUID,
		currency:            currency,
		itemsBreakdown:      itemsBreakdown,
		deliveryBreakdown:   deliveryBreakdown,
		totalBreakdown:      totalBreakdown,
		commissionNetAmount: commissionNetAmount,
		orderedAt:           orderedAt,
	}, nil
}

func (o Order) UUID() OrderUUID {
	return o.orderUUID
}

func (o Order) ShortID() string {
	return o.orderUUID.String()[len(o.orderUUID.String())-8:]
}

func (o Order) RestaurantUUID() domain.LegalEntityUUID {
	return o.restaurantUUID
}

func (o Order) CourierUUID() domain.LegalEntityUUID {
	return o.courierUUID
}

func (o Order) Currency() shared.Currency {
	return o.currency
}

func (o Order) ItemsBreakdown() AmountBreakdown {
	return o.itemsBreakdown
}

func (o Order) DeliveryBreakdown() AmountBreakdown {
	return o.deliveryBreakdown
}

func (o Order) TotalBreakdown() AmountBreakdown {
	return o.totalBreakdown
}

func (o Order) CommissionNetAmount() decimal.Decimal {
	return o.commissionNetAmount
}

func (o Order) OrderedAt() time.Time {
	return o.orderedAt
}

// UnmarshalOrder rebuilds an Order from already-validated state.
func UnmarshalOrder(
	orderUUID OrderUUID,
	restaurantUUID domain.LegalEntityUUID,
	courierUUID domain.LegalEntityUUID,
	currency shared.Currency,
	itemsBreakdown AmountBreakdown,
	deliveryBreakdown AmountBreakdown,
	totalBreakdown AmountBreakdown,
	commissionNetAmount decimal.Decimal,
	orderedAt time.Time,
) Order {
	return Order{
		orderUUID:           orderUUID,
		restaurantUUID:      restaurantUUID,
		courierUUID:         courierUUID,
		currency:            currency,
		commissionNetAmount: commissionNetAmount,
		itemsBreakdown:      itemsBreakdown,
		deliveryBreakdown:   deliveryBreakdown,
		totalBreakdown:      totalBreakdown,
		orderedAt:           orderedAt,
	}
}
