// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package tests_test

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"testing"
	"time"

	gofakeit "github.com/brianvoe/gofakeit/v7"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	billingclient "eats/backend/billing/api/http/client"
	"eats/backend/common"
	"eats/backend/common/shared"
	"eats/backend/common/testutils"
	ordersclient "eats/backend/orders/api/http/client"
	"eats/backend/orders/app"
	http2 "eats/backend/settlements/api/http"
	settlementclient "eats/backend/settlements/api/http/client"
	"eats/backend/settlements/app/models"
	"eats/backend/settlements/domain"
)

var orderCmpOpts = []cmp.Option{
	cmpopts.EquateComparable(shared.SharedTypes...),
	cmpopts.EquateApproxTime(time.Second),
}

func assertRestaurantMenuPublished(ctx context.Context, t *testing.T, clients testClients, restaurantUUID app.RestaurantUUID, restaurant ordersclient.OnboardRestaurant) {
	t.Helper()

	resp, err := clients.Orders.CustomerGetRestaurantMenuWithResponse(ctx, restaurantUUID, &ordersclient.CustomerGetRestaurantMenuParams{
		CustomerUUID: app.CustomerUUID{common.NewUUIDv7()},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())

	require.Empty(
		t,
		cmp.Diff(
			restaurant.MenuItems,
			resp.JSON200.Items,
			cmpopts.EquateComparable(app.ItemCategory{}),
			cmpopts.SortSlices(func(a, b ordersclient.MenuItem) bool {
				return a.Uuid.String() < b.Uuid.String()
			}),
		),
	)
	require.Equal(t, resp.JSON200.RestaurantName, restaurant.Name)
	require.Equal(t, resp.JSON200.Address, restaurant.Address)
	require.Equal(t, resp.JSON200.Description, restaurant.Description)
	require.Equal(t, resp.JSON200.Currency, restaurant.Currency)
}

type testPlatformEntity struct {
	UUID        models.PlatformEntityUUID
	BankAccount string
}

func createPlatformEntity(
	ctx context.Context,
	t *testing.T,
	clients testClients,
) testPlatformEntity {
	t.Helper()

	// Pre-generate the platform entity UUID so we can link the bank account to it
	platformUUID := domain.LegalEntityUUID{UUID: common.NewUUIDv7()}

	name := "Eats Platform LLC"
	address := testutils.GenerateRandomOpenapiAddress(shared.MustNewCountryCode("US"))
	taxID := gofakeit.Numerify("###-###-###")

	// Create bank account with merchant_id = platform UUID for payment routing
	bankAccount, _ := createBankAccount(ctx, t, platformUUID.String())

	request := settlementclient.CreatePlatformEntityJSONRequestBody{
		PlatformEntityUuid: platformUUID,
		Address:            settlementclient.Address(address),
		BankAccountIban:    bankAccount,
		BusinessName:       name,
		Currency:           shared.MustNewCurrency("USD"),
		TaxId:              taxID,
	}

	resp, err := clients.Settlements.CreatePlatformEntityWithResponse(
		ctx,
		&settlementclient.CreatePlatformEntityParams{
			OperatorUUID: common.NewUUIDv7(),
		},
		request,
	)
	require.NoError(t, err)

	require.Equal(t, http.StatusCreated, resp.StatusCode())
	require.NotNil(t, resp.JSON201)

	return testPlatformEntity{
		UUID:        resp.JSON201.PlatformEntityUuid,
		BankAccount: bankAccount,
	}
}

func createPlatformEntityWithCurrency(
	ctx context.Context,
	t *testing.T,
	clients testClients,
	currency shared.Currency,
) testPlatformEntity {
	t.Helper()

	platformUUID := domain.LegalEntityUUID{UUID: common.NewUUIDv7()}

	name := "Eats Platform LLC"
	address := testutils.GenerateRandomOpenapiAddress(shared.MustNewCountryCode("US"))
	taxID := gofakeit.Numerify("###-###-###")

	bankAccount, _ := createBankAccount(ctx, t, platformUUID.String())

	request := settlementclient.CreatePlatformEntityJSONRequestBody{
		PlatformEntityUuid: platformUUID,
		Address:            settlementclient.Address(address),
		BankAccountIban:    bankAccount,
		BusinessName:       name,
		Currency:           currency,
		TaxId:              taxID,
	}

	resp, err := clients.Settlements.CreatePlatformEntityWithResponse(
		ctx,
		&settlementclient.CreatePlatformEntityParams{
			OperatorUUID: common.NewUUIDv7(),
		},
		request,
	)
	require.NoError(t, err)

	require.Equal(t, http.StatusCreated, resp.StatusCode())
	require.NotNil(t, resp.JSON201)

	return testPlatformEntity{
		UUID:        resp.JSON201.PlatformEntityUuid,
		BankAccount: bankAccount,
	}
}

type testRestaurant struct {
	UUID        app.RestaurantUUID
	Data        ordersclient.OnboardRestaurant
	BankAccount string
}

func onboardRestaurant(
	ctx context.Context,
	t *testing.T,
	clients testClients,
	platformEntityUUID models.PlatformEntityUUID,
	country shared.CountryCode,
) testRestaurant {
	t.Helper()

	var category app.ItemCategory
	if rand.Intn(2) == 0 {
		category = app.ItemCategoryFood
	} else {
		category = app.ItemCategoryBeverage
	}

	var menuItems []ordersclient.MenuItem
	for i := 0; i < 5; i++ {
		menuItems = append(menuItems, ordersclient.MenuItem{
			Uuid:       app.RestaurantMenuItemUUID{common.NewUUIDv7()},
			Name:       gofakeit.Lunch(),
			Category:   category,
			GrossPrice: randomPrice(),
			Ordering:   rand.Float32(),
		})
	}

	name := ""
	if rand.Intn(2) == 0 {
		name += gofakeit.FirstName() + "'s "
	}
	name += gofakeit.HipsterWord()

	address := testutils.GenerateRandomOpenapiAddress(country)

	restaurantToCreate := ordersclient.OnboardRestaurant{
		Address:     address,
		Description: gofakeit.HipsterSentence(),
		MenuItems:   menuItems,
		Name:        cases.Title(language.Und).String(name),
		Currency:    currencyForCountry(t, country),
	}

	restaurantUUID := app.RestaurantUUID{common.NewUUIDv7()}
	resp, err := clients.Orders.OnboardRestaurantWithResponse(
		ctx,
		restaurantUUID,
		&ordersclient.OnboardRestaurantParams{
			OperatorUUID: common.NewUUIDv7(),
		},
		restaurantToCreate,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, resp.StatusCode())

	taxID := gofakeit.Numerify("###-###-###")

	bankAccount, _ := createBankAccount(ctx, t, restaurantUUID.String())

	partnerRequest := settlementclient.OnboardPartnerJSONRequestBody{
		PartnerUuid:        http2.LegalEntityUUID{restaurantUUID.UUID},
		PlatformEntityUuid: platformEntityUUID,
		PartnerType:        domain.PartnerTypeRestaurant,
		BusinessName:       addCompanySuffix(name),
		TaxId:              taxID,
		BankAccountIban:    bankAccount,
		Currency:           shared.MustNewCurrency("USD"),
		Address: settlementclient.Address{
			Line1:       address.Line1,
			Line2:       address.Line2,
			City:        address.City,
			PostalCode:  address.PostalCode,
			CountryCode: address.CountryCode,
		},
	}

	onboardResp, err := clients.Settlements.OnboardPartnerWithResponse(
		ctx,
		&settlementclient.OnboardPartnerParams{
			OperatorUUID: common.NewUUIDv7(),
		},
		partnerRequest,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, onboardResp.StatusCode())

	return testRestaurant{
		UUID:        restaurantUUID,
		Data:        restaurantToCreate,
		BankAccount: bankAccount,
	}
}

func registerCustomer(ctx context.Context, t *testing.T, clients testClients, country shared.CountryCode) ordersclient.CustomerUUID {
	t.Helper()

	customerToCreate := ordersclient.RegisterCustomer{
		Name:        gofakeit.Name(),
		Email:       openapi_types.Email(gofakeit.Email()),
		Address:     testutils.GenerateRandomOpenapiAddress(country),
		PhoneNumber: gofakeit.Phone(),
	}

	resp, err := clients.Orders.RegisterCustomerWithResponse(ctx, customerToCreate)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode())
	require.NotNil(t, resp.JSON201)

	return resp.JSON201.CustomerUuid
}

