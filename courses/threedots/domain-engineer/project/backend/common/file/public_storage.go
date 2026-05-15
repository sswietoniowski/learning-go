package file

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/ThreeDotsLabs/the-domain-engineer/clients"
	"github.com/ThreeDotsLabs/the-domain-engineer/clients/files"
	"github.com/google/uuid"
)

type PublicStorage struct {
	clients *clients.Clients
}

func NewPublicStorage(clients *clients.Clients) *PublicStorage {
	if clients == nil {
		panic("nil clients")
	}
	return &PublicStorage{clients: clients}
}

func (p *PublicStorage) StoreFile(ctx context.Context, path string, fileContent []byte) (string, error) {
	fileID := uuid.NewString()
	filename := filepath.Base(path)

	resp, err := p.clients.Files.PutFilesFileIdContentWithBodyWithResponse(
		ctx,
		fileID,
		&files.PutFilesFileIdContentParams{XFilename: filename},
		"application/octet-stream",
		bytes.NewReader(fileContent),
	)
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	if resp.StatusCode() != http.StatusCreated {
		return "", fmt.Errorf("failed to upload file: unexpected status %d", resp.StatusCode())
	}

	return resp.JSON201.Path, nil
}
