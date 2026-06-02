package query

import "eats/backend/billing/domain"

type Handlers struct {
	documentRepository domain.DocumentRepository
	documentFactory    *domain.DocumentFactory
}

func NewHandlers(
	documentRepository domain.DocumentRepository,
	taxRateProvider domain.TaxRateProvider,
) *Handlers {
	return &Handlers{
		documentRepository: documentRepository,
		documentFactory:    domain.NewDocumentFactory(taxRateProvider),
	}
}
