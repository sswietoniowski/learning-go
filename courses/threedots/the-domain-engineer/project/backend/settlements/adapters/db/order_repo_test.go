// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
//go:build integration

package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"eats/backend/billing/api/module/client"
	"eats/backend/common"
	"eats/backend/common/shared"
	"eats/backend/common/testutils"
	"eats/backend/settlements/adapters/db"
	"eats/backend/settlements/app/models"
	"eats/backend/settlements/domain"
)

func TestSaveOrder_PersistsOrderAndBreakdownsAtomically(t *testing.T) {
	ctx := context.Background()
	pool := testutils.NewDB(t)

	restaurantUUID, courierUUID := setupRestaurantAndCourier(t, ctx, pool)

	repo := db.NewOrderRepository(pool)
	order := newOrderTestOrder(t, restaurantUUID, courierUUID)

	require.NoError(t, repo.SaveOrder(ctx, order))

	assert.Equal(t, 1, countOrders(t, ctx, pool, order.UUID()))
	assert.Equal(t, 3, countOrderBreakdowns(t, ctx, pool, order.UUID()))
}

func TestSaveOrder_IsIdempotent_SameOrderUUID(t *testing.T) {
	ctx := context.Background()
	pool := testutils.NewDB(t)

	restaurantUUID, courierUUID := setupRestaurantAndCourier(t, ctx, pool)

	repo := db.NewOrderRepository(pool)
	order := newOrderTestOrder(t, restaurantUUID, courierUUID)

	require.NoError(t, repo.SaveOrder(ctx, order))
	require.NoError(t, repo.SaveOrder(ctx, order))

	// ON CONFLICT DO NOTHING means a second call leaves row counts unchanged.
	assert.Equal(t, 1, countOrders(t, ctx, pool, order.UUID()))
	assert.Equal(t, 3, countOrderBreakdowns(t, ctx, pool, order.UUID()))
}

func TestSaveOrder_RejectsUnknownRestaurant(t *testing.T) {
	ctx := context.Background()
	pool := testutils.NewDB(t)

	_, courierUUID := setupRestaurantAndCourier(t, ctx, pool)
	bogusRestaurant := domain.LegalEntityUUID{UUID: common.NewUUIDv7()}

	repo := db.NewOrderRepository(pool)
	order := newOrderTestOrder(t, bogusRestaurant, courierUUID)

	err := repo.SaveOrder(ctx, order)
	require.Error(t, err)

	// Tx rolled back: no rows persisted.
	assert.Equal(t, 0, countOrders(t, ctx, pool, order.UUID()))
	assert.Equal(t, 0, countOrderBreakdowns(t, ctx, pool, order.UUID()))
}

func TestSaveOrder_RejectsUnknownCourier(t *testing.T) {
	ctx := context.Background()
	pool := testutils.NewDB(t)

	restaurantUUID, _ := setupRestaurantAndCourier(t, ctx, pool)
	bogusCourier := domain.LegalEntityUUID{UUID: common.NewUUIDv7()}

	repo := db.NewOrderRepository(pool)
	order := newOrderTestOrder(t, restaurantUUID, bogusCourier)

	err := repo.SaveOrder(ctx, order)
	require.Error(t, err)

	assert.Equal(t, 0, countOrders(t, ctx, pool, order.UUID()))
	assert.Equal(t, 0, countOrderBreakdowns(t, ctx, pool, order.UUID()))
}

func TestSaveOrder_PersistsAllThreeBreakdownTypes(t *testing.T) {
	ctx := context.Background()
	pool := testutils.NewDB(t)

	restaurantUUID, courierUUID := setupRestaurantAndCourier(t, ctx, pool)

	repo := db.NewOrderRepository(pool)
	order := newOrderTestOrder(t, restaurantUUID, courierUUID)

	require.NoError(t, repo.SaveOrder(ctx, order))

	for _, brType := range []string{"items", "delivery", "total"} {
		row := pool.QueryRow(ctx,
			`SELECT 1 FROM settlements.order_breakdowns WHERE order_uuid = $1 AND breakdown_type = $2`,
			order.UUID().UUID, brType,
		)
		var x int
		err := row.Scan(&x)
		require.NoError(t, err, "breakdown_type %s missing", brType)
	}
}

