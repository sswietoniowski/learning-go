package testutils

import (
	"math/rand"

	"github.com/brianvoe/gofakeit/v7"

	"eats/backend/common/shared"
	"eats/backend/orders/api/http/client"
)

func GenerateRandomCountry() shared.CountryCode {
	values := shared.CountryCodeType("").Values()
	value := values[rand.Intn(len(values))]

	return shared.MustNewCountryCode(value)
}

func GenerateRandomOpenapiAddress(country shared.CountryCode) client.Address {
	address := gofakeit.Address()

	addr := client.Address{
		City:        address.City,
		CountryCode: country,
		Line1:       address.Street,
		Line2:       address.Unit,
		PostalCode:  address.Zip,
	}

	return addr
}

func GenerateRandomAddress(country shared.CountryCode) shared.Address {
	address := gofakeit.Address()

	addr, err := shared.NewAddress(
		address.Street,
		address.Unit,
		address.Zip,
		address.City,
		country,
	)
	if err != nil {
		panic(err)
	}

	return addr
}

func GenerateOpenapiAddressInCity(country shared.CountryCode, city string) client.Address {
	address := gofakeit.Address()

	addr := client.Address{
		City:        city,
		CountryCode: country,
		Line1:       address.Street,
		Line2:       address.Unit,
		PostalCode:  address.Zip,
	}

	return addr
}
