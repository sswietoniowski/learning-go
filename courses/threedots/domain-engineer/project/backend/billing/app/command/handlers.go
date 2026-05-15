package command

import "eats/backend/billing/domain"

type Handlers struct {
	documentRepository domain.DocumentRepository
	documentPrinter    documentPrinter
	fileStorage        fileStorage
	documentFactory    *domain.DocumentFactory
}

func NewHandlers(
	documentRepository domain.DocumentRepository,
	documentPrinter documentPrinter,
	fileStorage fileStorage,
	taxRateProvider domain.TaxRateProvider,
) *Handlers {
	return &Handlers{
		documentRepository: documentRepository,
		documentPrinter:    documentPrinter,
		fileStorage:        fileStorage,
		documentFactory:    domain.NewDocumentFactory(taxRateProvider),
	}
}
