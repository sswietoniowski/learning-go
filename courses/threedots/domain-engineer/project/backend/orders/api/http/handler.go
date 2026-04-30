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
	ListCustomerOrders(ctx context.Context, customerUUID app.CustomerUUID) ([]CustomerOrder, error)
	ListRestaurantOrders(ctx context.Context, restaurantUUID app.RestaurantUUID) ([]RestaurantOrder, error)
	ListAssignedCourierOrders(ctx context.Context, courierUUID app.CourierUUID) ([]CourierOrder, error)
	ListAvailableOrdersForCourier(ctx context.Context, courierUUID app.CourierUUID) ([]CourierOrder, error)
}

type RestaurantReader interface {
	RestaurantName(ctx context.Context, restaurantUUID app.RestaurantUUID) (string, error)
	ListRestaurants(ctx context.Context) ([]app.Restaurant, error)
	GetRestaurantMenu(ctx context.Context, restaurantUUID app.RestaurantUUID) (app.RestaurantMenu, error)
}

type Handler struct {
	service          *app.Service
	readModel        ReadModel
	restaurantReader RestaurantReader
}

func NewHandler(
	service *app.Service,
	readModel ReadModel,
	restaurantReader RestaurantReader,
) Handler {
	if service == nil {
		panic("service cannot be nil")
	}
	if readModel == nil {
		panic("readModel cannot be nil")
	}
	if restaurantReader == nil {
		panic("restaurant reader cannot be nil")
	}

	return Handler{
		service:          service,
		readModel:        readModel,
		restaurantReader: restaurantReader,
	}
}

func (h Handler) CourierGetCurrentOrders(ctx context.Context, request CourierGetCurrentOrdersRequestObject) (CourierGetCurrentOrdersResponseObject, error) {
	if request.Params.CourierUUID.IsZero() {
		return nil, common.NewUnauthorizedError("missing-courier-uuid", "courier UUID is required")
	}

	orders, err := h.readModel.ListAssignedCourierOrders(ctx, request.Params.CourierUUID)
	if err != nil {
		return nil, err
	}

	return CourierGetCurrentOrders200JSONResponse{
		Orders: orders,
	}, nil
}

