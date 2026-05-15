package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"eats/backend/billing/api/module/client"
	"eats/backend/common"
	"eats/backend/common/log"
	"eats/backend/common/shared"
	deliveryModule "eats/backend/delivery/api/module/client"
)

type QuoteUUID struct {
	common.UUID
}

type Quote struct {
	QuoteUUID      QuoteUUID
	CustomerUUID   CustomerUUID
	RestaurantUUID RestaurantUUID

	DeliveryAddress shared.Address

	ItemsSubtotalGross decimal.Decimal
	ServiceFeeGross    decimal.Decimal
	DeliveryFeeGross   decimal.Decimal
	TotalAmountGross   decimal.Decimal
	TotalTax           decimal.Decimal

	Currency shared.Currency

	CreatedAt time.Time
}

func (c Quote) Expired() bool {
	return time.Now().After(c.ExpirationTime())
}

func (c Quote) ExpirationTime() time.Time {
	return c.CreatedAt.Add(15 * time.Minute)
}

type QuoteMenuItem struct {
	MenuItemUUID RestaurantMenuItemUUID

	Category   ItemCategory
	GrossPrice decimal.Decimal
	Quantity   int
}

type OrderUUID struct {
	common.UUID
}

func NewOrderFromQuote(quote Quote) (Order, error) {
	var err error

	if quote.QuoteUUID.IsZero() {
		err = errors.Join(err, fmt.Errorf("quote UUID cannot be empty"))
	}
	if quote.CustomerUUID.IsZero() {
		err = errors.Join(err, fmt.Errorf("customer UUID cannot be empty"))
	}
	if quote.RestaurantUUID.IsZero() {
		err = errors.Join(err, fmt.Errorf("restaurant UUID cannot be empty"))
	}
	if quote.DeliveryAddress.IsZero() {
		err = errors.Join(err, fmt.Errorf("delivery address cannot be empty"))
	}
	if quote.ServiceFeeGross.IsZero() {
		err = errors.Join(err, fmt.Errorf("service fee cannot be zero"))
	}
	if quote.TotalAmountGross.IsZero() {
		err = errors.Join(err, fmt.Errorf("total amount cannot be zero"))
	}
	if quote.ItemsSubtotalGross.IsZero() {
		err = errors.Join(err, fmt.Errorf("items subtotal gross cannot be zero"))
	}
	if quote.DeliveryFeeGross.IsZero() {
		err = errors.Join(err, fmt.Errorf("delivery fee cannot be zero"))
	}
	// TotalTax may be zero for countries with 0% tax on food (e.g. GB).
	if quote.Currency.IsZero() {
		err = errors.Join(err, fmt.Errorf("currency cannot be empty"))
	}
	if err != nil {
		// it's not common.NewInvalidInputError because it's internal error if any of these happens here
		return Order{}, fmt.Errorf("invalid quote for creating order: %w", err)
	}

	return Order{
		OrderUUID{common.NewUUIDv7()},
		quote.QuoteUUID,
		quote.CustomerUUID,
		quote.RestaurantUUID,
		nil,
		quote.DeliveryAddress,
		time.Now(),
		nil,
		nil,
		nil,
		nil,
		nil,
		quote.ItemsSubtotalGross,
		quote.ServiceFeeGross,
		quote.DeliveryFeeGross,
		quote.TotalAmountGross,
		quote.TotalTax,
		quote.Currency,
	}, nil
}

type Order struct {
	OrderUUID OrderUUID
	QuoteUUID QuoteUUID

	CustomerUUID   CustomerUUID
	RestaurantUUID RestaurantUUID
	CourierUUID    *CourierUUID

	DeliveryAddress shared.Address

	OrderedAt             time.Time
	RestaurantConfirmedAt *time.Time
	CourierAcceptedAt     *time.Time
	RestaurantPreparedAt  *time.Time
	PickedUpAt            *time.Time
	DeliveredAt           *time.Time

	ItemsSubtotal    decimal.Decimal
	ServiceFeeGross  decimal.Decimal
	DeliveryFeeGross decimal.Decimal
	TotalAmountGross decimal.Decimal
	TotalTax         decimal.Decimal

	Currency shared.Currency
}

