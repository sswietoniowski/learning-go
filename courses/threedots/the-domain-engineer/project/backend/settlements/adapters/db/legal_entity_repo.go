package db

import (
	"context"
	"fmt"

	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"eats/backend/common"
	"eats/backend/settlements/adapters/db/dbmodels"
	"eats/backend/settlements/app/models"
	"eats/backend/settlements/domain"
)

type LegalEntityRepository struct {
	db *pgxpool.Pool
}

func NewLegalEntityRepository(db *pgxpool.Pool) *LegalEntityRepository {
	if db == nil {
		panic("db is nil")
	}

	return &LegalEntityRepository{
		db: db,
	}
}

func (r *LegalEntityRepository) LegalEntityByUUID(ctx context.Context, uuid domain.LegalEntityUUID) (models.LegalEntity, error) {
	queries := dbmodels.New(r.db)

	legalEntity, err := queries.LegalEntityByUUID(ctx, uuid)
	if err != nil {
		return models.LegalEntity{}, fmt.Errorf("error getting legal entity by uuid: %w", err)
	}

	return legalEntityFromDB(legalEntity), nil
}

func legalEntityFromDB(l dbmodels.SettlementsLegalEntity) models.LegalEntity {
	return models.LegalEntity{
		UUID:              l.LegalEntityUuid,
		Type:              l.LegalEntityType,
		BusinessName:      l.BusinessName,
		TaxID:             l.TaxID,
		Address:           l.Address,
		BankAccountNumber: domain.UnmarshalIBAN(l.BankAccountNumber),
		Currency:          l.Currency,
	}
}

func (r *LegalEntityRepository) SavePlatformEntity(ctx context.Context, platform models.LegalEntity) error {
	queries := dbmodels.New(r.db)

	err := queries.SaveLegalEntity(ctx, dbmodels.SaveLegalEntityParams{
		LegalEntityUuid:   platform.UUID,
		LegalEntityType:   models.LegalEntityPlatform,
		BusinessName:      platform.BusinessName,
		TaxID:             platform.TaxID,
		Address:           platform.Address,
		BankAccountNumber: platform.BankAccountNumber.String(),
		Currency:          platform.Currency,
	})
	if err != nil {
		return fmt.Errorf("error saving legal entity: %w", err)
	}

	return nil
}

func (r *LegalEntityRepository) PartnerByUUID(ctx context.Context, uuid domain.LegalEntityUUID) (models.Partner, error) {
	queries := dbmodels.New(r.db)

	legalEntity, err := queries.LegalEntityByUUID(ctx, uuid)
	if err != nil {
		return models.Partner{}, fmt.Errorf("error getting legal entity by uuid: %w", err)
	}

	platformUUID, err := queries.PlatformByPartnerUUID(ctx, uuid)
	if err != nil {
		return models.Partner{}, fmt.Errorf("error getting platform uuid by partner uuid: %w", err)
	}

	return models.Partner{
		LegalEntity:        legalEntityFromDB(legalEntity),
		PlatformEntityUUID: platformUUID,
	}, nil
}

func (r *LegalEntityRepository) SavePartner(ctx context.Context, partner models.Partner, billingCycle *domain.BillingCycle) error {
	return common.UpdateInTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		queries := dbmodels.New(tx)

		err := queries.SaveLegalEntity(ctx, dbmodels.SaveLegalEntityParams{
			LegalEntityUuid:   partner.LegalEntity.UUID,
			LegalEntityType:   models.LegalEntityPartner,
			BusinessName:      partner.LegalEntity.BusinessName,
			TaxID:             partner.LegalEntity.TaxID,
			Address:           partner.LegalEntity.Address,
			BankAccountNumber: partner.LegalEntity.BankAccountNumber.String(),
			Currency:          partner.LegalEntity.Currency,
		})
		if err != nil {
			return fmt.Errorf("error saving legal entity: %w", err)
		}

		platform, err := queries.LegalEntityByUUID(ctx, partner.PlatformEntityUUID.LegalEntityUUID)
		if err != nil {
			return fmt.Errorf("error getting platform legal entity: %w", err)
		}

		if platform.LegalEntityType != models.LegalEntityPlatform {
			return fmt.Errorf("legal entity %s is not a platform", partner.PlatformEntityUUID)
		}

		err = queries.SavePartnerPlatformMapping(ctx, dbmodels.SavePartnerPlatformMappingParams{
			PartnerUuid:        partner.LegalEntity.UUID,
			PlatformEntityUuid: partner.PlatformEntityUUID,
		})
		if err != nil {
			return fmt.Errorf("error saving partner-platform mapping: %w", err)
		}

		err = queries.SaveBillingCycle(ctx, newBillingCycleSaveParams(billingCycle))
		if err != nil {
			return fmt.Errorf("error saving billing cycle: %w", err)
		}

		return nil
	})
}
