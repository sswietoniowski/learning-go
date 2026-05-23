package command

import (
	"context"
	"fmt"
	"path"

	"eats/backend/billing/domain"
)

type documentPrinter interface {
	PrintDocument(ctx context.Context, doc *domain.Document) ([]byte, error)
}

type fileStorage interface {
	StoreFile(ctx context.Context, filename string, data []byte) (string, error)
}

type PrintDocument struct {
	DocumentUUID domain.DocumentUUID
}

func (h *Handlers) PrintDocument(ctx context.Context, cmd PrintDocument) error {
	doc, err := h.documentRepository.DocumentByUUID(ctx, cmd.DocumentUUID)
	if err != nil {
		return fmt.Errorf("could not get document by uuid: %w", err)
	}

	rendered, err := h.documentPrinter.PrintDocument(ctx, doc)
	if err != nil {
		return fmt.Errorf("could not print document: %w", err)
	}

	var subdir string
	switch doc.DocumentType() {
	case domain.DocumentTypeReceipt:
		subdir = "receipts"
	default:
		return fmt.Errorf("unknown document type: %s", doc.DocumentType())
	}

	fileName := doc.DocumentNumber().String() + ".html"
	filePath := path.Join("documents", subdir, fileName)

	storagePath, err := h.fileStorage.StoreFile(ctx, filePath, rendered)
	if err != nil {
		return fmt.Errorf("could not store document: %w", err)
	}

	// Should be idempotent
	err = h.documentRepository.UpdateFileUrl(ctx, cmd.DocumentUUID, storagePath)
	if err != nil {
		return fmt.Errorf("could not update document file url: %w", err)
	}

	return nil
}