type OrderRepository interface {
	GetRestaurant(
		ctx context.Context,
		restaurantUUID RestaurantUUID,
	) (Restaurant, error)

	GetMenuItems(
		ctx context.Context,
		restaurantUUID RestaurantUUID,
		menuItemUUIDs []RestaurantMenuItemUUID,
	) (map[RestaurantMenuItemUUID]MenuItem, error)

	CreateQuote(
		ctx context.Context,
		restaurantUUID RestaurantUUID,
		menuItems CreateQuoteItems,
		updateFn func(
			ctx context.Context,
			menuItems map[RestaurantMenuItemUUID]MenuItem,
			restaurant Restaurant,
		) (Quote, []QuoteMenuItem, error),
	) (Quote, error)

	QuoteWithMenuItems(ctx context.Context, quoteUUID QuoteUUID) (Quote, map[RestaurantMenuItemUUID]MenuItem, error)

	SaveOrder(ctx context.Context, order Order) error

	UpdateOrder(
		ctx context.Context,
		orderUUID OrderUUID,
		updateFn func(ctx context.Context, order Order) (Order, error),
	) error

	OrderByID(ctx context.Context, orderUUID OrderUUID) (Order, error)
	OrderItemsByOrderID(ctx context.Context, orderUUID OrderUUID) ([]OrderItem, error)
}

type CreateQuote struct {
	CustomerUUID    CustomerUUID
	RestaurantUUID  RestaurantUUID
	QuoteItems      []CreateQuoteItem
	DeliveryAddress shared.Address
}

type CreateQuoteItem struct {
	MenuItemUUID RestaurantMenuItemUUID
	Quantity     int
}

type OrderItem struct {
	Name       string
	Category   ItemCategory
	GrossPrice decimal.Decimal
	Quantity   int
}

type CreateQuoteItems []CreateQuoteItem

func (c CreateQuoteItems) MenuItemUUIDs() []RestaurantMenuItemUUID {
	uuids := make([]RestaurantMenuItemUUID, 0, len(c))
	for _, item := range c {
		uuids = append(uuids, item.MenuItemUUID)
	}
	return uuids
}