func registerCustomerInCity(ctx context.Context, t *testing.T, clients testClients, country shared.CountryCode, city string) ordersclient.CustomerUUID {
	t.Helper()

	customerToCreate := ordersclient.RegisterCustomer{
		Name:        gofakeit.Name(),
		Email:       openapi_types.Email(gofakeit.Email()),
		Address:     testutils.GenerateOpenapiAddressInCity(country, city),
		PhoneNumber: gofakeit.Phone(),
	}

	resp, err := clients.Orders.RegisterCustomerWithResponse(ctx, customerToCreate)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode())
	require.NotNil(t, resp.JSON201)

	return resp.JSON201.CustomerUuid
}

type testCourier struct {
	UUID        ordersclient.CourierUUID
	BankAccount string
}

func registerCourierInCity(
	ctx context.Context,
	t *testing.T,
	clients testClients,
	platformEntityUUID models.PlatformEntityUUID,
	country shared.CountryCode,
	city string,
) testCourier {
	t.Helper()

	courierToCreate := ordersclient.RegisterCourier{
		Name:        gofakeit.Name(),
		PhoneNumber: gofakeit.Phone(),
		City:        city,
	}

	resp, err := clients.Orders.RegisterCourierWithResponse(ctx, courierToCreate)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode())
	require.NotNil(t, resp.JSON201)

	taxID := gofakeit.Numerify("###-###-###")
	address := fakeAddressInUS()

	bankAccount, _ := createBankAccount(ctx, t, resp.JSON201.CourierUuid.String())

	partnerRequest := settlementclient.OnboardPartnerJSONRequestBody{
		PartnerUuid:        http2.LegalEntityUUID{resp.JSON201.CourierUuid.UUID},
		PlatformEntityUuid: platformEntityUUID,
		PartnerType:        domain.PartnerTypeCourier,
		BusinessName:       courierToCreate.Name + " LLC",
		TaxId:              taxID,
		BankAccountIban:    bankAccount,
		Currency:           shared.MustNewCurrency("USD"),
		Address: settlementclient.Address{
			Line1:       address.Street,
			Line2:       address.Unit,
			City:        city,
			PostalCode:  address.Zip,
			CountryCode: country,
		},
	}

	onboardResp, err := clients.Settlements.OnboardPartnerWithResponse(
		ctx,
		&settlementclient.OnboardPartnerParams{
			OperatorUUID: common.NewUUIDv7(),
		},
		partnerRequest,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, onboardResp.StatusCode())

	return testCourier{
		UUID:        resp.JSON201.CourierUuid,
		BankAccount: bankAccount,
	}
}

