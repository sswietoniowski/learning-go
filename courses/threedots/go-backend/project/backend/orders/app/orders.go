package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/shopspring/decimal"

	"eats/backend/common"
	"eats/backend/common/shared"
	"eats/backend/delivery/api/module/client"
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

	GrossPrice decimal.Decimal
	Quantity   int
}

type OrderRepository interface {
	CreateQuote(
		ctx context.Context,
		restaurantUUID RestaurantUUID,
		menuItems CreateQuoteItems,
		updateFn func(
			ctx context.Context,
			menuItems map[RestaurantMenuItemUUID]MenuItem,
			r Restaurant,
		) (Quote, []QuoteMenuItem, error),
	) (Quote, error)
	GetRestaurant(ctx context.Context, restaurantUUID RestaurantUUID) (Restaurant, error)
	GetQuote(ctx context.Context, quoteUUID QuoteUUID) (Quote, error)
	GetMenuItemsForQuote(ctx context.Context, quoteUUID QuoteUUID, restaurantUUID RestaurantUUID) (map[RestaurantMenuItemUUID]MenuItem, error)
	PlaceOrder(ctx context.Context, order Order) error
	OrderByID(ctx context.Context, orderUUID OrderUUID) (Order, error)
	UpdateOrder(ctx context.Context, orderUUID OrderUUID, updateFn func(ctx context.Context, order Order) (Order, error)) error
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

	deliveryFee, err := s.modules.CalculateDeliveryFee(ctx, client.CalculateDeliveryFeeRequest{
		RestaurantAddress: restaurant.Address,
		DeliveryAddress:   req.DeliveryAddress,
		Currency:          restaurant.Currency,
		When:              time.Now(),
	})
	if err != nil {
		return Quote{}, fmt.Errorf("error calculating delivery fee for quote: %w", err)
	}

	return s.orderRepository.CreateQuote(
		ctx,
		req.RestaurantUUID,
		quoteItems,
		func(
			ctx context.Context,
			menuItems map[RestaurantMenuItemUUID]MenuItem,
			r Restaurant,
		) (Quote, []QuoteMenuItem, error) {
			// Re-validate inside the transaction for consistency: menu items or restaurant data
			// may have changed between the pre-transaction reads and the commit.
			if err := ensureQuoteItemsAreNotArchived(menuItems); err != nil {
				return Quote{}, nil, err
			}

			if r.Address.City != req.DeliveryAddress.City {
				return Quote{}, nil, common.NewInvalidInputError(
					"address-out-of-delivery-zone",
					"restaurant does not deliver to the provided address",
				).WithDetails([]common.ErrorDetails{{
					EntityType: "quote",
					ErrorSlug:  "address-out-of-delivery-zone",
					Message:    fmt.Sprintf("restaurant delivers to %s only", r.Address.City),
				}})
			}

			itemsSubtotal := decimal.Zero
			quoteItemPositions := make([]QuoteMenuItem, 0, len(menuItems))

			for _, item := range quoteItems {
				menuItem := menuItems[item.MenuItemUUID]
				grossPriceTotal := menuItem.GrossPrice.Mul(decimal.NewFromInt(int64(item.Quantity)))
				itemsSubtotal = itemsSubtotal.Add(grossPriceTotal)

				quoteItemPositions = append(quoteItemPositions, QuoteMenuItem{
					MenuItemUUID: item.MenuItemUUID,
					GrossPrice:   grossPriceTotal,
					Quantity:     item.Quantity,
				})
			}

			serviceFeeGross := itemsSubtotal.Mul(decimal.RequireFromString("0.06")).RoundBank(2) // 6%

			totalAmount := itemsSubtotal.Add(serviceFeeGross).Add(deliveryFee.GrossFee)

			return Quote{
				QuoteUUID:      QuoteUUID{common.NewUUIDv7()},
				CustomerUUID:   req.CustomerUUID,
				RestaurantUUID: req.RestaurantUUID,

				DeliveryAddress: req.DeliveryAddress,

				ItemsSubtotalGross: itemsSubtotal,
				ServiceFeeGross:    serviceFeeGross,
				DeliveryFeeGross:   deliveryFee.GrossFee,
				TotalAmountGross:   totalAmount,

				TotalTax: totalAmount.Div(decimal.RequireFromString("1.23")).RoundBank(2),

				Currency: r.Currency,
			}, quoteItemPositions, nil
		},
	)
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

type OrderUUID struct {
	common.UUID
}

type Order struct {
	OrderUUID             OrderUUID
	QuoteUUID             QuoteUUID
	CustomerUUID          CustomerUUID
	RestaurantUUID        RestaurantUUID
	RestaurantName        string
	CourierUUID           *CourierUUID
	DeliveryAddress       shared.Address
	OrderedAt             time.Time
	RestaurantConfirmedAt *time.Time
	CourierAcceptedAt     *time.Time
	RestaurantPreparedAt  *time.Time
	PickedUpAt            *time.Time
	DeliveredAt           *time.Time
	ItemsSubtotalGross    decimal.Decimal
	ServiceFeeGross       decimal.Decimal
	DeliveryFeeGross      decimal.Decimal
	TotalAmountGross      decimal.Decimal
	TotalTax              decimal.Decimal
	Currency              shared.Currency
}

func NewOrderFromQuote(quote Quote) (Order, error) {
	return Order{
		OrderUUID:          OrderUUID{common.NewUUIDv7()},
		QuoteUUID:          quote.QuoteUUID,
		CustomerUUID:       quote.CustomerUUID,
		RestaurantUUID:     quote.RestaurantUUID,
		DeliveryAddress:    quote.DeliveryAddress,
		OrderedAt:          time.Now(),
		ItemsSubtotalGross: quote.ItemsSubtotalGross,
		ServiceFeeGross:    quote.ServiceFeeGross,
		DeliveryFeeGross: quote.DeliveryFeeGross,
		TotalAmountGross: quote.TotalAmountGross,
		TotalTax:         quote.TotalTax,
		Currency:         quote.Currency,
	}, nil
}

func (s *Service) PlaceOrder(ctx context.Context, customerUUID CustomerUUID, quoteUUID QuoteUUID, paymentNonce string) (Order, error) {
	quote, err := s.orderRepository.GetQuote(ctx, quoteUUID)
	if err != nil {
		return Order{}, err
	}

	if quote.CustomerUUID != customerUUID {
		return Order{}, common.NewForbiddenError("invalid-customer", "customer does not match the quote")
	}

	if quote.Expired() {
		return Order{}, common.NewExpiredError("quote-expired", "quote has expired")
	}

	menuItems, err := s.orderRepository.GetMenuItemsForQuote(ctx, quoteUUID, quote.RestaurantUUID)
	if err != nil {
		return Order{}, err
	}

	if err := ensureQuoteItemsAreNotArchived(menuItems); err != nil {
		return Order{}, err
	}

	restaurant, err := s.orderRepository.GetRestaurant(ctx, quote.RestaurantUUID)
	if err != nil {
		return Order{}, err
	}

	if err := s.paymentsClient.CapturePayment(ctx, paymentNonce, quote.TotalAmountGross, quote.RestaurantUUID.String()); err != nil {
		return Order{}, fmt.Errorf("payment capture failed: %w", err)
	}

	order, err := NewOrderFromQuote(quote)
	if err != nil {
		return Order{}, err
	}
	order.CustomerUUID = customerUUID
	order.RestaurantName = restaurant.Name

	if err := s.orderRepository.PlaceOrder(ctx, order); err != nil {
		return Order{}, err
	}

	return order, nil
}

func (s *Service) AcceptOrder(ctx context.Context, restaurantUUID RestaurantUUID, orderUUID OrderUUID) error {
	return s.orderRepository.UpdateOrder(ctx, orderUUID, func(ctx context.Context, order Order) (Order, error) {
		if order.RestaurantUUID != restaurantUUID {
			return Order{}, common.NewForbiddenError("wrong-restaurant", "restaurant does not own this order")
		}
		if order.RestaurantConfirmedAt != nil {
			slog.WarnContext(ctx, "order already accepted", "order_uuid", orderUUID)
			return order, nil
		}
		now := time.Now()
		order.RestaurantConfirmedAt = &now
		return order, nil
	})
}

func (s *Service) MarkOrderReadyForPickup(ctx context.Context, restaurantUUID RestaurantUUID, orderUUID OrderUUID) error {
	return s.orderRepository.UpdateOrder(ctx, orderUUID, func(ctx context.Context, order Order) (Order, error) {
		if order.RestaurantUUID != restaurantUUID {
			return Order{}, common.NewForbiddenError("wrong-restaurant", "restaurant does not own this order")
		}
		if order.RestaurantPreparedAt != nil {
			slog.WarnContext(ctx, "order already marked ready for pickup", "order_uuid", orderUUID)
			return order, nil
		}
		now := time.Now()
		order.RestaurantPreparedAt = &now
		return order, nil
	})
}
