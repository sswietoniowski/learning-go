package command

import (
	"context"
	"fmt"

	"eats/backend/common/shared"
	"eats/backend/settlements/app/models"
	"eats/backend/settlements/domain"
)

type CreatePlatformEntity struct {
	PlatformEntityUUID domain.LegalEntityUUID
	BusinessName       string
	TaxID              shared.TaxID
	Address            shared.Address
	BankAccountNumber  domain.IBAN
	Currency           shared.Currency
}

func (h *Handlers) CreatePlatformEntity(ctx context.Context, cmd CreatePlatformEntity) (models.PlatformEntityUUID, error) {
	legalEntity, err := models.NewLegalEntity(
		cmd.PlatformEntityUUID,
		models.LegalEntityPlatform,
		cmd.BusinessName,
		cmd.TaxID,
		cmd.Address,
		cmd.BankAccountNumber,
		cmd.Currency,
	)
	if err != nil {
		return models.PlatformEntityUUID{}, fmt.Errorf("error creating legal entity: %w", err)
	}

	err = h.legalEntityRepository.SavePlatformEntity(ctx, legalEntity)
	if err != nil {
		return models.PlatformEntityUUID{}, fmt.Errorf("could not save legal entity: %w", err)
	}

	return models.PlatformEntityUUID{legalEntity.UUID}, nil
}