func registerCourierInCityWithCurrency(
	ctx context.Context,
	t *testing.T,
	clients testClients,
	platformEntityUUID models.PlatformEntityUUID,
	country shared.CountryCode,
	city string,
	currency shared.Currency,
) testCourier {
	t.Helper()

	courierToCreate := ordersclient.RegisterCourier{
		Name:        gofakeit.Name(),
		PhoneNumber: gofakeit.Phone(),
		City:        city,
	}

	resp, err := clients.Orders.RegisterCourierWithResponse(ctx, courierToCreate)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode())
	require.NotNil(t, resp.JSON201)

	taxID := gofakeit.Numerify("###-###-###")
	address := fakeAddressInUS()

	bankAccount, _ := createBankAccount(ctx, t, resp.JSON201.CourierUuid.String())

	partnerRequest := settlementclient.OnboardPartnerJSONRequestBody{
		PartnerUuid:        http2.LegalEntityUUID{resp.JSON201.CourierUuid.UUID},
		PlatformEntityUuid: platformEntityUUID,
		PartnerType:        domain.PartnerTypeCourier,
		BusinessName:       courierToCreate.Name + " LLC",
		TaxId:              taxID,
		BankAccountIban:    bankAccount,
		Currency:           currency,
		Address: settlementclient.Address{
			Line1:       address.Street,
			Line2:       address.Unit,
			City:        city,
			PostalCode:  address.Zip,
			CountryCode: country,
		},
	}

	onboardResp, err := clients.Settlements.OnboardPartnerWithResponse(
		ctx,
		&settlementclient.OnboardPartnerParams{
			OperatorUUID: common.NewUUIDv7(),
		},
		partnerRequest,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, onboardResp.StatusCode())

	return testCourier{
		UUID:        resp.JSON201.CourierUuid,
		BankAccount: bankAccount,
	}
}

