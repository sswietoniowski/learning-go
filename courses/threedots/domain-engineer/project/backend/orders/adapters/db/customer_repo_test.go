// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
//go:build integration

package db_test

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"

	"eats/backend/common"
	"eats/backend/common/shared"
	"eats/backend/common/testutils"
	"eats/backend/orders/adapters/db"
	"eats/backend/orders/adapters/db/dbmodels"
	"eats/backend/orders/app"
)

func TestRegisterCustomer(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPool := testutils.NewDB(t)
	customerRepo := db.NewCustomerRepository(dbPool)

	// Create a customer
	customerUUID := app.CustomerUUID{common.NewUUIDv7()}

	address, err := shared.NewAddress(gofakeit.Street(), "10", gofakeit.Zip(), gofakeit.City(), shared.MustNewCountryCode("US"))
	require.NoError(t, err)

	customer := app.Customer{
		CustomerUUID: customerUUID,
		Name:         gofakeit.Name(),
		Email:        gofakeit.Email(),
		Address:      address,
		PhoneNumber:  gofakeit.Phone(),
	}

	err = customerRepo.RegisterCustomer(ctx, customer)
	require.NoError(t, err)

	queries := dbmodels.New(dbPool)

	dbCustomer, err := queries.GetCustomerByUUID(ctx, customerUUID)
	require.NoError(t, err)

	if diff := cmp.Diff(
		dbmodels.OrdersCustomer{
			CustomerUuid: customerUUID,
			Name:         customer.Name,
			Email:        customer.Email,
			Address:      customer.Address,
			PhoneNumber:  customer.PhoneNumber,
		},
		dbCustomer,
		cmpopts.EquateComparable(shared.SharedTypes...),
	); diff != "" {
		t.Errorf("customer mismatch (-want +got):\n%s", diff)
	}
}
