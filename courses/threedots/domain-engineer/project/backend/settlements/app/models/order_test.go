// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package models_test

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"eats/backend/billing/api/module/client"
	"eats/backend/common"
	"eats/backend/common/shared"
	"eats/backend/settlements/app/models"
	"eats/backend/settlements/domain"
)

func validOrderArgs(t *testing.T) (
	models.OrderUUID,
	domain.LegalEntityUUID,
	domain.LegalEntityUUID,
	shared.Currency,
	time.Time,
	client.DocumentReadModel,
) {
	t.Helper()

	orderUUID := models.OrderUUID{UUID: common.NewUUIDv7()}
	restaurantUUID := domain.LegalEntityUUID{UUID: common.NewUUIDv7()}
	courierUUID := domain.LegalEntityUUID{UUID: common.NewUUIDv7()}
	currency := shared.MustNewCurrency("EUR")
	orderedAt := time.Now()

	receipt := client.DocumentReadModel{
		UUID:           "doc-uuid",
		DocumentNumber: "INV/2026/01/0001",
		LineItems: []client.LineItemReadModel{
			{
				Name:        "Pizza",
				Type:        shared.LineItemTypeFood,
				Quantity:    1,
				NetAmount:   decimal.NewFromFloat(10.00),
				TaxAmount:   decimal.NewFromFloat(2.30),
				GrossAmount: decimal.NewFromFloat(12.30),
			},
			{
				Name:        "Cola",
				Type:        shared.LineItemTypeBeverage,
				Quantity:    1,
				NetAmount:   decimal.NewFromFloat(2.50),
				TaxAmount:   decimal.NewFromFloat(0.58),
				GrossAmount: decimal.NewFromFloat(3.08),
			},
			{
				Name:        "Delivery",
				Type:        shared.LineItemTypeDelivery,
				Quantity:    1,
				NetAmount:   decimal.NewFromFloat(4.00),
				TaxAmount:   decimal.NewFromFloat(0.92),
				GrossAmount: decimal.NewFromFloat(4.92),
			},
		},
		NetTotal:   decimal.NewFromFloat(16.50),
		TaxTotal:   decimal.NewFromFloat(3.80),
		GrossTotal: decimal.NewFromFloat(20.30),
	}

	return orderUUID, restaurantUUID, courierUUID, currency, orderedAt, receipt
}

func TestNewOrder_Valid(t *testing.T) {
	orderUUID, restaurantUUID, courierUUID, currency, orderedAt, receipt := validOrderArgs(t)

	order, err := models.NewOrder(orderUUID, restaurantUUID, courierUUID, currency, orderedAt, receipt)
	require.NoError(t, err)

	assert.Equal(t, orderUUID, order.UUID())
	assert.Equal(t, restaurantUUID, order.RestaurantUUID())
	assert.Equal(t, courierUUID, order.CourierUUID())
	assert.Equal(t, currency, order.Currency())
	assert.Equal(t, orderedAt, order.OrderedAt())
}

func TestNewOrder_RejectsZeroFields(t *testing.T) {
	t.Run("zero_order_uuid", func(t *testing.T) {
		_, restaurantUUID, courierUUID, currency, orderedAt, receipt := validOrderArgs(t)
		_, err := models.NewOrder(models.OrderUUID{}, restaurantUUID, courierUUID, currency, orderedAt, receipt)
		require.Error(t, err)
	})

	t.Run("zero_restaurant_uuid", func(t *testing.T) {
		orderUUID, _, courierUUID, currency, orderedAt, receipt := validOrderArgs(t)
		_, err := models.NewOrder(orderUUID, domain.LegalEntityUUID{}, courierUUID, currency, orderedAt, receipt)
		require.Error(t, err)
	})

	t.Run("zero_courier_uuid", func(t *testing.T) {
		orderUUID, restaurantUUID, _, currency, orderedAt, receipt := validOrderArgs(t)
		_, err := models.NewOrder(orderUUID, restaurantUUID, domain.LegalEntityUUID{}, currency, orderedAt, receipt)
		require.Error(t, err)
	})

	t.Run("zero_currency", func(t *testing.T) {
		orderUUID, restaurantUUID, courierUUID, _, orderedAt, receipt := validOrderArgs(t)
		_, err := models.NewOrder(orderUUID, restaurantUUID, courierUUID, shared.Currency{}, orderedAt, receipt)
		require.Error(t, err)
	})

	t.Run("zero_ordered_at", func(t *testing.T) {
		orderUUID, restaurantUUID, courierUUID, currency, _, receipt := validOrderArgs(t)
		_, err := models.NewOrder(orderUUID, restaurantUUID, courierUUID, currency, time.Time{}, receipt)
		require.Error(t, err)
	})
}

