package query

import "eats/backend/billing/domain"

type Handlers struct {
	documentRepository domain.DocumentRepository
}

func NewHandlers(
	documentRepository domain.DocumentRepository,
) *Handlers {
	return &Handlers{
		documentRepository: documentRepository,
	}
}
