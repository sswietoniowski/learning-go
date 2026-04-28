package app

import (
	"context"
	"strings"

	"eats/backend/common"
)

type CourierUUID struct {
	common.UUID
}

type Courier struct {
	CourierUUID CourierUUID
	Name        string
	PhoneNumber string
	City        string
}

type CourierRepository interface {
	RegisterCourier(ctx context.Context, courier Courier) error
}

func (s *Service) RegisterCourier(ctx context.Context, courier Courier) error {
	errDetails := []common.ErrorDetails{}

	if courier.CourierUUID.IsZero() {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "courier",
			EntityID:   "",
			ErrorSlug:  "empty-uuid",
			Message:    "UUID cannot be empty",
		})
	}

	if strings.TrimSpace(courier.Name) == "" {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "courier",
			EntityID:   courier.CourierUUID.String(),
			ErrorSlug:  "empty-name",
			Message:    "Name cannot be empty",
		})
	}

	if strings.TrimSpace(courier.PhoneNumber) == "" {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "courier",
			EntityID:   courier.CourierUUID.String(),
			ErrorSlug:  "empty-phone-number",
			Message:    "Phone number cannot be empty",
		})
	}

	if strings.TrimSpace(courier.City) == "" {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "courier",
			EntityID:   courier.CourierUUID.String(),
			ErrorSlug:  "empty-city",
			Message:    "City cannot be empty",
		})
	}

	if len(errDetails) > 0 {
		return common.NewInvalidInputError(
			"invalid_courier_data",
			"Invalid courier data",
		).WithDetails(errDetails)
	}

	return s.courierRepository.RegisterCourier(ctx, courier)
}
