package app

import (
	"context"
	"strings"

	"github.com/shopspring/decimal"

	"eats/backend/common"
	"eats/backend/common/shared"
)

type RestaurantRepository interface {
	UpsertRestaurant(ctx context.Context, restaurantUUID RestaurantUUID, restaurant OnboardRestaurant) error
}

type RestaurantUUID struct {
	common.UUID
}

type OnboardRestaurant struct {
	Name        string
	Address     shared.Address
	Currency    shared.Currency
	Description string
	MenuItems   []MenuItem
}

type RestaurantMenuItemUUID struct {
	common.UUID
}

type RestaurantMenu struct {
	RestaurantName string
	Address        shared.Address
	Description    string
	Currency       shared.Currency
	Positions      []MenuItem
}

type MenuItem struct {
	MenuItemUUID RestaurantMenuItemUUID

	Name       string
	Ordering   float64
	GrossPrice decimal.Decimal

	IsArchived bool
}

type Restaurant struct {
	RestaurantUUID RestaurantUUID
	Name           string
	Address        shared.Address
	Description    string
	Currency       shared.Currency
}

func (s *Service) OnboardRestaurant(ctx context.Context, restaurantUUID RestaurantUUID, req OnboardRestaurant) error {
	errorDetails := []common.ErrorDetails{}

	if restaurantUUID.IsZero() {
		errorDetails = append(errorDetails, common.ErrorDetails{
			EntityType: "restaurant",
			ErrorSlug:  "invalid-uuid",
			Message:    "restaurant UUID cannot be empty",
		})
	}

	if strings.TrimSpace(req.Name) == "" {
		errorDetails = append(errorDetails, common.ErrorDetails{
			EntityType: "restaurant",
			EntityID:   restaurantUUID.String(),
			ErrorSlug:  "invalid-name",
			Message:    "restaurant name cannot be empty",
		})
	}
	if req.Address.IsZero() {
		errorDetails = append(errorDetails, common.ErrorDetails{
			EntityType: "restaurant",
			EntityID:   restaurantUUID.String(),
			ErrorSlug:  "invalid-address",
			Message:    "restaurant address cannot be empty",
		})
	}
	if strings.TrimSpace(req.Description) == "" {
		errorDetails = append(errorDetails, common.ErrorDetails{
			EntityType: "restaurant",
			EntityID:   restaurantUUID.String(),
			ErrorSlug:  "invalid-description",
			Message:    "restaurant description cannot be empty",
		})
	}
	if len(req.MenuItems) == 0 {
		errorDetails = append(errorDetails, common.ErrorDetails{
			EntityType: "restaurant",
			EntityID:   restaurantUUID.String(),
			ErrorSlug:  "invalid-menu",
			Message:    "restaurant must have at least one menu position",
		})
	}

	// Here's a good example how encapsulation could help: a constructor making these
	// validations would keep this section simpler.
	for _, pos := range req.MenuItems {
		if strings.TrimSpace(pos.Name) == "" {
			errorDetails = append(errorDetails, common.ErrorDetails{
				EntityType: "menu_item",
				EntityID:   pos.MenuItemUUID.String(),
				ErrorSlug:  "invalid-name",
				Message:    "menu position name cannot be empty",
			})
		}
		if pos.GrossPrice.LessThanOrEqual(decimal.Zero) {
			errorDetails = append(errorDetails, common.ErrorDetails{
				EntityType: "menu_item",
				EntityID:   pos.MenuItemUUID.String(),
				ErrorSlug:  "invalid-price",
				Message:    "menu position price must be greater than zero",
			})
		}
	}
	if len(errorDetails) > 0 {
		return common.NewInvalidInputError(
			"invalid-restaurant-data",
			"invalid restaurant data",
		).WithDetails(errorDetails)
	}

	return s.restaurantRepository.UpsertRestaurant(ctx, restaurantUUID, req)
}
