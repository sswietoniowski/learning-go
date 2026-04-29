package app

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

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
	RegisterCourier(ctx context.Context, courierOrUUID any, maybeCourier ...Courier) error
	GetCourierCity(ctx context.Context, courierUUID CourierUUID) (string, error)
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

func (s *Service) AcceptDelivery(ctx context.Context, courierUUID CourierUUID, orderUUID OrderUUID) error {
	city, err := s.courierRepository.GetCourierCity(ctx, courierUUID)
	if err != nil {
		return err
	}

	return s.orderRepository.UpdateOrder(ctx, orderUUID, func(ctx context.Context, order Order) (Order, error) {
		if order.CourierAcceptedAt != nil {
			return Order{}, common.NewConflictError("already-accepted", "order already accepted by a courier")
		}
		if order.DeliveryAddress.City != city {
			return Order{}, common.NewInvalidInputError(
				"courier-out-of-delivery-zone",
				"courier cannot accept orders outside their delivery zone",
			).WithDetails([]common.ErrorDetails{{
				EntityType: "order",
				ErrorSlug:  "courier-out-of-delivery-zone",
				Message:    fmt.Sprintf("courier operates in %s only", city),
			}})
		}
		now := time.Now()
		order.CourierUUID = &courierUUID
		order.CourierAcceptedAt = &now
		return order, nil
	})
}

func (s *Service) ReportPickup(ctx context.Context, courierUUID CourierUUID, orderUUID OrderUUID) error {
	return s.orderRepository.UpdateOrder(ctx, orderUUID, func(ctx context.Context, order Order) (Order, error) {
		if order.CourierUUID == nil {
			return Order{}, common.NewConflictError("no-courier-assigned", "no courier is assigned to this order")
		}
		if *order.CourierUUID != courierUUID {
			return Order{}, common.NewForbiddenError("wrong-courier", "courier is not assigned to this order")
		}
		if order.PickedUpAt != nil {
			slog.WarnContext(ctx, "order already picked up", "order_uuid", orderUUID)
			return order, nil
		}
		now := time.Now()
		order.PickedUpAt = &now
		return order, nil
	})
}

func (s *Service) ReportDelivery(ctx context.Context, courierUUID CourierUUID, orderUUID OrderUUID) error {
	return s.orderRepository.UpdateOrder(ctx, orderUUID, func(ctx context.Context, order Order) (Order, error) {
		if order.CourierUUID == nil {
			return Order{}, common.NewConflictError("no-courier-assigned", "no courier is assigned to this order")
		}
		if *order.CourierUUID != courierUUID {
			return Order{}, common.NewForbiddenError("wrong-courier", "courier is not assigned to this order")
		}
		if order.DeliveredAt != nil {
			slog.WarnContext(ctx, "order already delivered", "order_uuid", orderUUID)
			return order, nil
		}
		now := time.Now()
		order.DeliveredAt = &now
		return order, nil
	})
}