func (s *Service) CreateQuote(ctx context.Context, req CreateQuote) (Quote, error) {
	var validationErrors []common.ErrorDetails

	if len(req.QuoteItems) == 0 {
		validationErrors = append(validationErrors, common.ErrorDetails{
			EntityType: "quote",
			ErrorSlug:  "empty-order",
			Message:    "at least one menu position must be included in the quote",
		})
	}
	for _, pos := range req.QuoteItems {
		if pos.Quantity <= 0 {
			validationErrors = append(validationErrors, common.ErrorDetails{
				EntityType: "menu_item",
				EntityID:   pos.MenuItemUUID.String(),
				ErrorSlug:  "invalid-quantity",
				Message:    "menu position quantity must be greater than zero",
			})
		}
	}

	if req.DeliveryAddress.IsZero() {
		validationErrors = append(validationErrors, common.ErrorDetails{
			EntityType: "quote",
			ErrorSlug:  "empty-delivery-address",
			Message:    "delivery address cannot be empty",
		})
	}
	if len(validationErrors) > 0 {
		return Quote{}, common.NewInvalidInputError(
			"invalid-quote-data",
			"invalid quote data",
		).WithDetails(validationErrors)
	}

	quoteItems := make(CreateQuoteItems, 0, len(req.QuoteItems))
	for _, item := range req.QuoteItems {
		quoteItems = append(quoteItems, CreateQuoteItem{
			MenuItemUUID: item.MenuItemUUID,
			Quantity:     item.Quantity,
		})
	}

	restaurant, err := s.orderRepository.GetRestaurant(ctx, req.RestaurantUUID)
	if err != nil {
		return Quote{}, err
	}

	menuItems, err := s.orderRepository.GetMenuItems(ctx, req.RestaurantUUID, quoteItems.MenuItemUUIDs())
	if err != nil {
		return Quote{}, err
	}

	if err := ensureQuoteItemsAreNotArchived(menuItems); err != nil {
		return Quote{}, err
	}

	if restaurant.Address.City() != req.DeliveryAddress.City() {
		return Quote{}, common.NewInvalidInputError(
			"address-out-of-delivery-zone",
			"restaurant does not deliver to the provided address",
		).WithDetails([]common.ErrorDetails{{
			EntityType: "quote",
			ErrorSlug:  "address-out-of-delivery-zone",
			Message:    fmt.Sprintf("restaurant delivers to %s only", restaurant.Address.City()),
		}})
	}

	itemsSubtotal := decimal.Zero
	quoteItemPositions := make([]QuoteMenuItem, 0, len(menuItems))
	billingLineItems := make([]client.LineItem, 0, len(menuItems))

	for _, item := range quoteItems {
		menuItem := menuItems[item.MenuItemUUID]
		grossPriceTotal := menuItem.GrossPrice.Mul(decimal.NewFromInt(int64(item.Quantity)))
		itemsSubtotal = itemsSubtotal.Add(grossPriceTotal)

		quoteItemPositions = append(quoteItemPositions, QuoteMenuItem{
			MenuItemUUID: item.MenuItemUUID,
			Category:     menuItem.Category,
			GrossPrice:   menuItem.GrossPrice,
			Quantity:     item.Quantity,
		})

		itemType, err := lineItemTypeFromCategory(menuItem.Category)
		if err != nil {
			return Quote{}, err
		}

		billingLineItems = append(billingLineItems, client.LineItem{
			Name:       menuItem.Name,
			Type:       itemType,
			Quantity:   item.Quantity,
			UnitAmount: shared.NewGrossAmount(menuItem.GrossPrice),
		})
	}

	serviceFeeGross := itemsSubtotal.Mul(decimal.RequireFromString("0.06")).RoundBank(2) // 6%

	billingLineItems = append(billingLineItems, client.LineItem{
		Name:       "Service Fee",
		Type:       shared.LineItemTypeService,
		Quantity:   1,
		UnitAmount: shared.NewGrossAmount(serviceFeeGross),
	})

	// Call the delivery and billing services before starting the transaction. If these ran inside
	// the transaction, a slow response would hold a database connection the entire time. If
	// other modules share the same database, exhausting the pool this way is a self-inflicted DDoS.
	// In production, use a separate database user per module with its own connection limit.
	deliveryFee, err := s.modules.CalculateDeliveryFee(
		ctx,
		deliveryModule.CalculateDeliveryFeeRequest{
			RestaurantAddress: restaurant.Address,
			DeliveryAddress:   req.DeliveryAddress,
			Currency:          restaurant.Currency,
			When:              time.Now(),
		},
	)
	if err != nil {
		return Quote{}, fmt.Errorf("error calculating delivery fee for quote: %w", err)
	}

	billingLineItems = append(billingLineItems, client.LineItem{
		Name:       "Delivery Fee",
		Type:       shared.LineItemTypeDelivery,
		Quantity:   1,
		UnitAmount: shared.NewGrossAmount(deliveryFee.GrossFee),
	})

	billingRes, err := s.modules.CalculateTaxes(ctx, client.CalculateTaxesRequest{
		Currency:          restaurant.Currency,
		BuyerCountryCode:  req.DeliveryAddress.CountryCode(),
		BuyerTaxID:        nil,
		SellerCountryCode: req.DeliveryAddress.CountryCode(),
		LineItems:         billingLineItems,
	})
	if err != nil {
		return Quote{}, fmt.Errorf("error calculating taxes for quote: %w", err)
	}

	return s.orderRepository.CreateQuote(
		ctx,
		req.RestaurantUUID,
		quoteItems,
		func(
			ctx context.Context,
			menuItems map[RestaurantMenuItemUUID]MenuItem,
			restaurant Restaurant,
		) (Quote, []QuoteMenuItem, error) {
			// Re-validate inside the transaction for consistency: menu items or restaurant data
			// may have changed between the pre-transaction reads and the commit.
			if err := ensureQuoteItemsAreNotArchived(menuItems); err != nil {
				return Quote{}, nil, err
			}

			if restaurant.Address.City() != req.DeliveryAddress.City() {
				return Quote{}, nil, common.NewInvalidInputError(
					"address-out-of-delivery-zone",
					"restaurant does not deliver to the provided address",
				).WithDetails([]common.ErrorDetails{{
					EntityType: "quote",
					ErrorSlug:  "address-out-of-delivery-zone",
					Message:    fmt.Sprintf("restaurant delivers to %s only", restaurant.Address.City()),
				}})
			}

			return Quote{
				QuoteUUID:      QuoteUUID{common.NewUUIDv7()},
				CustomerUUID:   req.CustomerUUID,
				RestaurantUUID: req.RestaurantUUID,

				DeliveryAddress: req.DeliveryAddress,

				ItemsSubtotalGross: itemsSubtotal,
				ServiceFeeGross:    serviceFeeGross,
				DeliveryFeeGross:   deliveryFee.GrossFee,
				TotalAmountGross:   billingRes.GrossTotal,

				TotalTax: billingRes.TaxTotal,

				Currency: restaurant.Currency,
			}, quoteItemPositions, nil
		},
	)
}

type PlaceOrder struct {
	CustomerUUID CustomerUUID
	QuoteUUID    QuoteUUID
	PaymentNonce string
}