func TestNewOrder_ItemsBreakdown_AggregatesFoodAndBeverage(t *testing.T) {
	orderUUID, restaurantUUID, courierUUID, currency, orderedAt, receipt := validOrderArgs(t)

	order, err := models.NewOrder(orderUUID, restaurantUUID, courierUUID, currency, orderedAt, receipt)
	require.NoError(t, err)

	// Food (10.00 + 2.30 + 12.30) + Beverage (2.50 + 0.58 + 3.08)
	assert.True(t, decimal.NewFromFloat(12.50).Equal(order.ItemsBreakdown().Net), "items net %s", order.ItemsBreakdown().Net)
	assert.True(t, decimal.NewFromFloat(2.88).Equal(order.ItemsBreakdown().Tax), "items tax %s", order.ItemsBreakdown().Tax)
	assert.True(t, decimal.NewFromFloat(15.38).Equal(order.ItemsBreakdown().Gross), "items gross %s", order.ItemsBreakdown().Gross)
}

func TestNewOrder_DeliveryBreakdown_AggregatesDeliveryLineItems(t *testing.T) {
	orderUUID, restaurantUUID, courierUUID, currency, orderedAt, receipt := validOrderArgs(t)

	order, err := models.NewOrder(orderUUID, restaurantUUID, courierUUID, currency, orderedAt, receipt)
	require.NoError(t, err)

	assert.True(t, decimal.NewFromFloat(4.00).Equal(order.DeliveryBreakdown().Net), "delivery net %s", order.DeliveryBreakdown().Net)
	assert.True(t, decimal.NewFromFloat(0.92).Equal(order.DeliveryBreakdown().Tax), "delivery tax %s", order.DeliveryBreakdown().Tax)
	assert.True(t, decimal.NewFromFloat(4.92).Equal(order.DeliveryBreakdown().Gross), "delivery gross %s", order.DeliveryBreakdown().Gross)
}

func TestNewOrder_TotalBreakdown_MatchesReceiptTotals(t *testing.T) {
	orderUUID, restaurantUUID, courierUUID, currency, orderedAt, receipt := validOrderArgs(t)

	order, err := models.NewOrder(orderUUID, restaurantUUID, courierUUID, currency, orderedAt, receipt)
	require.NoError(t, err)

	assert.True(t, receipt.NetTotal.Equal(order.TotalBreakdown().Net), "total net")
	assert.True(t, receipt.TaxTotal.Equal(order.TotalBreakdown().Tax), "total tax")
	assert.True(t, receipt.GrossTotal.Equal(order.TotalBreakdown().Gross), "total gross")
}

func TestNewOrder_Commission_IsTwentyPercentOfItemsNet_RoundedToCents(t *testing.T) {
	orderUUID, restaurantUUID, courierUUID, currency, orderedAt, receipt := validOrderArgs(t)

	order, err := models.NewOrder(orderUUID, restaurantUUID, courierUUID, currency, orderedAt, receipt)
	require.NoError(t, err)

	// items.Net = 12.50; 20% = 2.50
	assert.True(t, decimal.NewFromFloat(2.50).Equal(order.CommissionNetAmount()), "commission %s", order.CommissionNetAmount())
}

