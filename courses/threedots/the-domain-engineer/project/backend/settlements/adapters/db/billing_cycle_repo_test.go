// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
//go:build integration

package db_test

import (
	"context"
	"testing"
	"time"

	gofakeit "github.com/brianvoe/gofakeit/v7"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"

	"eats/backend/common"
	"eats/backend/common/shared"
	"eats/backend/common/testutils"
	"eats/backend/settlements/adapters/db"
	"eats/backend/settlements/app/models"
	"eats/backend/settlements/domain"
)

func TestBillingCycleRepository_Lifecycle(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	dbPool := testutils.NewDB(t)

	partnerRepo := db.NewLegalEntityRepository(dbPool)
	orderRepo := db.NewOrderRepository(dbPool)
	repo := db.NewBillingCycleRepository(dbPool)

	platformUUID := newTestPlatformEntity(t, partnerRepo)
	restaurant := newTestPartner(t, partnerRepo, platformUUID, domain.PartnerTypeRestaurant)
	courier := newTestPartner(t, partnerRepo, platformUUID, domain.PartnerTypeCourier)

	// Billing cycle UUIDs from DB (created by newTestPartner).
	restaurantBillingCycleUUID := currentBillingCycleUUID(t, dbPool, restaurant.UUID)
	courierBillingCycleUUID := currentBillingCycleUUID(t, dbPool, courier.UUID)

	ordersCount := 5

	t.Run("AddOrderToCurrentBillingCycle", func(t *testing.T) {
		g := errgroup.Group{}
		for i := 0; i < ordersCount; i++ {
			g.Go(func() error {
				order := newTestOrder(t, orderRepo, restaurant, courier)
				return repo.AddOrderToCurrentBillingCycle(ctx, restaurant.UUID, order.UUID())
			})
		}

		err := g.Wait()
		require.NoError(t, err)

		restaurant2 := newTestPartner(t, partnerRepo, platformUUID, domain.PartnerTypeRestaurant)
		restaurant3 := newTestPartner(t, partnerRepo, platformUUID, domain.PartnerTypeRestaurant)

		g = errgroup.Group{}
		for i := 0; i < ordersCount; i++ {
			g.Go(func() error {
				var rest models.LegalEntity

				// Create orders for different restaurants, so we can test aggregation later
				if i == 0 {
					rest = restaurant
				} else if i <= 2 {
					rest = restaurant2
				} else {
					rest = restaurant3
				}

				order := newTestOrder(t, orderRepo, rest, courier)

				return repo.AddOrderToCurrentBillingCycle(ctx, courier.UUID, order.UUID())
			})
		}

		err = g.Wait()
		require.NoError(t, err)
	})

	t.Run("BillingCycleOrders", func(t *testing.T) {
		restaurantOrders, err := repo.BillingCycleOrders(ctx, restaurantBillingCycleUUID)
		require.NoError(t, err)

		courierOrders, err := repo.BillingCycleOrders(ctx, courierBillingCycleUUID)
		require.NoError(t, err)

		require.Len(t, restaurantOrders, ordersCount)
		require.Len(t, courierOrders, ordersCount)
	})
}

// TestAddOrderToCurrentBillingCycle_Concurrent verifies that concurrent calls to
// AddOrderToCurrentBillingCycle correctly add all orders to the billing cycle.
// AddOrderToCurrentBillingCycle uses Serializable isolation to prevent race conditions
// with CloseBillingCycle (see TestAddOrderToCurrentBillingCycle_RaceWithClose).
// Conflicting serializable transactions are retried with exponential backoff.
func TestAddOrderToCurrentBillingCycle_Concurrent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPool := testutils.NewDB(t)

	partnerRepo := db.NewLegalEntityRepository(dbPool)
	orderRepo := db.NewOrderRepository(dbPool)
	repo := db.NewBillingCycleRepository(dbPool)

	platformUUID := newTestPlatformEntity(t, partnerRepo)
	restaurant := newTestPartner(t, partnerRepo, platformUUID, domain.PartnerTypeRestaurant)
	courier := newTestPartner(t, partnerRepo, platformUUID, domain.PartnerTypeCourier)

	// Get the billing cycle created by newTestPartner.
	billingCycleUUID := currentBillingCycleUUID(t, dbPool, restaurant.UUID)

	// Use a higher number of concurrent operations to stress test the isolation level
	concurrentOrders := 20

	// Pre-create orders so we can track their UUIDs
	orders := make([]models.Order, concurrentOrders)
	for i := 0; i < concurrentOrders; i++ {
		orders[i] = newTestOrder(t, orderRepo, restaurant, courier)
	}

	// Concurrently add all orders to the billing cycle
	g := errgroup.Group{}
	for i := 0; i < concurrentOrders; i++ {
		order := orders[i]
		g.Go(func() error {
			return repo.AddOrderToCurrentBillingCycle(ctx, restaurant.UUID, order.UUID())
		})
	}

	err := g.Wait()
	require.NoError(t, err)

	// Verify all orders were added to the billing cycle
	billingCycleOrders, err := repo.BillingCycleOrders(ctx, billingCycleUUID)
	require.NoError(t, err)

	assert.Len(t, billingCycleOrders, concurrentOrders, "all concurrent orders should be in the billing cycle")

	// Verify each order is present
	orderUUIDs := make(map[string]bool)
	for _, o := range billingCycleOrders {
		orderUUIDs[o.UUID().String()] = true
	}

	for _, order := range orders {
		assert.True(t, orderUUIDs[order.UUID().String()], "order %s should be in billing cycle", order.UUID())
	}
}

