package models

import (
	"context"
	"errors"

	"eats/backend/common"
	"eats/backend/common/shared"
	"eats/backend/settlements/domain"
)

type LegalEntityRepository interface {
	LegalEntityByUUID(ctx context.Context, uuid domain.LegalEntityUUID) (LegalEntity, error)
	PartnerByUUID(ctx context.Context, uuid domain.LegalEntityUUID) (Partner, error)
	SavePartner(ctx context.Context, partner Partner, billingCycle *domain.BillingCycle) error
	SavePlatformEntity(ctx context.Context, platform LegalEntity) error
}

type PlatformEntityUUID struct {
	domain.LegalEntityUUID
}

type LegalEntityType struct {
	common.Enum[LegalEntityTypeValues]
}

type LegalEntityTypeValues struct{}

func (l LegalEntityTypeValues) Values() []string {
	return []string{"platform", "partner"}
}

var (
	LegalEntityPlatform = common.MustEnum[LegalEntityType]("platform")
	LegalEntityPartner  = common.MustEnum[LegalEntityType]("partner")
)

// LegalEntity here is a settlements-side record, not a domain-layer entity
// (the billing LegalEntity is the encapsulated one). It's immutable and has
// no complex logic, so we pragmatically skip encapsulation: exported fields,
// no getters.
type LegalEntity struct {
	UUID              domain.LegalEntityUUID
	Type              LegalEntityType
	BusinessName      string
	TaxID             shared.TaxID
	Address           shared.Address
	BankAccountNumber domain.IBAN
	Currency          shared.Currency
}

func NewLegalEntity(
	uuid domain.LegalEntityUUID,
	legalEntityType LegalEntityType,
	businessName string,
	taxID shared.TaxID,
	address shared.Address,
	bankAccountNumber domain.IBAN,
	currency shared.Currency,
) (LegalEntity, error) {
	if uuid.IsZero() {
		return LegalEntity{}, errors.New("uuid cannot be zero")
	}

	if legalEntityType.IsZero() {
		return LegalEntity{}, errors.New("legal entity type cannot be empty")
	}

	if businessName == "" {
		return LegalEntity{}, errors.New("business name cannot be empty")
	}

	if taxID.IsZero() {
		return LegalEntity{}, errors.New("taxID cannot be zero")
	}

	if address.IsZero() {
		return LegalEntity{}, errors.New("address cannot be zero")
	}

	if bankAccountNumber.IsZero() {
		return LegalEntity{}, errors.New("bank account cannot be zero")
	}

	if currency.IsZero() {
		return LegalEntity{}, errors.New("currency cannot be zero")
	}

	return LegalEntity{
		UUID:              uuid,
		Type:              legalEntityType,
		BusinessName:      businessName,
		TaxID:             taxID,
		Address:           address,
		BankAccountNumber: bankAccountNumber,
		Currency:          currency,
	}, nil
}
