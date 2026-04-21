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
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/require"

	"eats/backend/common"
	"eats/backend/common/shared"
	"eats/backend/common/testutils"
	"eats/backend/orders/adapters/db"
	"eats/backend/orders/adapters/db/dbmodels"
	"eats/backend/orders/api/http"
)

func TestRegisterCustomer(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPool := testutils.NewDB(t)
	customerRepo := db.NewCustomerRepository(dbPool)

	// Create a customer
	customerUUID := common.NewUUIDv7()
	customer := http.RegisterCustomer{
		Name:  gofakeit.Name(),
		Email: openapi_types.Email(gofakeit.Email()),
		Address: http.Address{
			Line1:       gofakeit.Street(),
			Line2:       "10",
			PostalCode:  gofakeit.Zip(),
			City:        gofakeit.City(),
			CountryCode: shared.MustNewCountryCode("US"),
		},
		PhoneNumber: gofakeit.Phone(),
	}

	err := customerRepo.RegisterCustomer(ctx, customerUUID, customer)
	require.NoError(t, err)

	queries := dbmodels.New(dbPool)

	dbCustomer, err := queries.GetCustomerByUUID(ctx, customerUUID)
	require.NoError(t, err)

	if diff := cmp.Diff(
		dbmodels.OrdersCustomer{
			CustomerUuid: customerUUID,
			Name:         customer.Name,
			Email:        string(customer.Email),
			Address: shared.Address{
				Line1:       customer.Address.Line1,
				Line2:       customer.Address.Line2,
				PostalCode:  customer.Address.PostalCode,
				City:        customer.Address.City,
				CountryCode: customer.Address.CountryCode,
			},
			PhoneNumber: customer.PhoneNumber,
		},
		dbCustomer,
		cmpopts.EquateComparable(shared.SharedTypes...),
	); diff != "" {
		t.Errorf("customer mismatch (-want +got):\n%s", diff)
	}
}