// TestSaveOrder_PersistsBreakdownAmounts verifies the persisted breakdown rows
// hold the exact net/tax/gross values from the order, not zeros or swapped columns.
// Without this, broken SaveOrder code (dropped precision, wrong column mapping,
// hardcoded values) surfaces only much later as a wrong payout amount during settle.
//
// The order is built inline rather than via newTestOrder to avoid coupling to a
// helper that's renamed in later exercises in the chain.
func TestSaveOrder_PersistsBreakdownAmounts(t *testing.T) {
	ctx := context.Background()
	pool := testutils.NewDB(t)

	restaurantUUID, courierUUID := setupRestaurantAndCourier(t, ctx, pool)

	// Expected amounts: food line net=10.00/tax=2.30/gross=12.30 → items breakdown,
	// delivery line net=4.00/tax=0.92/gross=4.92 → delivery breakdown, and
	// receipt totals net=14.00/tax=3.22/gross=17.22 → total breakdown.
	receipt := client.DocumentReadModel{
		UUID:           "doc-uuid",
		DocumentNumber: "INV/2026/01/0002",
		LineItems: []client.LineItemReadModel{
			{
				Type:        shared.LineItemTypeFood,
				NetAmount:   decimal.NewFromFloat(10.00),
				TaxAmount:   decimal.NewFromFloat(2.30),
				GrossAmount: decimal.NewFromFloat(12.30),
			},
			{
				Type:        shared.LineItemTypeDelivery,
				NetAmount:   decimal.NewFromFloat(4.00),
				TaxAmount:   decimal.NewFromFloat(0.92),
				GrossAmount: decimal.NewFromFloat(4.92),
			},
		},
		NetTotal:   decimal.NewFromFloat(14.00),
		TaxTotal:   decimal.NewFromFloat(3.22),
		GrossTotal: decimal.NewFromFloat(17.22),
	}

	order, err := models.NewOrder(
		models.OrderUUID{UUID: common.NewUUIDv7()},
		restaurantUUID,
		courierUUID,
		shared.MustNewCurrency("EUR"),
		time.Now(),
		receipt,
	)
	require.NoError(t, err)

	repo := db.NewOrderRepository(pool)
	require.NoError(t, repo.SaveOrder(ctx, order))

	expected := map[string]struct {
		net, tax, gross decimal.Decimal
	}{
		"items":    {decimal.NewFromFloat(10.00), decimal.NewFromFloat(2.30), decimal.NewFromFloat(12.30)},
		"delivery": {decimal.NewFromFloat(4.00), decimal.NewFromFloat(0.92), decimal.NewFromFloat(4.92)},
		"total":    {decimal.NewFromFloat(14.00), decimal.NewFromFloat(3.22), decimal.NewFromFloat(17.22)},
	}

	rows, err := pool.Query(ctx,
		`SELECT breakdown_type, net_amount, tax_amount, gross_amount
		 FROM settlements.order_breakdowns
		 WHERE order_uuid = $1`,
		order.UUID().UUID,
	)
	require.NoError(t, err)
	defer rows.Close()

	seen := map[string]bool{}
	for rows.Next() {
		var (
			brType                string
			netAmt, taxAmt, gross decimal.Decimal
		)
		require.NoError(t, rows.Scan(&brType, &netAmt, &taxAmt, &gross))

		want, ok := expected[brType]
		require.True(t, ok, "unexpected breakdown_type %q persisted", brType)
		seen[brType] = true

		assert.True(t, netAmt.Equal(want.net),
			"breakdown %s: net mismatch (got %s, want %s)", brType, netAmt, want.net)
		assert.True(t, taxAmt.Equal(want.tax),
			"breakdown %s: tax mismatch (got %s, want %s)", brType, taxAmt, want.tax)
		assert.True(t, gross.Equal(want.gross),
			"breakdown %s: gross mismatch (got %s, want %s)", brType, gross, want.gross)
	}
	require.NoError(t, rows.Err())

	for brType := range expected {
		assert.True(t, seen[brType], "breakdown_type %q never persisted", brType)
	}
}