func placeOrder(
	ctx context.Context,
	t *testing.T,
	clients testClients,
	customerUUID app.CustomerUUID,
	restaurantUUID app.RestaurantUUID,
	restaurant ordersclient.OnboardRestaurant,
	country shared.CountryCode,
	cardNumber string,
) app.OrderUUID {
	t.Helper()

	// Select 1-3 random items from the menu
	numItems := rand.Intn(3) + 1
	if numItems > len(restaurant.MenuItems) {
		numItems = len(restaurant.MenuItems)
	}

	var orderItems []ordersclient.OrderItem
	for i := 0; i < numItems; i++ {
		orderItems = append(orderItems, ordersclient.OrderItem{
			MenuItemUuid: restaurant.MenuItems[i].Uuid,
			Quantity:     rand.Intn(3) + 1,
		})
	}

	createQuoteRequest := ordersclient.CreateQuoteRequest{
		RestaurantUuid:  restaurantUUID,
		Items:           orderItems,
		DeliveryAddress: testutils.GenerateOpenapiAddressInCity(country, restaurant.Address.City),
	}

	quoteResp, err := clients.Orders.CustomerCreateQuoteWithResponse(
		ctx,
		&ordersclient.CustomerCreateQuoteParams{
			CustomerUUID: customerUUID,
		},
		createQuoteRequest,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, quoteResp.StatusCode())
	require.Equal(t, quoteResp.JSON201.Currency, restaurant.Currency)
	require.NotNil(t, quoteResp.JSON201)
	require.False(t, quoteResp.JSON201.DeliveryFeeGross.IsZero())
	require.False(t, quoteResp.JSON201.ServiceFeeGross.IsZero())
	require.False(t, quoteResp.JSON201.TotalGross.IsZero())

	nonce := preauthPayment(
		ctx,
		t,
		cardNumber,
		quoteResp.JSON201.TotalGross,
		quoteResp.JSON201.Currency.String(),
		quoteResp.JSON201.QuoteUuid.String(),
	)

	placeOrderRequest := ordersclient.PlaceOrder{
		QuoteUuid:    quoteResp.JSON201.QuoteUuid,
		PaymentNonce: nonce,
	}

	orderResp, err := clients.Orders.CustomerPlaceOrderWithResponse(
		ctx,
		&ordersclient.CustomerPlaceOrderParams{
			CustomerUUID: customerUUID,
		},
		placeOrderRequest,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, orderResp.StatusCode())

	assertOrderMatchesQuote(t, orderResp.JSON201, quoteResp.JSON201)

	require.Empty(
		t,
		cmp.Diff(
			&ordersclient.CustomerOrder{
				CourierAcceptedAt:     nil,
				CourierUuid:           nil,
				DeliveredAt:           nil,
				DeliveryAddress:       createQuoteRequest.DeliveryAddress,
				DeliveryFeeGross:      quoteResp.JSON201.DeliveryFeeGross,
				ItemsSubtotalGross:    quoteResp.JSON201.ItemsSubtotalGross,
				OrderUuid:             orderResp.JSON201.OrderUuid,
				OrderedAt:             time.Now(),
				PickedUpAt:            nil,
				RestaurantConfirmedAt: nil,
				RestaurantName:        restaurant.Name,
				RestaurantPreparedAt:  nil,
				RestaurantUuid:        restaurantUUID,
				ServiceFeeGross:       quoteResp.JSON201.ServiceFeeGross,
				TotalGross:            quoteResp.JSON201.TotalGross,
				TotalTax:              quoteResp.JSON201.TotalTax,
				Currency:              quoteResp.JSON201.Currency,
			},
			orderResp.JSON201,
			orderCmpOpts...,
		),
	)

	customerOrdersResp, err := clients.Orders.CustomerGetOrdersWithResponse(ctx, &ordersclient.CustomerGetOrdersParams{
		CustomerUUID: customerUUID,
	})
	require.NoError(t, err)
	require.NotEmpty(t, customerOrdersResp.JSON200.Orders, "Expected at least one order for the customer")
	require.Empty(
		t,
		cmp.Diff(
			*orderResp.JSON201,
			customerOrdersResp.JSON200.Orders[0],
			orderCmpOpts...,
		),
	)

	restaurantOrdersResp, err := clients.Orders.RestaurantGetOrdersWithResponse(ctx, &ordersclient.RestaurantGetOrdersParams{
		RestaurantUUID: restaurantUUID,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, restaurantOrdersResp.StatusCode())
	require.NotEmpty(t, restaurantOrdersResp.JSON200.Orders, "Expected at least one order")

	orderUUID := app.OrderUUID{common.NewUUIDv7()}
	for _, order := range restaurantOrdersResp.JSON200.Orders {
		if order.CustomerUuid == customerUUID {
			orderUUID = order.OrderUuid
			break
		}
	}
	require.NotEqual(t, openapi_types.UUID{}, orderUUID, "Placed order not found in restaurant's order list")

	return orderUUID
}

func assertOrderVisibleToRestaurant(ctx context.Context, t *testing.T, clients testClients, restaurantUUID app.RestaurantUUID, orderUUID app.OrderUUID) {
	t.Helper()

	resp, err := clients.Orders.RestaurantGetOrdersWithResponse(ctx, &ordersclient.RestaurantGetOrdersParams{
		RestaurantUUID: restaurantUUID,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())

	// Verify the specific order is in the list
	found := false
	for _, order := range resp.JSON200.Orders {
		if order.OrderUuid == orderUUID {
			found = true
			break
		}
	}
	require.True(t, found, "Order %s should be visible to restaurant", orderUUID)
}

func restaurantAcceptOrder(ctx context.Context, t *testing.T, clients testClients, restaurantUUID app.RestaurantUUID, orderUUID app.OrderUUID) {
	t.Helper()

	resp, err := clients.Orders.RestaurantAcceptOrderWithResponse(
		ctx,
		&ordersclient.RestaurantAcceptOrderParams{
			RestaurantUUID: restaurantUUID,
		},
		ordersclient.AcceptOrder{
			OrderUuid: orderUUID,
		},
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusAccepted, resp.StatusCode())
}

func assertOrderVisibleToCourier(ctx context.Context, t *testing.T, clients testClients, courierUUID app.CourierUUID, orderUUID app.OrderUUID) {
	t.Helper()

	resp, err := clients.Orders.CourierGetAvailableOrdersWithResponse(ctx, &ordersclient.CourierGetAvailableOrdersParams{
		CourierUUID: courierUUID,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())

	// Verify the specific order is in the list
	found := false
	for _, order := range resp.JSON200.Orders {
		if order.OrderUuid == orderUUID {
			found = true
			break
		}
	}
	require.True(t, found, "Order %s should be visible to courier", orderUUID)
}

func courierAcceptDelivery(ctx context.Context, t *testing.T, clients testClients, courierUUID app.CourierUUID, orderUUID app.OrderUUID) {
	t.Helper()

	resp, err := clients.Orders.CourierAcceptDeliveryWithResponse(
		ctx,
		&ordersclient.CourierAcceptDeliveryParams{
			CourierUUID: courierUUID,
		},
		ordersclient.AcceptDelivery{
			OrderUuid: orderUUID,
		},
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusAccepted, resp.StatusCode())
}

func restaurantMarkOrderReady(ctx context.Context, t *testing.T, clients testClients, restaurantUUID app.RestaurantUUID, orderUUID app.OrderUUID) {
	t.Helper()

	resp, err := clients.Orders.RestaurantMarkOrderReadyForPickupWithResponse(
		ctx,
		&ordersclient.RestaurantMarkOrderReadyForPickupParams{
			RestaurantUUID: restaurantUUID,
		},
		ordersclient.MarkOrderReady{
			OrderUuid: orderUUID,
		},
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusAccepted, resp.StatusCode())
}

func assertOrderReadyForCourier(ctx context.Context, t *testing.T, clients testClients, courierUUID app.CourierUUID, orderUUID app.OrderUUID) {
	t.Helper()

	resp, err := clients.Orders.CourierGetCurrentOrdersWithResponse(ctx, &ordersclient.CourierGetCurrentOrdersParams{
		CourierUUID: courierUUID,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())

	// Find the specific order and verify it's marked as ready
	found := false
	for _, order := range resp.JSON200.Orders {
		if order.OrderUuid == orderUUID {
			require.NotNil(t, order.RestaurantPreparedAt, "Order should be marked as ready")
			found = true
			break
		}
	}

	require.True(t, found, "Order %s not found in courier's order list", orderUUID)
}

func courierReportPickup(
	ctx context.Context,
	t *testing.T,
	clients testClients,
	courierUUID app.CourierUUID,
	orderUUID app.OrderUUID,
) {
	t.Helper()

	resp, err := clients.Orders.CourierReportPickupWithResponse(
		ctx,
		&ordersclient.CourierReportPickupParams{
			CourierUUID: courierUUID,
		},
		ordersclient.ReportPickup{
			OrderUuid: orderUUID,
		},
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusAccepted, resp.StatusCode())
}

func assertPickupReported(
	ctx context.Context,
	t *testing.T,
	clients testClients,
	courierUUID app.CourierUUID,
	orderUUID app.OrderUUID,
) {
	t.Helper()

	resp, err := clients.Orders.CourierGetCurrentOrdersWithResponse(ctx, &ordersclient.CourierGetCurrentOrdersParams{
		CourierUUID: courierUUID,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())

	// Find the specific order and verify it's marked as picked up
	found := false
	for _, order := range resp.JSON200.Orders {
		if order.OrderUuid == orderUUID {
			require.NotNil(t, order.PickedUpAt, "Order should be marked as picked up")
			found = true
			break
		}
	}

	require.True(t, found, "Order %s not found in courier's order list", orderUUID)
}

func courierReportDelivered(
	ctx context.Context,
	t *testing.T,
	clients testClients,
	courierUUID app.CourierUUID,
	orderUUID app.OrderUUID,
) {
	t.Helper()

	resp, err := clients.Orders.CourierReportDeliveryWithResponse(
		ctx,
		&ordersclient.CourierReportDeliveryParams{
			CourierUUID: courierUUID,
		},
		ordersclient.ReportDelivery{
			OrderUuid: orderUUID,
		},
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusAccepted, resp.StatusCode())
}

func assertDeliveryReported(
	ctx context.Context,
	t *testing.T,
	clients testClients,
	courierUUID app.CourierUUID,
	orderUUID app.OrderUUID,
) {
	t.Helper()

	resp, err := clients.Orders.CourierGetCurrentOrdersWithResponse(ctx, &ordersclient.CourierGetCurrentOrdersParams{
		CourierUUID: courierUUID,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())

	// Find the specific order and verify it's marked as delivered
	found := false
	for _, order := range resp.JSON200.Orders {
		if order.OrderUuid == orderUUID {
			require.NotNil(t, order.DeliveredAt, "Order should be marked as delivered")
			found = true
			break
		}
	}

	require.True(t, found, "Order %s not found in courier's order list", orderUUID)
}

func updateRestaurantMenu(
	ctx context.Context,
	t *testing.T,
	clients testClients,
	restaurantUUID app.RestaurantUUID,
	restaurant ordersclient.OnboardRestaurant,
) {
	t.Helper()

	resp, err := clients.Orders.OnboardRestaurantWithResponse(
		ctx,
		restaurantUUID,
		&ordersclient.OnboardRestaurantParams{
			OperatorUUID: common.NewUUIDv7(),
		},
		restaurant,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, resp.StatusCode())
}

func assertMenuItemsEquals(
	ctx context.Context,
	t *testing.T,
	clients testClients,
	restaurantUUID app.RestaurantUUID,
	expectedMenuItems []ordersclient.MenuItem,
) {
	t.Helper()

	menuResp, err := clients.Orders.CustomerGetRestaurantMenuWithResponse(ctx, restaurantUUID, &ordersclient.CustomerGetRestaurantMenuParams{
		CustomerUUID: app.CustomerUUID{common.NewUUIDv7()},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, menuResp.StatusCode())

	require.Len(t, menuResp.JSON200.Items, len(expectedMenuItems), "Menu should contain expected number of items")

	expectedUUIDs := make([]app.RestaurantMenuItemUUID, len(expectedMenuItems))
	for i, item := range expectedMenuItems {
		expectedUUIDs[i] = item.Uuid
	}

	actualUUIDs := make([]app.RestaurantMenuItemUUID, len(menuResp.JSON200.Items))
	for i, item := range menuResp.JSON200.Items {
		actualUUIDs[i] = item.Uuid
	}

	require.ElementsMatch(t, expectedUUIDs, actualUUIDs, "Menu should contain only expected items")
}

func assertArchivedMenuItemsNotInMenu(
	ctx context.Context,
	t *testing.T,
	clients testClients,
	restaurantUUID app.RestaurantUUID,
	archivedItems []ordersclient.MenuItem,
) {
	t.Helper()

	menuResp, err := clients.Orders.CustomerGetRestaurantMenuWithResponse(ctx, restaurantUUID, &ordersclient.CustomerGetRestaurantMenuParams{
		CustomerUUID: app.CustomerUUID{common.NewUUIDv7()},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, menuResp.StatusCode())

	for _, archivedItem := range archivedItems {
		for _, actualItem := range menuResp.JSON200.Items {
			require.NotEqual(
				t,
				archivedItem.Uuid,
				actualItem.Uuid,
				"Archived item %s should not be in the menu",
				archivedItem.Uuid,
			)
		}
	}
}

func assertOrderWithArchivedItemFails(
	ctx context.Context,
	t *testing.T,
	clients testClients,
	customerUUID app.CustomerUUID,
	restaurantUUID app.RestaurantUUID,
	archivedItems ordersclient.MenuItem,
	country shared.CountryCode,
	city string,
) {
	t.Helper()

	// Try to create an order offer with an archived item
	createOfferRequest := ordersclient.CreateQuoteRequest{
		RestaurantUuid: restaurantUUID,
		Items: []ordersclient.OrderItem{
			{
				MenuItemUuid: archivedItems.Uuid,
				Quantity:     1,
			},
		},
		DeliveryAddress: testutils.GenerateOpenapiAddressInCity(country, city),
	}

	offerResp, err := clients.Orders.CustomerCreateQuoteWithResponse(
		ctx,
		&ordersclient.CustomerCreateQuoteParams{
			CustomerUUID: customerUUID,
		},
		createOfferRequest,
	)
	require.NoError(t, err)
	require.Equal(
		t,
		http.StatusGone,
		offerResp.StatusCode(),
		"Should return 400 when creating offer with archived items",
	)

	require.NotNil(t, offerResp.JSON410, "Should have error response body")
	require.NotEmpty(t, offerResp.JSON410.Details, "Should have errors in response")
	require.Equal(
		t,
		"archived-menu-position",
		offerResp.JSON410.Details[0].ErrorSlug,
		"Error slug should indicate unavailable menu items",
	)
	require.Equal(
		t,
		// it's part of the public contract - we want to assert it
		fmt.Sprintf("menu position '%s' is archived", archivedItems.Name),
		offerResp.JSON410.Details[0].Message,
		"Error message should indicate items not available",
	)
}

func addCompanySuffix(name string) string {
	companySuffixes := []string{"Corp.", "Inc.", "Ltd.", "Co."}

	return fmt.Sprintf("%s %s", name, companySuffixes[rand.Intn(len(companySuffixes))])
}

func randomPrice() decimal.Decimal {
	return decimal.New(int64(rand.Intn(200)+50), -1)
}

func onboardRestaurantWithData(
	ctx context.Context,
	t *testing.T,
	clients testClients,
	restaurantUUID app.RestaurantUUID,
	restaurant ordersclient.OnboardRestaurant,
) {
	t.Helper()

	resp, err := clients.Orders.OnboardRestaurantWithResponse(
		ctx,
		restaurantUUID,
		&ordersclient.OnboardRestaurantParams{
			OperatorUUID: common.NewUUIDv7(),
		},
		restaurant,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, resp.StatusCode())
}

func onboardPartnerForRestaurant(
	ctx context.Context,
	t *testing.T,
	clients testClients,
	platformEntityUUID models.PlatformEntityUUID,
	restaurantUUID app.RestaurantUUID,
	businessName string,
) string {
	t.Helper()

	address := fakeAddressInUS()

	taxID := gofakeit.Numerify("###-###-###")

	bankAccount, _ := createBankAccount(ctx, t, restaurantUUID.String())

	partnerRequest := settlementclient.OnboardPartnerJSONRequestBody{
		PartnerUuid:        http2.LegalEntityUUID{restaurantUUID.UUID},
		PlatformEntityUuid: platformEntityUUID,
		PartnerType:        domain.PartnerTypeRestaurant,
		BusinessName:       businessName,
		TaxId:              taxID,
		BankAccountIban:    bankAccount,
		Address: settlementclient.Address{
			Line1:       address.Street,
			Line2:       address.Unit,
			City:        address.City,
			PostalCode:  address.Zip,
			CountryCode: shared.MustNewCountryCode(address.Country),
		},
		Currency: shared.MustNewCurrency("USD"),
	}

	resp, err := clients.Settlements.OnboardPartnerWithResponse(
		ctx,
		&settlementclient.OnboardPartnerParams{
			OperatorUUID: common.NewUUIDv7(),
		},
		partnerRequest,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, resp.StatusCode())

	return bankAccount
}

func onboardPartnerForRestaurantWithCurrency(
	ctx context.Context,
	t *testing.T,
	clients testClients,
	platformEntityUUID models.PlatformEntityUUID,
	restaurantUUID app.RestaurantUUID,
	businessName string,
	currency shared.Currency,
) string {
	t.Helper()

	address := fakeAddressInUS()

	taxID := gofakeit.Numerify("###-###-###")

	bankAccount, _ := createBankAccount(ctx, t, restaurantUUID.String())

	partnerRequest := settlementclient.OnboardPartnerJSONRequestBody{
		PartnerUuid:        http2.LegalEntityUUID{restaurantUUID.UUID},
		PlatformEntityUuid: platformEntityUUID,
		PartnerType:        domain.PartnerTypeRestaurant,
		BusinessName:       businessName,
		TaxId:              taxID,
		BankAccountIban:    bankAccount,
		Address: settlementclient.Address{
			Line1:       address.Street,
			Line2:       address.Unit,
			City:        address.City,
			PostalCode:  address.Zip,
			CountryCode: shared.MustNewCountryCode(address.Country),
		},
		Currency: currency,
	}

	resp, err := clients.Settlements.OnboardPartnerWithResponse(
		ctx,
		&settlementclient.OnboardPartnerParams{
			OperatorUUID: common.NewUUIDv7(),
		},
		partnerRequest,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, resp.StatusCode())

	return bankAccount
}

func createQuote(
	ctx context.Context,
	t *testing.T,
	clients testClients,
	customerUUID app.CustomerUUID,
	restaurantUUID app.RestaurantUUID,
	orderItems []ordersclient.OrderItem,
	deliveryAddress ordersclient.Address,
) *ordersclient.CreateQuoteResponse {
	t.Helper()

	createQuoteRequest := ordersclient.CreateQuoteRequest{
		RestaurantUuid:  restaurantUUID,
		Items:           orderItems,
		DeliveryAddress: deliveryAddress,
	}

	resp, err := clients.Orders.CustomerCreateQuoteWithResponse(
		ctx,
		&ordersclient.CustomerCreateQuoteParams{
			CustomerUUID: customerUUID,
		},
		createQuoteRequest,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode())
	require.NotNil(t, resp.JSON201)

	return resp.JSON201
}

func placeOrderFromQuote(
	ctx context.Context,
	t *testing.T,
	clients testClients,
	customerUUID app.CustomerUUID,
	quote *ordersclient.CreateQuoteResponse,
	cardNumber string,
) *ordersclient.CustomerOrder {
	t.Helper()

	nonce := preauthPayment(ctx, t, cardNumber, quote.TotalGross, quote.Currency.String(), quote.QuoteUuid.String())

	placeOrderRequest := ordersclient.PlaceOrder{
		QuoteUuid:    quote.QuoteUuid,
		PaymentNonce: nonce,
	}

	resp, err := clients.Orders.CustomerPlaceOrderWithResponse(
		ctx,
		&ordersclient.CustomerPlaceOrderParams{
			CustomerUUID: customerUUID,
		},
		placeOrderRequest,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode())
	require.NotNil(t, resp.JSON201)

	return resp.JSON201
}

func assertOrderMatchesQuote(t *testing.T, order *ordersclient.CustomerOrder, quote *ordersclient.CreateQuoteResponse) {
	t.Helper()

	assert.Equal(
		t,
		quote.ItemsSubtotalGross.String(),
		order.ItemsSubtotalGross.String(),
		"order items subtotal should match quote",
	)
	assert.Equal(
		t,
		quote.ServiceFeeGross.String(),
		order.ServiceFeeGross.String(),
		"order service fee should match quote",
	)
	assert.Equal(
		t,
		quote.DeliveryFeeGross.String(),
		order.DeliveryFeeGross.String(),
		"order delivery fee should match quote",
	)
	assert.Equal(
		t,
		quote.TotalGross.String(),
		order.TotalGross.String(),
		"order total gross should match quote",
	)
	assert.Equal(
		t,
		quote.TotalTax.String(),
		order.TotalTax.String(),
		"order total tax should match quote",
	)
}

func currencyForCountry(t *testing.T, country shared.CountryCode) shared.Currency {
	t.Helper()
	switch country.Code() {
	case "US":
		return shared.MustNewCurrency("USD")
	case "DE":
		return shared.MustNewCurrency("EUR")
	case "GB":
		return shared.MustNewCurrency("GBP")
	case "JP":
		return shared.MustNewCurrency("JPY")
	case "PL":
		return shared.MustNewCurrency("PLN")
	default:
		t.Fatalf("unsupported country for currency mapping: %s", country.Code())
		return shared.Currency{} // unreachable
	}
}

func assertJsonReprEqual(t *testing.T, expected, actual any) {
	t.Helper()

	expectedJSON, err := json.Marshal(expected)
	require.NoError(t, err)

	actualJSON, err := json.Marshal(actual)
	require.NoError(t, err)

	require.JSONEq(t, string(expectedJSON), string(actualJSON))
}

func createBankAccount(
	ctx context.Context,
	t *testing.T,
	merchantID string,
) (string, string) {
	t.Helper()

	return createBankAccountWithBalance(ctx, t, decimal.Zero, merchantID)
}

func createBankAccountWithBalance(
	ctx context.Context,
	t *testing.T,
	balance decimal.Decimal,
	merchantID string,
) (string, string) {
	t.Helper()

	accountNumber, cardNumber := stubs.Payments.CreateAccount(merchantID, balance)
	return accountNumber, cardNumber
}

func preauthPayment(
	ctx context.Context,
	t *testing.T,
	cardNumber string,
	amount decimal.Decimal,
	currency string,
	idempotencyKey string,
) string {
	return stubs.Payments.PreauthorizePayment(cardNumber, amount, currency)
}

type bankTransfer struct {
	Amount                decimal.Decimal
	Currency              string
	ExternalAccountNumber string
}

func assertAccountBalance(ctx context.Context, t *testing.T, number string, amount decimal.Decimal, transfers []bankTransfer) {
	balance, err := stubs.Payments.GetAccountBalance(number)
	require.NoError(t, err)

	require.True(t, balance.Equal(amount),
		"Account %s balance expected to be %s, got %s",
		number, amount.String(), balance.String())

	history, err := stubs.Payments.GetAccountHistory(number)
	require.NoError(t, err)

	amounts := make([]string, 0, len(history))
	for i, h := range history {
		amounts = append(amounts, fmt.Sprintf("\t- #%v: %v %v %v %v (%v)", i, h.Amount.String(), h.Currency, h.ExternalAccountNumber, h.Reference, h.ReceiverDetails))
	}

	for _, expectedTransfer := range transfers {
		found := false
		for _, h := range history {
			if h.Amount.Equal(expectedTransfer.Amount) &&
				h.Currency == expectedTransfer.Currency &&
				h.ExternalAccountNumber == expectedTransfer.ExternalAccountNumber {
				found = true
				break
			}
		}

		if !found {
			t.Errorf(
				"Account %s does not contain balance change: %v %v %v\nAmounts found:\n%v\n",
				number,
				expectedTransfer.ExternalAccountNumber,
				expectedTransfer.Amount,
				expectedTransfer.Currency,
				strings.Join(amounts, "\n"),
			)
		}
	}
}

func fakeAddressInUS() *gofakeit.AddressInfo {
	address := gofakeit.Address()
	address.Country = "US"
	return address
}

func onboardRestaurantWithName(
	ctx context.Context,
	t *testing.T,
	clients testClients,
	country shared.CountryCode,
	name string,
) (app.RestaurantUUID, ordersclient.OnboardRestaurant) {
	t.Helper()

	var category app.ItemCategory
	if rand.Intn(2) == 0 {
		category = app.ItemCategoryFood
	} else {
		category = app.ItemCategoryBeverage
	}

	var menuItems []ordersclient.MenuItem
	for i := 0; i < 5; i++ {
		menuItems = append(menuItems, ordersclient.MenuItem{
			Uuid:       app.RestaurantMenuItemUUID{common.NewUUIDv7()},
			Name:       gofakeit.Lunch(),
			Category:   category,
			GrossPrice: randomPrice(),
			Ordering:   rand.Float32(),
		})
	}

	restaurantToCreate := ordersclient.OnboardRestaurant{
		Address:     testutils.GenerateRandomOpenapiAddress(country),
		Description: gofakeit.HipsterSentence(),
		MenuItems:   menuItems,
		Name:        name,
		Currency:    currencyForCountry(t, country),
	}

	restaurantUUID := app.RestaurantUUID{common.NewUUIDv7()}
	resp, err := clients.Orders.OnboardRestaurantWithResponse(
		ctx,
		restaurantUUID,
		&ordersclient.OnboardRestaurantParams{
			OperatorUUID: common.NewUUIDv7(),
		},
		restaurantToCreate,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, resp.StatusCode())

	return restaurantUUID, restaurantToCreate
}

func onboardRestaurantWithItems(
	ctx context.Context,
	t *testing.T,
	clients testClients,
	country shared.CountryCode,
	name string,
	itemNames []string,
) (app.RestaurantUUID, ordersclient.OnboardRestaurant) {
	t.Helper()

	menuItems := make([]ordersclient.MenuItem, 0, len(itemNames))
	for i, itemName := range itemNames {
		menuItems = append(menuItems, ordersclient.MenuItem{
			Uuid:       app.RestaurantMenuItemUUID{common.NewUUIDv7()},
			Name:       itemName,
			Category:   app.ItemCategoryFood,
			GrossPrice: decimal.NewFromFloat(10.00 + float64(i)),
			Ordering:   float32(i + 1),
		})
	}

	restaurantToCreate := ordersclient.OnboardRestaurant{
		Address:     testutils.GenerateRandomOpenapiAddress(country),
		Description: gofakeit.HipsterSentence(),
		MenuItems:   menuItems,
		Name:        name,
		Currency:    currencyForCountry(t, country),
	}

	restaurantUUID := app.RestaurantUUID{common.NewUUIDv7()}
	resp, err := clients.Orders.OnboardRestaurantWithResponse(
		ctx,
		restaurantUUID,
		&ordersclient.OnboardRestaurantParams{
			OperatorUUID: common.NewUUIDv7(),
		},
		restaurantToCreate,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, resp.StatusCode())

	return restaurantUUID, restaurantToCreate
}

func issueReceipt(
	ctx context.Context,
	t *testing.T,
	clients testClients,
) string {
	t.Helper()

	sellerTaxID, err := shared.NewTaxID("12-3456789")
	require.NoError(t, err)

	resp, err := clients.Billing.CreateReceiptWithResponse(ctx, billingclient.CreateDocument{
		IssueDate: time.Now(),
		Currency:  shared.MustNewCurrency("USD"),
		Seller: billingclient.LegalEntity{
			Name: "Eats Inc.",
			Address: billingclient.Address{
				Line1:       "123 Main St",
				Line2:       "Suite 100",
				City:        "New York",
				PostalCode:  "10001",
				CountryCode: shared.MustNewCountryCode("US"),
			},
			TaxId: &sellerTaxID,
		},
		Buyer: billingclient.LegalEntity{
			Name: "John Doe",
			Address: billingclient.Address{
				Line1:       "456 Oak Ave",
				City:        "New York",
				PostalCode:  "10002",
				CountryCode: shared.MustNewCountryCode("US"),
			},
		},
		LineItems: []billingclient.LineItem{
			{
				Name:         "Classic Burger",
				LineItemType: shared.LineItemTypeFood,
				Quantity:     2,
				UnitAmount:   decimal.RequireFromString("12.50"),
				IsGross:      true,
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode(), "creating receipt failed: %s", string(resp.Body))
	require.NotNil(t, resp.JSON201)

	return resp.JSON201.DocumentUuid.String()
}

func issueInvoice(
	ctx context.Context,
	t *testing.T,
	clients testClients,
) billingclient.DocumentUUID {
	t.Helper()

	sellerTaxID, err := shared.NewTaxID("12-3456789")
	require.NoError(t, err)

	buyerTaxID, err := shared.NewTaxID("98-7654321")
	require.NoError(t, err)

	resp, err := clients.Billing.CreateInvoiceWithResponse(ctx, billingclient.CreateDocument{
		IssueDate: time.Now(),
		Currency:  shared.MustNewCurrency("USD"),
		Seller: billingclient.LegalEntity{
			Name: "Eats Inc.",
			Address: billingclient.Address{
				Line1:       "123 Main St",
				Line2:       "Suite 100",
				City:        "New York",
				PostalCode:  "10001",
				CountryCode: shared.MustNewCountryCode("US"),
			},
			TaxId: &sellerTaxID,
		},
		Buyer: billingclient.LegalEntity{
			Name: "Acme Corp.",
			Address: billingclient.Address{
				Line1:       "789 Business Blvd",
				City:        "Chicago",
				PostalCode:  "60601",
				CountryCode: shared.MustNewCountryCode("US"),
			},
			TaxId: &buyerTaxID,
		},
		LineItems: []billingclient.LineItem{
			{
				Name:         "Platform Service Fee",
				LineItemType: shared.LineItemTypeService,
				Quantity:     1,
				UnitAmount:   decimal.RequireFromString("250.00"),
				IsGross:      false,
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode(), "creating invoice failed: %s", string(resp.Body))
	require.NotNil(t, resp.JSON201)

	return resp.JSON201.DocumentUuid
}
