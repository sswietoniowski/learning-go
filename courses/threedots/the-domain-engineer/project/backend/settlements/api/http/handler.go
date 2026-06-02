package http

import (
	"context"
	"fmt"

	"eats/backend/common"
	"eats/backend/common/shared"
	"eats/backend/settlements/app/command"
	"eats/backend/settlements/app/query"
	"eats/backend/settlements/domain"
)

type Handler struct {
	commandHandler *command.Handlers
	queryHandler   *query.Handlers
}

func NewHandler(commandHandler *command.Handlers, queryHandler *query.Handlers) *Handler {
	if commandHandler == nil {
		panic("command handler is required")
	}
	if queryHandler == nil {
		panic("query handler is required")
	}

	return &Handler{
		commandHandler: commandHandler,
		queryHandler:   queryHandler,
	}
}

func (h Handler) OnboardPartner(ctx context.Context, request OnboardPartnerRequestObject) (OnboardPartnerResponseObject, error) {
	if request.Params.OperatorUUID.IsZero() {
		return nil, common.NewUnauthorizedError("missing-operator-uuid", "operator UUID is required")
	}

	address, err := shared.NewAddress(
		request.Body.Address.Line1,
		request.Body.Address.Line2,
		request.Body.Address.PostalCode,
		request.Body.Address.City,
		request.Body.Address.CountryCode,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating address: %w", err)
	}

	taxID, err := shared.NewTaxID(request.Body.TaxId)
	if err != nil {
		return nil, fmt.Errorf("error creating tax ID: %w", err)
	}

	bankAccount, err := domain.NewIBAN(request.Body.BankAccountIban)
	if err != nil {
		return nil, fmt.Errorf("error creating bank account IBAN: %w", err)
	}

	cmd := command.OnboardPartner{
		PartnerUUID:        request.Body.PartnerUuid,
		PlatformEntityUUID: request.Body.PlatformEntityUuid,
		PartnerType:        request.Body.PartnerType,
		BusinessName:       request.Body.BusinessName,
		TaxID:              taxID,
		Address:            address,
		BankAccountNumber:  bankAccount,
		Currency:           request.Body.Currency,
	}

	if err := h.commandHandler.OnboardPartner(ctx, cmd); err != nil {
		return nil, err
	}

	return OnboardPartner204Response{}, nil
}

func (h Handler) CreatePlatformEntity(ctx context.Context, request CreatePlatformEntityRequestObject) (CreatePlatformEntityResponseObject, error) {
	if request.Params.OperatorUUID.IsZero() {
		return nil, common.NewUnauthorizedError("missing-operator-uuid", "operator UUID is required")
	}

	address, err := shared.NewAddress(
		request.Body.Address.Line1,
		request.Body.Address.Line2,
		request.Body.Address.PostalCode,
		request.Body.Address.City,
		request.Body.Address.CountryCode,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating address: %w", err)
	}

	taxID, err := shared.NewTaxID(request.Body.TaxId)
	if err != nil {
		return nil, fmt.Errorf("error creating tax ID: %w", err)
	}

	bankAccount, err := domain.NewIBAN(request.Body.BankAccountIban)
	if err != nil {
		return nil, fmt.Errorf("error creating bank account IBAN: %w", err)
	}

	cmd := command.CreatePlatformEntity{
		PlatformEntityUUID: request.Body.PlatformEntityUuid,
		BusinessName:       request.Body.BusinessName,
		TaxID:              taxID,
		Address:            address,
		BankAccountNumber:  bankAccount,
		Currency:           request.Body.Currency,
	}

	uuid, err := h.commandHandler.CreatePlatformEntity(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return CreatePlatformEntity201JSONResponse{
		PlatformEntityUuid: uuid,
	}, nil
}

func (h Handler) GetBillingCyclesByPartner(ctx context.Context, request GetBillingCyclesByPartnerRequestObject) (GetBillingCyclesByPartnerResponseObject, error) {
	cycles, err := h.queryHandler.BillingCycleByPartner(ctx, query.BillingCycleByPartner{
		PartnerUUID: request.PartnerUuid,
	})
	if err != nil {
		return nil, err
	}

	result := make([]BillingCycle, 0, len(cycles))
	for _, c := range cycles {
		bc := BillingCycle{
			BillingCycleUuid:   c.BillingCycleUUID,
			PartnerUuid:        c.PartnerUUID,
			BillingCycleNumber: c.BillingCycleNumber,
			StartDate:          c.StartDate,
			EndDate:            c.EndDate,
			Closed:             c.Closed,
			Settled:            c.Settled,
		}
		result = append(result, bc)
	}

	return GetBillingCyclesByPartner200JSONResponse(result), nil
}

func Register(e EchoRouter, commandHandlers *command.Handlers, queryHandlers *query.Handlers) error {
	handler := NewHandler(commandHandlers, queryHandlers)

	RegisterHandlers(e, NewStrictHandler(handler, nil))

	return nil
}
