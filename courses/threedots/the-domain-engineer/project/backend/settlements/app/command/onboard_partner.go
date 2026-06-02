package command

import (
	"context"
	"fmt"

	"eats/backend/common/shared"
	"eats/backend/settlements/app/models"
	"eats/backend/settlements/domain"
)

type OnboardPartner struct {
	PartnerUUID        domain.LegalEntityUUID
	PlatformEntityUUID models.PlatformEntityUUID
	PartnerType        domain.PartnerType
	BusinessName       string
	TaxID              shared.TaxID
	Address            shared.Address
	BankAccountNumber  domain.IBAN
	Currency           shared.Currency
}

func (h *Handlers) OnboardPartner(ctx context.Context, cmd OnboardPartner) error {
	if cmd.PlatformEntityUUID.IsZero() {
		return fmt.Errorf("platform entity uuid cannot be zero")
	}

	legalEntity, err := models.NewLegalEntity(
		cmd.PartnerUUID,
		models.LegalEntityPartner,
		cmd.BusinessName,
		cmd.TaxID,
		cmd.Address,
		cmd.BankAccountNumber,
		cmd.Currency,
	)
	if err != nil {
		return fmt.Errorf("error creating legal entity: %w", err)
	}

	partner := models.NewPartner(legalEntity, cmd.PlatformEntityUUID)

	billingCycle, err := domain.NewInitialBillingCycle(legalEntity.UUID, cmd.PartnerType)
	if err != nil {
		return fmt.Errorf("could not create initial billing cycle: %w", err)
	}

	err = h.legalEntityRepository.SavePartner(ctx, partner, billingCycle)
	if err != nil {
		return fmt.Errorf("could not save partner: %w", err)
	}

	return nil
}
