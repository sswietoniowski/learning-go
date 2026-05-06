package shared

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"eats/backend/common"
)

type Address struct {
	line1       string
	line2       string
	postalCode  string
	city        string
	countryCode CountryCode
}

func NewAddress(line1, line2, postalCode, city string, countryCode CountryCode) (Address, error) {
	var errDetails []common.ErrorDetails

	if line1 == "" {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "address",
			ErrorSlug:  "address-line1-required",
			Message:    "address line 1 is required",
		})
	}
	if postalCode == "" {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "address",
			ErrorSlug:  "address-postal-code-required",
			Message:    "postal code is required",
		})
	}
	if city == "" {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "address",
			ErrorSlug:  "address-city-required",
			Message:    "city is required",
		})
	}
	if countryCode.IsZero() {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "address",
			ErrorSlug:  "address-country-code-required",
			Message:    "country code is required",
		})
	}

	if len(errDetails) > 0 {
		return Address{}, common.NewInvalidInputError(
			"invalid-address",
			"invalid address data",
		).WithDetails(errDetails)
	}

	return Address{
		line1:       line1,
		line2:       line2,
		postalCode:  postalCode,
		city:        city,
		countryCode: countryCode,
	}, nil
}

func (a Address) IsZero() bool {
	return a == Address{}
}

func (a Address) Line1() string {
	return a.line1
}

func (a Address) Line2() string {
	return a.line2
}

func (a Address) PostalCode() string {
	return a.postalCode
}

func (a Address) City() string {
	return a.city
}

func (a Address) CountryCode() CountryCode {
	return a.countryCode
}

type addressDbDTO struct {
	Line1       string      `json:"line_1"`
	Line2       string      `json:"line_2"`
	PostalCode  string      `json:"postal_code"`
	City        string      `json:"city"`
	CountryCode CountryCode `json:"country_code"`
}

func (a *Address) Scan(src any) error {
	text, ok := src.(string)
	if !ok {
		return fmt.Errorf("invalid type for %T, expected string", src)
	}

	err := a.UnmarshalJSON([]byte(text))
	if err != nil {
		return fmt.Errorf("error unmarshalling %T from json: %w", a, err)
	}

	return nil
}

func (a Address) Value() (driver.Value, error) {
	data, err := a.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("error marshalling %T to json: %w", a, err)
	}

	return string(data), nil
}

func (a Address) MarshalJSON() ([]byte, error) {
	m := addressDbDTO{
		Line1:       a.line1,
		Line2:       a.line2,
		PostalCode:  a.postalCode,
		City:        a.city,
		CountryCode: a.countryCode,
	}

	return json.Marshal(m)
}

func (a *Address) UnmarshalJSON(data []byte) error {
	m := addressDbDTO{}
	err := json.Unmarshal(data, &m)
	if err != nil {
		return fmt.Errorf("error unmarshalling %T from json: %w", a, err)
	}

	a.line1 = m.Line1
	a.line2 = m.Line2
	a.postalCode = m.PostalCode
	a.city = m.City
	a.countryCode = m.CountryCode

	return nil
}
