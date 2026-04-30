package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"eats/backend/common"
	"eats/backend/common/log"
)

type CourierUUID struct {
	common.UUID
}

type RegisterCourier struct {
	Name        string
	PhoneNumber string
	City        string
}

type CourierRepository interface {
	RegisterCourier(ctx context.Context, courierUUID CourierUUID, courier RegisterCourier) error
	GetCourierCity(ctx context.Context, courierUUID CourierUUID) (string, error)
}

func (s *Service) RegisterCourier(ctx context.Context, req RegisterCourier) (CourierUUID, error) {
	courierUUID := CourierUUID{common.NewUUIDv7()}

	var validationDetails []common.ErrorDetails

	if strings.TrimSpace(req.Name) == "" {
		validationDetails = append(validationDetails, common.ErrorDetails{
			EntityType: "courier",
			EntityID:   courierUUID.String(),
			ErrorSlug:  "invalid-name",
			Message:    "courier name cannot be empty",
		})
	}
	if strings.TrimSpace(req.PhoneNumber) == "" {
		validationDetails = append(validationDetails, common.ErrorDetails{
			EntityType: "courier",
			EntityID:   courierUUID.String(),
			ErrorSlug:  "invalid-phone-number",
			Message:    "courier phone number cannot be empty",
		})
	}
	if strings.TrimSpace(req.City) == "" {
		validationDetails = append(validationDetails, common.ErrorDetails{
			EntityType: "courier",
			EntityID:   courierUUID.String(),
			ErrorSlug:  "invalid-city",
			Message:    "courier city cannot be empty",
		})
	}
	if len(validationDetails) > 0 {
		return CourierUUID{}, common.NewInvalidInputError(
			"invalid-courier-data",
			"invalid courier data",
		).WithDetails(validationDetails)
	}

	err := s.courierRepository.RegisterCourier(ctx, courierUUID, req)
	if err != nil {
		return CourierUUID{}, err
	}

	return courierUUID, nil
}

func (s *Service) CourierAcceptDelivery(ctx context.Context, courierUUID CourierUUID, orderUUID OrderUUID) error {
	courierCity, err := s.courierRepository.GetCourierCity(ctx, courierUUID)
	if err != nil {
		return fmt.Errorf("failed to get courier city: %w", err)
	}

	return s.orderRepository.UpdateOrder(
		ctx,
		orderUUID,
		func(ctx context.Context, order Order) (Order, error) {
			if order.CourierUUID != nil {
				return Order{}, common.NewConflictError(
					"already-accepted",
					"order already accepted by another courier",
				).WithInternalError(fmt.Errorf(
					"order courier %s does not match provided courier %s",
					order.CourierUUID,
					courierUUID,
				))
			}

			if order.DeliveryAddress.City != courierCity {
				return Order{}, common.NewInvalidInputError(
					"courier-out-of-delivery-zone",
					"courier cannot accept orders outside their delivery zone",
				).WithDetails([]common.ErrorDetails{{
					EntityType: "order",
					ErrorSlug:  "courier-out-of-delivery-zone",
					Message:    fmt.Sprintf("courier operates in %s only", courierCity),
				}})
			}

			order.CourierUUID = common.ToPtr(courierUUID)
			order.CourierAcceptedAt = common.ToPtr(time.Now())

			return order, nil
		},
	)
}

func (s *Service) CourierReportDeliveryPickup(ctx context.Context, courierUUID CourierUUID, orderUUID OrderUUID) error {
	return s.orderRepository.UpdateOrder(
		ctx,
		orderUUID,
		func(ctx context.Context, order Order) (Order, error) {
			if err := checkCourierMatch(order.CourierUUID, courierUUID); err != nil {
				return Order{}, err
			}

			if order.PickedUpAt != nil {
				// Idempotent: the first pickup timestamp matters, we don't want to overwrite it.
				log.FromContext(ctx).With("order_uuid", orderUUID).Warn("Order already marked as picked up")
				return order, nil
			}
			order.PickedUpAt = common.ToPtr(time.Now())

			return order, nil
		},
	)
}

func (s *Service) CourierReportDelivery(ctx context.Context, courierUUID CourierUUID, orderUUID OrderUUID) error {
	return s.orderRepository.UpdateOrder(
		ctx,
		orderUUID,
		func(ctx context.Context, order Order) (Order, error) {
			if err := checkCourierMatch(order.CourierUUID, courierUUID); err != nil {
				return Order{}, err
			}

			if order.DeliveredAt != nil {
				// Idempotent: the first delivery timestamp matters, we don't want to overwrite it.
				log.FromContext(ctx).With("order_uuid", orderUUID).Warn("Order already marked as delivered")
				return order, nil
			}
			order.DeliveredAt = common.ToPtr(time.Now())

			return order, nil
		},
	)
}

func checkCourierMatch(orderCourier *CourierUUID, courierUUID CourierUUID) error {
	if orderCourier == nil {
		return common.NewConflictError(
			"no-courier-assigned",
			"order does not have assigned courier",
		).WithInternalError(fmt.Errorf("order courier is nil, provided courier %s", courierUUID))
	}

	if orderCourier.Equals(courierUUID.UUID) {
		return nil
	}

	return common.NewForbiddenError(
		"invalid-courier",
		"order does not belong to the courier",
	).WithInternalError(fmt.Errorf(
		"order courier %s does not match provided courier %s",
		orderCourier,
		courierUUID,
	))
}

func checkCustomerMatch(orderCustomer CustomerUUID, customerUUID CustomerUUID) error {
	if orderCustomer.Equals(customerUUID.UUID) {
		return nil
	}

	return common.NewForbiddenError(
		"invalid-customer",
		"order does not belong to the customer",
	).WithInternalError(fmt.Errorf(
		"order customer %s does not match provided customer %s",
		orderCustomer,
		customerUUID,
	))
}
