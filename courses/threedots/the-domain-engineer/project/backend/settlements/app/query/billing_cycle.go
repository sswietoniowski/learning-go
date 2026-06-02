package query

import (
	"context"
	"time"

	"eats/backend/settlements/domain"
)

type BillingCycleByPartner struct {
	PartnerUUID domain.LegalEntityUUID
}

type BillingCycleReadModel struct {
	BillingCycleUUID   domain.BillingCycleUUID
	PartnerUUID        domain.LegalEntityUUID
	BillingCycleNumber int
	StartDate          time.Time
	EndDate            *time.Time
	Closed             bool
	Settled            bool
}

type Handlers struct {
	billingCycleRepository BillingCycleRepository
}

func NewHandlers(billingCycleRepository BillingCycleRepository) *Handlers {
	if billingCycleRepository == nil {
		panic("billingCycleRepository is required")
	}

	return &Handlers{
		billingCycleRepository: billingCycleRepository,
	}
}

func (h *Handlers) BillingCycleByPartner(ctx context.Context, query BillingCycleByPartner) ([]BillingCycleReadModel, error) {
	return h.billingCycleRepository.BillingCyclesForPartner(ctx, query.PartnerUUID)
}

type BillingCycleRepository interface {
	BillingCyclesForPartner(ctx context.Context, partnerUUID domain.LegalEntityUUID) ([]BillingCycleReadModel, error)
}
