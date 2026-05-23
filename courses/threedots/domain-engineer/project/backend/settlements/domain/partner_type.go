package domain

import "eats/backend/common"

type PartnerType struct {
	common.Enum[PartnerTypeValues]
}

type PartnerTypeValues string

func (p PartnerTypeValues) Values() []string {
	return []string{"restaurant", "courier"}
}

var (
	PartnerTypeRestaurant = common.MustEnum[PartnerType]("restaurant")
	PartnerTypeCourier    = common.MustEnum[PartnerType]("courier")
)
