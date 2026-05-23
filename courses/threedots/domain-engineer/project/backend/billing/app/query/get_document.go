package query

import (
	"context"
	"fmt"

	"eats/backend/billing/domain"
)

type GetDocumentByUUID struct {
	DocumentUUID domain.DocumentUUID
}

func (h *Handlers) GetDocumentByUUID(ctx context.Context, query GetDocumentByUUID) (*domain.Document, error) {
	doc, err := h.documentRepository.DocumentByUUID(ctx, query.DocumentUUID)
	if err != nil {
		return nil, fmt.Errorf("error getting document by uuid: %w", err)
	}

	return doc, nil
}
