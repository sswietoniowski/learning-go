package domain

import (
	"errors"
	"unicode"
)

// IBAN represents an International Bank Account Number.
type IBAN struct {
	iban string
}

const (
	minIBANLength = 15 // Norway has the shortest IBANs
	maxIBANLength = 34 // Maximum IBAN length per ISO 13616
)

func NewIBAN(iban string) (IBAN, error) {
	if iban == "" {
		return IBAN{}, errors.New("iban is empty")
	}

	if len(iban) < minIBANLength {
		return IBAN{}, errors.New("iban is too short")
	}

	if len(iban) > maxIBANLength {
		return IBAN{}, errors.New("iban is too long")
	}

	// First two characters must be letters (country code)
	for _, r := range iban[:2] {
		if !unicode.IsLetter(r) {
			return IBAN{}, errors.New("iban country code is invalid")
		}
	}

	// Rest must be alphanumeric
	for _, r := range iban[2:] {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return IBAN{}, errors.New("iban contains invalid characters")
		}
	}

	return IBAN{
		iban: iban,
	}, nil
}

func (b IBAN) IsZero() bool {
	return b.iban == ""
}

func (b IBAN) String() string {
	return b.iban
}

func UnmarshalIBAN(iban string) IBAN {
	return IBAN{
		iban: iban,
	}
}
