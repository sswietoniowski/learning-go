package command

import "eats/backend/billing/domain"

type Handlers struct {
	documentRepository domain.DocumentRepository
	documentPrinter    documentPrinter
	fileStorage        fileStorage
}

func NewHandlers(
	documentRepository domain.DocumentRepository,
	documentPrinter documentPrinter,
	fileStorage fileStorage,
) *Handlers {
	return &Handlers{
		documentRepository: documentRepository,
		documentPrinter:    documentPrinter,
		fileStorage:        fileStorage,
	}
}
