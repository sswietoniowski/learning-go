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
	order := newTestOrder(t, restaurantUUID, courierUUID)

	require.NoError(t, repo.SaveOrder(ctx, order))

	assert.Equal(t, 1, countOrders(t, ctx, pool, order.UUID()))
	assert.Equal(t, 3, countOrderBreakdowns(t, ctx, pool, order.UUID()))
}

func TestSaveOrder_IsIdempotent_SameOrderUUID(t *testing.T) {
	ctx := context.Background()
	pool := testutils.NewDB(t)

	restaurantUUID, courierUUID := setupRestaurantAndCourier(t, ctx, pool)

	repo := db.NewOrderRepository(pool)
	order := newTestOrder(t, restaurantUUID, courierUUID)

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
	order := newTestOrder(t, bogusRestaurant, courierUUID)

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
	order := newTestOrder(t, restaurantUUID, bogusCourier)

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
	order := newTestOrder(t, restaurantUUID, courierUUID)

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

// setupRestaurantAndCourier persists a platform plus two partners (restaurant, courier)
// so SaveOrder's FK constraints can resolve.
func setupRestaurantAndCourier(t *testing.T, ctx context.Context, pool *pgxpool.Pool) (domain.LegalEntityUUID, domain.LegalEntityUUID) {
	t.Helper()

	repo := db.NewLegalEntityRepository(pool)

	platform := newOrderTestPlatform(t)
	require.NoError(t, repo.SavePlatformEntity(ctx, platform))

	restaurant := newOrderTestPartner(t, "Mama's Pizzeria", models.PlatformEntityUUID{LegalEntityUUID: platform.UUID})
	require.NoError(t, repo.SavePartner(ctx, restaurant))

	courier := newOrderTestPartner(t, "Speedy Couriers", models.PlatformEntityUUID{LegalEntityUUID: platform.UUID})
	require.NoError(t, repo.SavePartner(ctx, courier))

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

func newTestOrder(t *testing.T, restaurantUUID, courierUUID domain.LegalEntityUUID) models.Order {
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
