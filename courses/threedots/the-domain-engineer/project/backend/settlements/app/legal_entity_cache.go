package app

import (
	"context"

	"eats/backend/settlements/app/models"
	"eats/backend/settlements/domain"
)

type legalEntityFinder interface {
	LegalEntityByUUID(ctx context.Context, uuid domain.LegalEntityUUID) (models.LegalEntity, error)
}

// CachedLegalEntityFinder decorates legal entity finder with in-memory cache.
// You can use it when you need to resolve multiple legal entities by UUIDs
// within the same request scope.
type CachedLegalEntityFinder struct {
	finder legalEntityFinder
	cache  map[domain.LegalEntityUUID]models.LegalEntity
}

func NewCachedLegalEntityFinder(finder legalEntityFinder) *CachedLegalEntityFinder {
	return &CachedLegalEntityFinder{
		finder: finder,
		cache:  map[domain.LegalEntityUUID]models.LegalEntity{},
	}
}

func (c *CachedLegalEntityFinder) LegalEntityByUUID(ctx context.Context, uuid domain.LegalEntityUUID) (models.LegalEntity, error) {
	if le, ok := c.cache[uuid]; ok {
		return le, nil
	}

	le, err := c.finder.LegalEntityByUUID(ctx, uuid)
	if err != nil {
		return models.LegalEntity{}, err
	}

	c.cache[uuid] = le

	return le, nil
}