// TestAddOrderToCurrentBillingCycle_Idempotent verifies that adding the same order
// to a billing cycle multiple times is idempotent (does not error and results in
// only one entry).
func TestAddOrderToCurrentBillingCycle_Idempotent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPool := testutils.NewDB(t)

	partnerRepo := db.NewLegalEntityRepository(dbPool)
	orderRepo := db.NewOrderRepository(dbPool)
	repo := db.NewBillingCycleRepository(dbPool)

	platformUUID := newTestPlatformEntity(t, partnerRepo)
	restaurant := newTestPartner(t, partnerRepo, platformUUID, domain.PartnerTypeRestaurant)
	courier := newTestPartner(t, partnerRepo, platformUUID, domain.PartnerTypeCourier)

	// Get the billing cycle created by newTestPartner.
	billingCycleUUID := currentBillingCycleUUID(t, dbPool, restaurant.UUID)

	// Create a single order
	order := newTestOrder(t, orderRepo, restaurant, courier)

	// Add the same order multiple times - should be idempotent
	for i := 0; i < 3; i++ {
		err := repo.AddOrderToCurrentBillingCycle(ctx, restaurant.UUID, order.UUID())
		require.NoError(t, err, "adding order attempt %d should not error", i+1)
	}

	// Verify only one order is in the billing cycle
	billingCycleOrders, err := repo.BillingCycleOrders(ctx, billingCycleUUID)
	require.NoError(t, err)

	assert.Len(t, billingCycleOrders, 1, "should have exactly one order despite multiple adds")
	assert.True(t, billingCycleOrders[0].UUID().Equals(order.UUID().UUID), "the order should be the one we added")
}

func newTestPartner(
	t *testing.T,
	legalEntityRepo models.LegalEntityRepository,
	platformEntityUUID models.PlatformEntityUUID,
	partnerType domain.PartnerType,
) models.LegalEntity {
	le := newTestLegalEntity(t)

	partner := models.NewPartner(le, platformEntityUUID)

	billingCycle, err := domain.NewInitialBillingCycle(partner.LegalEntity.UUID, partnerType)
	require.NoError(t, err)

	err = legalEntityRepo.SavePartner(context.Background(), partner, billingCycle)
	require.NoError(t, err)

	return le
}

func newTestPlatformEntity(
	t *testing.T,
	legalEntityRepo models.LegalEntityRepository,
) models.PlatformEntityUUID {
	le := newTestLegalEntity(t)

	err := legalEntityRepo.SavePlatformEntity(context.Background(), le)
	require.NoError(t, err)

	return models.PlatformEntityUUID{le.UUID}
}

func newTestLegalEntity(t *testing.T) models.LegalEntity {
	address, err := shared.NewAddress(
		gofakeit.Street(),
		"",
		gofakeit.City(),
		gofakeit.Zip(),
		shared.MustNewCountryCode("US"),
	)
	require.NoError(t, err)

	taxID, err := shared.NewTaxID(gofakeit.Numerify("#########"))
	require.NoError(t, err)

	le, err := models.NewLegalEntity(
		domain.LegalEntityUUID{common.NewUUIDv7()},
		models.LegalEntityPartner,
		gofakeit.Company(),
		taxID,
		address,
		domain.UnmarshalIBAN(gofakeit.Numerify("## #### #### #### ####")),
		shared.MustNewCurrency("PLN"),
	)
	require.NoError(t, err)

	return le
}

func newTestOrder(t *testing.T, repo models.OrderRepository, restaurant models.LegalEntity, courier models.LegalEntity) models.Order {
	order := models.UnmarshalOrder(
		models.OrderUUID{common.NewUUIDv7()},
		restaurant.UUID,
		courier.UUID,
		shared.MustNewCurrency("EUR"),
		models.AmountBreakdown{
			Net:   decimal.NewFromFloat(100.0),
			Tax:   decimal.NewFromFloat(10.0),
			Gross: decimal.NewFromFloat(110.0),
		},
		models.AmountBreakdown{
			Net:   decimal.NewFromFloat(9.0),
			Tax:   decimal.NewFromFloat(0.8),
			Gross: decimal.NewFromFloat(9.8),
		},
		models.AmountBreakdown{
			Net:   decimal.NewFromFloat(109.0),
			Tax:   decimal.NewFromFloat(10.8),
			Gross: decimal.NewFromFloat(119.8),
		},
		decimal.NewFromFloat(20.0),
		time.Now(),
	)

	err := repo.SaveOrder(context.Background(), order)
	require.NoError(t, err)

	return order
}

func currentBillingCycleUUID(t *testing.T, dbPool *pgxpool.Pool, partnerUUID domain.LegalEntityUUID) domain.BillingCycleUUID {
	t.Helper()

	var billingCycleUUID domain.BillingCycleUUID
	err := dbPool.QueryRow(
		context.Background(),
		"SELECT billing_cycle_uuid FROM settlements.billing_cycles WHERE partner_uuid = $1 AND closed = false ORDER BY billing_cycle_number DESC LIMIT 1",
		partnerUUID,
	).Scan(&billingCycleUUID)
	require.NoError(t, err)

	return billingCycleUUID
}
