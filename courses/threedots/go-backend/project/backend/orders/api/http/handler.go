package http

import (
	"context"

	"eats/backend/common"
)

type Handler struct{}

func NewHandler() Handler {
	return Handler{}
}

func Register(ctx context.Context, e common.EchoRouter, handler Handler) error {
	return nil
}
