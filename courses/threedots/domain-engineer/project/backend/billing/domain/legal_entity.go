package domain

import (
	"errors"

	"eats/backend/common/shared"
)

type LegalEntity struct {
	name    string
	address shared.Address
	taxID   *shared.TaxID
}

func NewLegalEntity(name string, address shared.Address, taxID *shared.TaxID) (LegalEntity, error) {
	if name == "" {
		return LegalEntity{}, errors.New("name can't be empty")
	}

	if address.IsZero() {
		return LegalEntity{}, errors.New("address can't be empty")
	}

	return LegalEntity{
		name:    name,
		address: address,
		taxID:   taxID,
	}, nil
}

func (l LegalEntity) Name() string {
	return l.name
}

func (l LegalEntity) Address() shared.Address {
	return l.address
}

func (l LegalEntity) TaxID() *shared.TaxID {
	return l.taxID
}

func (l LegalEntity) IsZero() bool {
	return l == LegalEntity{}
}