func (s *Service) PlaceOrder(ctx context.Context, req PlaceOrder) (Order, error) {
	quote, menuItems, err := s.orderRepository.QuoteWithMenuItems(ctx, req.QuoteUUID)
	if err != nil {
		return Order{}, fmt.Errorf("error reading quote: %w", err)
	}

	if err := checkCustomerMatch(quote.CustomerUUID, req.CustomerUUID); err != nil {
		return Order{}, err
	}

	if quote.Expired() {
		// Frontend should handle by requesting a new quote and retrying.
		return Order{}, common.NewExpiredError(
			"quote-expired",
			"quote has expired",
		)
	}

	if err := ensureQuoteItemsAreNotArchived(menuItems); err != nil {
		return Order{}, err
	}

	order, err := NewOrderFromQuote(quote)
	if err != nil {
		return Order{}, fmt.Errorf("error creating order from quote: %w", err)
	}

	// CapturePayment is called outside the transaction.
	// CapturePayment is idempotent (nonce ensures single charge), so retrying the whole
	// PlaceOrder is safe for the payment side.
	//
	// If CapturePayment succeeds but SaveOrder fails, the payment was captured but the order
	// is not saved. A reconciliation process is needed to handle this edge case.
	// An event-driven approach would be the proper solution. See https://threedots.tech/event-driven/
	err = s.paymentsService.CapturePayment(ctx, req.PaymentNonce, quote.TotalAmountGross, quote.RestaurantUUID.String())
	if err != nil {
		return Order{}, fmt.Errorf("error charging card for order: %w", err)
	}

	// SaveOrder persists the order inside a transaction.
	err = s.orderRepository.SaveOrder(ctx, order)
	if err != nil {
		return Order{}, fmt.Errorf("error saving order: %w", err)
	}

	return order, nil
}

func ensureQuoteItemsAreNotArchived(menuItems map[RestaurantMenuItemUUID]MenuItem) error {
	var archivedPositions []MenuItem
	for _, item := range menuItems {
		if item.IsArchived {
			archivedPositions = append(archivedPositions, item)
		}
	}

	if len(archivedPositions) == 0 {
		return nil
	}

	details := make([]common.ErrorDetails, 0, len(archivedPositions))
	for _, item := range archivedPositions {
		details = append(details, common.ErrorDetails{
			EntityType: "menu_item",
			EntityID:   item.MenuItemUUID.String(),
			ErrorSlug:  "archived-menu-position",
			Message:    fmt.Sprintf("menu position '%s' is archived", item.Name),
		})
	}

	return common.NewExpiredError(
		"unavailable-menu-items",
		"one or more menu items are not available",
	).WithInternalError(fmt.Errorf(
		"archived menu items in order: %v",
		archivedPositions,
	)).WithDetails(details)
}

type AvailableDelivery struct {
	RestaurantUUID    uuid.UUID
	OrderUUID         uuid.UUID
	RestaurantName    string
	RestaurantAddress string
}

func (s *Service) AcceptOrder(
	ctx context.Context,
	restaurantUUID RestaurantUUID,
	orderUUID OrderUUID,
) error {
	return s.orderRepository.UpdateOrder(
		ctx,
		orderUUID,
		func(ctx context.Context, order Order) (Order, error) {
			if err := checkRestaurantMatch(order.RestaurantUUID, restaurantUUID); err != nil {
				return Order{}, err
			}

			if order.RestaurantConfirmedAt != nil {
				log.FromContext(ctx).With("order_uuid", orderUUID).Warn("Order already confirmed")
				return order, nil
			}
			order.RestaurantConfirmedAt = common.ToPtr(time.Now())

			return order, nil
		},
	)
}

func (s *Service) MarkAsPrepared(
	ctx context.Context,
	restaurantUUID RestaurantUUID,
	orderUUID OrderUUID,
) error {
	return s.orderRepository.UpdateOrder(
		ctx,
		orderUUID,
		func(ctx context.Context, order Order) (Order, error) {
			if err := checkRestaurantMatch(order.RestaurantUUID, restaurantUUID); err != nil {
				return Order{}, err
			}

			if order.RestaurantPreparedAt != nil {
				log.FromContext(ctx).With("order_uuid", orderUUID).Warn("Order already marked as ready")
				return order, nil
			}
			order.RestaurantPreparedAt = common.ToPtr(time.Now())

			return order, nil
		},
	)
}

func checkRestaurantMatch(orderRestaurant RestaurantUUID, restaurantUUID RestaurantUUID) error {
	if orderRestaurant.Equals(restaurantUUID.UUID) {
		return nil
	}

	return common.NewForbiddenError(
		"invalid-restaurant",
		"order does not belong to the restaurant",
	).WithInternalError(fmt.Errorf(
		"order restaurant %s does not match provided restaurant %s",
		orderRestaurant,
		restaurantUUID,
	))
}
