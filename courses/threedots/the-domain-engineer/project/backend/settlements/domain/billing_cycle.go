package domain

import (
	"errors"
	"time"

	"eats/backend/common"
)

type BillingCycleUUID struct {
	common.UUID
}

type BillingCycle struct {
	uuid        BillingCycleUUID
	partnerUUID LegalEntityUUID
	partnerType PartnerType
	number      int
	closed      bool
	settled     bool
	startDate   time.Time
	endDate     *time.Time
}

func NewInitialBillingCycle(partnerUUID LegalEntityUUID, partnerType PartnerType) (*BillingCycle, error) {
	startDate := time.Now().UTC()

	if partnerUUID.IsZero() {
		return nil, errors.New("partner UUID is zero")
	}

	if partnerType.IsZero() {
		return nil, errors.New("partner type is zero")
	}

	return &BillingCycle{
		uuid:        BillingCycleUUID{common.NewUUIDv7()},
		partnerUUID: partnerUUID,
		partnerType: partnerType,
		number:      1,
		closed:      false,
		startDate:   startDate,
	}, nil
}

func NewNextBillingCycle(previous *BillingCycle) (*BillingCycle, error) {
	if previous == nil {
		return nil, errors.New("previous billing cycle is nil")
	}

	if !previous.Closed() {
		return nil, errors.New("previous billing cycle is not closed yet")
	}

	number := previous.number + 1
	startDate := previous.endDate.AddDate(0, 0, 1)

	return &BillingCycle{
		uuid:        BillingCycleUUID{common.NewUUIDv7()},
		partnerUUID: previous.partnerUUID,
		partnerType: previous.partnerType,
		number:      number,
		closed:      false,
		startDate:   startDate,
	}, nil
}

func (bc *BillingCycle) UUID() BillingCycleUUID {
	return bc.uuid
}

func (bc *BillingCycle) PartnerUUID() LegalEntityUUID {
	return bc.partnerUUID
}

func (bc *BillingCycle) PartnerType() PartnerType {
	return bc.partnerType
}

func (bc *BillingCycle) Number() int {
	return bc.number
}

func (bc *BillingCycle) Closed() bool {
	return bc.closed
}

func (bc *BillingCycle) StartDate() time.Time {
	return bc.startDate
}

func (bc *BillingCycle) EndDate() *time.Time {
	return bc.endDate
}

func (bc *BillingCycle) Settled() bool {
	return bc.settled
}

func (bc *BillingCycle) Close() error {
	if bc.closed {
		return errors.New("billing cycle already closed")
	}

	bc.closed = true

	endDate := time.Now().UTC()
	bc.endDate = &endDate

	return nil
}

func (bc *BillingCycle) Settle() error {
	if !bc.closed {
		return errors.New("billing cycle is not closed")
	}
	if bc.settled {
		return errors.New("billing cycle already settled")
	}
	bc.settled = true
	return nil
}

func UnmarshalBillingCycle(
	uuid BillingCycleUUID,
	partnerUUID LegalEntityUUID,
	partnerType PartnerType,
	number int,
	closed bool,
	settled bool,
	startDate time.Time,
	endDate *time.Time,
) *BillingCycle {
	return &BillingCycle{
		uuid:        uuid,
		partnerUUID: partnerUUID,
		partnerType: partnerType,
		number:      number,
		closed:      closed,
		settled:     settled,
		startDate:   startDate,
		endDate:     endDate,
	}
}
