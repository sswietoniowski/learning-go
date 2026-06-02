package models

type Partner struct {
	LegalEntity        LegalEntity
	PlatformEntityUUID PlatformEntityUUID
}

func NewPartner(legalEntity LegalEntity, platformEntityUUID PlatformEntityUUID) Partner {
	return Partner{
		LegalEntity:        legalEntity,
		PlatformEntityUUID: platformEntityUUID,
	}
}