func TestNewOrder_Commission_RoundsToCents(t *testing.T) {
	orderUUID, restaurantUUID, courierUUID, currency, orderedAt, _ := validOrderArgs(t)

	// items.Net = 13.33; 20% = 2.666 -> rounded to 2.67
	receipt := client.DocumentReadModel{
		LineItems: []client.LineItemReadModel{
			{
				Type:        shared.LineItemTypeFood,
				NetAmount:   decimal.NewFromFloat(13.33),
				TaxAmount:   decimal.Zero,
				GrossAmount: decimal.NewFromFloat(13.33),
			},
		},
		NetTotal:   decimal.NewFromFloat(13.33),
		TaxTotal:   decimal.Zero,
		GrossTotal: decimal.NewFromFloat(13.33),
	}

	order, err := models.NewOrder(orderUUID, restaurantUUID, courierUUID, currency, orderedAt, receipt)
	require.NoError(t, err)

	assert.True(t, decimal.NewFromFloat(2.67).Equal(order.CommissionNetAmount()), "commission %s", order.CommissionNetAmount())
}

func TestAmountBreakdown_Add_IsPure(t *testing.T) {
	original := models.NewAmountBreakdown()
	added := original.Add(decimal.NewFromFloat(1.00), decimal.NewFromFloat(0.23), decimal.NewFromFloat(1.23))

	// Original must remain zero (Add returns a new value).
	assert.True(t, original.Net.IsZero())
	assert.True(t, original.Tax.IsZero())
	assert.True(t, original.Gross.IsZero())

	assert.True(t, decimal.NewFromFloat(1.00).Equal(added.Net))
	assert.True(t, decimal.NewFromFloat(0.23).Equal(added.Tax))
	assert.True(t, decimal.NewFromFloat(1.23).Equal(added.Gross))
}

func TestUnmarshalOrder_RoundTrips(t *testing.T) {
	orderUUID := models.OrderUUID{UUID: common.NewUUIDv7()}
	restaurantUUID := domain.LegalEntityUUID{UUID: common.NewUUIDv7()}
	courierUUID := domain.LegalEntityUUID{UUID: common.NewUUIDv7()}
	currency := shared.MustNewCurrency("EUR")
	orderedAt := time.Now()

	items := models.AmountBreakdown{
		Net:   decimal.NewFromFloat(10.00),
		Tax:   decimal.NewFromFloat(2.30),
		Gross: decimal.NewFromFloat(12.30),
	}
	delivery := models.AmountBreakdown{
		Net:   decimal.NewFromFloat(4.00),
		Tax:   decimal.NewFromFloat(0.92),
		Gross: decimal.NewFromFloat(4.92),
	}
	total := models.AmountBreakdown{
		Net:   decimal.NewFromFloat(14.00),
		Tax:   decimal.NewFromFloat(3.22),
		Gross: decimal.NewFromFloat(17.22),
	}
	commission := decimal.NewFromFloat(2.00)

	order := models.UnmarshalOrder(
		orderUUID,
		restaurantUUID,
		courierUUID,
		currency,
		items,
		delivery,
		total,
		commission,
		orderedAt,
	)

	assert.Equal(t, orderUUID, order.UUID())
	assert.Equal(t, restaurantUUID, order.RestaurantUUID())
	assert.Equal(t, courierUUID, order.CourierUUID())
	assert.Equal(t, currency, order.Currency())
	assert.Equal(t, items, order.ItemsBreakdown())
	assert.Equal(t, delivery, order.DeliveryBreakdown())
	assert.Equal(t, total, order.TotalBreakdown())
	assert.True(t, commission.Equal(order.CommissionNetAmount()))
	assert.Equal(t, orderedAt, order.OrderedAt())
}