// setupRestaurantAndCourier persists a platform plus two partners (restaurant, courier)
// so SaveOrder's FK constraints can resolve.
func setupRestaurantAndCourier(t *testing.T, ctx context.Context, pool *pgxpool.Pool) (domain.LegalEntityUUID, domain.LegalEntityUUID) {
	t.Helper()

	repo := db.NewLegalEntityRepository(pool)

	platform := newOrderTestPlatform(t)
	require.NoError(t, repo.SavePlatformEntity(ctx, platform))

	restaurant := newOrderTestPartner(t, "Mama's Pizzeria", models.PlatformEntityUUID{LegalEntityUUID: platform.UUID})
	restaurantBillingCycle, err := domain.NewInitialBillingCycle(restaurant.LegalEntity.UUID, domain.PartnerTypeRestaurant)
	require.NoError(t, err)
	require.NoError(t, repo.SavePartner(ctx, restaurant, restaurantBillingCycle))

	courier := newOrderTestPartner(t, "Speedy Couriers", models.PlatformEntityUUID{LegalEntityUUID: platform.UUID})
	courierBillingCycle, err := domain.NewInitialBillingCycle(courier.LegalEntity.UUID, domain.PartnerTypeCourier)
	require.NoError(t, err)
	require.NoError(t, repo.SavePartner(ctx, courier, courierBillingCycle))

	return restaurant.LegalEntity.UUID, courier.LegalEntity.UUID
}

func newOrderTestPlatform(t *testing.T) models.LegalEntity {
	t.Helper()

	taxID, err := shared.NewTaxID("PL1111111111")
	require.NoError(t, err)

	address, err := shared.NewAddress("Platform 1", "", "00-100", "Warsaw", shared.MustNewCountryCode("PL"))
	require.NoError(t, err)

	iban, err := domain.NewIBAN("DE89370400440532013222")
	require.NoError(t, err)

	entity, err := models.NewLegalEntity(
		domain.LegalEntityUUID{UUID: common.NewUUIDv7()},
		models.LegalEntityPlatform,
		"Three Dots Eats Platform",
		taxID,
		address,
		iban,
		shared.MustNewCurrency("EUR"),
	)
	require.NoError(t, err)

	return entity
}

func newOrderTestPartner(t *testing.T, name string, platformUUID models.PlatformEntityUUID) models.Partner {
	t.Helper()

	taxID, err := shared.NewTaxID("PL2222222222")
	require.NoError(t, err)

	address, err := shared.NewAddress("Partner 1", "", "00-200", "Warsaw", shared.MustNewCountryCode("PL"))
	require.NoError(t, err)

	iban, err := domain.NewIBAN("DE89370400440532013333")
	require.NoError(t, err)

	legalEntity, err := models.NewLegalEntity(
		domain.LegalEntityUUID{UUID: common.NewUUIDv7()},
		models.LegalEntityPartner,
		name,
		taxID,
		address,
		iban,
		shared.MustNewCurrency("EUR"),
	)
	require.NoError(t, err)

	return models.NewPartner(legalEntity, platformUUID)
}

func newOrderTestOrder(t *testing.T, restaurantUUID, courierUUID domain.LegalEntityUUID) models.Order {
	t.Helper()

	receipt := client.DocumentReadModel{
		UUID:           "doc-uuid",
		DocumentNumber: "INV/2026/01/0001",
		LineItems: []client.LineItemReadModel{
			{
				Type:        shared.LineItemTypeFood,
				NetAmount:   decimal.NewFromFloat(10.00),
				TaxAmount:   decimal.NewFromFloat(2.30),
				GrossAmount: decimal.NewFromFloat(12.30),
			},
			{
				Type:        shared.LineItemTypeDelivery,
				NetAmount:   decimal.NewFromFloat(4.00),
				TaxAmount:   decimal.NewFromFloat(0.92),
				GrossAmount: decimal.NewFromFloat(4.92),
			},
		},
		NetTotal:   decimal.NewFromFloat(14.00),
		TaxTotal:   decimal.NewFromFloat(3.22),
		GrossTotal: decimal.NewFromFloat(17.22),
	}

	order, err := models.NewOrder(
		models.OrderUUID{UUID: common.NewUUIDv7()},
		restaurantUUID,
		courierUUID,
		shared.MustNewCurrency("EUR"),
		time.Now(),
		receipt,
	)
	require.NoError(t, err)

	return order
}

func countOrders(t *testing.T, ctx context.Context, pool *pgxpool.Pool, orderUUID models.OrderUUID) int {
	t.Helper()

	var n int
	err := pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM settlements.orders WHERE order_uuid = $1`,
		orderUUID.UUID,
	).Scan(&n)
	require.NoError(t, err)
	return n
}

func countOrderBreakdowns(t *testing.T, ctx context.Context, pool *pgxpool.Pool, orderUUID models.OrderUUID) int {
	t.Helper()

	var n int
	err := pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM settlements.order_breakdowns WHERE order_uuid = $1`,
		orderUUID.UUID,
	).Scan(&n)
	require.NoError(t, err)
	return n
}
