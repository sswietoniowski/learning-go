// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
//go:build integration

package db_test

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"eats/backend/common"
	"eats/backend/common/testutils"
	"eats/backend/orders/adapters/db"
	"eats/backend/orders/adapters/db/dbmodels"
	"eats/backend/orders/app"
)

func TestRegisterCourier(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPool := testutils.NewDB(t)
	courierRepo := db.NewCourierRepository(dbPool)

	courierUUID := app.CourierUUID{UUID: common.NewUUIDv7()}
	courier := app.Courier{
		CourierUUID: courierUUID,
		Name:        gofakeit.Name(),
		PhoneNumber: gofakeit.Phone(),
		City:        gofakeit.City(),
	}

	err := courierRepo.RegisterCourier(ctx, courier)
	require.NoError(t, err)

	queries := dbmodels.New(dbPool)

	dbCourier, err := queries.GetCourierByUUID(ctx, courierUUID)
	require.NoError(t, err)

	if diff := cmp.Diff(
		dbmodels.OrdersCourier{
			CourierUuid: courierUUID,
			Name:        courier.Name,
			PhoneNumber: courier.PhoneNumber,
			City:        courier.City,
		},
		dbCourier,
	); diff != "" {
		t.Errorf("courier mismatch (-want +got):\n%s", diff)
	}
}
