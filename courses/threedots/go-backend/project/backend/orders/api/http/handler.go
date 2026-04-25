package http

import (
	"context"

	"eats/backend/common"
	"eats/backend/common/shared"
	"eats/backend/orders/app"
)

// ListMenuItemsFilter contains optional filters for the menu items query.
type ListMenuItemsFilter struct {
	RestaurantName *string
	Search         *string
	OrderBy        *string
}

// ReadModel is an interface for the read model that lists menu items.
// It is defined here (consumer side) to allow for easy testing and decoupling.
type ReadModel interface {
	ListMenuItemsWithRestaurant(ctx context.Context, filter ListMenuItemsFilter) ([]MenuItemWithRestaurant, error)
}

type Handler struct {
	service   *app.Service
	readModel ReadModel
}

func NewHandler(
	service *app.Service,
	readModel ReadModel,
) Handler {
	if service == nil {
		panic("service cannot be nil")
	}
	if readModel == nil {
		panic("readModel cannot be nil")
	}

	return Handler{
		service:   service,
		readModel: readModel,
	}
}

func (h Handler) RegisterCustomer(ctx context.Context, request RegisterCustomerRequestObject) (RegisterCustomerResponseObject, error) {
	addr, err := openapiAddressToSharedAddress(request.Body.Address)
	if err != nil {
		return nil, common.NewInvalidInputError("invalid-address", "invalid address: %s", err)
	}

	customerUUID := CustomerUUID{common.NewUUIDv7()}

	err = h.service.RegisterCustomer(ctx, app.Customer{
		CustomerUUID: customerUUID,
		Name:         request.Body.Name,
		Email:        string(request.Body.Email),
		// address should be ideally normalized to ensure consistent city names and postal codes
		// across customers, restaurants, and delivery addresses
		Address:     addr,
		PhoneNumber: request.Body.PhoneNumber,
	})
	if err != nil {
		return nil, err
	}

	return RegisterCustomer201JSONResponse{
		CustomerUuid: customerUUID,
	}, nil
}

func openapiAddressToSharedAddress(addr Address) (shared.Address, error) {
	sharedAddr, err := shared.NewAddress(
		addr.Line1,
		addr.Line2,
		addr.PostalCode,
		addr.City,
		addr.CountryCode,
	)
	if err != nil {
		return shared.Address{}, err
	}

	return sharedAddr, nil
}

func (h Handler) CustomerCreateQuote(ctx context.Context, request CustomerCreateQuoteRequestObject) (CustomerCreateQuoteResponseObject, error) {
	if request.Params.CustomerUUID.IsZero() {
		return nil, common.NewUnauthorizedError("missing-customer-uuid", "customer UUID is required")
	}

	var items []app.CreateQuoteItem
	for _, item := range request.Body.Items {
		items = append(items, app.CreateQuoteItem{
			MenuItemUUID: item.MenuItemUuid,
			Quantity:     item.Quantity,
		})
	}

	addr, err := openapiAddressToSharedAddress(request.Body.DeliveryAddress)
	if err != nil {
		return nil, common.NewInvalidInputError("invalid-address", "invalid address: %s", err)
	}

	quote, err := h.service.CreateQuote(ctx, app.CreateQuote{
		request.Params.CustomerUUID,
		request.Body.RestaurantUuid,
		items,
		addr,
	})
	if err != nil {
		return nil, err
	}

	return CustomerCreateQuote201JSONResponse{
		quote.Currency,
		quote.DeliveryFeeGross,
		quote.ExpirationTime(),
		quote.ItemsSubtotalGross,
		quote.QuoteUUID,
		quote.ServiceFeeGross,
		quote.TotalAmountGross,
		quote.TotalTax,
	}, nil
}

func (h Handler) OnboardRestaurant(ctx context.Context, request OnboardRestaurantRequestObject) (OnboardRestaurantResponseObject, error) {
	if request.Params.OperatorUUID.IsZero() {
		return nil, common.NewUnauthorizedError("missing-operator-uuid", "operator UUID is required")
	}

	var menuItems []app.MenuItem
	for _, item := range request.Body.MenuItems {
		menuItems = append(menuItems, app.MenuItem{
			MenuItemUUID: item.Uuid,
			Name:         item.Name,
			GrossPrice:   item.GrossPrice,
			Ordering:     float64(item.Ordering),
		})
	}

	addr, err := openapiAddressToSharedAddress(request.Body.Address)
	if err != nil {
		return nil, common.NewInvalidInputError("invalid-address", "invalid address: %s", err)
	}

	err = h.service.OnboardRestaurant(
		ctx,
		request.RestaurantUuid,
		app.OnboardRestaurant{
			request.Body.Name,
			addr,
			request.Body.Currency,
			request.Body.Description,
			menuItems,
		},
	)
	if err != nil {
		return nil, err
	}

	return OnboardRestaurant204Response{}, nil
}

// ListMenuItems returns all active menu items with their restaurant information.
// Supports optional filtering by restaurant name and ordering.
func (h Handler) ListMenuItems(ctx context.Context, request ListMenuItemsRequestObject) (ListMenuItemsResponseObject, error) {
	var orderBy *string
	if request.Params.OrderBy != nil {
		s := string(*request.Params.OrderBy)
		orderBy = &s
	}

	filter := ListMenuItemsFilter{
		RestaurantName: request.Params.RestaurantName,
		Search:         request.Params.Search,
		OrderBy:        orderBy,
	}

	items, err := h.readModel.ListMenuItemsWithRestaurant(ctx, filter)
	if err != nil {
		return nil, err
	}

	return ListMenuItems200JSONResponse(items), nil
}

func Register(ctx context.Context, e EchoRouter, handler Handler) error {
	RegisterHandlers(e, NewStrictHandler(handler, nil))

	return nil
}