func (h Handler) CourierGetAvailableOrders(ctx context.Context, request CourierGetAvailableOrdersRequestObject) (CourierGetAvailableOrdersResponseObject, error) {
	if request.Params.CourierUUID.IsZero() {
		return nil, common.NewUnauthorizedError("missing-courier-uuid", "courier UUID is required")
	}

	orders, err := h.readModel.ListAvailableOrdersForCourier(ctx, request.Params.CourierUUID)
	if err != nil {
		return nil, err
	}

	return CourierGetAvailableOrders200JSONResponse{
		Orders: orders,
	}, nil
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

func (h Handler) RegisterCourier(ctx context.Context, request RegisterCourierRequestObject) (RegisterCourierResponseObject, error) {
	courierUUID, err := h.service.RegisterCourier(ctx, app.RegisterCourier{
		Name:        request.Body.Name,
		PhoneNumber: request.Body.PhoneNumber,
		// city should be ideally normalized to ensure consistent city names
		// across customers, restaurants, and delivery addresses
		City: request.Body.City,
	})
	if err != nil {
		return nil, err
	}

	return RegisterCourier201JSONResponse{
		CourierUuid: courierUUID,
	}, nil
}

func (h Handler) RestaurantGetOrders(ctx context.Context, request RestaurantGetOrdersRequestObject) (RestaurantGetOrdersResponseObject, error) {
	if request.Params.RestaurantUUID.IsZero() {
		return nil, common.NewUnauthorizedError("missing-restaurant-uuid", "restaurant UUID is required")
	}

	// This handler uses the read model directly instead of going through the service layer.
	// This is the CQRS pattern: queries bypass the domain and read directly from optimized read models,
	// while commands go through the service layer to enforce business rules.
	orders, err := h.readModel.ListRestaurantOrders(ctx, request.Params.RestaurantUUID)
	if err != nil {
		return nil, err
	}

	return RestaurantGetOrders200JSONResponse{
		Orders: orders,
	}, nil
}

func (h Handler) CourierAcceptDelivery(ctx context.Context, request CourierAcceptDeliveryRequestObject) (CourierAcceptDeliveryResponseObject, error) {
	if request.Params.CourierUUID.IsZero() {
		return nil, common.NewUnauthorizedError("missing-courier-uuid", "courier UUID is required")
	}

	err := h.service.CourierAcceptDelivery(
		ctx,
		request.Params.CourierUUID,
		request.Body.OrderUuid,
	)
	if err != nil {
		return nil, err
	}

	return CourierAcceptDelivery202Response{}, nil
}

func (h Handler) RestaurantAcceptOrder(ctx context.Context, request RestaurantAcceptOrderRequestObject) (RestaurantAcceptOrderResponseObject, error) {
	if request.Params.RestaurantUUID.IsZero() {
		return nil, common.NewUnauthorizedError("missing-restaurant-uuid", "restaurant UUID is required")
	}

	err := h.service.AcceptOrder(
		ctx,
		request.Params.RestaurantUUID,
		request.Body.OrderUuid,
	)
	if err != nil {
		return nil, err
	}

	return RestaurantAcceptOrder202Response{}, nil
}

func (h Handler) RestaurantMarkOrderReadyForPickup(ctx context.Context, request RestaurantMarkOrderReadyForPickupRequestObject) (RestaurantMarkOrderReadyForPickupResponseObject, error) {
	if request.Params.RestaurantUUID.IsZero() {
		return nil, common.NewUnauthorizedError("missing-restaurant-uuid", "restaurant UUID is required")
	}

	err := h.service.MarkOrderAsReadyForPickup(
		ctx,
		request.Params.RestaurantUUID,
		request.Body.OrderUuid,
	)
	if err != nil {
		return nil, err
	}

	return RestaurantMarkOrderReadyForPickup202Response{}, nil
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

func (h Handler) CustomerPlaceOrder(ctx context.Context, request CustomerPlaceOrderRequestObject) (CustomerPlaceOrderResponseObject, error) {
	if request.Params.CustomerUUID.IsZero() {
		return nil, common.NewUnauthorizedError("missing-customer-uuid", "customer UUID is required")
	}

	order, err := h.service.PlaceOrder(ctx, app.PlaceOrder{
		CustomerUUID: request.Params.CustomerUUID,
		QuoteUUID:    request.Body.QuoteUuid,
		PaymentNonce: request.Body.PaymentNonce,
	})
	if err != nil {
		return nil, err
	}

	restaurantName, err := h.restaurantReader.RestaurantName(ctx, order.RestaurantUUID)
	if err != nil {
		return nil, err
	}

	return appOrderToHttpOrder(order, restaurantName), nil
}

func (h Handler) CustomerGetOrders(ctx context.Context, request CustomerGetOrdersRequestObject) (CustomerGetOrdersResponseObject, error) {
	if request.Params.CustomerUUID.IsZero() {
		return nil, common.NewUnauthorizedError("missing-customer-uuid", "customer UUID is required")
	}

	orders, err := h.readModel.ListCustomerOrders(ctx, request.Params.CustomerUUID)
	if err != nil {
		return nil, err
	}

	return CustomerGetOrders200JSONResponse{
		Orders: orders,
	}, nil
}

func sharedAddressToOpenapiAddress(addr shared.Address) Address {
	return Address{
		Line1:       addr.Line1,
		Line2:       addr.Line2,
		PostalCode:  addr.PostalCode,
		City:        addr.City,
		CountryCode: addr.CountryCode,
	}
}

func appOrderToHttpOrder(order app.Order, restaurantName string) CustomerPlaceOrder201JSONResponse {
	return CustomerPlaceOrder201JSONResponse{
		order.CourierAcceptedAt,
		order.CourierUUID,
		order.Currency,
		order.DeliveredAt,
		sharedAddressToOpenapiAddress(order.DeliveryAddress),
		order.DeliveryFeeGross,
		order.ItemsSubtotal,
		order.OrderUUID,
		order.OrderedAt,
		order.PickedUpAt,
		order.RestaurantConfirmedAt,
		restaurantName,
		order.RestaurantPreparedAt,
		order.RestaurantUUID,
		order.ServiceFeeGross,
		order.TotalAmountGross,
		order.TotalTax,
	}
}

func (h Handler) CourierReportDelivery(ctx context.Context, request CourierReportDeliveryRequestObject) (CourierReportDeliveryResponseObject, error) {
	if request.Params.CourierUUID.IsZero() {
		return nil, common.NewUnauthorizedError("missing-courier-uuid", "courier UUID is required")
	}

	err := h.service.CourierReportDelivery(
		ctx,
		request.Params.CourierUUID,
		request.Body.OrderUuid,
	)
	if err != nil {
		return nil, err
	}

	return CourierReportDelivery202Response{}, nil
}

func (h Handler) CourierReportPickup(ctx context.Context, request CourierReportPickupRequestObject) (CourierReportPickupResponseObject, error) {
	if request.Params.CourierUUID.IsZero() {
		return nil, common.NewUnauthorizedError("missing-courier-uuid", "courier UUID is required")
	}

	err := h.service.CourierReportDeliveryPickup(
		ctx,
		request.Params.CourierUUID,
		request.Body.OrderUuid,
	)
	if err != nil {
		return nil, err
	}

	return CourierReportPickup202Response{}, nil
}

func (h Handler) CustomerListRestaurants(ctx context.Context, request CustomerListRestaurantsRequestObject) (CustomerListRestaurantsResponseObject, error) {
	restaurants, err := h.restaurantReader.ListRestaurants(ctx)
	if err != nil {
		return nil, err
	}

	var httpRestaurants []Restaurant
	for _, r := range restaurants {
		httpRestaurants = append(httpRestaurants, Restaurant{
			Uuid:        r.RestaurantUUID,
			Name:        r.Name,
			Address:     sharedAddressToOpenapiAddress(r.Address),
			Description: r.Description,
		})
	}

	return CustomerListRestaurants200JSONResponse{
		Restaurants: httpRestaurants,
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

func (h Handler) CustomerGetRestaurantMenu(ctx context.Context, request CustomerGetRestaurantMenuRequestObject) (CustomerGetRestaurantMenuResponseObject, error) {
	menu, err := h.restaurantReader.GetRestaurantMenu(ctx, request.RestaurantUuid)
	if err != nil {
		return nil, err
	}

	var httpMenuItems []MenuItem
	for _, item := range menu.Positions {
		httpMenuItems = append(httpMenuItems, MenuItem{
			Uuid:       item.MenuItemUUID,
			Name:       item.Name,
			GrossPrice: item.GrossPrice,
			Ordering:   float32(item.Ordering),
		})
	}

	return CustomerGetRestaurantMenu200JSONResponse{
		Address:        sharedAddressToOpenapiAddress(menu.Address),
		Currency:       menu.Currency,
		Description:    menu.Description,
		Items:          httpMenuItems,
		RestaurantName: menu.RestaurantName,
		RestaurantUuid: request.RestaurantUuid,
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

// ListMenuItems returns all active menu items with their restaurant information.
// Supports optional filtering by restaurant name, full-text search, and ordering.
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
