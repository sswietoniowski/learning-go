package shared

import (
	"eats/backend/common"
)

type LineItemType struct {
	common.Enum[LineItemTypeValues]
}

type LineItemTypeValues string

func (d LineItemTypeValues) Values() []string {
	return []string{"food", "beverage", "delivery", "service"}
}

var (
	LineItemTypeFood     = common.MustEnum[LineItemType]("food")
	LineItemTypeBeverage = common.MustEnum[LineItemType]("beverage")
	LineItemTypeDelivery = common.MustEnum[LineItemType]("delivery")
	LineItemTypeService  = common.MustEnum[LineItemType]("service")
)
