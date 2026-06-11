// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
//go:build integration

package db_test

import (
	"context"
	"fmt"
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

	t.Run("CloseBillingCycle_restaurant", func(t *testing.T) {
		bc, _, err := repo.CloseBillingCycle(ctx, restaurant.UUID)
		require.NoError(t, err)
		assert.Equal(t, 1, bc.Number())
		assert.True(t, bc.Closed())

		// Close again (cycle 2)
		bc, _, err = repo.CloseBillingCycle(ctx, restaurant.UUID)
		require.NoError(t, err)
		assert.Equal(t, 2, bc.Number())
		assert.True(t, bc.Closed())

		// Once more (cycle 3)
		bc, _, err = repo.CloseBillingCycle(ctx, restaurant.UUID)
		require.NoError(t, err)
		assert.Equal(t, 3, bc.Number())
		assert.True(t, bc.Closed())
	})

	t.Run("CloseBillingCycle_courier", func(t *testing.T) {
		bc, _, err := repo.CloseBillingCycle(ctx, courier.UUID)
		require.NoError(t, err)
		assert.Equal(t, 1, bc.Number())
		assert.True(t, bc.Closed())

		bc, _, err = repo.CloseBillingCycle(ctx, courier.UUID)
		require.NoError(t, err)
		assert.Equal(t, 2, bc.Number())
		assert.True(t, bc.Closed())
	})

	t.Run("CalculateCommissionInvoiceData", func(t *testing.T) {
		invoice, err := repo.CalculateCommissionInvoiceData(ctx, restaurantBillingCycleUUID, platformUUID)
		require.NoError(t, err)
		assert.Equal(t, invoice.BuyerUUID, restaurant.UUID)
		assert.Equal(t, platformUUID.LegalEntityUUID, invoice.SellerUUID, "platform should be the seller of the commission invoice")
		require.Len(t, invoice.LineItems, 1, "commission invoice should have one aggregated line item")

		// Restaurant cycle has 5 orders (added above), each with commission=$20 set in
		// newTestOrder. The aggregation should sum the per-order commissions:
		//   NetAmount = 5 × $20 = $100.
		// Quantity is always 1 because NetAmount is already the aggregated total.
		// The order count goes into the line item name instead.
		// Catches: SQL missing SUM/GROUP BY, wrong column aggregated, returning a
		// single order's commission instead of the cycle total.
		expectedNet := decimal.NewFromFloat(100.0)
		assert.True(t, invoice.LineItems[0].NetAmount.Equal(expectedNet),
			"commission net amount: expected %s, got %s", expectedNet, invoice.LineItems[0].NetAmount)
		assert.Equal(t, 1, invoice.LineItems[0].Quantity, "quantity should be 1 because net amount is already the aggregated total")
		assert.Contains(t, invoice.LineItems[0].Name, fmt.Sprintf("%d orders", ordersCount), "line item name should contain the order count")

		_, err = repo.CalculateCommissionInvoiceData(ctx, courierBillingCycleUUID, platformUUID)
		require.Error(t, err, "Courier billing cycle should not have commission invoice")
	})

	t.Run("CalculateDeliveryInvoicesData", func(t *testing.T) {
		invoices, err := repo.CalculateDeliveryInvoicesData(ctx, courierBillingCycleUUID)
		require.NoError(t, err)
		require.Len(t, invoices, 3, "There should be 3 delivery invoices for 3 different restaurants")

		// Courier cycle has 5 orders across 3 restaurants, each with delivery net=$9.
		// Aggregated per-buyer, the totals should sum to 5 × $9 = $45 across 3 invoices
		// (1 + 2 + 2 orders by buyer restaurant). Catches: SQL summing the wrong column,
		// returning per-order rows instead of per-buyer aggregates, wrong seller assignment.
		totalNet := decimal.Zero
		for _, inv := range invoices {
			require.Len(t, inv.LineItems, 1, "each delivery invoice should have one aggregated line item")
			assert.Equal(t, courier.UUID, inv.SellerUUID, "courier should be the seller of every delivery invoice")
			assert.Equal(t, 1, inv.LineItems[0].Quantity, "quantity should be 1 because net amount is already the aggregated total")
			assert.Contains(t, inv.LineItems[0].Name, "orders", "line item name should contain the order count")
			totalNet = totalNet.Add(inv.LineItems[0].NetAmount)
		}
		expectedTotalNet := decimal.NewFromFloat(45.0)
		assert.True(t, totalNet.Equal(expectedTotalNet),
			"sum of delivery invoice net amounts: expected %s, got %s", expectedTotalNet, totalNet)

		invoices, err = repo.CalculateDeliveryInvoicesData(ctx, restaurantBillingCycleUUID)
		require.NoError(t, err)
		require.Len(t, invoices, 0, "Restaurant billing cycle should not have delivery invoices")
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

// TestAddOrderToCurrentBillingCycle_RaceWithClose tests the scenario where orders
// are being added concurrently while the billing cycle is being closed.
// The concern is: can an order end up in a closed billing cycle?
//
// Timeline of the race condition:
// 1. TX1 (AddOrder): BEGIN, reads current billing cycle (#1, open)
// 2. TX2 (Close): BEGIN, reads cycle #1, reads orders for snapshot, closes cycle #1, creates cycle #2, COMMIT
// 3. TX1: INSERT order into billing_cycle_orders for cycle #1, COMMIT
// Result: Order is in closed cycle #1 but wasn't included in the statement snapshot.
func TestAddOrderToCurrentBillingCycle_RaceWithClose(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPool := testutils.NewDB(t)

	partnerRepo := db.NewLegalEntityRepository(dbPool)
	orderRepo := db.NewOrderRepository(dbPool)
	repo := db.NewBillingCycleRepository(dbPool)

	platformUUID := newTestPlatformEntity(t, partnerRepo)
	restaurant := newTestPartner(t, partnerRepo, platformUUID, domain.PartnerTypeRestaurant)
	courier := newTestPartner(t, partnerRepo, platformUUID, domain.PartnerTypeCourier)

	// Run multiple iterations to verify the race condition is properly handled.
	// With Serializable isolation, conflicting transactions will retry automatically.
	const iterations = 5
	const ordersPerIteration = 10

	for iter := 0; iter < iterations; iter++ {
		// Create a fresh billing cycle for each iteration
		if iter > 0 {
			// Close the previous cycle to create a new one
			_, _, err := repo.CloseBillingCycle(ctx, restaurant.UUID)
			require.NoError(t, err)
		}

		// Get the current billing cycle UUID before we start
		currentCycleUUID := currentBillingCycleUUID(t, dbPool, restaurant.UUID)
		require.False(t, currentCycleUUID.IsZero(), "should have an open billing cycle")

		// Pre-create orders
		orders := make([]models.Order, ordersPerIteration)
		for i := 0; i < ordersPerIteration; i++ {
			orders[i] = newTestOrder(t, orderRepo, restaurant, courier)
		}

		// Track which orders were in the snapshot at close time
		var ordersInSnapshot []models.Order

		g := errgroup.Group{}

		// Goroutine: Close billing cycle at a random point
		g.Go(func() error {
			// Wait a bit to let some orders start adding
			time.Sleep(time.Duration(iter%5) * time.Millisecond)

			// Orders are read within the same serializable transaction, so the snapshot
			// is guaranteed to be consistent with the closed billing cycle.
			_, snapshotOrders, err := repo.CloseBillingCycle(ctx, restaurant.UUID)
			ordersInSnapshot = snapshotOrders
			return err
		})

		// Goroutines: Add orders concurrently
		for i := 0; i < ordersPerIteration; i++ {
			order := orders[i]
			g.Go(func() error {
				return repo.AddOrderToCurrentBillingCycle(ctx, restaurant.UUID, order.UUID())
			})
		}

		err := g.Wait()
		require.NoError(t, err)

		// Now check: are there orders in the closed billing cycle that weren't in the snapshot?
		ordersInClosedCycle, err := repo.BillingCycleOrders(ctx, currentCycleUUID)
		require.NoError(t, err)

		// Build sets for comparison
		snapshotOrderUUIDs := make(map[string]bool)
		for _, o := range ordersInSnapshot {
			snapshotOrderUUIDs[o.UUID().String()] = true
		}

		closedCycleOrderUUIDs := make(map[string]bool)
		for _, o := range ordersInClosedCycle {
			closedCycleOrderUUIDs[o.UUID().String()] = true
		}

		// Find orders that are in the closed cycle but weren't in the snapshot
		var missingFromSnapshot []string
		for uuid := range closedCycleOrderUUIDs {
			if !snapshotOrderUUIDs[uuid] {
				missingFromSnapshot = append(missingFromSnapshot, uuid)
			}
		}

		// This assertion verifies the race condition:
		// If there are orders in the closed billing cycle that weren't in the snapshot,
		// those orders would be "lost" - they're in a closed cycle but not in the statement.
		assert.Empty(t, missingFromSnapshot,
			"iteration %d: found %d orders in closed billing cycle that weren't in the snapshot - these orders would be lost!",
			iter, len(missingFromSnapshot))

		if len(missingFromSnapshot) > 0 {
			t.Logf("Race condition detected in iteration %d: %d orders in snapshot, %d in closed cycle, %d missing",
				iter, len(ordersInSnapshot), len(ordersInClosedCycle), len(missingFromSnapshot))
			// Fail fast on first detection to make debugging easier
			break
		}
	}
}

func TestCloseBillingCycle_Idempotent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPool := testutils.NewDB(t)

	partnerRepo := db.NewLegalEntityRepository(dbPool)
	orderRepo := db.NewOrderRepository(dbPool)
	repo := db.NewBillingCycleRepository(dbPool)

	platformUUID := newTestPlatformEntity(t, partnerRepo)
	restaurant := newTestPartner(t, partnerRepo, platformUUID, domain.PartnerTypeRestaurant)
	courier := newTestPartner(t, partnerRepo, platformUUID, domain.PartnerTypeCourier)

	// Add orders to cycle 1
	order := newTestOrder(t, orderRepo, restaurant, courier)
	err := repo.AddOrderToCurrentBillingCycle(ctx, restaurant.UUID, order.UUID())
	require.NoError(t, err)

	// First call: closes cycle 1, creates cycle 2
	bc, orders, err := repo.CloseBillingCycle(ctx, restaurant.UUID)
	require.NoError(t, err)
	assert.Equal(t, 1, bc.Number())
	assert.True(t, bc.Closed())
	assert.False(t, bc.Settled())
	assert.Len(t, orders, 1)

	// Settle cycle 1
	err = repo.SettleBillingCycle(ctx, bc.UUID())
	require.NoError(t, err)

	// Second call (retry): closes cycle 2 (empty), creates cycle 3
	bc2, orders2, err := repo.CloseBillingCycle(ctx, restaurant.UUID)
	require.NoError(t, err)
	assert.Equal(t, 2, bc2.Number())
	assert.True(t, bc2.Closed())
	assert.Len(t, orders2, 0)

	// Verify UnsettledClosedCycles returns only cycle 2 (cycle 1 is settled)
	unsettled, err := repo.UnsettledClosedCycles(ctx, restaurant.UUID)
	require.NoError(t, err)
	require.Len(t, unsettled, 1)
	assert.Equal(t, 2, unsettled[0].Number())

	// Cycle 2 has 0 orders
	cycle2Orders, err := repo.BillingCycleOrders(ctx, unsettled[0].UUID())
	require.NoError(t, err)
	assert.Len(t, cycle2Orders, 0)
}

func TestUnsettledClosedCycles(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPool := testutils.NewDB(t)

	partnerRepo := db.NewLegalEntityRepository(dbPool)
	orderRepo := db.NewOrderRepository(dbPool)
	repo := db.NewBillingCycleRepository(dbPool)

	platformUUID := newTestPlatformEntity(t, partnerRepo)
	restaurant := newTestPartner(t, partnerRepo, platformUUID, domain.PartnerTypeRestaurant)
	courier := newTestPartner(t, partnerRepo, platformUUID, domain.PartnerTypeCourier)

	// Add order and close cycle 1
	order := newTestOrder(t, orderRepo, restaurant, courier)
	err := repo.AddOrderToCurrentBillingCycle(ctx, restaurant.UUID, order.UUID())
	require.NoError(t, err)

	bc1, _, err := repo.CloseBillingCycle(ctx, restaurant.UUID)
	require.NoError(t, err)

	// Close cycle 2 (empty)
	bc2, _, err := repo.CloseBillingCycle(ctx, restaurant.UUID)
	require.NoError(t, err)

	// Close cycle 3 (empty)
	_, _, err = repo.CloseBillingCycle(ctx, restaurant.UUID)
	require.NoError(t, err)

	// All three closed cycles should be unsettled
	unsettled, err := repo.UnsettledClosedCycles(ctx, restaurant.UUID)
	require.NoError(t, err)
	require.Len(t, unsettled, 3)
	assert.Equal(t, 1, unsettled[0].Number())
	assert.Equal(t, 2, unsettled[1].Number())
	assert.Equal(t, 3, unsettled[2].Number())

	// Settle cycle 1
	err = repo.SettleBillingCycle(ctx, bc1.UUID())
	require.NoError(t, err)

	// Only cycle 2 and 3 should be unsettled now
	unsettled, err = repo.UnsettledClosedCycles(ctx, restaurant.UUID)
	require.NoError(t, err)
	require.Len(t, unsettled, 2)
	assert.Equal(t, 2, unsettled[0].Number())
	assert.True(t, unsettled[0].UUID().Equals(bc2.UUID().UUID))
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
